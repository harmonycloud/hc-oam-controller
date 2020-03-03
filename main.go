package main

import (
	"flag"
	"hc-oam-controller/controllers"
	"log"

	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	"github.com/oam-dev/oam-go-sdk/pkg/client/clientset/versioned"
	"github.com/oam-dev/oam-go-sdk/pkg/oam"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	ctrl.SetLogger(zap.Logger(true))
	_ = v1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	metricsAddr = ""
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.Parse()
	options := ctrl.Options{Scheme: scheme, MetricsBindAddress: metricsAddr}
	// init
	oam.InitMgr(ctrl.GetConfigOrDie(), options)
	clientset, err := kubernetes.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		panic(err)
	}
	oamclient, err := versioned.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		log.Fatal("create client err: ", err)
	}
	// register workloadtpye & trait hooks and handlers
	oam.RegisterHandlers(oam.STypeApplicationConfiguration, &controllers.Handler{Name: "application-configuration-handler", Oamclient: oamclient, K8sclient: clientset})

	// reconcilers must register manualy
	// cloudnativeapp/oam-runtime/pkg/oam as a pkg should not do os.Exit(), instead of
	// panic or returning Error could be better
	err = oam.Run(oam.WithApplicationConfiguration())
	if err != nil {
		panic(err)
	}
}
