module hc-oam-controller

go 1.12

require (
	github.com/oam-dev/oam-go-sdk v0.0.0-20200311031835-ce9ec52bd420
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.4.0
)

replace github.com/oam-dev/oam-go-sdk => github.com/chenbilong/oam-go-sdk v0.0.0-20200416154853-f4529ed960a7
