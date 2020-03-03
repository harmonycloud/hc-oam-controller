package controllers

import (
	v12 "k8s.io/api/batch/v1"
	v1beta12 "k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"

	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func convertDeployment(owner v1.OwnerReference, compConf v1alpha1.ComponentConfiguration, comp v1alpha1.ComponentSchematic) (*appsv1.Deployment, []apiv1.ConfigMap) {
	containers, configMaps := convertContainers(owner, comp.Spec.Containers, parseParameters(compConf.ParameterValues))
	deployment := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: compConf.InstanceName,
			OwnerReferences: []v1.OwnerReference{
				owner,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": compConf.InstanceName,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"app": compConf.InstanceName,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: containers,
					Volumes:    nil,
				},
			},
		},
	}
	return deployment, configMaps
}

func convertService(owner v1.OwnerReference, compConf v1alpha1.ComponentConfiguration, comp v1alpha1.ComponentSchematic) *apiv1.Service {
	service := &apiv1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name: compConf.InstanceName,
			OwnerReferences: []v1.OwnerReference{
				owner,
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: convertsServicePorts(comp.Spec.Containers),
			Selector: map[string]string{
				"app": compConf.InstanceName,
			},
			Type: "ClusterIP",
		},
	}
	return service
}
func convertsServicePorts(oamContainers []v1alpha1.Container) []apiv1.ServicePort {
	var servicePorts []apiv1.ServicePort

	for _, c := range oamContainers {
		for _, p := range c.Ports {
			servicePort := &apiv1.ServicePort{
				Name: p.Name,
				Port: p.ContainerPort,
				TargetPort: intstr.IntOrString{
					Type:   0,
					IntVal: p.ContainerPort,
					StrVal: "",
				},
				Protocol: apiv1.Protocol(p.Protocol),
			}
			servicePorts = append(servicePorts, *servicePort)
		}
	}
	return servicePorts
}

func convertJob(owner v1.OwnerReference, compConf v1alpha1.ComponentConfiguration, comp v1alpha1.ComponentSchematic) (*v12.Job, []apiv1.ConfigMap) {
	containers, configMaps := convertContainers(owner, comp.Spec.Containers, parseParameters(compConf.ParameterValues))
	job := &v12.Job{
		ObjectMeta: v1.ObjectMeta{
			Name: compConf.InstanceName,
			OwnerReferences: []v1.OwnerReference{
				owner,
			},
		},
		Spec: v12.JobSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: containers,
				},
			},
		},
	}
	return job, configMaps
}

func convertCornJob(owner v1.OwnerReference, compConf v1alpha1.ComponentConfiguration, comp v1alpha1.ComponentSchematic) (*v1beta12.CronJob, []apiv1.ConfigMap) {
	containers, configMaps := convertContainers(owner, comp.Spec.Containers, parseParameters(compConf.ParameterValues))
	cronJob := &v1beta12.CronJob{
		ObjectMeta: v1.ObjectMeta{
			Name: compConf.InstanceName,
			OwnerReferences: []v1.OwnerReference{
				owner,
			},
		},
		Spec: v1beta12.CronJobSpec{
			Schedule: "",
			JobTemplate: v1beta12.JobTemplateSpec{
				Spec: v12.JobSpec{
					Template: apiv1.PodTemplateSpec{
						Spec: apiv1.PodSpec{
							Containers: containers,
						},
					},
				},
			},
		},
	}
	return cronJob, configMaps
}

func convertContainers(owner v1.OwnerReference, oamContainers []v1alpha1.Container, parameterMap map[string]string) ([]apiv1.Container, []apiv1.ConfigMap) {
	var containerlist []apiv1.Container
	var configMapList []apiv1.ConfigMap
	for _, c := range oamContainers {
		container := &apiv1.Container{
			Name:           c.Name,
			Image:          c.Image,
			Command:        c.Cmd,
			Args:           c.Args,
			Ports:          convertContainerPorts(c),
			Env:            convertEnvs(c.Env, parameterMap),
			Resources:      convertResources(&c.Resources),
			VolumeMounts:   convertVolumeMounts(c.Resources.Volumes, c.Config),
			LivenessProbe:  convertProbe(c.LivenessProbe),
			ReadinessProbe: convertProbe(c.ReadinessProbe),
		}
		containerlist = append(containerlist, *container)
		cm := convertConfigMap(owner, c.Config, parameterMap, owner.Name+"-"+c.Name+"-config")
		if cm != nil {
			configMapList = append(configMapList, *cm)
		}
	}
	return containerlist, configMapList
}

func convertContainerPorts(oamContainer v1alpha1.Container) []apiv1.ContainerPort {
	var containerPorts []apiv1.ContainerPort
	for _, p := range oamContainer.Ports {
		containerPort := &apiv1.ContainerPort{
			Name:          p.Name,
			ContainerPort: p.ContainerPort,
			Protocol:      apiv1.Protocol(p.Protocol),
		}
		containerPorts = append(containerPorts, *containerPort)
	}
	return containerPorts
}

