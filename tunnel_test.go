package k8sportforward

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"k8s.io/client-go/rest"
)

func Test_NoGlogOutput(t *testing.T) {
	//client := fake.NewSimpleClientset()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	cfg := Config{
		RestConfig: &rest.Config{
			Host: ts.URL,
		},
	}

	f, err := New(cfg)
	if err != nil {
		t.Fatalf("could not create forwarder %v", err)
	}

	tunnelCfg := TunnelConfig{
		Remote:    1,
		Namespace: "test",
		PodName:   "test",
	}

	_, err = f.ForwardPort(tunnelCfg)
	if err != nil {
		t.Fatalf("could not create port forward %v", err)
	}
}
