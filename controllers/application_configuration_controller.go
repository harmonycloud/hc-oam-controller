package controllers

import (
	"errors"
	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	"github.com/oam-dev/oam-go-sdk/pkg/oam"
	hcv1alpha1 "hc-oam-controller/api/harmonycloud.cn/v1alpha1"
	hcv1beta1 "hc-oam-controller/api/harmonycloud.cn/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta2"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	//"k8s.io/api/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme     = runtime.NewScheme()
	handlerLog = ctrl.Log.WithName("application-configuration-handler")
)

func (s *Handler) Id() string {
	return "Handler"
}

func (s *Handler) Handle(ctx *oam.ActionContext, obj runtime.Object, eType oam.EType) error {
	ac, ok := obj.(*v1alpha1.ApplicationConfiguration)
	if !ok {
		return errors.New("type mismatch")
	}
	handlerLog.Info("Received ApplicationConfiguration.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name)

	owner := v1.OwnerReference{
		APIVersion: ac.APIVersion,
		Kind:       ac.Kind,
		Name:       ac.Name,
		UID:        ac.UID,
	}

	for _, compConf := range ac.Spec.Components {
		comp, err := s.Oamclient.CoreV1alpha1().ComponentSchematics(ac.Namespace).Get(compConf.ComponentName, v1.GetOptions{})
		if err != nil {
			return err
		}

		parameterMap := parseParameters(compConf.ParameterValues, ac.Spec.Variables)

		//create or update configmaps before create workloads
		configMaps := convertConfigMaps(owner, compConf, *comp, parameterMap)
		if err := createOrUpdateConfigMaps(s, ac.Namespace, ac.Name, compConf.ComponentName, configMaps); err != nil {
			handlerLog.Info("Create or update configMaps error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
		}

		//create pvcs before create workloads
		pvcs := convertPvcsFromVolumeMounters(owner, *comp, compConf.Traits)
		if err := createOrUpdatePvcs(s, ac.Namespace, ac.Name, compConf.ComponentName, pvcs); err != nil {
			handlerLog.Info("Create pvcs error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
		}

		switch comp.Spec.WorkloadType {
		case WorkloadTypeServer, WorkloadTypeSingletonServer, WorkloadTypeWorker, WorkloadTypeSingletonWorker:
			deployment := convertDeployment(owner, compConf, *comp, parameterMap)
			var apiVersion string

			//manuel-scaler trait
			replicas := *getManuelScale(compConf.Traits)
			if comp.Spec.WorkloadType == WorkloadTypeSingletonServer || comp.Spec.WorkloadType == WorkloadTypeSingletonWorker {
				replicas = 1
			}
			deployment.Spec.Replicas = &replicas

			//volume-mounter trait
			volumes := getVolumesFromVolumeMounters(compConf.Traits)
			volumes = append(volumes, convertVolumesFromConfig(configMaps)...)
			deployment.Spec.Template.Spec.Volumes = volumes

			//log-pilot trait
			volumes = append(volumes, getLogPilotVolumes(compConf.Traits)...)
			deployment.Spec.Template.Spec.Volumes = volumes
			for i, _ := range deployment.Spec.Template.Spec.Containers {
				injectLogPilotConfigs(&deployment.Spec.Template.Spec.Containers[i], compConf.Traits)
			}

			if err := createOrUpdateDeployment(s, ac.Namespace, ac.Name, compConf.ComponentName, deployment); err != nil {
				handlerLog.Info("Create or update deployment error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
			}

			if comp.Spec.WorkloadType == WorkloadTypeServer || comp.Spec.WorkloadType == WorkloadTypeSingletonServer {
				service := convertService(owner, compConf, *comp)
				if err := createOrUpdateService(s, ac.Namespace, ac.Name, compConf.ComponentName, service); err != nil {
					handlerLog.Info("Create or update service error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
				}

				//ingress trait
				ingress := convertIngress(owner, compConf.InstanceName, compConf.Traits)
				if err := createOrUpdateIngress(s, ac.Namespace, ac.Name, compConf.ComponentName, ingress); err != nil {
					handlerLog.Info("Create or update ingress error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
				}

			}

			//auto-scaler trait
			if comp.Spec.WorkloadType == WorkloadTypeServer || comp.Spec.WorkloadType == WorkloadTypeWorker {
				//apiVersion = "extensions/v1beta1"
				apiVersion = "apps/v1"
				hpa := convertHpa(owner, "Deployment", apiVersion, compConf.InstanceName, compConf.Traits)
				if err := createOrUpdateHpa(s, ac.Namespace, ac.Name, compConf.ComponentName, hpa); err != nil {
					handlerLog.Info("Create or update hpa error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
				}

			}

			//better-auto-scaler trait
			if comp.Spec.WorkloadType == WorkloadTypeServer || comp.Spec.WorkloadType == WorkloadTypeWorker {
				apiVersion = "apps/v1"
				hcHpa := convertHcHpa(owner, "Deployment", apiVersion, compConf.InstanceName, compConf.Traits)
				if err := createOrUpdateHcHpa(s, ac.Namespace, ac.Name, compConf.ComponentName, hcHpa); err != nil {
					handlerLog.Info("Create or update hcHpa error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
				}
			}

		case WorkloadTypeTask, WorkloadTypeSingletonTask:
			job := convertJob(owner, compConf, *comp, parameterMap)
			var apiVersion string

			//manuel-scaler trait
			parallelism := *getManuelScale(compConf.Traits)
			if comp.Spec.WorkloadType == WorkloadTypeSingletonTask {
				parallelism = 1
			}
			job.Spec.Parallelism = &parallelism

			//volume-mounter trait
			volumes := getVolumesFromVolumeMounters(compConf.Traits)
			volumes = append(volumes, convertVolumesFromConfig(configMaps)...)
			job.Spec.Template.Spec.Volumes = volumes

			//log-pilot trait
			volumes = append(volumes, getLogPilotVolumes(compConf.Traits)...)
			job.Spec.Template.Spec.Volumes = volumes
			for i, _ := range job.Spec.Template.Spec.Containers {
				injectLogPilotConfigs(&job.Spec.Template.Spec.Containers[i], compConf.Traits)
			}

			if err := createOrUpdateJob(s, ac.Namespace, ac.Name, compConf.ComponentName, job); err != nil {
				handlerLog.Info("Create or update job error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
			}

			//auto-scaler trait
			if comp.Spec.WorkloadType == WorkloadTypeTask {
				apiVersion = "batch/v1"
				hpa := convertHpa(owner, "Job", apiVersion, compConf.InstanceName, compConf.Traits)
				if err := createOrUpdateHpa(s, ac.Namespace, ac.Name, compConf.ComponentName, hpa); err != nil {
					handlerLog.Info("Create or update hpa error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
				}
			}

			//better-auto-scaler trait
			if comp.Spec.WorkloadType == WorkloadTypeTask {
				apiVersion = "batch/v1"
				hcHpa := convertHcHpa(owner, "Job", apiVersion, compConf.InstanceName, compConf.Traits)
				if err := createOrUpdateHcHpa(s, ac.Namespace, ac.Name, compConf.ComponentName, hcHpa); err != nil {
					handlerLog.Info("Create or update hcHpa error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
				}
			}

		case WorkloadTypeMysqlCluster:
			mysqlCluster, mysqlCm, mysqlPvc, err := convertMysqlCluster(owner, compConf, *comp, parameterMap)
			if err != nil {
				handlerLog.Info("Convert configuration for MysqlCluster failed", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
			}

			if err := createOrUpdateConfigMap(s, ac.Namespace, ac.Name, compConf.ComponentName, *mysqlCm); err != nil {
				handlerLog.Info("Create or update configMap for MysqlCluster failed", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
			}

			//volume-mounter trait
			if err := createOrUpdatePvc(s, ac.Namespace, ac.Name, compConf.ComponentName, *mysqlPvc); err != nil {
				handlerLog.Info("Create or update pvc for MysqlCluster failed", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
			}

			//manuel-scaler trait
			mysqlReplicas := *getManuelScale(compConf.Traits)
			mysqlCluster.Spec.Replicas = &mysqlReplicas

			if err := createOrUpdateMysqlCluster(s, ac.Namespace, ac.Name, compConf.ComponentName, mysqlCluster); err != nil {
				handlerLog.Info("Create or update MysqlCluster error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
			}

		default:
			//You could launch you own CRD here according to workloadType
			return errors.New("WorkloadType " + comp.Spec.WorkloadType + " is undefined")
		}
	}
	return nil
}

func isOwnerEqual(ownerA v1.OwnerReference, ownerB v1.OwnerReference) bool {
	if ownerA.UID == ownerB.UID && ownerA.Name == ownerB.Name && ownerA.Kind == ownerB.Kind && ownerA.APIVersion == ownerB.APIVersion {
		return true
	}
	return false
}

func createOrUpdateConfigMaps(s *Handler, namespace string, applicationConfiguration string, component string, configMaps []apiv1.ConfigMap) error {
	for _, configmap := range configMaps {
		if err := createOrUpdateConfigMap(s, namespace, applicationConfiguration, component, configmap); err != nil {
			return err
		}
	}
	return nil
}

func createOrUpdateConfigMap(s *Handler, namespace string, applicationConfiguration string, component string, configMap apiv1.ConfigMap) error {
	configMapClient := s.K8sclient.CoreV1().ConfigMaps(namespace)
	tmpCm, _ := configMapClient.Get(configMap.Name, v1.GetOptions{})
	if tmpCm.OwnerReferences != nil && isOwnerEqual(configMap.OwnerReferences[0], tmpCm.OwnerReferences[0]) {
		cmResult, err := configMapClient.Update(&configMap)
		if err != nil {
			handlerLog.Info("ConfigMap update failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "ConfigMap", configMap.Name, "Error", err)
			return err
		} else {
			handlerLog.Info("ConfigMap updated.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "ConfigMap", cmResult.Name)
		}
	} else {
		cmResult, err := configMapClient.Create(&configMap)
		if err != nil {
			handlerLog.Info("ConfigMap create failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "ConfigMap", configMap.Name, "Error", err)
			return err
		} else {
			handlerLog.Info("ConfigMap created.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "ConfigMap", cmResult.Name)
		}
	}
	return nil
}

func createOrUpdatePvcs(s *Handler, namespace string, applicationConfiguration string, component string, pvcs []apiv1.PersistentVolumeClaim) error {
	for _, pvc := range pvcs {
		if err := createOrUpdatePvc(s, namespace, applicationConfiguration, component, pvc); err != nil {
			return err
		}
	}
	return nil
}

func createOrUpdatePvc(s *Handler, namespace string, applicationConfiguration string, component string, pvc apiv1.PersistentVolumeClaim) error {
	pvcsClient := s.K8sclient.CoreV1().PersistentVolumeClaims(namespace)
	pvcResult, err := pvcsClient.Create(&pvc)
	if err != nil {
		handlerLog.Info("Pvc create failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Pvc", pvc.Name, "Error", err)
		return err
	} else {
		handlerLog.Info("Pvc created.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Pvc", pvcResult.Name)
	}
	return nil
}

func createOrUpdateDeployment(s *Handler, namespace string, applicationConfiguration string, component string, deployment *appsv1.Deployment) error {
	if deployment == nil {
		return nil
	}
	deploymentsClient := s.K8sclient.AppsV1().Deployments(namespace)
	tmpDeploy, _ := deploymentsClient.Get(deployment.Name, v1.GetOptions{})
	if tmpDeploy.OwnerReferences != nil && isOwnerEqual(deployment.OwnerReferences[0], tmpDeploy.OwnerReferences[0]) {
		deployResult, err := deploymentsClient.Update(deployment)
		if err != nil {
			handlerLog.Info("Deployment update failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Deployment", deployment.Name, "Error", err)
			return nil
		} else {
			handlerLog.Info("Deployment updated.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Deployment", deployResult.Name)
		}
	} else {
		deployResult, err := deploymentsClient.Create(deployment)
		if err != nil {
			handlerLog.Info("Deployment create failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Deployment", deployment.Name, "Error", err)
			return nil
		} else {
			handlerLog.Info("Deployment created.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Deployment", deployResult.Name)
		}
	}
	return nil
}

func createOrUpdateJob(s *Handler, namespace string, applicationConfiguration string, component string, job *batchv1.Job) error {
	if job == nil {
		return nil
	}
	jobsClient := s.K8sclient.BatchV1().Jobs(namespace)

	tmpJob, _ := jobsClient.Get(job.Name, v1.GetOptions{})
	if tmpJob.OwnerReferences != nil && isOwnerEqual(job.OwnerReferences[0], tmpJob.OwnerReferences[0]) {
		jobResult, err := jobsClient.Update(job)
		if err != nil {
			handlerLog.Info("Job update failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Job", job.Name, "Error", err)
			return err
		} else {
			handlerLog.Info("Job updated.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Job", jobResult.Name)
		}
	} else {
		jobResult, err := jobsClient.Create(job)
		if err != nil {
			handlerLog.Info("Job create failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Job", job.Name, "Error", err)
			return err
		} else {
			handlerLog.Info("Job created.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Job", jobResult.Name)
		}
	}
	return nil
}

func createOrUpdateMysqlCluster(s *Handler, namespace string, applicationConfiguration string, component string, mysqlCluster *hcv1alpha1.MysqlCluster) error {
	if mysqlCluster == nil {
		return nil
	}
	mysqlClustersClient := s.Hcclient.HarmonycloudV1alpha1().MysqlClusters(namespace)
	tmpMysqlCluster, _ := mysqlClustersClient.Get(nil, mysqlCluster.Name, v1.GetOptions{})
	if tmpMysqlCluster.OwnerReferences != nil && isOwnerEqual(mysqlCluster.OwnerReferences[0], tmpMysqlCluster.OwnerReferences[0]) {
		mysqlClusterResult, err := mysqlClustersClient.Update(nil, mysqlCluster, v1.UpdateOptions{})
		if err != nil {
			handlerLog.Info("MysqlCluster update failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "MysqlCluster", mysqlCluster.Name, "Error", err)
			return nil
		} else {
			handlerLog.Info("MysqlCluster updated.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Deployment", mysqlClusterResult.Name)
		}
	} else {
		mysqlClusterResult, err := mysqlClustersClient.Create(nil, mysqlCluster, v1.CreateOptions{})
		if err != nil {
			handlerLog.Info("MysqlCluster create failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "MysqlCluster", mysqlCluster.Name, "Error", err)
			return nil
		} else {
			handlerLog.Info("MysqlCluster created.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "MysqlCluster", mysqlClusterResult.Name)
		}
	}
	return nil
}

func createOrUpdateService(s *Handler, namespace string, applicationConfiguration string, component string, service *apiv1.Service) error {
	if service == nil {
		return nil
	}
	serviceClient := s.K8sclient.CoreV1().Services(namespace)
	tmpsvc, _ := serviceClient.Get(service.Name, v1.GetOptions{})
	if tmpsvc.OwnerReferences != nil && isOwnerEqual(service.OwnerReferences[0], tmpsvc.OwnerReferences[0]) {
		service.ResourceVersion = tmpsvc.ResourceVersion
		service.Spec.ClusterIP = tmpsvc.Spec.ClusterIP
		svcResult, err := serviceClient.Update(service)
		if err != nil {
			handlerLog.Info("Service update failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Service", service.Name, "Error", err)
		} else {
			handlerLog.Info("Service updated.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Service", svcResult.Name)
		}
	} else {
		svcResult, err := serviceClient.Create(service)
		if err != nil {
			handlerLog.Info("Service create failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Service", service.Name, "Error", err)
		} else {
			handlerLog.Info("Service created.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Service", svcResult.Name)
		}
	}
	return nil
}

func createOrUpdateIngress(s *Handler, namespace string, applicationConfiguration string, component string, ingress *extensionsv1beta1.Ingress) error {
	if ingress == nil {
		return nil
	}
	//ingressClient := s.K8sclient.NetworkingV1beta1().Ingresses(namespace)
	ingressClient := s.K8sclient.ExtensionsV1beta1().Ingresses(namespace)
	tmpIng, _ := ingressClient.Get(ingress.Name, v1.GetOptions{})
	if tmpIng.OwnerReferences != nil && isOwnerEqual(ingress.OwnerReferences[0], tmpIng.OwnerReferences[0]) {
		ingResult, err := ingressClient.Update(ingress)
		if err != nil {
			handlerLog.Info("Ingress update failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Ingress", ingress.Name, "Error", err)
		} else {
			handlerLog.Info("Ingress updated.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Ingress", ingResult.Name)
		}
	} else {
		ingResult, err := ingressClient.Create(ingress)
		if err != nil {
			handlerLog.Info("Ingress create failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Ingress", ingress.Name, "Error", err)
		} else {
			handlerLog.Info("Ingress created.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Ingress", ingResult.Name)
		}
	}
	return nil

}

func createOrUpdateHpa(s *Handler, namespace string, applicationConfiguration string, component string, hpa *v2beta2.HorizontalPodAutoscaler) error {
	if hpa == nil {
		return nil
	}
	hpaClient := s.K8sclient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace)
	tmpHpa, _ := hpaClient.Get(hpa.Name, v1.GetOptions{})
	if tmpHpa.OwnerReferences != nil && isOwnerEqual(hpa.OwnerReferences[0], tmpHpa.OwnerReferences[0]) {
		hpaResult, err := hpaClient.Update(hpa)
		if err != nil {
			handlerLog.Info("Hpa update failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Hpa", hpa.Name, "Error", err)
			return err
		} else {
			handlerLog.Info("Hpa updated.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Hpa", hpaResult.Name)
		}
	} else {
		hpaResult, err := hpaClient.Create(hpa)
		if err != nil {
			handlerLog.Info("Hpa create failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Hpa", hpa.Name, "Error", err)
			return err
		} else {
			handlerLog.Info("Hpa created.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "Hpa", hpaResult.Name)
		}
	}
	return nil
}

func createOrUpdateHcHpa(s *Handler, namespace string, applicationConfiguration string, component string, hcHpa *hcv1beta1.HorizontalPodAutoscaler) error {
	if hcHpa == nil {
		return nil
	}
	hcHpaClient := s.Hcclient.HarmonycloudV1beta1().HorizontalPodAutoscalers(namespace)
	tmpHcHpa, _ := hcHpaClient.Get(nil, hcHpa.Name, v1.GetOptions{})
	if tmpHcHpa.OwnerReferences != nil && isOwnerEqual(hcHpa.OwnerReferences[0], tmpHcHpa.OwnerReferences[0]) {
		hcHpaResult, err := hcHpaClient.Update(nil, hcHpa, v1.UpdateOptions{})
		if err != nil {
			handlerLog.Info("HcHpa update failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "HcHpa", hcHpa.Name, "Error", err)
			return err
		} else {
			handlerLog.Info("HcHpa updated.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "HcHpa", hcHpaResult.Name)
		}
	} else {
		hcHpaResult, err := hcHpaClient.Create(nil, hcHpa, v1.CreateOptions{})
		if err != nil {
			handlerLog.Info("HcHpa create failed.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "HcHpa", hcHpa.Name, "Error", err)
			return err
		} else {
			handlerLog.Info("HcHpa created.", "Namespace", namespace, "ApplicationConfiguration", applicationConfiguration, "Component", component, "HcHpa", hcHpaResult.Name)
		}
	}
	return nil
}
