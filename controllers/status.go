package controllers

import (
	"errors"
	"fmt"
	"github.com/oam-dev/oam-go-sdk/pkg/oam"
	"hc-oam-controller/api/harmonycloud.cn/v1alpha1"
	hcv1beta1 "hc-oam-controller/api/harmonycloud.cn/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	statusLog = ctrl.Log.WithName("status-handler")
)

func (s *DeploymentHandler) Handle(ctx *oam.ActionContext, obj runtime.Object, eType oam.EType) error {
	deployment, ok := obj.(*appsv1.Deployment)
	if !ok {
		return errors.New("type mismatch")
	}
	if deployment.OwnerReferences == nil {
		return nil
	}
	for _, o := range deployment.OwnerReferences {
		if o.Kind == "ApplicationConfiguration" {
			ac, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(deployment.Namespace).Get(o.Name, v1.GetOptions{})
			if err != nil {
				return err
			}
			for k, m := range ac.Status.Modules {
				if m.Kind == "Deployment" && m.NamespacedName == deployment.Name {
					ac.Status.Modules[k].Status = fmt.Sprintf("Ready: %v/%v, Up-to-date: %v, Available: %v",
						deployment.Status.ReadyReplicas, deployment.Status.ReadyReplicas, deployment.Status.UpdatedReplicas, deployment.Status.AvailableReplicas)
					if _, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(deployment.Namespace).UpdateStatus(ac); err != nil {
						statusLog.Info("Update status failed", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name)
						return err
					}
					return nil
				}
			}
		}
	}

	return nil
}

func (s *ServiceHandler) Handle(ctx *oam.ActionContext, obj runtime.Object, eType oam.EType) error {
	service, ok := obj.(*corev1.Service)
	if !ok {
		return errors.New("type mismatch")
	}
	if service.OwnerReferences == nil {
		return nil
	}
	for _, o := range service.OwnerReferences {
		if o.Kind == "ApplicationConfiguration" {
			ac, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(service.Namespace).Get(o.Name, v1.GetOptions{})
			if err != nil {
				return err
			}
			var ports string
			for j, p := range service.Spec.Ports {
				if service.Spec.Type == "NodePort" {
					ports += fmt.Sprintf("%v:%v/%s", p.Port, p.NodePort, p.Protocol)
				} else {
					ports += fmt.Sprintf("%v/%s", p.Port, p.Protocol)
				}
				if j < len(service.Spec.Ports)-1 {
					ports += ","
				}
			}
			for k, m := range ac.Status.Modules {
				if m.Kind == "Service" && m.NamespacedName == service.Name {
					ac.Status.Modules[k].Status = fmt.Sprintf("Type: %s, Cluster-IP: %s, Port(s): %s",
						service.Spec.Type, service.Spec.ClusterIP, ports)
					if _, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(service.Namespace).UpdateStatus(ac); err != nil {
						statusLog.Info("Update status failed", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name)
						return err
					}
					return nil
				}
			}
		}
	}

	return nil
}

func (s *ConfigMapHandler) Handle(ctx *oam.ActionContext, obj runtime.Object, eType oam.EType) error {
	configmap, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return errors.New("type mismatch")
	}
	if configmap.OwnerReferences == nil {
		return nil
	}
	for _, o := range configmap.OwnerReferences {
		if o.Kind == "ApplicationConfiguration" {
			ac, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(configmap.Namespace).Get(o.Name, v1.GetOptions{})
			if err != nil {
				return err
			}
			for k, m := range ac.Status.Modules {
				if m.Kind == "ConfigMap" && m.NamespacedName == configmap.Name {
					ac.Status.Modules[k].Status = fmt.Sprintf("Data: %v", len(configmap.Data))
					if _, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(configmap.Namespace).UpdateStatus(ac); err != nil {
						statusLog.Info("Update status failed", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name)
						return err
					}
					return nil
				}
			}
		}
	}

	return nil
}

