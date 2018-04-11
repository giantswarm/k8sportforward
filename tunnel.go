package k8sportforward

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/giantswarm/microerror"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type Config struct {
	K8sClient  rest.Interface
	RestConfig *rest.Config

	Namespace string
	// Remote port to connect to.
	Remote int
	// Resource includes type and name of the resource, like svc/cnr-server or
	// pod/mypod-3445kdfg.
	Resource string
}

// Tunnel describes a ssh-like tunnel to a kubernetes pod.
type Tunnel struct {
	Local        int
	Remote       int
	Namespace    string
	ResourceName string
	ResourceType string
	Out          io.Writer
	stopChan     chan struct{}
	readyChan    chan struct{}
	restCfg      *rest.Config
	client       rest.Interface
}

// NewTunnel creates a new tunnel.
func NewTunnel(config *Config) (*Tunnel, error) {
	items := strings.Split(config.Resource, "/")
	if len(items) != 2 {
		return nil, microerror.Mask(invalidConfigError)
	}

	return &Tunnel{
		restCfg:      config.RestConfig,
		client:       config.K8sClient,
		Namespace:    config.Namespace,
		ResourceType: items[0],
		ResourceName: items[1],
		Remote:       config.Remote,
		stopChan:     make(chan struct{}, 1),
		readyChan:    make(chan struct{}, 1),
		Out:          ioutil.Discard,
	}, nil
}

// Close disconnects a tunnel connection.
func (t *Tunnel) Close() {
	close(t.stopChan)
}

// ForwardPort opens a tunnel to a kubernetes pod.
func (t *Tunnel) ForwardPort() error {
	// Build a url to the portforward endpoint.
	// Example: http://localhost:8080/api/v1/namespaces/helm/pods/tiller-deploy-9itlq/portforward
	u := t.client.Post().
		Resource(t.ResourceType).
		Namespace(t.Namespace).
		Name(t.ResourceName).
		SubResource("portforward").URL()

	transport, upgrader, err := spdy.RoundTripperFor(t.restCfg)
	if err != nil {
		return microerror.Mask(err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", u)

	local, err := getAvailablePort()
	if err != nil {
		return microerror.Mask(err)
	}
	t.Local = local

	ports := []string{fmt.Sprintf("%d:%d", t.Local, t.Remote)}

	pf, err := portforward.New(dialer, ports, t.stopChan, t.readyChan, t.Out, t.Out)
	if err != nil {
		return microerror.Mask(err)
	}

	errChan := make(chan error)
	go func() {
		errChan <- pf.ForwardPorts()
	}()

	select {
	case err = <-errChan:
		return microerror.Mask(err)
	case <-pf.Ready:
		return nil
	}
}

func getAvailablePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, microerror.Mask(err)
	}
	defer l.Close()

	_, p, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return 0, microerror.Mask(err)
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return 0, microerror.Mask(err)
	}
	return port, microerror.Mask(err)
}
