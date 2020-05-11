package main

import (
	"flag"
	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	"github.com/oam-dev/oam-go-sdk/pkg/client/clientset/versioned"
	"github.com/oam-dev/oam-go-sdk/pkg/oam"
	hcv1alpha1 "hc-oam-controller/api/harmonycloud.cn/v1alpha1"
	hcv1beta1 "hc-oam-controller/api/harmonycloud.cn/v1beta1"
	hcversioned "hc-oam-controller/client/clientset/versioned"
	"hc-oam-controller/controllers"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	kubescheme "k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"log"
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
	_ = hcv1alpha1.AddToScheme(scheme)
	_ = hcv1beta1.AddToScheme(scheme)
	_ = v1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)
	_ = v1beta1.AddToScheme(scheme)
	_ = v2beta2.AddToScheme(scheme)

	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	metricsAddr = ""
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.Parse()
	options := ctrl.Options{Scheme: scheme, MetricsBindAddress: metricsAddr}
	//options := ctrl.Options{Scheme: scheme}

	// init
	// set up signals so we handle the first shutdown signal gracefully

	oam.InitMgr(ctrl.GetConfigOrDie(), options)
	clientset, err := kubernetes.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		panic(err)
	}
	oamclient, err := versioned.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		log.Fatal("create oam client err: ", err)
	}

	hcClient, err := hcversioned.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		log.Fatal("create hc client err: ", err)
	}

	//event
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: clientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(kubescheme.Scheme, corev1.EventSource{Component: "hc-oam-controller"})

	// register workloadtpye & trait hooks and handlers
	oam.RegisterHandlers(oam.STypeApplicationConfiguration,
		&controllers.ApplicationConfigurationHandler{Name: "application-configuration-handler", Oamclient: oamclient, K8sclient: clientset, Hcclient: hcClient, Recorder: recorder})
	oam.RegisterObject("deployment", new(v1.Deployment))
	oam.RegisterHandlers("deployment", &controllers.DeploymentHandler{Name: "deployment-handler", Oamclient: oamclient, K8sclient: clientset})
	oam.RegisterObject("service", new(corev1.Service))
	oam.RegisterHandlers("service", &controllers.ServiceHandler{Name: "service-handler", Oamclient: oamclient, K8sclient: clientset})
	oam.RegisterObject("configmap", new(corev1.ConfigMap))
	oam.RegisterHandlers("configmap", &controllers.ConfigMapHandler{Name: "configmap-handler", Oamclient: oamclient, K8sclient: clientset})
	oam.RegisterObject("persistentvolumeclaim", new(corev1.PersistentVolumeClaim))
	oam.RegisterHandlers("persistentvolumeclaim", &controllers.PvcHandler{Name: "pvc-handler", Oamclient: oamclient, K8sclient: clientset})
	oam.RegisterObject("job", new(batchv1.Job))
	oam.RegisterHandlers("job", &controllers.JobHandler{Name: "job-handler", Oamclient: oamclient, K8sclient: clientset})
	oam.RegisterObject("mysqlcluster", new(hcv1alpha1.MysqlCluster))
	oam.RegisterObject("ingress", new(v1beta1.Ingress))
	oam.RegisterHandlers("ingress", &controllers.IngressHandler{Name: "ingress-handler", Oamclient: oamclient, K8sclient: clientset})
	oam.RegisterObject("hpa", new(v2beta2.HorizontalPodAutoscaler))
	oam.RegisterHandlers("hpa", &controllers.HpaHandler{Name: "hpa-handler", Oamclient: oamclient, K8sclient: clientset})
	oam.RegisterObject("hchpa", new(hcv1beta1.HorizontalPodAutoscaler))
	oam.RegisterHandlers("hchpa", &controllers.HcHpaHandler{Name: "hchpa-handler", Oamclient: oamclient, K8sclient: clientset})

	// reconcilers must register manualy
	// cloudnativeapp/oam-runtime/pkg/oam as a pkg should not do os.Exit(), instead of
	// panic or returning Error could be better
	err = oam.Run(oam.WithApplicationConfiguration(),
		oam.WithSpec("deployment"),
		oam.WithSpec("service"),
		oam.WithSpec("configmap"),
		oam.WithSpec("persistentvolumeclaim"),
		oam.WithSpec("job"),
		//oam.WithSpec("mysqlcluster"),
		//oam.WithSpec("hpa"),
		oam.WithSpec("hchpa"),
		oam.WithSpec("ingress"),
	)

	if err != nil {
		panic(err)
	}
}
