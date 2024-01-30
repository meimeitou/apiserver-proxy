package apiserverproxy

import (
	"fmt"
	"net/http"
	"net/url"

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

	KubeConfig   string `json:"kubeconfig,omitempty"`
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

func (h *HTTPTransport) SetScheme(req *http.Request) {
	u, _ := url.Parse(h.Config.Host)
	req.URL.Scheme = u.Scheme
	req.URL.Host = u.Host
}

// RoundTrip implements http.RoundTripper.
func (h *HTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if h.Config.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.Config.BearerToken)
	}
	h.SetScheme(req)
	h.logger.Info("request", zap.String("req", fmt.Sprintf("%+v", req)))
	return h.RoundTripper.RoundTrip(req)
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
