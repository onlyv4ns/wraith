package transport

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

func NewClient() *http.Client {
	return &http.Client{Transport: newTransport()}
}

type roundTripper struct {
	h2 *http2.Transport
	h1 *http.Transport
}

func newTransport() *roundTripper {
	t := &roundTripper{}

	t.h2 = &http2.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			return dialUTLS(ctx, network, addr, []string{"h2"})
		},
		MaxHeaderListSize: 262144,
	}

	t.h1 = &http.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialUTLS(ctx, network, addr, []string{"http/1.1"})
		},
	}

	return t
}

func (t *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "https" {
		return http.DefaultTransport.RoundTrip(req)
	}
	resp, err := t.h2.RoundTrip(req)
	if err != nil {
		return t.h1.RoundTrip(req)
	}
	return resp, nil
}

func dialUTLS(ctx context.Context, network, addr string, alpn []string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	raw, err := (&net.Dialer{}).DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	conn := utls.UClient(raw, &utls.Config{
		ServerName: host,
		NextProtos: alpn,
	}, utls.HelloChrome_Auto)

	if err := conn.HandshakeContext(ctx); err != nil {
		raw.Close()
		return nil, err
	}

	return conn, nil
}
