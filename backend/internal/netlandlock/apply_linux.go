//go:build linux && (amd64 || arm64)

package netlandlock

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

const (
	createRuleset = 444
	addRule       = 445
	restrictSelf  = 446
	versionFlag   = 1
	netPortRule   = 2
	noNewPrivs    = 38
	bindTCP       = 1 << 0
	connectTCP    = 1 << 1
	bindUDP       = 1 << 2
	connectUDP    = 1 << 3
)

type ruleset struct {
	filesystem uint64
	network    uint64
	scoped     uint64
}
type networkPort struct {
	access uint64
	port   uint64
}

// Reexec lets every runtime thread inherit the policy; applying it in place
// would leave other Go threads unrestricted.
func Reexec(databasePort, listenPort uint16) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := apply(databasePort, listenPort); err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	// Landlock does not mediate MPTCP sockets, which Go enables by default.
	env := make([]string, 0, len(os.Environ())+2)
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "GODEBUG=") {
			env = append(env, entry)
		}
	}
	debug := os.Getenv("GODEBUG")
	if debug != "" {
		debug += ","
	}
	env = append(env, "GODEBUG="+debug+"multipathtcp=0", "PUBLICATION_NETWORK_LANDLOCK_APPLIED=1")
	return syscall.Exec(executable, os.Args, env)
}

// Filesystem restrictions would impede bwrap mounts.
func apply(databasePort, listenPort uint16) error {
	version, _, errno := syscall.RawSyscall6(createRuleset, 0, 0, versionFlag, 0, 0, 0)
	if errno != 0 {
		return fmt.Errorf("query Landlock ABI: %w", errno)
	}
	if version < 10 {
		return fmt.Errorf("network Landlock requires ABI 10; this kernel has %d", version)
	}
	policy := ruleset{network: bindTCP | connectTCP | bindUDP | connectUDP}
	fd, _, errno := syscall.RawSyscall6(createRuleset, uintptr(unsafe.Pointer(&policy)), unsafe.Sizeof(policy), 0, 0, 0, 0)
	runtime.KeepAlive(&policy)
	if errno != 0 {
		return fmt.Errorf("create Landlock ruleset: %w", errno)
	}
	defer syscall.Close(int(fd))
	for _, permission := range []networkPort{{bindTCP, uint64(listenPort)}, {connectTCP, uint64(databasePort)}} {
		_, _, errno = syscall.RawSyscall6(addRule, fd, netPortRule, uintptr(unsafe.Pointer(&permission)), 0, 0, 0)
		runtime.KeepAlive(&permission)
		if errno != 0 {
			return fmt.Errorf("add Landlock port rule: %w", errno)
		}
	}
	if _, _, errno = syscall.RawSyscall6(syscall.SYS_PRCTL, noNewPrivs, 1, 0, 0, 0, 0); errno != 0 {
		return fmt.Errorf("set no_new_privs: %w", errno)
	}
	if _, _, errno = syscall.RawSyscall6(restrictSelf, fd, 0, 0, 0, 0, 0); errno != 0 {
		return fmt.Errorf("enforce Landlock network policy: %w", errno)
	}
	return nil
}
