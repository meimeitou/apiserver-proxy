package apiserverproxy

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
)

// UnmarshalCaddyfile deserializes Caddyfile tokens into h.
//
//	transport http {
//	    kubeconfig <kubeconfig>
//	}
func (h *HTTPTransport) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "kubeconfig":
				if d.NextArg() {
					h.KubeConfig = d.Val()
				}
				if h.TLS == nil {
					h.TLS = new(reverseproxy.TLSConfig)
				}
			default:
				return d.Errf("apiserver unrecognized subdirective %s", d.Val())
			}
		}
	}
	return nil
}
