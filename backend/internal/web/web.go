package web

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"216chan/backend/internal/model"
	"216chan/backend/internal/publication"
)

//go:embed templates/*.html static
var assets embed.FS

var SiteOrigin = "https://216chan.org"

type Handler struct {
	pages  map[string]*template.Template
	static http.Handler
}
type pageLink struct {
	Number int
	URL    string
	Active bool
}
type page struct {
	Kind, Title, Slug, Error, Notice, Body, Action string
	Receipt                                        string
	Description, Canonical                         string
	ThreadID                                       int64
	Boards                                         []model.Board
	Stats                                          *model.Stats
	Board                                          *model.ThreadsPage
	Thread                                         *model.ThreadPage
	Staff                                          *StaffPage
	model.Pagination
	PageLinks []pageLink
}

func NewRenderer() *Handler {
	funcs := template.FuncMap{"fmtTime": func(t time.Time) string { return t.UTC().Format("02/01/06, 15:04") }, "isoTime": func(t time.Time) string { return t.UTC().Format(time.RFC3339) }, "number": func(n int64) string {
		s := strconv.FormatInt(n, 10)
		for i := len(s) - 3; i > 0; i -= 3 {
			s = s[:i] + "," + s[i:]
		}
		return s
	}, "bytes": func(n int64) string {
		const unit = 1024
		if n < unit {
			return fmt.Sprintf("%d B", n)
		}
		value, exponent := float64(n)/unit, 0
		for value >= unit && exponent < 3 {
			value /= unit
			exponent++
		}
		return fmt.Sprintf("%.1f %ciB", value, "KMGT"[exponent])
	}, "origin": func() string { return SiteOrigin }, "markup": markup}
	base := template.Must(template.New("").Funcs(funcs).ParseFS(assets, "templates/layout.html"))
	pages := map[string]*template.Template{}
	for _, kind := range []string{"home", "rules", "board", "thread", "staff"} {
		pages[kind] = template.Must(template.Must(base.Clone()).ParseFS(assets, "templates/"+kind+".html"))
	}
	static, _ := fs.Sub(assets, "static")
	return &Handler{pages: pages, static: revalidated(static)}
}

// Embedded files have no modification time; ETags prevent repeated downloads.
func revalidated(files fs.FS) http.Handler {
	tags := map[string]string{}
	fs.WalkDir(files, ".", func(name string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			data, _ := fs.ReadFile(files, name)
			sum := sha256.Sum256(data)
			tags["/"+name] = entityTag(sum[:])
		}
		return nil
	})
	server := http.FileServerFS(files)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tag, ok := tags[r.URL.Path]; ok {
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("ETag", tag)
		}
		server.ServeHTTP(w, r)
	})
}

var quoteLink = regexp.MustCompile(`&gt;&gt;([1-9][0-9]{0,18})\b`)

// User input must not introduce markup.
func markup(body string) template.HTML {
	lines := strings.Split(template.HTMLEscapeString(body), "\n")
	for i, line := range lines {
		line = quoteLink.ReplaceAllString(line, `<a href="/posts/$1">&gt;&gt;$1</a>`)
		if strings.HasPrefix(line, "&gt;") {
			line = `<span class="quote">` + line + `</span>`
		}
		lines[i] = line
	}
	return template.HTML(strings.Join(lines, "\n"))
}

func entityTag(digest []byte) string {
	return `"` + hex.EncodeToString(digest[:16]) + `"`
}

func fail(w http.ResponseWriter, r *http.Request, err error) {
	logFailure(w, r, err)
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeJSON(w, publication.Status(err), map[string]string{"error": publication.Message(err)})
		return
	}
	http.Error(w, publication.Message(err), publication.Status(err))
}

func unavailable(w http.ResponseWriter, r *http.Request, err error, message string) {
	logFailure(w, r, err)
	http.Error(w, message, 503)
}

func logFailure(w http.ResponseWriter, r *http.Request, err error) {
	if publication.Status(err) >= 500 {
		slog.Error("request failed", "request_id", w.Header().Get("X-Request-ID"), "route", r.Pattern, "error", err)
	}
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, status int, p page) {
	renderTemplate(w, r, h.pages[p.Kind], "layout", status, p)
}

func renderTemplate(w http.ResponseWriter, r *http.Request, templates *template.Template, name string, status int, value any) {
	var b bytes.Buffer
	if err := templates.ExecuteTemplate(&b, name, value); err != nil {
		logFailure(w, r, fmt.Errorf("web render: %w", err))
		http.Error(w, "Unable to render page", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(b.Bytes())
}
