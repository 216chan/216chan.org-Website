package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"

	"216chan/backend/internal/config"
	"216chan/backend/internal/geoip"
	"216chan/backend/internal/handler"
	"216chan/backend/internal/middleware"
	"216chan/backend/internal/repository"
	"216chan/backend/internal/service"
	"216chan/backend/internal/storage"
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

	if err := runMigrations(db); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("migrations ok")

	geo := geoip.New()
	defer geo.Close()

	st := storage.New(cfg.UploadDir)

	statsRepo := repository.NewStatsRepo(db)
	statsSvc := service.NewStatsService(statsRepo)
	statsHandler := handler.NewStatsHandler(statsSvc)

	boardRepo := repository.NewBoardRepo(db)
	boardSvc := service.NewBoardService(boardRepo)
	boardHandler := handler.NewBoardHandler(boardSvc, st, geo)

	mux := http.NewServeMux()
	mux.Handle("/api/stats", statsHandler)
	mux.Handle("/api/", boardHandler)

	uploadDir := cfg.UploadDir
	if uploadDir == "" {
		uploadDir = "./uploads"
		if _, err := os.Stat("/data"); err == nil {
			uploadDir = "/data/uploads"
		}
	}
	if os.Getenv("S3_BUCKET") == "" {
		_ = os.MkdirAll(uploadDir, 0755)
		mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))
		log.Printf("serving uploads from %s", uploadDir)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, middleware.CORS(mux)); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func runMigrations(db *sql.DB) error {
	dirs := []string{"migrations", "./migrations", "../migrations", "backend/migrations", "/app/migrations"}
	var files []string
	var chosen string
	for _, d := range dirs {
		matches, _ := filepath.Glob(filepath.Join(d, "*.sql"))
		if len(matches) > 0 {
			files = matches
			chosen = d
			break
		}
	}
	if len(files) == 0 {
		log.Printf("no migration files found, skipping")
		return nil
	}
	sort.Strings(files)
	log.Printf("running migrations from %s", chosen)
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		sqlStr := string(b)
		if strings.TrimSpace(sqlStr) == "" {
			continue
		}
		if _, err := db.Exec(sqlStr); err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		log.Printf("applied %s", filepath.Base(f))
	}
	return nil
}