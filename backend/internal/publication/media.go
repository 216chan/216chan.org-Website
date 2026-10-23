package publication

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const MaxOriginalBytes = 16 << 20
const MaxMediaBytes = 8 << 20

// Coalescing materializes every frame at canvas size, so decode cost scales with
// frames × canvas pixels. Identify reads at most MaxFrames+1 frames and policy.xml
// list-length (an exclusive bound) is MaxFrames+2, so oversized input reaches the
// Go check as a 400 while the policy remains a backstop.
const MaxFrames = 300
const MaxDecodedPixels = 1 << 24
const maxImageDimension = 4096

//go:embed policy.xml
var imagePolicy []byte

type OriginalDigest [32]byte

var tooLarge = &Error{413, "Converted image exceeds 8 MiB."}

type ApprovedMedia struct {
	bytes    []byte
	original OriginalDigest
}

func webpChunks(data []byte) ([]string, error) {
	kinds := []string{}
	for pos := 12; pos < len(data); {
		if len(data)-pos < 8 {
			return nil, fmt.Errorf("truncated image container")
		}
		size := uint64(binary.LittleEndian.Uint32(data[pos+4 : pos+8]))
		kind := string(data[pos : pos+4])
		end := uint64(pos) + 8 + size + size%2
		if end > uint64(len(data)) {
			return nil, fmt.Errorf("truncated image chunk")
		}
		kinds = append(kinds, kind)
		pos = int(end)
	}
	return kinds, nil
}
func imageFormat(data []byte) (string, error) {
	switch {
	case bytes.HasPrefix(data, []byte{255, 216, 255}):
		return "JPEG", nil
	case bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")):
		return "PNG", nil
	case bytes.HasPrefix(data, []byte("GIF87a")), bytes.HasPrefix(data, []byte("GIF89a")):
		return "GIF", nil
	case len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		return "WEBP", nil
	}
	return "", fmt.Errorf("only JPEG, PNG, GIF and WebP are accepted")
}

// Require an explicit choice to avoid silently dropping namespace isolation.
type Isolation string

const (
	Bubblewrap Isolation = "bwrap"
	Unisolated Isolation = "none"
)

func (iso Isolation) input(dir string) string {
	if iso == Unisolated {
		return filepath.Join(dir, "input")
	}
	return "/input"
}

// Retain enough stderr to diagnose failures while draining the child pipe.
type imageDiagnostics struct {
	data [4096]byte
	n    int
}

func (d *imageDiagnostics) Write(p []byte) (int, error) {
	d.n += copy(d.data[d.n:], p)
	return len(p), nil
}

type imageOperation uint8

const (
	inspectImage imageOperation = iota
	convertImage
)

