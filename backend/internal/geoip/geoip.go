package geoip

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"

	"github.com/oschwald/maxminddb-golang"
)

type Observer struct {
	db      *maxminddb.Reader
	trusted []netip.Prefix
}

func Open(path, proxies string) (*Observer, error) {
	o := &Observer{}
	if proxies != "" {
		for _, raw := range strings.Split(proxies, ",") {
			p, err := netip.ParsePrefix(strings.TrimSpace(raw))
			if err != nil {
				return nil, fmt.Errorf("invalid GEOIP_TRUSTED_PROXIES CIDR")
			}
			if p.Addr().Is4In6() {
				if p.Bits() < 96 {
					return nil, fmt.Errorf("invalid GEOIP_TRUSTED_PROXIES CIDR")
				}
				p = netip.PrefixFrom(p.Addr().Unmap(), p.Bits()-96)
			}
			o.trusted = append(o.trusted, p.Masked())
		}
	}
	if path == "" {
		return o, nil
	}
	db, err := maxminddb.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open GEOIP_DB: %w", err)
	}
	if db.Metadata.DatabaseType != "GeoIP2-Country" && db.Metadata.DatabaseType != "GeoLite2-Country" {
		db.Close()
		return nil, fmt.Errorf("GEOIP_DB must be a country database")
	}
	o.db = db
	return o, nil
}

func (o *Observer) Close() error {
	if o == nil || o.db == nil {
		return nil
	}
	return o.db.Close()
}

func validAddress(a netip.Addr) bool {
	a = a.Unmap()
	return a.IsGlobalUnicast() && !a.IsPrivate()
}

// Trust X-Real-IP only when ingress overwrites it for every request.
func (o *Observer) Address(r *http.Request) netip.Addr {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}
	}
	peer, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}
	}
	peer = peer.Unmap()
	trusted := false
	if o != nil {
		for _, p := range o.trusted {
			if p.Contains(peer) {
				trusted = true
				break
			}
		}
	}
	if !trusted {
		return peer
	}
	values := r.Header.Values("X-Real-IP")
	if len(values) != 1 || values[0] != strings.TrimSpace(values[0]) || strings.ContainsAny(values[0], ", \t\r\n") {
		return netip.Addr{}
	}
	address, err := netip.ParseAddr(values[0])
	if err != nil || address.Zone() != "" {
		return netip.Addr{}
	}
	return address.Unmap()
}

func normalize(code string) string {
	if len(code) != 2 || code == "XX" {
		return ""
	}
	for i := 0; i < 2; i++ {
		if code[i] < 'A' || code[i] > 'Z' {
			return ""
		}
	}
	return code
}

func (o *Observer) Country(r *http.Request) string {
	if o == nil || o.db == nil {
		return ""
	}
	address := o.Address(r)
	if !validAddress(address) {
		return ""
	}
	var result struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}
	if err := o.db.Lookup(net.IP(address.AsSlice()), &result); err != nil {
		return ""
	}
	return normalize(result.Country.ISOCode)
}
