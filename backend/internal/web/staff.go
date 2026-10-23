package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"216chan/backend/internal/publication"
)

type StaffConfig struct {
	Origin  string            `json:"origin"`
	CSRFKey string            `json:"csrf_key"`
	Users   map[string]string `json:"users"`
	// Forwarded TLS is trustworthy only when the proxy is the sole ingress.
	TLSProxy bool `json:"tls_proxy,omitempty"`
}

var staffName = regexp.MustCompile(`^[A-Za-z0-9_-]{1,32}$`)

type StaffHTTP struct {
	store    *publication.Staff
	config   StaffConfig
	key      []byte
	renderer *Handler
}
type StaffForm struct {
	Action           string
	Target, Revision int64
	Token            string
}
type StaffPage struct {
	Actor      string
	Queue      []publication.Review
	Detail     *publication.Detail
	Forms      []StaffForm
	Page, Next int
}

func LoadStaffConfig(path string) (StaffConfig, error) {
	var c StaffConfig
	info, err := os.Stat(path)
	if err != nil {
		return c, err
	}
	if info.Mode().Perm()&0077 != 0 {
		return c, fmt.Errorf("staff configuration must be mode 0600")
	}
	file, err := os.Open(path)
	if err != nil {
		return c, err
	}
	defer file.Close()
	return DecodeStaffConfig(file)
}

func DecodeStaffConfig(source io.Reader) (StaffConfig, error) {
	var c StaffConfig
	decoder := json.NewDecoder(source)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&c); err != nil {
		return c, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		if err == nil {
			return c, fmt.Errorf("staff configuration must contain one JSON value")
		}
		return c, err
	}
	return c, nil
}
func NewStaffHTTP(store *publication.Staff, c StaffConfig) (*StaffHTTP, error) {
	origin, err := url.Parse(c.Origin)
	if err != nil || origin.Host == "" || origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" || origin.User != nil {
		return nil, fmt.Errorf("invalid staff origin")
	}
	local := origin.Hostname() == "localhost" || net.ParseIP(origin.Hostname()).IsLoopback()
	if origin.Scheme != "https" && !(local && origin.Scheme == "http") {
		return nil, fmt.Errorf("staff origin must use TLS or loopback")
	}
	if c.TLSProxy && origin.Scheme != "https" {
		return nil, fmt.Errorf("staff tls_proxy requires an https origin")
	}
	key, err := base64.RawURLEncoding.DecodeString(c.CSRFKey)
	if err != nil || len(key) != 32 || len(c.Users) == 0 {
		return nil, fmt.Errorf("staff key and accounts required")
	}
	for user, digest := range c.Users {
		data, e := hex.DecodeString(digest)
		if !staffName.MatchString(user) || e != nil || len(data) != 32 {
			return nil, fmt.Errorf("invalid staff account")
		}
	}
	return &StaffHTTP{store: store, config: c, key: key}, nil
}
func (h *StaffHTTP) token(actor, action string, target, revision, expiry int64) string {
	message := fmt.Sprintf("%s:%s:%d:%d:%d", actor, action, target, revision, expiry)
	mac := hmac.New(sha256.New, h.key)
	mac.Write([]byte(message))
	return fmt.Sprintf("%d.%s", expiry, hex.EncodeToString(mac.Sum(nil)))
}

func (h *StaffHTTP) routes(mux *http.ServeMux, renderer *Handler) {
	h.renderer = renderer
	mux.HandleFunc("GET /staff", h.guard(h.page))
	mux.HandleFunc("POST /staff", h.guard(h.decide))
	mux.HandleFunc("GET /staff/media/{id}", h.guard(h.media))
}

