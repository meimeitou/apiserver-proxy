package apiserverproxy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

func init() {
	caddy.RegisterModule(HTTPTransport{})
}

type HTTPTransport struct {
	TLS *reverseproxy.TLSConfig `json:"tls,omitempty"`

	KubeConfig string `json:"kubeconfig,omitempty"`
	// ^/api/.*/pods/.*/exec,^/api/.*/pods/.*/attach
	RejectPaths []*regexp.Regexp `json:"reject_paths,omitempty"`
	// ^localhost$,^127\.0\.0\.1$,^\[::1\]$
	AcceptHosts  []*regexp.Regexp `json:"accept_hosts,omitempty"`
	RoundTripper http.RoundTripper
	Config       *rest.Config
	logger       *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (HTTPTransport) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.reverse_proxy.transport.apiserver",
		New: func() caddy.Module { return new(HTTPTransport) },
	}
}

func (h *HTTPTransport) Provision(ctx caddy.Context) error {
	h.logger = ctx.Logger()
	var err error
	if h.Config, err = buildConfig(h.KubeConfig); err != nil {
		return err
	}
	if h.RoundTripper, err = rest.TransportFor(h.Config); err != nil {
		return err
	}
	return nil
}

func (h *HTTPTransport) SetRequest(req *http.Request) {
	u, _ := url.Parse(h.Config.Host)
	req.URL.Scheme = u.Scheme
	req.URL.Host = u.Host
}

// RoundTrip implements http.RoundTripper.
func (h *HTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if h.Config.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.Config.BearerToken)
	}
	h.logger.Info("request", zap.String("req", fmt.Sprintf("%+v", req)))
	h.SetRequest(req)
	if !h.RequestAccept(req) {
		resp := &http.Response{
			Status:     "200 OK",
			StatusCode: 403,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`RBAC Unauthorized`)),
		}
		return resp, nil
	}
	return h.RoundTripper.RoundTrip(req)
}

func (h *HTTPTransport) RequestAccept(req *http.Request) bool {
	if len(h.RejectPaths) > 0 {
		for _, reg := range h.RejectPaths {
			if reg.MatchString(req.URL.Path) {
				return false
			}
		}
	}
	if len(h.AcceptHosts) > 0 {
		for _, reg := range h.AcceptHosts {
			if reg.MatchString(req.Host) {
				return true
			}
		}
		return false
	}
	return true
}

// TLSEnabled returns true if TLS is enabled.
func (h HTTPTransport) TLSEnabled() bool {
	return true
}

// EnableTLS enables TLS on the transport.
func (h *HTTPTransport) EnableTLS(base *reverseproxy.TLSConfig) error {
	h.TLS = base
	return nil
}

func (h HTTPTransport) Cleanup() error {
	if h.RoundTripper == nil {
		return nil
	}
	h.RoundTripper.(*http.Transport).CloseIdleConnections()
	return nil
}

// Interface guards
var (
	_ caddyfile.Unmarshaler     = (*HTTPTransport)(nil)
	_ caddy.Provisioner         = (*HTTPTransport)(nil)
	_ http.RoundTripper         = (*HTTPTransport)(nil)
	_ caddy.CleanerUpper        = (*HTTPTransport)(nil)
	_ reverseproxy.TLSTransport = (*HTTPTransport)(nil)
)
