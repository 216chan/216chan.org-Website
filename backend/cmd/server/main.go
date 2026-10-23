package main

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"image"
	"image/png"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"216chan/backend/internal/geoip"
	"216chan/backend/internal/netlandlock"
	"216chan/backend/internal/publication"
	"216chan/backend/internal/web"
	_ "github.com/lib/pq"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	if err := run(os.Args[1:]); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
func open(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("database URL required")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(4)
	return db, nil
}
func run(args []string) error {
	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	install := flags.Bool("install", false, "install a fresh dedicated database using PUBLICATION_OWNER_URL and existing request roles")
	check := flags.Bool("check", false, "verify request roles and all serving configuration, then exit")
	if err := flags.Parse(args); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if *install {
		db, err := open(os.Getenv("PUBLICATION_OWNER_URL"))
		if err != nil {
			return err
		}
		defer db.Close()
		return publication.Install(ctx, db, os.Getenv("PUBLICATION_READER_ROLE"), os.Getenv("PUBLICATION_WRITER_ROLE"), os.Getenv("PUBLICATION_STAFF_ROLE"))
	}
	address := os.Getenv("PUBLICATION_LISTEN")
	if address == "" {
		address = "127.0.0.1:8081"
	}
	dsns, err := databaseURLs()
	if err != nil {
		return err
	}
	landlock, err := envBool("PUBLICATION_NETWORK_LANDLOCK")
	if err != nil {
		return err
	}
	if landlock {
		databasePort, listenPort, err := netlandlock.Ports(address, os.Getenv("PUBLICATION_DATABASE_URL"))
		if err != nil {
			return err
		}
		if os.Getenv("PUBLICATION_NETWORK_LANDLOCK_APPLIED") != "1" {
			return netlandlock.Reexec(databasePort, listenPort)
		}
	}
	var dbs [3]*sql.DB
	for i, dsn := range dsns {
		if dbs[i], err = open(dsn); err != nil {
			return err
		}
		defer dbs[i].Close()
	}
	if err = publication.CheckRoles(ctx, dbs[0], dbs[1], dbs[2]); err != nil {
		return err
	}
	store, staffStore := publication.New(dbs[0], dbs[1]), publication.NewStaff(dbs[2])
	observer, err := geoip.Open(os.Getenv("GEOIP_DB"), os.Getenv("GEOIP_TRUSTED_PROXIES"))
	if err != nil {
		return err
	}
	defer observer.Close()
	iso, err := imageIsolation(ctx)
	if err != nil {
		return err
	}
	staffConfig, err := staffConfiguration()
	if err != nil {
		return err
	}
	staffHandler, err := web.NewStaffHTTP(staffStore, staffConfig)
	if err != nil {
		return err
	}
	postingDisabled, err := envBool("POSTING_DISABLED")
	if err != nil {
		return err
	}
	if *check { // A successful check must imply serving prerequisites passed.
		return nil
	}
	if origin := os.Getenv("SITE_ORIGIN"); origin != "" {
		web.SiteOrigin = strings.TrimSuffix(origin, "/")
	}
	server := &http.Server{Addr: address, Handler: accessLog(web.NewPublication(store, staffHandler, postingDisabled, observer, iso)), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 35 * time.Second, IdleTimeout: 60 * time.Second}
	stopping, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServe() }()
	slog.Info("listening", "address", address, "image_isolation", iso, "network_landlock", landlock, "geoip", os.Getenv("GEOIP_DB") != "", "posting_disabled", postingDisabled, "staff_tls_proxy", staffConfig.TLSProxy)
	select {
	case err := <-done:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-stopping.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(ctx)
	}
}

func envBool(name string) (bool, error) {
	switch os.Getenv(name) {
	case "", "false":
		return false, nil
	case "true":
		return true, nil
	}
	return false, fmt.Errorf("%s must be true or false", name)
}

func databaseURLs() ([3]string, error) {
	var dsns [3]string
	target, err := url.Parse(os.Getenv("PUBLICATION_DATABASE_URL"))
	if err != nil || (target.Scheme != "postgres" && target.Scheme != "postgresql") || target.Hostname() == "" || target.Path == "" || target.Path == "/" || target.User != nil || target.Fragment != "" {
		return dsns, fmt.Errorf("PUBLICATION_DATABASE_URL must name one PostgreSQL host and database without credentials")
	}
	query, err := url.ParseQuery(target.RawQuery)
	if err != nil {
		return dsns, fmt.Errorf("invalid database URL options")
	}
	for _, key := range []string{"host", "hostaddr", "port", "dbname", "database", "service", "servicefile", "user", "password"} {
		if _, exists := query[key]; exists {
			return dsns, fmt.Errorf("database URL cannot override %s", key)
		}
	}
	for i, kind := range []string{"READER", "WRITER", "STAFF"} {
		if os.Getenv("PUBLICATION_"+kind+"_URL") != "" {
			return dsns, fmt.Errorf("replace per-role URLs with PUBLICATION_DATABASE_URL and role credentials")
		}
		role, password := os.Getenv("PUBLICATION_"+kind+"_ROLE"), os.Getenv("PUBLICATION_"+kind+"_PASSWORD")
		if role == "" || password == "" {
			return dsns, fmt.Errorf("%s role and password required", kind)
		}
		login := *target
		login.User = url.UserPassword(role, password)
		dsns[i] = login.String()
	}
	return dsns, nil
}

// Probe at startup so converter failures surface before accepting uploads.
func imageIsolation(ctx context.Context) (publication.Isolation, error) {
	var iso publication.Isolation
	switch value := os.Getenv("IMAGE_SANDBOX"); value {
	case "", "bwrap":
		iso = publication.Bubblewrap
	case "none":
		iso = publication.Unisolated
		slog.Warn("image conversion runs without namespace isolation; rlimits and the ImageMagick policy still apply")
	default:
		return "", fmt.Errorf("IMAGE_SANDBOX must be bwrap or none")
	}
	var probe bytes.Buffer
	if err := png.Encode(&probe, image.NewGray(image.Rect(0, 0, 1, 1))); err != nil {
		return "", err
	}
	if _, err := publication.ApproveMedia(ctx, iso, &probe); err != nil {
		return "", fmt.Errorf("image converter probe (IMAGE_SANDBOX=%s): %w", iso, err)
	}
	return iso, nil
}

func staffConfiguration() (web.StaffConfig, error) {
	path, inline := os.Getenv("STAFF_CONFIG"), os.Getenv("STAFF_CONFIG_JSON")
	switch {
	case path != "" && inline != "":
		return web.StaffConfig{}, fmt.Errorf("set STAFF_CONFIG or STAFF_CONFIG_JSON, not both")
	case inline != "":
		return web.DecodeStaffConfig(strings.NewReader(inline))
	default:
		return web.LoadStaffConfig(path)
	}
}

type responseLog struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (w *responseLog) WriteHeader(status int) {
	if !w.wrote {
		w.wrote, w.status = true, status
		w.ResponseWriter.WriteHeader(status)
	}
}
func (w *responseLog) Write(b []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(w.status)
	}
	return w.ResponseWriter.Write(b)
}
func (w *responseLog) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func accessLog(next http.Handler) http.Handler {
	var sequence atomic.Uint64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := fmt.Sprintf("%x-%x", time.Now().UnixNano(), sequence.Add(1))
		w.Header().Set("X-Request-ID", id)
		rec := &responseLog{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r)
		slog.Info("request", "request_id", id, "method", r.Method, "path", r.URL.Path, "status", rec.status, "duration_ms", time.Since(start).Milliseconds())
	})
}
