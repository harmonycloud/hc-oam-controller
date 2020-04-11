package controllers

import (
	hcversioned "hc-oam-controller/client/clientset/versioned"
	"k8s.io/client-go/tools/record"

	"github.com/oam-dev/oam-go-sdk/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

type ApplicationConfigurationHandler struct {
	Name      string
	Oamclient *versioned.Clientset
	K8sclient *kubernetes.Clientset
	Hcclient  *hcversioned.Clientset
	Recorder  record.EventRecorder
}

type DeploymentHandler struct {
	Name      string
	Oamclient *versioned.Clientset
	K8sclient *kubernetes.Clientset
}

type ServiceHandler struct {
	Name      string
	Oamclient *versioned.Clientset
	K8sclient *kubernetes.Clientset
}

type ConfigMapHandler struct {
	Name      string
	Oamclient *versioned.Clientset
	K8sclient *kubernetes.Clientset
}

type PvcHandler struct {
	Name      string
	Oamclient *versioned.Clientset
	K8sclient *kubernetes.Clientset
}

type JobHandler struct {
	Name      string
	Oamclient *versioned.Clientset
	K8sclient *kubernetes.Clientset
}

type MysqlClusterHandler struct {
	Name      string
	Oamclient *versioned.Clientset
	K8sclient *kubernetes.Clientset
}

type IngressHandler struct {
	Name      string
	Oamclient *versioned.Clientset
	K8sclient *kubernetes.Clientset
}

type HpaHandler struct {
	Name      string
	Oamclient *versioned.Clientset
	K8sclient *kubernetes.Clientset
}

type HcHpaHandler struct {
	Name      string
	Oamclient *versioned.Clientset
	K8sclient *kubernetes.Clientset
}

func (s *ApplicationConfigurationHandler) Id() string {
	return "application-configuration-handler"
}

func (s *DeploymentHandler) Id() string {
	return "deployment-handler"
}

func (s *ServiceHandler) Id() string {
	return "service-handler"
}

func (s *ConfigMapHandler) Id() string {
	return "configmap-handler"
}

func (s *PvcHandler) Id() string {
	return "pvc-handler"
}

func (s *JobHandler) Id() string {
	return "job-handler"
}

func (s *MysqlClusterHandler) Id() string {
	return "mysqlcluster-handler"
}

func (s *IngressHandler) Id() string {
	return "ingress-handler"
}

func (s *HpaHandler) Id() string {
	return "hpa-handler"
}

func (s *HcHpaHandler) Id() string {
	return "hchpa-handler"
}
