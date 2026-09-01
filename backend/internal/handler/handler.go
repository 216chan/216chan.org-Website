package handler

import (
	"encoding/json"
	"net/http"

	"216chan/backend/internal/service"
)

type StatsHandler struct {
	svc *service.StatsService
}

func NewStatsHandler(svc *service.StatsService) *StatsHandler {
	return &StatsHandler{svc: svc}
}

func (h *StatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetStats()
	if err != nil {
		http.Error(w, `{"error":"failed to fetch stats"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
