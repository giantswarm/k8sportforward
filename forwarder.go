package k8sportforward

import (
	"net/http"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport/spdy"
)

type ForwarderConfig struct {
	RestConfig *rest.Config
}

type Forwarder struct {
	restConfig *rest.Config

	restClient rest.Interface
}

func NewForwarder(config ForwarderConfig) (*Forwarder, error) {
	if config.RestConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RestConfig must not be empty", config)
	}

	var err error

	var restClient rest.Interface
	{
		restConfigShallowCopy := *config.RestConfig

		// We need to configre the config in order to generate correct
		// URLs for dialer.
		setConfigDefaults(&restConfigShallowCopy)

		restClient, err = rest.RESTClientFor(&restConfigShallowCopy)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	f := &Forwarder{
		restConfig: config.RestConfig,

		restClient: restClient,
	}

	return f, nil
}

// ForwardPort opens a tunnel to a kubernetes pod.
func (f *Forwarder) ForwardPort(namespace string, podName string, remotePort int) (*Tunnel, error) {
	transport, upgrader, err := spdy.RoundTripperFor(f.restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	config := tunnelConfig{
		Dialer: spdy.NewDialer(
			upgrader,
			&http.Client{
				Transport: transport,
			},
			"POST",
			// Build a url to the portforward endpoint.
			// Example: http://localhost:8080/api/v1/namespaces/helm/pods/tiller-deploy-9itlq/portforward
			f.restClient.Post().Resource("pods").Namespace(namespace).Name(podName).SubResource("portforward").URL(),
		),

		RemotePort: remotePort,
	}

	tunnel, err := newTunnel(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tunnel, nil
}

// setConfigDefaults is copied and adjusted from client-go core/v1.
func setConfigDefaults(config *rest.Config) error {
	config.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	config.APIPath = "/api"
	{
		s := runtime.NewScheme()
		c := serializer.NewCodecFactory(s)
		config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: c}
	}
	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}
