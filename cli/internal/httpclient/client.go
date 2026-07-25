// Package httpclient provides HTTP clients that can recover from broken local
// DNS resolvers without weakening HTTPS verification.
package httpclient

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

var fallbackResolver = &net.Resolver{
	PreferGo: true,
	Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
		// Query Cloudflare's public resolver directly only after the system
		// resolver has failed. The HTTP request still uses its original host,
		// so TLS SNI and certificate validation are unchanged.
		return (&net.Dialer{}).DialContext(ctx, "udp", "1.1.1.1:53")
	},
}

// New returns a client that uses the system resolver first. If a network's
// DNS proxy incorrectly reports that a hostname does not exist, it retries
// that lookup via a public resolver before failing the connection.
func New(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		conn, err := dialer.DialContext(ctx, network, address)
		if err == nil {
			return conn, nil
		}
		var dnsErr *net.DNSError
		if !errors.As(err, &dnsErr) || !dnsErr.IsNotFound {
			return nil, err
		}

		host, port, splitErr := net.SplitHostPort(address)
		if splitErr != nil {
			return nil, err
		}
		ips, lookupErr := fallbackResolver.LookupIPAddr(ctx, host)
		if lookupErr != nil {
			return nil, err
		}
		for _, ip := range ips {
			conn, fallbackErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
			if fallbackErr == nil {
				return conn, nil
			}
		}
		return nil, err
	}

	return &http.Client{Timeout: timeout, Transport: transport}
}
