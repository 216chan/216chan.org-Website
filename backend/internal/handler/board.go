package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"216chan/backend/internal/geoip"
	"216chan/backend/internal/service"
	"216chan/backend/internal/storage"
)

type BoardHandler struct {
	svc     *service.BoardService
	storage storage.Storage
	geo     *geoip.Resolver
}

func NewBoardHandler(svc *service.BoardService, st storage.Storage, g *geoip.Resolver) *BoardHandler {
	return &BoardHandler{svc: svc, storage: st, geo: g}
}

func (h *BoardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	board := parts[0]
	if !isValidBoard(board) {
		http.Error(w, `{"error":"invalid board"}`, http.StatusNotFound)
		return
	}
	if len(parts) == 2 && parts[1] == "threads" {
		switch r.Method {
		case http.MethodGet:
			h.listThreads(w, r, board)
		case http.MethodPost:
			h.createThread(w, r, board)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
		return
	}
	if len(parts) == 3 && parts[1] == "threads" {
		id, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			http.Error(w, `{"error":"invalid thread id"}`, http.StatusBadRequest)
			return
		}
		if r.Method == http.MethodGet {
			h.getThread(w, r, board, id)
			return
		}
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	if len(parts) == 4 && parts[1] == "threads" && parts[3] == "replies" {
		id, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			http.Error(w, `{"error":"invalid thread id"}`, http.StatusBadRequest)
			return
		}
		if r.Method == http.MethodPost {
			h.createReply(w, r, board, id)
			return
		}
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	http.NotFound(w, r)
}

func isValidBoard(s string) bool {
	if len(s) < 1 || len(s) > 10 {
		return false
	}
	for _, c := range s {
		if c < 'a' || c > 'z' {
			return false
		}
	}
	return true
}

func (h *BoardHandler) listThreads(w http.ResponseWriter, r *http.Request, board string) {
	page := atoi(r.URL.Query().Get("page"), 1)
	limit := atoi(r.URL.Query().Get("limit"), 10)
	res, err := h.svc.ListThreads(board, page, limit)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, `{"error":"board not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"failed"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *BoardHandler) getThread(w http.ResponseWriter, r *http.Request, board string, id int64) {
	page := atoi(r.URL.Query().Get("page"), 1)
	limit := atoi(r.URL.Query().Get("limit"), 15)
	res, err := h.svc.GetThread(board, id, page, limit)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"failed"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *BoardHandler) createThread(w http.ResponseWriter, r *http.Request, board string) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, `{"error":"invalid form"}`, http.StatusBadRequest)
		return
	}
	body := strings.TrimSpace(r.FormValue("body"))
	if len(body) > 5000 {
		http.Error(w, `{"error":"body too long"}`, http.StatusBadRequest)
		return
	}
	hasImage := false
	var url *string
	var imageName *string
	var fileSize *int
	if fh, ok := r.MultipartForm.File["image"]; ok && len(fh) > 0 && fh[0].Size > 0 {
		fh0 := fh[0]
		if fh0.Size > 8<<20 {
			http.Error(w, `{"error":"image too large (max 8MB)"}`, http.StatusBadRequest)
			return
		}
		ext := strings.ToLower(filepath.Ext(fh0.Filename))
		if !isAllowedExt(ext) {
			http.Error(w, `{"error":"invalid image type"}`, http.StatusBadRequest)
			return
		}
		f, err := fh0.Open()
		if err != nil {
			http.Error(w, `{"error":"upload failed"}`, http.StatusInternalServerError)
			return
		}
		defer f.Close()
		filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), randomHex(6), ext)
		saved, err := h.storage.Save(r.Context(), filename, f, fh0.Size, fh0.Header.Get("Content-Type"))
		if err != nil {
			http.Error(w, `{"error":"upload failed"}`, http.StatusInternalServerError)
			return
		}
		url = &saved
		imageName = &fh0.Filename
		sz := int(fh0.Size)
		fileSize = &sz
		hasImage = true
	}
	if body == "" || !hasImage {
		if url != nil {
			_ = h.storage.Delete(r.Context(), *url)
		}
		if body == "" && !hasImage {
			http.Error(w, `{"error":"body and image required"}`, http.StatusBadRequest)
		} else if body == "" {
			http.Error(w, `{"error":"body required"}`, http.StatusBadRequest)
		} else {
			http.Error(w, `{"error":"image required"}`, http.StatusBadRequest)
		}
		return
	}
	ipHash := hashIP(r.RemoteAddr)
	country := h.geo.Country(r)
	tid, err := h.svc.CreateThread(board, body, url, imageName, fileSize, country, ipHash)
	if err != nil {
		if url != nil {
			_ = h.storage.Delete(r.Context(), *url)
		}
		http.Error(w, `{"error":"create failed"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": tid})
}

func (h *BoardHandler) createReply(w http.ResponseWriter, r *http.Request, board string, threadID int64) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, `{"error":"invalid form"}`, http.StatusBadRequest)
		return
	}
	body := strings.TrimSpace(r.FormValue("body"))
	if body == "" && r.MultipartForm.File["image"] == nil {
		http.Error(w, `{"error":"body or image required"}`, http.StatusBadRequest)
		return
	}
	if len(body) > 5000 {
		http.Error(w, `{"error":"body too long"}`, http.StatusBadRequest)
		return
	}
	var imageURL *string
	var imageName *string
	var fileSize *int
	if fh, ok := r.MultipartForm.File["image"]; ok && len(fh) > 0 {
		fh0 := fh[0]
		if fh0.Size > 8<<20 {
			http.Error(w, `{"error":"image too large (max 8MB)"}`, http.StatusBadRequest)
			return
		}
		ext := strings.ToLower(filepath.Ext(fh0.Filename))
		if !isAllowedExt(ext) {
			http.Error(w, `{"error":"invalid image type"}`, http.StatusBadRequest)
			return
		}
		f, err := fh0.Open()
		if err != nil {
			http.Error(w, `{"error":"upload failed"}`, http.StatusInternalServerError)
			return
		}
		defer f.Close()
		filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), randomHex(6), ext)
		url, err := h.storage.Save(r.Context(), filename, f, fh0.Size, fh0.Header.Get("Content-Type"))
		if err != nil {
			http.Error(w, `{"error":"upload failed"}`, http.StatusInternalServerError)
			return
		}
		imageURL = &url
		imageName = &fh0.Filename
		sz := int(fh0.Size)
		fileSize = &sz
	}
	ipHash := hashIP(r.RemoteAddr)
	country := h.geo.Country(r)
	pid, err := h.svc.CreateReply(threadID, board, body, imageURL, imageName, fileSize, country, ipHash)
	if err != nil {
		if imageURL != nil {
			_ = h.storage.Delete(r.Context(), *imageURL)
		}
		if strings.Contains(err.Error(), "locked") {
			http.Error(w, `{"error":"thread locked"}`, http.StatusForbidden)
			return
		}
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, `{"error":"thread not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"create failed"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": pid})
}

func atoi(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func isAllowedExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func hashIP(ip string) string {
	h := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(h[:8])
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		for i := range b {
			b[i] = byte(time.Now().UnixNano() >> (i * 2))
		}
	}
	return hex.EncodeToString(b)
}