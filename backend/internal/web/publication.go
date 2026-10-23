package web

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"216chan/backend/internal/geoip"
	"216chan/backend/internal/model"
	"216chan/backend/internal/publication"
)

func NewPublication(store *publication.Store, staff *StaffHTTP, postingDisabled bool, observer *geoip.Observer, iso publication.Isolation) http.Handler {
	h := NewRenderer()
	mux := http.NewServeMux()
	if staff != nil {
		staff.routes(mux, h)
	}
	mux.HandleFunc("POST /reports/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := positive(r.PathValue("id"))
		if err := store.Report(r.Context(), publication.PostID(id)); err != nil {
			fail(w, r, err)
			return
		}
		// A hidden post has no public destination for the reporter.
		location, err := store.PostLocation(r.Context(), id)
		if err != nil {
			location = "/"
		}
		path, fragment, anchored := strings.Cut(location, "#")
		separator := "?"
		if strings.Contains(path, "?") {
			separator = "&"
		}
		target := path + separator + "reported=1"
		if anchored {
			target += "#" + fragment
		}
		http.Redirect(w, r, target, 303)
	})
	mux.HandleFunc("GET /media/{id}", func(w http.ResponseWriter, r *http.Request) {
		data, digest, err := store.Media(r.Context(), publication.PostID(positive(r.PathValue("id"))))
		if err != nil {
			fail(w, r, err)
			return
		}
		// Revalidate every view: hiding or erasure must take effect immediately.
		w.Header().Set("Content-Type", "image/webp")
		w.Header().Set("Content-Disposition", "inline")
		w.Header().Set("Cache-Control", "private, no-cache")
		w.Header().Set("ETag", entityTag(digest))
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(data))
	})
	for _, path := range []string{"/icons/{asset}", "/site.css", "/site.js"} {
		mux.Handle("GET "+path, h.static)
	}
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok\n")) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		checks := []func(context.Context) error{store.Ready}
		if staff != nil {
			checks = append(checks, staff.store.Ready)
		}
		for _, ready := range checks {
			if err := ready(r.Context()); err != nil {
				unavailable(w, r, err, "Database unavailable")
				return
			}
		}
		w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("GET /api/stats", func(w http.ResponseWriter, r *http.Request) {
		stats, err := store.Stats(r.Context())
		if err != nil {
			fail(w, r, err)
			return
		}
		writeJSON(w, 200, stats)
	})
	mux.HandleFunc("GET /posts/{id}", func(w http.ResponseWriter, r *http.Request) {
		location, err := store.PostLocation(r.Context(), positive(r.PathValue("id")))
		if err != nil {
			fail(w, r, err)
			return
		}
		http.Redirect(w, r, location, 303)
	})
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		p := newPage("home", "")
		p.Title = "216chan - Tunisian Imageboard"
		var err error
		p.Stats, err = store.Stats(r.Context())
		status := 200
		if err != nil {
			logFailure(w, r, err)
			status = publication.Status(err)
			p.Error = "Statistics are temporarily unavailable."
		}
		h.render(w, r, status, p)
	})
	mux.HandleFunc("GET /rules", func(w http.ResponseWriter, r *http.Request) {
		p := newPage("rules", "")
		p.Title = "Rules - 216chan"
		p.Description = "216chan rules - global and board rules. Tunisian imageboard."
		p.Canonical = SiteOrigin + "/rules"
		h.render(w, r, 200, p)
	})
	routes := &publicationHTTP{renderer: h, store: store, observer: observer, iso: iso, postingDisabled: postingDisabled, slots: make(chan struct{}, 2)}
	for _, pattern := range []string{"GET /{board}/{$}", "POST /{board}/{$}", "GET /{board}/{thread}", "POST /{board}/{thread}", "GET /api/{board}/threads", "POST /api/{board}/threads", "GET /api/{board}/threads/{thread}", "POST /api/{board}/threads/{thread}/replies"} {
		mux.HandleFunc(pattern, routes.page)
	}
	// Stops other sites posting or reporting through visitors' browsers.
	protection := http.NewCrossOriginProtection()
	protection.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fail(w, r, &publication.Error{Status: 403, Message: "Cross-origin request denied."})
	}))
	guarded := protection.Handler(mux)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w = &apiResponse{ResponseWriter: w}
		}
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		// The executable adds socket deadlines; this also bounds SQL across adapters.
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		guarded.ServeHTTP(w, r.WithContext(ctx))
	})
}