func (s *PvcHandler) Handle(ctx *oam.ActionContext, obj runtime.Object, eType oam.EType) error {
	pvc, ok := obj.(*corev1.PersistentVolumeClaim)
	if !ok {
		return errors.New("type mismatch")
	}

	if pvc.OwnerReferences == nil {
		return nil
	}
	for _, o := range pvc.OwnerReferences {
		if o.Kind == "ApplicationConfiguration" {
			ac, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(pvc.Namespace).Get(o.Name, v1.GetOptions{})
			if err != nil {
				return err
			}
			for k, m := range ac.Status.Modules {
				if m.Kind == "PersistentVolumeClaim" && m.NamespacedName == pvc.Name {
					var status string
					if pvc.Status.Phase != "Bound" {
						status = fmt.Sprintf("Status: %s", pvc.Status.Phase)
					} else {
						capacity := pvc.Status.Capacity["storage"]
						status = fmt.Sprintf("Status: %s, Volume: %s, Capacity: %s, AccessModes: %s",
							pvc.Status.Phase,
							pvc.Spec.VolumeName,
							capacity.String(),
							pvc.Status.AccessModes)
					}
					if pvc.Spec.StorageClassName != nil {
						status += fmt.Sprintf(", StorageClass: %s", *pvc.Spec.StorageClassName)
					}
					ac.Status.Modules[k].Status = status
					if _, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(pvc.Namespace).UpdateStatus(ac); err != nil {
						statusLog.Info("Update status failed", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name)
						return err
					}
					return nil
				}
			}
		}
	}

	return nil
}

func (s *JobHandler) Handle(ctx *oam.ActionContext, obj runtime.Object, eType oam.EType) error {
	job, ok := obj.(*batchv1.Job)
	if !ok {
		return errors.New("type mismatch")
	}
	if job.OwnerReferences == nil {
		return nil
	}
	for _, o := range job.OwnerReferences {
		if o.Kind == "ApplicationConfiguration" {
			ac, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(job.Namespace).Get(o.Name, v1.GetOptions{})
			if err != nil {
				return err
			}
			for k, m := range ac.Status.Modules {
				if m.Kind == "Job" && m.NamespacedName == job.Name {
					ac.Status.Modules[k].Status = fmt.Sprintf("Active: %v, Succeeded: %v, Failed: %v",
						job.Status.Active, job.Status.Succeeded, job.Status.Failed)
					if _, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(job.Namespace).UpdateStatus(ac); err != nil {
						statusLog.Info("Update status failed", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name)
						return err
					}
					return nil
				}
			}
		}
	}

	return nil
}

func (s *MysqlClusterHandler) Handle(ctx *oam.ActionContext, obj runtime.Object, eType oam.EType) error {
	mysqlCluster, ok := obj.(*v1alpha1.MysqlCluster)
	if !ok {
		return errors.New("type mismatch")
	}
	if mysqlCluster.OwnerReferences == nil {
		return nil
	}
	for _, o := range mysqlCluster.OwnerReferences {
		if o.Kind == "ApplicationConfiguration" {
			ac, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(mysqlCluster.Namespace).Get(o.Name, v1.GetOptions{})
			if err != nil {
				return err
			}
			for k, m := range ac.Status.Modules {
				if m.Kind == "Deployment" && m.NamespacedName == mysqlCluster.Name {
					ac.Status.Modules[k].Status = fmt.Sprintf("Phase: %s, Replicas: %v, CurrentRevision: %s, UpdateRevision: %s, CurrentSwitchedNum: %v, FailedCount: %v, Reason: %s",
						mysqlCluster.Status.Phase, mysqlCluster.Status.Replicas, mysqlCluster.Status.CurrentRevision, mysqlCluster.Status.UpdateRevision, mysqlCluster.Status.CurrentSwitchedNum, mysqlCluster.Status.FailedCount, mysqlCluster.Status.Reason)
					if _, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(mysqlCluster.Namespace).UpdateStatus(ac); err != nil {
						statusLog.Info("Update status failed", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name)
						return err
					}
					return nil
				}
			}
		}
	}

	return nil
}