func (h *StaffHTTP) guard(next func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		peer, _, _ := net.SplitHostPort(r.RemoteAddr)
		forwarded := r.Header.Values("X-Forwarded-Proto")
		proxiedTLS := h.config.TLSProxy && len(forwarded) == 1 && forwarded[0] == "https"
		if r.TLS == nil && !proxiedTLS && !net.ParseIP(peer).IsLoopback() {
			http.Error(w, "TLS or a loopback proxy is required", 403)
			return
		}
		user, password, ok := r.BasicAuth()
		digest := sha256.Sum256([]byte(password))
		expected, exists := h.config.Users[user]
		decoded, _ := hex.DecodeString(expected)
		if !ok || !exists || subtle.ConstantTimeCompare(digest[:], decoded) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="216chan staff", charset="UTF-8"`)
			http.Error(w, "Authentication required", 401)
			return
		}
		next(w, r, user)
	}
}

func (h *StaffHTTP) media(w http.ResponseWriter, r *http.Request, _ string) {
	data, err := h.store.Media(r.Context(), publication.PostID(positive(r.PathValue("id"))))
	if err != nil {
		fail(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "image/webp")
	w.Write(data)
}

func (h *StaffHTTP) decide(w http.ResponseWriter, r *http.Request, user string) {
	if r.Header.Get("Origin") != h.config.Origin {
		http.Error(w, "Invalid request origin", 403)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16384)
	if err := r.ParseForm(); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(w, "Form too large", 413)
		} else {
			http.Error(w, "Invalid form", 400)
		}
		return
	}
	target, targetErr := strconv.ParseInt(r.PostForm.Get("target"), 10, 64)
	revision, revisionErr := strconv.ParseInt(r.PostForm.Get("revision"), 10, 64)
	if targetErr != nil || revisionErr != nil || target < 1 || revision < 0 {
		http.Error(w, "Invalid target or revision", 400)
		return
	}
	action := r.PostForm.Get("action")
	token := r.PostForm.Get("csrf")
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		http.Error(w, "Invalid form token", 403)
		return
	}
	expiry, e := strconv.ParseInt(parts[0], 10, 64)
	if e != nil || expiry < time.Now().Unix() || expiry > time.Now().Add(30*time.Minute).Unix() || !hmac.Equal([]byte(token), []byte(h.token(user, action, target, revision, expiry))) {
		http.Error(w, "Expired or invalid form token", 403)
		return
	}
	if action == "erased" && r.PostForm.Get("confirm") != "erase" {
		http.Error(w, "Confirm erasure", 400)
		return
	}
	var err error
	reason := r.PostForm.Get("reason")
	switch action {
	case "hidden", "visible", "erased":
		err = h.store.SetPost(r.Context(), user, publication.PostDecision{Post: publication.PostID(target), Expected: revision, State: publication.Exposure(action), Reason: reason})
	case "configure":
		err = h.store.Configure(r.Context(), user, publication.ConversationDecision{Conversation: publication.ConversationID(target), Expected: revision, Board: r.PostForm.Get("board"), Locked: r.PostForm.Get("locked") == "on", Pinned: r.PostForm.Get("pinned") == "on", Reason: reason})
	case "close":
		err = h.store.CloseReview(r.Context(), user, publication.ReviewDecision{Post: publication.PostID(target), Expected: revision, Reason: reason})
	default:
		err = &publication.Error{Status: 400, Message: "Unknown staff action."}
	}
	if err != nil {
		logFailure(w, r, err)
		message := publication.Message(err)
		if publication.Status(err) >= 500 {
			message = "Decision outcome is uncertain. Reload the current state before retrying."
		}
		http.Error(w, message, publication.Status(err))
		return
	}
	http.Redirect(w, r, "/staff", 303)
}

func (h *StaffHTTP) page(w http.ResponseWriter, r *http.Request, user string) {
	p := StaffPage{Actor: user, Page: 1}
	if n, e := strconv.Atoi(r.URL.Query().Get("page")); e == nil && n > 0 && n < 100000 {
		p.Page = n
	}
	p.Next = p.Page + 1
	add := func(action string, target, revision int64) {
		p.Forms = append(p.Forms, StaffForm{action, target, revision, h.token(user, action, target, revision, time.Now().Add(30*time.Minute).Unix())})
	}
	if id := positive(r.URL.Query().Get("post")); id > 0 {
		d, err := h.store.Detail(r.Context(), id)
		if err != nil {
			fail(w, r, err)
			return
		}
		p.Detail = &d
		for _, action := range []string{"hidden", "visible", "erased"} {
			add(action, d.ID, d.Revision)
		}
		add("configure", d.ConversationID, d.ConversationRevision)
	} else {
		queue, err := h.store.Queue(r.Context(), p.Page)
		if err != nil {
			fail(w, r, err)
			return
		}
		p.Queue = queue
		for _, report := range queue {
			add("close", report.PostID, report.Revision)
		}
	}
	h.renderer.render(w, r, 200, page{Kind: "staff", Title: "216chan staff", Boards: publication.Boards, Staff: &p})
}
