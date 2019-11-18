module github.com/giantswarm/k8sportforward

go 1.13

require (
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/fortytw2/leaktest v1.3.1-0.20190606143808-d73c753520d9
	github.com/giantswarm/microerror v0.0.0-20191011121515-e0ebc4ecf5a5
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/juju/errgo v0.0.0-20140925100237-08cceb5d0b53 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/crypto v0.0.0-20191117063200-497ca9f6d64f // indirect
	golang.org/x/net v0.0.0-20191116160921-f9c825593386 // indirect
	golang.org/x/sys v0.0.0-20191118133127-cf1e2d577169 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.2.5 // indirect
	k8s.io/apimachinery v0.0.0 // see replace
	k8s.io/client-go v0.0.0 // see replace
	k8s.io/klog v1.0.0 // indirect
	k8s.io/utils v0.0.0 // indirect
)

replace (
	k8s.io/apimachinery v0.0.0 => k8s.io/apimachinery v0.0.0-20191028221656-72ed19daf4bb // kubernetes-1.16.3
	k8s.io/client-go v0.0.0 => k8s.io/client-go v0.0.0-20191114101535-6c5935290e33 // kubernetes-1.16.3
	k8s.io/utils v0.0.0 => k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6 // kubernetes-1.16.3
)