type publicationHTTP struct {
	renderer        *Handler
	store           *publication.Store
	observer        *geoip.Observer
	iso             publication.Isolation
	postingDisabled bool
	slots           chan struct{}
}

func (h *publicationHTTP) page(w http.ResponseWriter, r *http.Request) {
	board := r.PathValue("board")
	id := positive(r.PathValue("thread"))
	isThread := r.PathValue("thread") != ""
	api := strings.HasPrefix(r.URL.Path, "/api/")
	if isThread && id == 0 {
		fail(w, r, sql.ErrNoRows)
		return
	}
	p := newPage("board", board)
	if isThread {
		p.Kind = "thread"
		p.Description = fmt.Sprintf("216chan /%s/ thread %d - Tunisian imageboard.", board, id)
		p.ThreadID = id
		p.Action = fmt.Sprintf("/%s/%d", board, id)
	}
	status := 200
	if r.Method == http.MethodPost {
		admitted := false
		if !h.postingDisabled {
			select {
			case h.slots <- struct{}{}:
				defer func() { <-h.slots }()
				admitted = true
			default:
			}
		}
		var handled bool
		status, handled = h.publish(w, r, &p, api, admitted)
		if handled {
			return
		}
	}
	pageNumber := max(1, int(positive(r.URL.Query().Get("page"))))
	result, err := h.readPage(r.Context(), &p, pageNumber)
	if err != nil {
		status = publication.Status(err)
		if api {
			fail(w, r, err)
			return
		}
		logFailure(w, r, err)
		p.Error = publication.Message(err)
	} else if api {
		writeJSON(w, 200, result)
		return
	}
	if r.URL.Query().Has("reported") {
		p.Notice = "Report received. Staff will review it."
	}
	if p.Receipt == "" {
		p.Receipt, err = publication.NewReceipt()
		if err != nil {
			fail(w, r, err)
			return
		}
	}
	p.Canonical = SiteOrigin + p.Action
	if p.Page > 1 {
		p.Canonical += fmt.Sprintf("?page=%d", p.Page)
	}
	addPages(&p)
	h.renderer.render(w, r, status, p)
}

func (h *publicationHTTP) publish(w http.ResponseWriter, r *http.Request, p *page, api, admitted bool) (status int, handled bool) {
	var in publication.Submission
	var err error
	if admitted {
		in, err = parsePublication(w, r, h.iso)
	} else {
		err = &publication.Error{Status: 429, Message: "Posting is busy. Retry with the same receipt."}
		if h.postingDisabled {
			err = &publication.Error{Status: 503, Message: "Posting is paused. Retry later with the same receipt."}
		}
		if !api {
			in.Receipt, in.Body = draftFields(w, r)
		}
	}
	in.Board = p.Slug
	in.Conversation = publication.ConversationID(p.ThreadID)
	p.Body, p.Receipt = in.Body, in.Receipt
	if err == nil {
		in.CountryCode = h.observer.Country(r)
		var accepted publication.Accepted
		accepted, err = h.store.Publish(r.Context(), in)
		if err == nil {
			if api {
				writeJSON(w, 200, accepted)
				return 200, true
			}
			if accepted.PostID != nil {
				http.Redirect(w, r, fmt.Sprintf("/posts/%d", *accepted.PostID), 303)
				return 303, true
			}
			p.Error = "Your submission was accepted, but it is currently unavailable."
			p.Body, p.Receipt = "", ""
			return 200, false
		}
	}
	status = publication.Status(err)
	if api {
		fail(w, r, err)
		return status, true
	}
	if admitted {
		logFailure(w, r, err)
	}
	p.Error = publication.Message(err)
	if errors.Is(err, publication.ErrReceiptConflict) || !publication.ValidReceipt(p.Receipt) {
		p.Receipt = ""
	}
	return status, false
}

func (h *publicationHTTP) readPage(ctx context.Context, p *page, pageNumber int) (any, error) {
	var err error
	if p.ThreadID != 0 {
		p.Thread, err = h.store.Thread(ctx, p.Slug, p.ThreadID, pageNumber)
		if err == nil {
			p.Pagination = p.Thread.Pagination
		}
		return p.Thread, err
	}
	p.Board, err = h.store.Board(ctx, p.Slug, pageNumber)
	if err == nil {
		p.Pagination = p.Board.Pagination
		p.Description = fmt.Sprintf("216chan /%s/ - %s board. Post anonymously on the Tunisian imageboard.", p.Slug, p.Board.Board.Name)
	}
	return p.Board, err
}

