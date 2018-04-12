package k8sportforward

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"

	"github.com/giantswarm/microerror"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type Config struct {
	RestConfig *rest.Config
}

type Forwarder struct {
	k8sClient  kubernetes.Interface
	restConfig *rest.Config
}

func New(config Config) (*Forwarder, error) {
	if config.RestConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.RestConfig must not be empty")
	}

	k8sClient, err := kubernetes.NewForConfig(config.RestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &Forwarder{
		k8sClient:  k8sClient,
		restConfig: config.RestConfig,
	}, nil
}

// ForwardPort opens a tunnel to a kubernetes pod.
func (f *Forwarder) ForwardPort(cofnig TunnelConfig) (*Tunnel, error) {
	// Build a url to the portforward endpoint.
	// Example: http://localhost:8080/api/v1/namespaces/helm/pods/tiller-deploy-9itlq/portforward
	u := f.k8sClient.CoreV1().RESTClient.Post().
		Resource("pods").
		Namespace(config.Namespace).
		Name(config.PodName).
		SubResource("portforward").URL()

	transport, upgrader, err := spdy.RoundTripperFor(f.restConfig)
	if err != nil {
		return microerror.Mask(err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", u)

	local, err := getAvailablePort()
	if err != nil {
		return microerror.Mask(err)
	}

	tunnel := &Tunnel{
		TunnelConfig: config,
		Local:        local,

		stopChan: make(chan struct{}, 1),
	}

	out := ioutil.Discard
	ports := []string{fmt.Sprintf("%d:%d", tunnel.Local, tunnel.Remote)}
	readyChan := make(chan struct{}, 1)

	pf, err := portforward.New(dialer, ports, tunnel.stopChan, readyChan, out, out)
	if err != nil {
		return microerror.Mask(err)
	}

	errChan := make(chan error)
	go func() {
		errChan <- pf.ForwardPorts()
	}()

	select {
	case err = <-errChan:
		return nil, microerror.Mask(err)
	case <-pf.Ready:
		return tunnel, nil
	}
}

type TunnelConfig struct {
	Remote    int
	Namespace string
	PodName   string
}

type Tunnel struct {
	TunnelConfig
	Local int

	stopChan chan struct{}
}

// Close disconnects a tunnel connection.
func (t *Tunnel) Close() error {
	close(t.stopChan)
	return nil
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
