package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	"github.com/oam-dev/oam-go-sdk/pkg/oam"
	"github.com/oam-dev/oam-go-sdk/pkg/util"
	hcv1alpha1 "hc-oam-controller/api/harmonycloud.cn/v1alpha1"
	hcv1beta1 "hc-oam-controller/api/harmonycloud.cn/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta2"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/types"

	//"k8s.io/api/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme     = runtime.NewScheme()
	handlerLog = ctrl.Log.WithName("application-configuration-handler")
)

func (s *ApplicationConfigurationHandler) Handle(ctx *oam.ActionContext, obj runtime.Object, eType oam.EType) error {
	ac, ok := obj.(*v1alpha1.ApplicationConfiguration)
	if !ok {
		return errors.New("type mismatch")
	}
	handlerLog.Info("Received ApplicationConfiguration.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name)

	owner := *v1.NewControllerRef(ac, v1alpha1.SchemeGroupVersion.WithKind("ApplicationConfiguration"))
	for _, compConf := range ac.Spec.Components {
		labels := map[string]string{
			"application": ac.Name,
			"component":   compConf.ComponentName,
			"instance":    compConf.InstanceName,
		}
		comp, err := s.Oamclient.CoreV1alpha1().ComponentSchematics(ac.Namespace).Get(compConf.ComponentName, v1.GetOptions{})
		if err != nil {
			return err
		}

		parameterMap := parseParameters(compConf.ParameterValues, ac.Spec.Variables)

		//create or update configmaps before create workloads
		configMaps := convertConfigMaps(owner, labels, compConf, *comp, parameterMap)
		if err := createOrUpdateConfigMaps(s, ac, compConf.ComponentName, configMaps); err != nil {
			handlerLog.Info("Create or update configMaps error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
		}

		//create pvcs before create workloads
		pvcs := convertPvcsFromVolumeMounters(owner, labels, *comp, compConf.Traits)
		if err := createOrUpdatePvcs(s, ac, compConf.ComponentName, pvcs); err != nil {
			handlerLog.Info("Create pvcs error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
		}

		switch comp.Spec.WorkloadType {
		case WorkloadTypeServer, WorkloadTypeSingletonServer, WorkloadTypeWorker, WorkloadTypeSingletonWorker:
			deployment := convertDeployment(owner, labels, compConf, *comp, parameterMap)
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

			if err := createOrUpdateDeployment(s, ac, compConf.ComponentName, deployment); err != nil {
				handlerLog.Info("Create or update deployment error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
			}

			if comp.Spec.WorkloadType == WorkloadTypeServer || comp.Spec.WorkloadType == WorkloadTypeSingletonServer {
				service := convertService(owner, labels, compConf, *comp)
				if err := createOrUpdateService(s, ac, compConf.ComponentName, service); err != nil {
					handlerLog.Info("Create or update service error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
				}

				//ingress trait
				ingress := convertIngress(owner, labels, compConf.InstanceName, compConf.Traits)
				if err := createOrUpdateIngress(s, ac, compConf.ComponentName, ingress); err != nil {
					handlerLog.Info("Create or update ingress error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
				}

			}

			//auto-scaler trait
			if comp.Spec.WorkloadType == WorkloadTypeServer || comp.Spec.WorkloadType == WorkloadTypeWorker {
				//apiVersion = "extensions/v1beta1"
				apiVersion = "apps/v1"
				hpa := convertHpa(owner, labels, "Deployment", apiVersion, compConf.InstanceName, compConf.Traits)
				if err := createOrUpdateHpa(s, ac, compConf.ComponentName, hpa); err != nil {
					handlerLog.Info("Create or update hpa error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
				}

			}

			//better-auto-scaler trait
			if comp.Spec.WorkloadType == WorkloadTypeServer || comp.Spec.WorkloadType == WorkloadTypeWorker {
				apiVersion = "apps/v1"
				hcHpa := convertHcHpa(owner, labels, "Deployment", apiVersion, compConf.InstanceName, compConf.Traits)
				if err := createOrUpdateHcHpa(s, ac, compConf.ComponentName, hcHpa); err != nil {
					handlerLog.Info("Create or update hcHpa error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
				}
			}

		case WorkloadTypeTask, WorkloadTypeSingletonTask:
			job := convertJob(owner, labels, compConf, *comp, parameterMap)
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

			if err := createOrUpdateJob(s, ac, compConf.ComponentName, job); err != nil {
				handlerLog.Info("Create or update job error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
			}

			//auto-scaler trait
			if comp.Spec.WorkloadType == WorkloadTypeTask {
				apiVersion = "batch/v1"
				hpa := convertHpa(owner, labels, "Job", apiVersion, compConf.InstanceName, compConf.Traits)
				if err := createOrUpdateHpa(s, ac, compConf.ComponentName, hpa); err != nil {
					handlerLog.Info("Create or update hpa error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
				}
			}

			//better-auto-scaler trait
			if comp.Spec.WorkloadType == WorkloadTypeTask {
				apiVersion = "batch/v1"
				hcHpa := convertHcHpa(owner, labels, "Job", apiVersion, compConf.InstanceName, compConf.Traits)
				if err := createOrUpdateHcHpa(s, ac, compConf.ComponentName, hcHpa); err != nil {
					handlerLog.Info("Create or update hcHpa error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
				}
			}

		case WorkloadTypeMysqlCluster:
			mysqlCluster, mysqlCm, mysqlPvc, err := convertMysqlCluster(owner, compConf, *comp, parameterMap)
			if err != nil {
				handlerLog.Info("Convert configuration for MysqlCluster failed", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
			}

			if err := createOrUpdateConfigMap(s, ac, compConf.ComponentName, *mysqlCm); err != nil {
				handlerLog.Info("Create or update configMap for MysqlCluster failed", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
			}

			//volume-mounter trait
			if err := createOrUpdatePvc(s, ac, compConf.ComponentName, *mysqlPvc); err != nil {
				handlerLog.Info("Create or update pvc for MysqlCluster failed", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
			}

			//manuel-scaler trait
			mysqlReplicas := *getManuelScale(compConf.Traits)
			mysqlCluster.Spec.Replicas = &mysqlReplicas

			if err := createOrUpdateMysqlCluster(s, ac, compConf.ComponentName, mysqlCluster); err != nil {
				handlerLog.Info("Create or update MysqlCluster error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
			}

		default:
			//You could launch you own CRD here according to workloadType
			return errors.New("WorkloadType " + comp.Spec.WorkloadType + " is undefined")
		}
	}

	// update status
	s.Oamclient.CoreV1alpha1().ApplicationConfigurations(ac.Namespace).UpdateStatus(ac)

	return nil
}

func createOrUpdateConfigMaps(s *ApplicationConfigurationHandler, applicationConfiguration *v1alpha1.ApplicationConfiguration, component string, configMaps []apiv1.ConfigMap) error {
	for _, configmap := range configMaps {
		if err := createOrUpdateConfigMap(s, applicationConfiguration, component, configmap); err != nil {
			return err
		}
	}
	return nil
}

func createOrUpdateConfigMap(s *ApplicationConfigurationHandler, applicationConfiguration *v1alpha1.ApplicationConfiguration, component string, configMap apiv1.ConfigMap) error {
	configMapClient := s.K8sclient.CoreV1().ConfigMaps(applicationConfiguration.Namespace)
	tmpCm, _ := configMapClient.Get(configMap.Name, v1.GetOptions{})
	applicationConfiguration.GetObjectMeta()
	if v1.IsControlledBy(tmpCm, applicationConfiguration.GetObjectMeta()) {
		patchDate, _ := json.Marshal(configMap)
		cmResult, err := configMapClient.Patch(configMap.Name, types.MergePatchType, patchDate)
		if util.SpecEqual(tmpCm, configMap.Data, false) {
			return nil
		}
		if err != nil {
			handlerLog.Info("ConfigMap patch failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "ConfigMap", configMap.Name, "Error", err)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return err
		} else {
			handlerLog.Info("ConfigMap patched.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "ConfigMap", cmResult.Name)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Patched, fmt.Sprintf(MessageResourcePatched, apiv1.ResourceConfigMaps, cmResult.Name))
		}
	} else {
		cmResult, err := configMapClient.Create(&configMap)
		if err != nil {
			handlerLog.Info("ConfigMap create failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "ConfigMap", configMap.Name, "Error", err)
			addModuleStatus(&applicationConfiguration.Status.Modules, configMap.Name, "ConfigMap", "v1", Failed)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return err
		} else {
			handlerLog.Info("ConfigMap created.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "ConfigMap", cmResult.Name)
			addModuleStatus(&applicationConfiguration.Status.Modules, configMap.Name, "ConfigMap", "v1", Created)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Created, fmt.Sprintf(MessageResourceCreated, apiv1.ResourceConfigMaps, cmResult.Name))
		}
	}
	return nil
}

func createOrUpdatePvcs(s *ApplicationConfigurationHandler, applicationConfiguration *v1alpha1.ApplicationConfiguration, component string, pvcs []apiv1.PersistentVolumeClaim) error {
	for _, pvc := range pvcs {
		if err := createOrUpdatePvc(s, applicationConfiguration, component, pvc); err != nil {
			return err
		}
	}
	return nil
}

func createOrUpdatePvc(s *ApplicationConfigurationHandler, applicationConfiguration *v1alpha1.ApplicationConfiguration, component string, pvc apiv1.PersistentVolumeClaim) error {
	pvcsClient := s.K8sclient.CoreV1().PersistentVolumeClaims(applicationConfiguration.Namespace)
	pvcResult, err := pvcsClient.Create(&pvc)
	if err != nil {
		handlerLog.Info("Pvc create failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Pvc", pvc.Name, "Error", err)
		addModuleStatus(&applicationConfiguration.Status.Modules, pvc.Name, "PersistentVolumeClaim", "v1", Failed)
		s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
		return err
	} else {
		handlerLog.Info("Pvc created.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Pvc", pvcResult.Name)
		addModuleStatus(&applicationConfiguration.Status.Modules, pvc.Name, "PersistentVolumeClaim", "v1", Created)
		s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Created, fmt.Sprintf(MessageResourceCreated, apiv1.ResourcePersistentVolumeClaims, pvcResult.Name))
	}
	return nil
}

func createOrUpdateDeployment(s *ApplicationConfigurationHandler, applicationConfiguration *v1alpha1.ApplicationConfiguration, component string, deployment *appsv1.Deployment) error {
	if deployment == nil {
		return nil
	}
	deploymentsClient := s.K8sclient.AppsV1().Deployments(applicationConfiguration.Namespace)
	tmpDeploy, _ := deploymentsClient.Get(deployment.Name, v1.GetOptions{})
	if v1.IsControlledBy(tmpDeploy, applicationConfiguration.GetObjectMeta()) {
		for index, c := range applicationConfiguration.Spec.Components {
			if c.ComponentName == component {
				for _, t := range applicationConfiguration.Spec.Components[index].Traits {
					if t.Name == "better-auto-scaler" || t.Name == "auto-scaler" {
						deployment.Spec.Replicas = tmpDeploy.Spec.Replicas
						break
					}
				}
				break
			}
		}
		patchData, _ := json.Marshal(deployment)
		deployResult, err := deploymentsClient.Patch(deployment.Name, types.MergePatchType, patchData)
		if err != nil {
			handlerLog.Info("Deployment patch failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Deployment", deployment.Name, "Error", err)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return nil
		} else {
			handlerLog.Info("Deployment patched.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Deployment", deployResult.Name)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Patched, fmt.Sprintf(MessageResourcePatched, "deployments", deployResult.Name))
		}
	} else {
		deployResult, err := deploymentsClient.Create(deployment)
		if err != nil {
			handlerLog.Info("Deployment create failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Deployment", deployment.Name, "Error", err)
			addModuleStatus(&applicationConfiguration.Status.Modules, deployment.Name, "Deployment", "apps/v1", Failed)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return err
		} else {
			handlerLog.Info("Deployment created.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Deployment", deployResult.Name)
			addModuleStatus(&applicationConfiguration.Status.Modules, deployment.Name, "Deployment", "apps/v1", Created)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Created, fmt.Sprintf(MessageResourceCreated, "deployments", deployResult.Name))
		}
	}
	return nil
}

func createOrUpdateJob(s *ApplicationConfigurationHandler, applicationConfiguration *v1alpha1.ApplicationConfiguration, component string, job *batchv1.Job) error {
	if job == nil {
		return nil
	}
	jobsClient := s.K8sclient.BatchV1().Jobs(applicationConfiguration.Namespace)

	tmpJob, _ := jobsClient.Get(job.Name, v1.GetOptions{})
	if v1.IsControlledBy(tmpJob, applicationConfiguration.GetObjectMeta()) {
		for index, c := range applicationConfiguration.Spec.Components {
			if c.ComponentName == component {
				for _, t := range applicationConfiguration.Spec.Components[index].Traits {
					if t.Name == "better-auto-scaler" || t.Name == "auto-scaler" {
						job.Spec.Parallelism = tmpJob.Spec.Parallelism
						break
					}
				}
				break
			}
		}
		patchData, _ := json.Marshal(job)
		jobResult, err := jobsClient.Patch(job.Name, types.MergePatchType, patchData)
		if err != nil {
			handlerLog.Info("Job patch failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Job", job.Name, "Error", err)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return err
		} else {
			handlerLog.Info("Job patched.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Job", jobResult.Name)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Patched, fmt.Sprintf(MessageResourcePatched, "jobs", jobResult.Name))
		}
	} else {
		jobResult, err := jobsClient.Create(job)
		if err != nil {
			handlerLog.Info("Job create failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Job", job.Name, "Error", err)
			addModuleStatus(&applicationConfiguration.Status.Modules, job.Name, "Deployment", "apps/v1", Failed)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return err
		} else {
			handlerLog.Info("Job created.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Job", jobResult.Name)
			addModuleStatus(&applicationConfiguration.Status.Modules, job.Name, "Deployment", "apps/v1", Created)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Created, fmt.Sprintf(MessageResourceCreated, "jobs", jobResult.Name))
		}
	}
	return nil
}

func createOrUpdateMysqlCluster(s *ApplicationConfigurationHandler, applicationConfiguration *v1alpha1.ApplicationConfiguration, component string, mysqlCluster *hcv1alpha1.MysqlCluster) error {
	if mysqlCluster == nil {
		return nil
	}
	mysqlClustersClient := s.Hcclient.HarmonycloudV1alpha1().MysqlClusters(applicationConfiguration.Namespace)
	tmpMysqlCluster, _ := mysqlClustersClient.Get(nil, mysqlCluster.Name, v1.GetOptions{})
	if v1.IsControlledBy(tmpMysqlCluster, applicationConfiguration.GetObjectMeta()) {
		patchData, _ := json.Marshal(mysqlCluster)
		mysqlClusterResult, err := mysqlClustersClient.Patch(nil, mysqlCluster.Name, types.MergePatchType, patchData, v1.PatchOptions{})
		if err != nil {
			handlerLog.Info("MysqlCluster patch failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "MysqlCluster", mysqlCluster.Name, "Error", err)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return err
		} else {
			handlerLog.Info("MysqlCluster patched.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Deployment", mysqlClusterResult.Name)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Patched, fmt.Sprintf(MessageResourcePatched, "mysqlclusters", mysqlClusterResult.Name))
		}
	} else {
		mysqlClusterResult, err := mysqlClustersClient.Create(nil, mysqlCluster, v1.CreateOptions{})
		if err != nil {
			handlerLog.Info("MysqlCluster create failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "MysqlCluster", mysqlCluster.Name, "Error", err)
			addModuleStatus(&applicationConfiguration.Status.Modules, mysqlCluster.Name, "MysqlCluster", "mysql.middleware.harmonycloud.cn/v1alpha1", Failed)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return err
		} else {
			handlerLog.Info("MysqlCluster created.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "MysqlCluster", mysqlClusterResult.Name)
			addModuleStatus(&applicationConfiguration.Status.Modules, mysqlCluster.Name, "MysqlCluster", "mysql.middleware.harmonycloud.cn/v1alpha1ap	fffaaa", Created)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Created, fmt.Sprintf(MessageResourceCreated, "mysqlclusters", mysqlClusterResult.Name))
		}
	}
	return nil
}

func createOrUpdateService(s *ApplicationConfigurationHandler, applicationConfiguration *v1alpha1.ApplicationConfiguration, component string, service *apiv1.Service) error {
	if service == nil {
		return nil
	}
	serviceClient := s.K8sclient.CoreV1().Services(applicationConfiguration.Namespace)
	tmpsvc, _ := serviceClient.Get(service.Name, v1.GetOptions{})
	if v1.IsControlledBy(tmpsvc, applicationConfiguration.GetObjectMeta()) {
		patchData, _ := json.Marshal(service)
		svcResult, err := serviceClient.Patch(service.Name, types.MergePatchType, patchData)
		if err != nil {
			handlerLog.Info("Service patch failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Service", service.Name, "Error", err)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return err
		} else {
			handlerLog.Info("Service patched.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Service", svcResult.Name)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Patched, fmt.Sprintf(MessageResourcePatched, apiv1.ResourceServices, svcResult.Name))
		}
	} else {
		svcResult, err := serviceClient.Create(service)
		if err != nil {
			handlerLog.Info("Service create failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Service", service.Name, "Error", err)
			addModuleStatus(&applicationConfiguration.Status.Modules, service.Name, "Service", "v1", Failed)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return err
		} else {
			handlerLog.Info("Service created.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Service", svcResult.Name)
			addModuleStatus(&applicationConfiguration.Status.Modules, service.Name, "Service", "v1", Created)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Created, fmt.Sprintf(MessageResourceCreated, apiv1.ResourceServices, svcResult.Name))
		}
	}
	return nil
}

func createOrUpdateIngress(s *ApplicationConfigurationHandler, applicationConfiguration *v1alpha1.ApplicationConfiguration, component string, ingress *extensionsv1beta1.Ingress) error {
	if ingress == nil {
		return nil
	}
	//ingressClient := s.K8sclient.NetworkingV1beta1().Ingresses(namespace)
	ingressClient := s.K8sclient.ExtensionsV1beta1().Ingresses(applicationConfiguration.Namespace)
	tmpIng, _ := ingressClient.Get(ingress.Name, v1.GetOptions{})
	if v1.IsControlledBy(tmpIng, applicationConfiguration.GetObjectMeta()) {
		patchData, _ := json.Marshal(ingress)
		ingResult, err := ingressClient.Patch(ingress.Name, types.MergePatchType, patchData)
		if err != nil {
			handlerLog.Info("Ingress patch failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Ingress", ingress.Name, "Error", err)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return err
		} else {
			handlerLog.Info("Ingress patched.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Ingress", ingResult.Name)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Patched, fmt.Sprintf(MessageResourcePatched, "ingresses", ingResult.Name))
		}
	} else {
		ingResult, err := ingressClient.Create(ingress)
		if err != nil {
			handlerLog.Info("Ingress create failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Ingress", ingress.Name, "Error", err)
			addModuleStatus(&applicationConfiguration.Status.Modules, ingress.Name, "Ingress", "extensions/v1beta1", Failed)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return err
		} else {
			handlerLog.Info("Ingress created.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Ingress", ingResult.Name)
			addModuleStatus(&applicationConfiguration.Status.Modules, ingress.Name, "Ingress", "extensions/v1beta1", Created)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Created, fmt.Sprintf(MessageResourceCreated, "ingresses", ingResult.Name))
		}
	}
	return nil

}

func createOrUpdateHpa(s *ApplicationConfigurationHandler, applicationConfiguration *v1alpha1.ApplicationConfiguration, component string, hpa *v2beta2.HorizontalPodAutoscaler) error {
	if hpa == nil {
		return nil
	}
	hpaClient := s.K8sclient.AutoscalingV2beta2().HorizontalPodAutoscalers(applicationConfiguration.Namespace)
	tmpHpa, _ := hpaClient.Get(hpa.Name, v1.GetOptions{})
	if v1.IsControlledBy(tmpHpa, applicationConfiguration.GetObjectMeta()) {
		patchData, _ := json.Marshal(hpa)
		hpaResult, err := hpaClient.Patch(hpa.Name, types.MergePatchType, patchData)
		if err != nil {
			handlerLog.Info("Hpa patch failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Hpa", hpa.Name, "Error", err)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return err
		} else {
			handlerLog.Info("Hpa patched.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Hpa", hpaResult.Name)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Patched, fmt.Sprintf(MessageResourcePatched, "hpas", hpaResult.Name))
		}
	} else {
		hpaResult, err := hpaClient.Create(hpa)
		if err != nil {
			handlerLog.Info("Hpa create failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Hpa", hpa.Name, "Error", err)
			addModuleStatus(&applicationConfiguration.Status.Modules, hpa.Name, "HorizontalPodAutoscaler", "autoscaling/v2beta2", Failed)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return err
		} else {
			handlerLog.Info("Hpa created.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "Hpa", hpaResult.Name)
			addModuleStatus(&applicationConfiguration.Status.Modules, hpa.Name, "HorizontalPodAutoscaler", "autoscaling/v2beta2", Created)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Created, fmt.Sprintf(MessageResourceCreated, "hpas", hpaResult.Name))
		}
	}
	return nil
}

func createOrUpdateHcHpa(s *ApplicationConfigurationHandler, applicationConfiguration *v1alpha1.ApplicationConfiguration, component string, hcHpa *hcv1beta1.HorizontalPodAutoscaler) error {
	if hcHpa == nil {
		return nil
	}
	hcHpaClient := s.Hcclient.HarmonycloudV1beta1().HorizontalPodAutoscalers(applicationConfiguration.Namespace)
	tmpHcHpa, _ := hcHpaClient.Get(nil, hcHpa.Name, v1.GetOptions{})
	if v1.IsControlledBy(tmpHcHpa, applicationConfiguration.GetObjectMeta()) {
		patchData, _ := json.Marshal(hcHpa)
		hcHpaResult, err := hcHpaClient.Patch(nil, hcHpa.Name, types.MergePatchType, patchData, v1.PatchOptions{})
		if err != nil {
			handlerLog.Info("HcHpa patch failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "HcHpa", hcHpa.Name, "Error", err)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return err
		} else {
			handlerLog.Info("HcHpa patched.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "HcHpa", hcHpaResult.Name)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Patched, fmt.Sprintf(MessageResourcePatched, "hchpas", hcHpaResult.Name))
		}
	} else {
		hcHpaResult, err := hcHpaClient.Create(nil, hcHpa, v1.CreateOptions{})
		if err != nil {
			handlerLog.Info("HcHpa create failed.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "HcHpa", hcHpa.Name, "Error", err)
			addModuleStatus(&applicationConfiguration.Status.Modules, hcHpa.Name, "HorizontalPodAutoscaler", "harmonycloud.cn/v1beta1", Failed)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeWarning, Failed, err.Error())
			return err
		} else {
			handlerLog.Info("HcHpa created.", "Namespace", applicationConfiguration.Namespace, "ApplicationConfiguration", applicationConfiguration.Name, "Component", component, "HcHpa", hcHpaResult.Name)
			addModuleStatus(&applicationConfiguration.Status.Modules, hcHpa.Name, "HorizontalPodAutoscaler", "harmonycloud.cn/v1beta1", Created)
			s.Recorder.Event(applicationConfiguration, apiv1.EventTypeNormal, Created, fmt.Sprintf(MessageResourceCreated, "hchpas", hcHpaResult.Name))
		}
	}
	return nil
}
