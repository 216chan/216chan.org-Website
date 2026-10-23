package netlandlock

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func Ports(address string, databaseURL string) (database, listen uint16, err error) {
	_, rawListen, err := net.SplitHostPort(address)
	if err != nil {
		return 0, 0, fmt.Errorf("network Landlock requires a host:port listen address: %w", err)
	}
	listen, err = port(rawListen)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid listen port: %w", err)
	}
	u, e := url.Parse(databaseURL)
	if e != nil {
		return 0, 0, fmt.Errorf("network Landlock requires PostgreSQL URL DSNs")
	}
	host := u.Hostname()
	if host != "localhost" && net.ParseIP(host) == nil {
		return 0, 0, fmt.Errorf("network Landlock requires a numeric database host or localhost, got %q", host)
	}
	rawPort := u.Port()
	if rawPort == "" {
		if strings.HasSuffix(u.Host, ":") {
			return 0, 0, fmt.Errorf("empty database port")
		}
		rawPort = "5432"
	}
	value, e := port(rawPort)
	if e != nil {
		return 0, 0, fmt.Errorf("invalid database port: %w", e)
	}
	return value, listen, nil
}

func port(raw string) (uint16, error) {
	n, err := strconv.ParseUint(raw, 10, 16)
	if err != nil || n == 0 {
		return 0, fmt.Errorf("port must be 1–65535")
	}
	return uint16(n), nil
}