func sandboxImage(ctx context.Context, iso Isolation, dir string, operation imageOperation, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	// Inspection only reads metadata under the same process/resource bounds.
	// Apply dimension limits after inspection so Go can classify them as input
	// errors; retain the decoder's dimension backstop during conversion.
	policy := imagePolicy
	if operation == convertImage {
		limits := fmt.Sprintf(`<policy domain="resource" name="width" value="%d"/><policy domain="resource" name="height" value="%d"/></policymap>`, maxImageDimension, maxImageDimension)
		policy = bytes.Replace(imagePolicy, []byte("</policymap>"), []byte(limits), 1)
	}
	if err := os.WriteFile(filepath.Join(dir, "policy.xml"), policy, 0600); err != nil {
		return nil, err
	}
	command := []string{"--as=805306368", "--cpu=10", fmt.Sprintf("--fsize=%d", MaxMediaBytes), "--nofile=64", "--core=0", "--"}
	env := []string{"PATH=/usr/bin:/bin"}
	switch iso {
	case Bubblewrap:
		command = append(command, "bwrap", "--unshare-all", "--unshare-user", "--disable-userns", "--assert-userns-disabled", "--die-with-parent", "--new-session", "--cap-drop", "ALL", "--clearenv")
		for _, path := range []string{"/usr", "/lib", "/lib64"} {
			if _, err := os.Stat(path); err == nil {
				command = append(command, "--ro-bind", path, path)
			}
		}
		command = append(command, "--ro-bind", filepath.Join(dir, "input"), "/input", "--ro-bind", filepath.Join(dir, "policy.xml"), "/policy/policy.xml", "--proc", "/proc", "--dev", "/dev", "--size", "1048576", "--tmpfs", "/tmp", "--remount-ro", "/", "--setenv", "PATH", "/usr/bin", "--setenv", "MAGICK_CONFIGURE_PATH", "/policy", "--setenv", "MAGICK_TEMPORARY_PATH", "/tmp", "--setenv", "OMP_NUM_THREADS", "1", "--")
	case Unisolated:
		env = append(env, "MAGICK_CONFIGURE_PATH="+dir, "MAGICK_TEMPORARY_PATH="+dir, "OMP_NUM_THREADS=1")
	default:
		return nil, fmt.Errorf("unknown image isolation %q", iso)
	}
	command = append(command, "/usr/bin/magick")
	command = append(command, args...)
	output, err := os.CreateTemp(dir, "output-")
	if err != nil {
		return nil, err
	}
	defer output.Close()
	cmd := exec.CommandContext(ctx, "prlimit", command...)
	cmd.Env = env
	cmd.Dir = dir
	cmd.Stdout = output
	var diagnostics imageDiagnostics
	cmd.Stderr = &diagnostics
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = time.Second
	if err = cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		// Reaching the output limit indicates oversized input, not an outage.
		if info, e := output.Stat(); e == nil && info.Size() >= MaxMediaBytes {
			return nil, tooLarge
		}
		if detail := strings.TrimSpace(string(diagnostics.data[:diagnostics.n])); detail != "" {
			return nil, fmt.Errorf("isolated image processing failed: %w: %s", err, detail)
		}
		return nil, fmt.Errorf("isolated image processing failed: %w", err)
	}
	if _, err = output.Seek(0, 0); err != nil {
		return nil, err
	}
	return io.ReadAll(io.LimitReader(output, MaxMediaBytes+1))
}
func ApproveMedia(ctx context.Context, iso Isolation, source io.Reader) (*ApprovedMedia, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	original, err := io.ReadAll(io.LimitReader(source, MaxOriginalBytes+1))
	if err != nil {
		return nil, err
	}
	if len(original) == 0 || len(original) > MaxOriginalBytes {
		return nil, &Error{400, "Image input must be 1 byte to 16 MiB."}
	}
	kind, err := imageFormat(original)
	if err != nil {
		return nil, &Error{400, err.Error()}
	}
	dir, err := os.MkdirTemp("", "216chan-approve-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	if err = os.WriteFile(filepath.Join(dir, "input"), original, 0600); err != nil {
		return nil, err
	}
	// Multi-frame input prints one descriptor line per frame; the first carries the canvas.
	descriptor, err := sandboxImage(ctx, iso, dir, inspectImage, "identify", "-ping", "-format", "%m %W %H %n\n", fmt.Sprintf("%s:%s[0-%d]", kind, iso.input(dir), MaxFrames))
	if err != nil {
		return nil, approvalError(ctx, err)
	}
	first, _, _ := strings.Cut(string(descriptor), "\n")
	fields := strings.Fields(first)
	if len(fields) != 4 || fields[0] != kind {
		return nil, &Error{400, "Unreadable image."}
	}
	width, e1 := strconv.Atoi(fields[1])
	height, e2 := strconv.Atoi(fields[2])
	frames, e3 := strconv.Atoi(fields[3])
	if e1 != nil || e2 != nil || e3 != nil || width < 1 || height < 1 || frames < 1 {
		return nil, &Error{400, "Unreadable image."}
	}
	if width > maxImageDimension || height > maxImageDimension {
		return nil, &Error{400, fmt.Sprintf("Image dimensions exceed %d pixels.", maxImageDimension)}
	}
	if frames > MaxFrames || width*height*frames > MaxDecodedPixels {
		return nil, &Error{400, fmt.Sprintf("Animation exceeds %d frames or %d decoded pixels.", MaxFrames, MaxDecodedPixels)}
	}
	data, err := sandboxImage(ctx, iso, dir, convertImage, kind+":"+iso.input(dir), "-coalesce", "-auto-orient", "-colorspace", "sRGB", "-strip", "-quality", "82", "-loop", "0", "WEBP:-")
	if err != nil {
		return nil, approvalError(ctx, err)
	}
	if len(data) > MaxMediaBytes {
		return nil, tooLarge
	}
	if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		return nil, fmt.Errorf("invalid approved output")
	}
	chunks, err := webpChunks(data)
	if err != nil {
		return nil, err
	}
	animated := false
	for _, chunk := range chunks {
		switch chunk {
		case "VP8 ", "VP8L", "VP8X", "ALPH":
		case "ANIM", "ANMF":
			animated = true
		default:
			return nil, fmt.Errorf("unexpected output metadata")
		}
	}
	if animated != (frames > 1) {
		return nil, fmt.Errorf("output animation does not match input")
	}
	return &ApprovedMedia{data, OriginalDigest(sha256.Sum256(original))}, nil
}

// The stored digest avoids rehashing each delivery.
func (s *Store) Media(ctx context.Context, id PostID) (data, digest []byte, err error) {
	err = s.reader.QueryRowContext(ctx, `SELECT bytes,digest FROM pub.visible_media WHERE post_id=$1`, id).Scan(&data, &digest)
	return data, digest, err
}

// Cancellation and unavailable sandbox infrastructure must not look like bad input.
func approvalError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	var missing *exec.Error
	if errors.As(err, &missing) {
		return &Error{503, "Image processor unavailable."}
	}
	// Exit status cannot distinguish decoder failure from sandbox startup failure;
	// treat both as operational failures.
	return fmt.Errorf("image approval failed: %w", err)
}
