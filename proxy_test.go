package apiserverproxy

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func TestTransport(t *testing.T) {
	config := `
        transport apiserver {
            kubeconfig  ~/.kube/config-k8s-dev
        }
    `
	disp := caddyfile.NewTestDispenser(config)
	hd := HTTPTransport{}
	err := hd.UnmarshalCaddyfile(disp)
	if err != nil {
		t.Fatal(err)
	}

	err = hd.Provision(caddy.Context{})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", hd.Config.Host, "/apis/apps/v1/namespaces/default/deployments"), nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := hd.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))

	err = hd.Cleanup()
	if err != nil {
		t.Fatal(err)
	}
}
