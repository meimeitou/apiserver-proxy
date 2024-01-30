package apiserverproxy

import (
	"regexp"

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
			case "reject_paths":
				for d.NextArg() {
					reg, err := regexp.Compile(d.Val())
					if err != nil {
						d.Errf("reject_paths regexp regexp format error %s", d.Val())
					}
					h.RejectPaths = append(h.RejectPaths, reg)
				}
			case "accept_hosts":
				for d.NextArg() {
					reg, err := regexp.Compile(d.Val())
					if err != nil {
						d.Errf("accept_hosts regexp regexp format error %s", d.Val())
					}
					h.AcceptHosts = append(h.AcceptHosts, reg)
				}
			default:
				return d.Errf("apiserver unrecognized subdirective %s", d.Val())
			}
		}
	}
	return nil
}
