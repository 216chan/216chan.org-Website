package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"216chan/backend/internal/config"
	"216chan/backend/internal/handler"
	"216chan/backend/internal/middleware"
	"216chan/backend/internal/repository"
	"216chan/backend/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("connected to postgres")

	statsRepo := repository.NewStatsRepo(db)
	statsSvc := service.NewStatsService(statsRepo)
	statsHandler := handler.NewStatsHandler(statsSvc)

	mux := http.NewServeMux()
	mux.Handle("/api/stats", statsHandler)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, middleware.CORS(mux)); err != nil {
		log.Fatalf("server: %v", err)
	}
}