func (s *IngressHandler) Handle(ctx *oam.ActionContext, obj runtime.Object, eType oam.EType) error {
	ingress, ok := obj.(*v1beta1.Ingress)
	if !ok {
		return errors.New("type mismatch")
	}
	if ingress.OwnerReferences == nil {
		return nil
	}
	for _, o := range ingress.OwnerReferences {
		if o.Kind == "ApplicationConfiguration" {
			ac, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(ingress.Namespace).Get(o.Name, v1.GetOptions{})
			if err != nil {
				return err
			}
			var hosts string
			for j, r := range ingress.Spec.Rules {
				hosts += r.Host
				if j < len(ingress.Spec.Rules)-1 {
					hosts += ","
				}

			}
			for k, m := range ac.Status.Modules {
				if m.Kind == "Ingress" && m.NamespacedName == ingress.Name {
					ac.Status.Modules[k].Status = fmt.Sprintf("Hosts: %s", hosts)
					if _, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(ingress.Namespace).UpdateStatus(ac); err != nil {
						statusLog.Info("Update status failed", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name)
						return err
					}
					return nil
				}
			}
		}
	}

	return nil
}

func (s *HpaHandler) Handle(ctx *oam.ActionContext, obj runtime.Object, eType oam.EType) error {
	hpa, ok := obj.(*v2beta2.HorizontalPodAutoscaler)
	if !ok {
		return errors.New("type mismatch")
	}
	if hpa.OwnerReferences == nil {
		return nil
	}
	for _, o := range hpa.OwnerReferences {
		if o.Kind == "ApplicationConfiguration" {
			ac, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(hpa.Namespace).Get(o.Name, v1.GetOptions{})
			if err != nil {
				return err
			}

			for k, m := range ac.Status.Modules {
				if m.Kind == "HorizontalPodAutoscaler" && m.GroupVersion == "autoscaling/v2beta2" && m.NamespacedName == hpa.Name {
					ac.Status.Modules[k].Status = fmt.Sprintf("CurrentReplicas: %v, DesiredReplicas: %v", hpa.Status.CurrentReplicas, hpa.Status.DesiredReplicas)
					if _, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(hpa.Namespace).UpdateStatus(ac); err != nil {
						statusLog.Info("Update status failed", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name)
						return err
					}
					return nil
				}
			}
		}
	}

	return nil
}

func (s *HcHpaHandler) Handle(ctx *oam.ActionContext, obj runtime.Object, eType oam.EType) error {
	hpa, ok := obj.(*hcv1beta1.HorizontalPodAutoscaler)
	if !ok {
		return errors.New("type mismatch")
	}
	if hpa.OwnerReferences == nil {
		return nil
	}
	for _, o := range hpa.OwnerReferences {
		if o.Kind == "ApplicationConfiguration" {
			ac, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(hpa.Namespace).Get(o.Name, v1.GetOptions{})
			if err != nil {
				return err
			}

			for k, m := range ac.Status.Modules {
				if m.Kind == "HorizontalPodAutoscaler" && m.GroupVersion == "harmonycloud.cn/v1beta1" && m.NamespacedName == hpa.Name {
					ac.Status.Modules[k].Status = fmt.Sprintf("CurrentReplicas: %v, DesiredReplicas: %v", hpa.Status.CurrentReplicas, hpa.Status.DesiredReplicas)
					if _, err := s.Oamclient.CoreV1alpha1().ApplicationConfigurations(hpa.Namespace).UpdateStatus(ac); err != nil {
						statusLog.Info("Update status failed", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name)
						return err
					}
					return nil
				}
			}
		}
	}

	return nil
}
