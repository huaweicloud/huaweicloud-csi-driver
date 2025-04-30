package transport

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/tjfoc/gmsm/gmtls"
)

func NewGMRoundTripper(cfg *gmtls.Config) WrapperFunc {
	return func(inner http.RoundTripper) http.RoundTripper {
		return &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.InsecureSkipVerify,
			},
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialer := &net.Dialer{}
				conn, err := gmtls.DialWithDialer(dialer, network, addr, cfg)
				if err != nil {
					return nil, err
				}
				return conn, nil
			},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialer := &net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 60 * time.Second,
				}
				conn, err := gmtls.DialWithDialer(dialer, network, addr, cfg)
				if err != nil {
					return nil, err
				}
				return conn, nil
			},
			TLSHandshakeTimeout: 15 * time.Second,
			IdleConnTimeout:     30 * time.Second,
		}
	}
}
