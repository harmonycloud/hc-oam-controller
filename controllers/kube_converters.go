package controllers

import (
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"

	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func convertDeployment(owner v1.OwnerReference, labels map[string]string, compConf v1alpha1.ComponentConfiguration, comp v1alpha1.ComponentSchematic, parameterMap map[string]string) *appsv1.Deployment {
	labels["role"] = "workload"
	labels["app"] = compConf.InstanceName
	containers := convertContainers(owner, compConf.InstanceName, comp.Spec.Containers, parameterMap)
	deployment := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: compConf.InstanceName,
			OwnerReferences: []v1.OwnerReference{
				owner,
			},
			Labels: labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &v1.LabelSelector{
				MatchLabels: labels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: labels,
				},
				Spec: apiv1.PodSpec{
					Containers: containers,
					Volumes:    nil,
				},
			},
		},
	}
	return deployment
}

func convertService(owner v1.OwnerReference, labels map[string]string, compConf v1alpha1.ComponentConfiguration, comp v1alpha1.ComponentSchematic) *apiv1.Service {
	labels["role"] = "workload"
	labels["app"] = compConf.InstanceName
	servicePorts := convertsServicePorts(comp.Spec.Containers)
	if servicePorts == nil || cap(servicePorts) == 0 {
		return nil
	}

	service := &apiv1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name: compConf.InstanceName,
			OwnerReferences: []v1.OwnerReference{
				owner,
			},
			Labels: labels,
		},
		Spec: apiv1.ServiceSpec{
			Ports:    servicePorts,
			Selector: labels,
			Type:     "ClusterIP",
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

func convertJob(owner v1.OwnerReference, labels map[string]string, compConf v1alpha1.ComponentConfiguration, comp v1alpha1.ComponentSchematic, parameterMap map[string]string) *batchv1.Job {
	labels["role"] = "workload"
	containers := convertContainers(owner, compConf.InstanceName, comp.Spec.Containers, parameterMap)
	job := &batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name: compConf.InstanceName,
			OwnerReferences: []v1.OwnerReference{
				owner,
			},
			Labels: labels,
		},
		Spec: batchv1.JobSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: containers,
				},
			},
		},
	}
	return job
}

func convertCornJob(owner v1.OwnerReference, labels map[string]string, compConf v1alpha1.ComponentConfiguration, comp v1alpha1.ComponentSchematic, parameterMap map[string]string) *batchv1beta1.CronJob {
	labels["role"] = "workload"
	containers := convertContainers(owner, compConf.InstanceName, comp.Spec.Containers, parameterMap)
	cronJob := &batchv1beta1.CronJob{
		ObjectMeta: v1.ObjectMeta{
			Name: compConf.InstanceName,
			OwnerReferences: []v1.OwnerReference{
				owner,
			},
			Labels: labels,
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule: "",
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: apiv1.PodTemplateSpec{
						Spec: apiv1.PodSpec{
							Containers: containers,
						},
					},
				},
			},
		},
	}
	return cronJob
}

func convertContainers(owner v1.OwnerReference, instanceName string, oamContainers []v1alpha1.Container, parameterMap map[string]string) []apiv1.Container {
	var containerlist []apiv1.Container
	for _, c := range oamContainers {
		container := &apiv1.Container{
			Name:           c.Name,
			Image:          c.Image,
			Command:        c.Cmd,
			Args:           c.Args,
			Ports:          convertContainerPorts(c),
			Env:            convertEnvs(c.Env, parameterMap),
			Resources:      convertResources(&c.Resources),
			VolumeMounts:   convertVolumeMounts(instanceName, c.Name, c.Resources.Volumes, c.Config),
			LivenessProbe:  convertProbe(c.LivenessProbe),
			ReadinessProbe: convertProbe(c.ReadinessProbe),
		}
		containerlist = append(containerlist, *container)
	}
	return containerlist
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

func convertConfigMaps(owner v1.OwnerReference, labels map[string]string, compConf v1alpha1.ComponentConfiguration, comp v1alpha1.ComponentSchematic, parameterMap map[string]string) []apiv1.ConfigMap {
	labels["role"] = "workload"
	var configMaps []apiv1.ConfigMap
	for _, container := range comp.Spec.Containers {
		if configMap := convertConfigMap(owner, labels, container.Config, parameterMap, compConf.InstanceName+"-"+container.Name); configMap != nil {
			configMaps = append(configMaps, *configMap)
		}
	}
	return configMaps
}

func convertConfigMap(owner v1.OwnerReference, labels map[string]string, oamConfigFile []v1alpha1.ConfigFile, parameterMap map[string]string, configMapName string) *apiv1.ConfigMap {
	if oamConfigFile == nil {
		return nil
	}
	configmap := &apiv1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name: configMapName,
			OwnerReferences: []v1.OwnerReference{
				owner,
			},
			Labels: labels,
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

func convertVolumeMounts(instanceName string, containerName string, oamVolumes []v1alpha1.Volume, oamConfigFile []v1alpha1.ConfigFile) []apiv1.VolumeMount {
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
			Name:      instanceName + "-" + containerName + "-config",
			MountPath: f.Path,
			SubPath:   getConfigFileName(f.Path),
		}
		volumeMounts = append(volumeMounts, volumeMount)
	}
	return volumeMounts
}

func convertVolumesFromConfig(configMaps []apiv1.ConfigMap) []apiv1.Volume {
	var volumes []apiv1.Volume
	for _, c := range configMaps {
		volume := apiv1.Volume{
			Name: c.Name + "-config",
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{Name: c.Name},
					Items:                convertConfigVolumeSourceItems(c),
				},
			},
		}
		volumes = append(volumes, volume)
	}
	return volumes
}

func convertConfigVolumeSourceItems(configMap apiv1.ConfigMap) []apiv1.KeyToPath {
	var keyToPaths []apiv1.KeyToPath
	for k := range configMap.Data {
		ktp := apiv1.KeyToPath{
			Key:  k,
			Path: k,
		}
		keyToPaths = append(keyToPaths, ktp)
	}
	return keyToPaths
}

func getReadOnly(accessMode v1alpha1.AccessMode) bool {
	if accessMode == "RO" {
		return true
	}
	return false
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