func createOrUpdateConfigMaps(s *Handler, namespace string, configMaps []apiv1.ConfigMap) error {
	configMapClient := s.K8sclient.CoreV1().ConfigMaps(namespace)
	for _, configmap := range configMaps {
		tmpCm, _ := configMapClient.Get(configmap.Name, v1.GetOptions{})
		if tmpCm.OwnerReferences != nil && isOwnerEqual(configmap.OwnerReferences[0], tmpCm.OwnerReferences[0]) {
			cmResult, err := configMapClient.Update(&configmap)
			if err != nil {
				handlerLog.Info("ConfigMap update failed.", "Namespace", namespace, "ConfigMap", configmap.Name, "Error", err)
				return err
			} else {
				handlerLog.Info("ConfigMap updated.", "Namespace", namespace, "ConfigMap", cmResult.Name)
			}
		} else {
			cmResult, err := configMapClient.Create(&configmap)
			if err != nil {
				handlerLog.Info("ConfigMap create failed.", "Namespace", namespace, "ConfigMap", configmap.Name, "Error", err)
				return err
			} else {
				handlerLog.Info("ConfigMap created.", "Namespace", namespace, "ConfigMap", cmResult.Name)
			}
		}
	}
	return nil
}

func convertConfigMap(owner v1.OwnerReference, oamConfigFile []v1alpha1.ConfigFile, parameterMap map[string]string, configMapName string) *apiv1.ConfigMap {
	if oamConfigFile == nil {
		return nil
	}
	configmap := &apiv1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name: configMapName,
			OwnerReferences: []v1.OwnerReference{
				owner,
			},
		},
		Data: map[string]string{},
	}
	for _, f := range oamConfigFile {
		value := f.Value
		if &f.FromParam != nil {
			value = parameterMap[f.FromParam]
		}
		configmap.Data[getConfigFileName(f.Path)] = value
	}
	return configmap
}

func getConfigFileName(path string) string {
	ss := strings.Split(path, "/")
	return ss[len(ss)-1]
}

func convertVolumeMounts(oamVolumes []v1alpha1.Volume, oamConfigFile []v1alpha1.ConfigFile) []apiv1.VolumeMount {
	var volumeMounts []apiv1.VolumeMount
	for _, v := range oamVolumes {
		volumeMount := apiv1.VolumeMount{
			Name:             v.Name,
			ReadOnly:         getReadOnly(v.AccessMode),
			MountPath:        v.MountPath,
			SubPath:          "",
			MountPropagation: nil,
			SubPathExpr:      "",
		}
		volumeMounts = append(volumeMounts, volumeMount)
	}
	for _, f := range oamConfigFile {
		volumeMount := apiv1.VolumeMount{

			Name:      "config",
			MountPath: f.Path,
			SubPath:   getConfigFileName(f.Path),
		}
		volumeMounts = append(volumeMounts, volumeMount)
	}
	return volumeMounts
}

func getReadOnly(accessMode v1alpha1.AccessMode) bool {
	if accessMode == "RW" {
		return false
	}
	return true
}

func convertProbe(oamProbe *v1alpha1.HealthProbe) *apiv1.Probe {
	if oamProbe == nil {
		return nil
	}
	probe := apiv1.Probe{
		Handler: apiv1.Handler{
			Exec: &apiv1.ExecAction{Command: oamProbe.Exec.Command},
			HTTPGet: &apiv1.HTTPGetAction{
				Path: oamProbe.HttpGet.Path,
				Port: intstr.IntOrString{
					Type:   0,
					IntVal: oamProbe.HttpGet.Port,
					StrVal: "",
				},
				HTTPHeaders: convertHttpHeaders(oamProbe.HttpGet.HttpHeaders),
			},
			TCPSocket: &apiv1.TCPSocketAction{
				Port: intstr.IntOrString{
					Type:   0,
					IntVal: oamProbe.TcpSocket.Port,
					StrVal: "",
				},
			},
		},
		InitialDelaySeconds: oamProbe.InitialDelaySeconds,
		TimeoutSeconds:      oamProbe.TimeoutSeconds,
		PeriodSeconds:       oamProbe.PeriodSeconds,
		SuccessThreshold:    oamProbe.SuccessThreshold,
		FailureThreshold:    oamProbe.FailureThreshold,
	}
	return &probe
}

func convertHttpHeaders(oamHeaders []v1alpha1.HttpHeader) []apiv1.HTTPHeader {
	var headers []apiv1.HTTPHeader
	for _, h := range oamHeaders {
		header := apiv1.HTTPHeader{
			Name:  h.Name,
			Value: h.Value,
		}
		headers = append(headers, header)
	}
	return headers
}

func convertEnvs(oamEnvs []v1alpha1.Env, parameterMap map[string]string) []apiv1.EnvVar {
	var envs []apiv1.EnvVar
	for _, e := range oamEnvs {
		env := apiv1.EnvVar{
			Name:  e.Name,
			Value: e.Value,
		}

		if &e.FromParam != nil {
			if v, ok := parameterMap[e.FromParam]; ok {
				env.Value = v
			}
		}
		envs = append(envs, env)
	}
	return envs
}

func convertResources(oamResources *v1alpha1.Resources) apiv1.ResourceRequirements {
	resources := apiv1.ResourceRequirements{
		Limits:   convertResourceList(oamResources),
		Requests: convertResourceList(oamResources),
	}
	return resources
}

func convertResourceList(oamResources *v1alpha1.Resources) apiv1.ResourceList {
	resourceList := map[apiv1.ResourceName]resource.Quantity{}
	if oamResources.Cpu.Required.Value() != 0 {
		resourceList["cpu"] = oamResources.Cpu.Required
	}
	if oamResources.Memory.Required.Value() != 0 {
		resourceList["memory"] = oamResources.Memory.Required
	}
	if oamResources.Gpu.Required.Value() != 0 {
		resourceList["nvidia/gpu"] = oamResources.Gpu.Required
	}
	for _, e := range oamResources.Extended {
		var q resource.Quantity
		q, _ = resource.ParseQuantity(e.Required)
		resourceList[apiv1.ResourceName(e.Name)] = q
	}
	return resourceList
}
