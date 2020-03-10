package controllers

import (
	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	"strings"
)

func injectLogPilotConfigs(container *apiv1.Container, traits []v1alpha1.TraitBinding) {
	for _, tr := range traits {
		if tr.Name != "log-pilot" {
			continue
		}
		container.Env = append(container.Env, getLogPilotEnvs(tr)...)
		container.VolumeMounts = append(container.VolumeMounts, getLogPilotVolumeMounts(tr))
	}
}

func getLogPilotEnvs(trait v1alpha1.TraitBinding) []apiv1.EnvVar {
	var envs []apiv1.EnvVar

	values, err := parsePropertiesOfTrait(trait)
	if err != nil {
		return nil
	}
	logDir := values["path"].(string)
	if !strings.HasSuffix(logDir, "/") {
		logDir += "/"
	}
	logDir += "*"
	index := values["name"].(string)
	tags := values["tags"].(string)

	pilotLogPrefix := "aliyun"
	prefix, ok := values["pilotLogPrefix"]
	if ok {
		pilotLogPrefix = prefix.(string)
	}
	envLog := apiv1.EnvVar{
		Name:  pilotLogPrefix + "_logs_" + index,
		Value: logDir,
	}
	envTag := apiv1.EnvVar{
		Name:  pilotLogPrefix + "_logs_" + index + "_tags",
		Value: tags,
	}
	envs = append(envs, envLog, envTag)
	return envs
}

func getLogPilotVolumeMounts(trait v1alpha1.TraitBinding) apiv1.VolumeMount {
	var volumeMount apiv1.VolumeMount

	values, err := parsePropertiesOfTrait(trait)
	if err != nil {
		return volumeMount
	}
	container := values["container"].(string)
	logDir := values["path"].(string)
	volumeMount = apiv1.VolumeMount{
		Name:      container + "-log",
		MountPath: logDir,
	}

	return volumeMount
}

func getLogPilotVolumes(traits []v1alpha1.TraitBinding) []apiv1.Volume {
	var volumes []apiv1.Volume
	for _, tr := range traits {
		if tr.Name != "log-pilot" {
			continue
		}
		values, err := parsePropertiesOfTrait(tr)
		if err != nil {
			continue
		}
		container := values["container"].(string)
		volume := apiv1.Volume{
			Name: container + "-log",
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		}
		volumes = append(volumes, volume)
	}
	return volumes
}
