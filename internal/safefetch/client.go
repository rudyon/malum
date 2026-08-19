// Package safefetch provides the default network boundary for URL ingestion.
package safefetch

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

var ErrPrivateDestination = errors.New("URL resolves to a non-public network address")

const (
	requestTimeout = 90 * time.Second
	maxRedirects   = 10
)

func NewClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = publicDialer(net.DefaultResolver, &net.Dialer{Timeout: 15 * time.Second, KeepAlive: 30 * time.Second})
	return &http.Client{
		Transport: transport,
		Timeout:   requestTimeout,
		CheckRedirect: func(request *http.Request, previous []*http.Request) error {
			if len(previous) >= maxRedirects {
				return errors.New("too many redirects")
			}
			return validateURL(request.URL)
		},
	}
}

type resolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}

type contextDialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
}

func publicDialer(resolver resolver, dialer contextDialer) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, _, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("parse destination address: %w", err)
		}
		addresses, err := resolve(ctx, resolver, host)
		if err != nil {
			return nil, err
		}
		if err := validateAddresses(addresses); err != nil {
			return nil, fmt.Errorf("refuse destination %q: %w", host, err)
		}
		var failures []error
		for _, resolved := range addresses {
			connection, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(resolved.IP.String(), port))
			if err == nil {
				return connection, nil
			}
			failures = append(failures, err)
		}
		return nil, fmt.Errorf("connect to %q: %w", host, errors.Join(failures...))
	}
}

func resolve(ctx context.Context, resolver resolver, host string) ([]net.IPAddr, error) {
	if parsed := net.ParseIP(strings.Trim(host, "[]")); parsed != nil {
		return []net.IPAddr{{IP: parsed}}, nil
	}
	addresses, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve destination %q: %w", host, err)
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("resolve destination %q: no addresses", host)
	}
	return addresses, nil
}

func validateAddresses(addresses []net.IPAddr) error {
	for _, address := range addresses {
		ip := address.IP
		if !globallyRoutable(ip) {
			return fmt.Errorf("%w: %s", ErrPrivateDestination, ip)
		}
	}
	return nil
}

var nonPublicPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("::/128"),
	netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("fc00::/7"),
	netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("ff00::/8"),
}

func globallyRoutable(ip net.IP) bool {
	address, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	address = address.Unmap()
	for _, prefix := range nonPublicPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return address.IsGlobalUnicast() && !address.IsPrivate()
}

func validateURL(value *url.URL) error {
	if value == nil || (value.Scheme != "http" && value.Scheme != "https") || value.Hostname() == "" {
		return errors.New("redirect destination must be an HTTP or HTTPS URL")
	}
	return nil
}