func positive(value string) int64 {
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n < 1 {
		return 0
	}
	return n
}
func newPage(kind, board string) page {
	return page{Kind: kind, Title: "/" + board + "/ - 216chan", Slug: board, Boards: publication.Boards, Pagination: model.Pagination{Page: 1, Pages: 1}, Action: "/" + board + "/", Canonical: SiteOrigin + "/", Description: "216chan is a Tunisian imageboard and forum."}
}

// ServeMux supplies routing errors itself. Normalize those at the API boundary
// while preserving JSON errors from the application and headers such as Allow.
type apiResponse struct {
	http.ResponseWriter
	replaced bool
	wrote    bool
}

func (w *apiResponse) WriteHeader(status int) {
	if w.wrote {
		return
	}
	if status >= 200 {
		w.wrote = true
	}
	if status >= 400 && w.Header().Get("Content-Type") != "application/json" {
		w.replaced = true
		w.Header().Del("Content-Length")
		writeJSON(w.ResponseWriter, status, map[string]string{"error": http.StatusText(status)})
		return
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *apiResponse) Write(p []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	if w.replaced {
		return len(p), nil
	}
	return w.ResponseWriter.Write(p)
}

func (w *apiResponse) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func parsePublication(w http.ResponseWriter, r *http.Request, iso publication.Isolation) (publication.Submission, error) {
	var in publication.Submission
	r.Body = http.MaxBytesReader(w, r.Body, publication.MaxOriginalBytes+(64<<10))
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		var body struct {
			Receipt string `json:"receipt"`
			Body    string `json:"body"`
		}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&body); err != nil {
			return in, &publication.Error{Status: 400, Message: "Invalid submission."}
		}
		if err := dec.Decode(new(any)); err != io.EOF {
			return in, &publication.Error{Status: 400, Message: "One submission is required."}
		}
		return publication.Submission{Receipt: body.Receipt, Body: body.Body}, nil
	}
	var err error
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		err = r.ParseMultipartForm(64 << 10)
		if r.MultipartForm != nil {
			defer r.MultipartForm.RemoveAll()
		}
	} else {
		err = r.ParseForm()
	}
	if err != nil {
		return in, &publication.Error{Status: 400, Message: "Invalid or oversized submission."}
	}
	in.Receipt = r.PostForm.Get("receipt")
	in.Body = r.PostForm.Get("body")
	if r.MultipartForm != nil {
		files := r.MultipartForm.File["image"]
		if len(files) > 1 || len(r.MultipartForm.File) > 1 {
			return in, &publication.Error{Status: 400, Message: "One image is allowed."}
		}
		if len(files) == 1 && files[0].Size > 0 {
			file, e := files[0].Open()
			if e != nil {
				return in, e
			}
			defer file.Close()
			in.Media, e = publication.ApproveMedia(r.Context(), iso, file)
			if e != nil {
				return in, e
			}
		}
	}
	return in, nil
}

// Avoid reading uploads so shed requests can retain the draft and receipt.
func draftFields(w http.ResponseWriter, r *http.Request) (receipt, body string) {
	r.Body = http.MaxBytesReader(w, r.Body, publication.MaxOriginalBytes+(64<<10))
	parts, err := r.MultipartReader()
	if err != nil {
		r.ParseForm()
		return r.PostForm.Get("receipt"), r.PostForm.Get("body")
	}
	for {
		part, err := parts.NextPart()
		if err != nil || part.FileName() != "" {
			return receipt, body
		}
		value, _ := io.ReadAll(io.LimitReader(part, 64<<10))
		switch part.FormName() {
		case "receipt":
			receipt = string(value)
		case "body":
			body = string(value)
		}
	}
}
func addPages(p *page) {
	first, last := min(max(p.Page-3, 1), p.Pages), min(p.Page+3, p.Pages)
	add := func(n int) {
		p.PageLinks = append(p.PageLinks, pageLink{n, fmt.Sprintf("%s?page=%d", p.Action, n), n == p.Page})
	}
	if first > 1 {
		add(1)
		if first > 2 {
			p.PageLinks = append(p.PageLinks, pageLink{})
		}
	}
	for n := first; n <= last; n++ {
		add(n)
	}
	if last < p.Pages {
		if last < p.Pages-1 {
			p.PageLinks = append(p.PageLinks, pageLink{})
		}
		add(p.Pages)
	}
}
