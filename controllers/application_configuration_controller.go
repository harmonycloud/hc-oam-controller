package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"reflect"

	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	"github.com/oam-dev/oam-go-sdk/pkg/oam"
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
		switch comp.Spec.WorkloadType {
		case WorkloadTypeServer, WorkloadTypeSingletonServer, WorkloadTypeWorker, WorkloadTypeSingletonWorker:
			deploymentsClient := s.K8sclient.AppsV1().Deployments(ac.Namespace)
			deployment, configMaps := convertDeployment(owner, compConf, *comp)

			replicas := *getManuelScale(compConf.Traits)
			if comp.Spec.WorkloadType == WorkloadTypeSingletonServer || comp.Spec.WorkloadType == WorkloadTypeSingletonWorker {
				replicas = 1
			}
			deployment.Spec.Replicas = &replicas

			if err := createOrUpdateConfigMaps(s, ac.Namespace, configMaps); err != nil {
				handlerLog.Info("Create or update configMaps error.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Error", err)
				log.Println(err)
			}

			tmpDeploy, _ := deploymentsClient.Get(deployment.Name, v1.GetOptions{})
			if tmpDeploy.OwnerReferences != nil && isOwnerEqual(deployment.OwnerReferences[0], tmpDeploy.OwnerReferences[0]) {
				deployResult, err := deploymentsClient.Update(deployment)
				if err != nil {
					handlerLog.Info("Deployment update failed.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Deployment", deployment.Name, "Error", err)
				} else {
					handlerLog.Info("Deployment updated.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Deployment", deployResult.Name)
				}
			} else {
				deployResult, err := deploymentsClient.Create(deployment)
				if err != nil {
					handlerLog.Info("Deployment create failed.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Deployment", deployment.Name, "Error", err)
				} else {
					handlerLog.Info("Deployment created.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Deployment", deployResult.Name)
				}
			}

			if comp.Spec.WorkloadType == WorkloadTypeServer || comp.Spec.WorkloadType == WorkloadTypeSingletonServer {
				serviceClient := s.K8sclient.CoreV1().Services(ac.Namespace)
				service := convertService(owner, compConf, *comp)
				tmpsvc, _ := serviceClient.Get(service.Name, v1.GetOptions{})
				if tmpsvc.OwnerReferences != nil && isOwnerEqual(service.OwnerReferences[0], tmpsvc.OwnerReferences[0]) {
					service.ResourceVersion = tmpsvc.ResourceVersion
					service.Spec.ClusterIP = tmpsvc.Spec.ClusterIP
					svcResult, err := serviceClient.Update(service)
					if err != nil {
						handlerLog.Info("Service update failed.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Service", service.Name, "Error", err)
					} else {
						handlerLog.Info("Service updated.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Service", svcResult.Name)
					}
				} else {
					svcResult, err := serviceClient.Create(service)
					if err != nil {
						handlerLog.Info("Service create failed.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Service", service.Name, "Error", err)
					} else {
						handlerLog.Info("Service created.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Service", svcResult.Name)
					}
				}
			}
		case WorkloadTypeTask, WorkloadTypeSingletonTask:
			job, configMaps := convertJob(owner, compConf, *comp)
			parallelism := *getManuelScale(compConf.Traits)
			if comp.Spec.WorkloadType == WorkloadTypeSingletonTask {
				parallelism = 1
			}
			job.Spec.Parallelism = &parallelism
			jobsClient := s.K8sclient.BatchV1().Jobs(ac.Namespace)

			if err := createOrUpdateConfigMaps(s, ac.Namespace, configMaps); err != nil {
				log.Println(err)
			}

			tmpJob, _ := jobsClient.Get(job.Name, v1.GetOptions{})
			if tmpJob.OwnerReferences != nil && isOwnerEqual(job.OwnerReferences[0], tmpJob.OwnerReferences[0]) {
				jobResult, err := jobsClient.Update(job)
				if err != nil {
					handlerLog.Info("Job update failed.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Job", job.Name, "Error", err)
				} else {
					handlerLog.Info("Job updated.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Job", jobResult.Name)
				}
			} else {
				jobResult, err := jobsClient.Create(job)
				if err != nil {
					handlerLog.Info("Job create failed.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Job", job.Name, "Error", err)
				} else {
					handlerLog.Info("Job created.", "Namespace", ac.Namespace, "ApplicationConfiguration", ac.Name, "Component", compConf.ComponentName, "Job", jobResult.Name)
				}
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

func getManuelScale(traits []v1alpha1.TraitBinding) *int32 {
	var def int32 = 1
	for _, tr := range traits {
		if tr.Name != "manual-scaler" {
			continue
		}
		values := make(map[string]interface{})
		err := json.Unmarshal(tr.Properties.Raw, &values)
		if err != nil {
			handlerLog.Info("traits value spec error", "Error", err)
			continue
		}
		f, ok := values["replicaCount"]
		if !ok {
			handlerLog.Info("replicaCount didn't exist error")
			continue
		}
		ff, ok := f.(float64)
		if !ok {
			handlerLog.Info("replicaCount type is " + reflect.TypeOf(f).Name())
			continue
		}
		def = int32(ff)
	}
	return &def
}
