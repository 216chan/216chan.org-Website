package geoip

import (
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

type Resolver struct {
	db *geoip2.Reader
}

func New() *Resolver {
	path := os.Getenv("GEOIP_DB")
	if path == "" {
		path = "./GeoLite2-Country.mmdb"
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if _, err2 := os.Stat("/data/GeoLite2-Country.mmdb"); err2 == nil {
			path = "/data/GeoLite2-Country.mmdb"
		}
	}
	db, err := geoip2.Open(path)
	if err != nil {
		return &Resolver{db: nil}
	}
	return &Resolver{db: db}
}

func (r *Resolver) Country(rh *http.Request) string {
	ipStr := clientIP(rh)
	ip := net.ParseIP(ipStr)
	if ip == nil || r.db == nil {
		return "XX"
	}
	record, err := r.db.Country(ip)
	if err != nil {
		return "XX"
	}
	if record.Country.IsoCode == "" {
		return "XX"
	}
	return record.Country.IsoCode
}

func (r *Resolver) Close() {
	if r.db != nil {
		r.db.Close()
	}
}

func clientIP(r *http.Request) string {
	if cf := r.Header.Get("CF-Connecting-IP"); cf != "" {
		return strings.TrimSpace(strings.Split(cf, ",")[0])
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}