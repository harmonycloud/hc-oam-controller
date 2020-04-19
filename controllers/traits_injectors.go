package controllers

import (
	"encoding/json"
	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	traits2 "hc-oam-controller/api/core.oam.dev/v1alpha1/traits"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
)

var (
	traitsInjectorLog = ctrl.Log.WithName("traits-injector")
)

func injectLogPilotConfigs(container *apiv1.Container, traits []v1alpha1.TraitBinding) {
	for _, tr := range traits {
		if tr.Name != "log-pilot" {
			continue
		}
		values, err := parsePropertiesOfTrait(tr)
		if err != nil || container.Name != values["container"].(string) {
			return
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

func injectHostPolicy(podSpec *apiv1.PodSpec, traits []v1alpha1.TraitBinding) {
	for _, tr := range traits {
		if tr.Name != "host-policy" {
			continue
		}
		hostPolicy := new(traits2.HostPolicy)
		if err := json.Unmarshal(tr.Properties.Raw, &hostPolicy); err != nil {
			traitsInjectorLog.Info(err.Error())
		}
		podSpec.HostNetwork = hostPolicy.HostNetwork
		podSpec.HostPID = hostPolicy.HostPid
		podSpec.HostIPC = hostPolicy.HostIpc
	}
}

func injectResourcesPolicy(container *apiv1.Container, traits []v1alpha1.TraitBinding) {
	for _, tr := range traits {
		if tr.Name != "resources-policy" {
			continue
		}
		resourcesPolicy := new(traits2.ResourcesPolicy)
		if err := json.Unmarshal(tr.Properties.Raw, &resourcesPolicy); err != nil {
			traitsInjectorLog.Info(err.Error())
		}
		if resourcesPolicy.Container != container.Name {
			continue
		}
		container.Resources.Limits = resourcesPolicy.Limits
	}
}

func injectSchedulePolicy(namespace string, spec *apiv1.PodSpec, traits []v1alpha1.TraitBinding) {
	for _, tr := range traits {
		if tr.Name != "schedule-policy" {
			continue
		}
		schedulePolicy := new(traits2.SchedulePolicy)
		if err := json.Unmarshal(tr.Properties.Raw, &schedulePolicy); err != nil {
			traitsInjectorLog.Info(err.Error())
		}

		var nodeSelectorTerms []apiv1.NodeSelectorTerm
		var nodePreferredSchedulingTerms []apiv1.PreferredSchedulingTerm
		var podAffinityTerms []apiv1.PodAffinityTerm
		var WeightedPodAffinityTerms []apiv1.WeightedPodAffinityTerm
		var podAntiAffinityTerms []apiv1.PodAffinityTerm
		var weightedPodAntiAffinityTerms []apiv1.WeightedPodAffinityTerm

		// NodeAffinity
		if schedulePolicy.NodeAffinity.Type == "required" {
			var matchExpressions []apiv1.NodeSelectorRequirement
			for k, v := range schedulePolicy.NodeAffinity.Selector {
				matchExpressions = append(matchExpressions, apiv1.NodeSelectorRequirement{
					Key:      k,
					Operator: apiv1.NodeSelectorOpIn,
					Values:   []string{v},
				})
			}
			nodeSelectorTerms = append(nodeSelectorTerms,
				apiv1.NodeSelectorTerm{
					MatchExpressions: matchExpressions,
				})
		} else {
			var matchExpressions []apiv1.NodeSelectorRequirement
			for k, v := range schedulePolicy.NodeAffinity.Selector {
				matchExpressions = append(matchExpressions, apiv1.NodeSelectorRequirement{
					Key:      k,
					Operator: apiv1.NodeSelectorOpIn,
					Values:   []string{v},
				})
			}
			nodePreferredSchedulingTerms = append(nodePreferredSchedulingTerms,
				apiv1.PreferredSchedulingTerm{
					Weight: 50,
					Preference: apiv1.NodeSelectorTerm{
						MatchExpressions: matchExpressions,
					},
				})
		}

		// PodAffinity
		if schedulePolicy.PodAffinity.Type == "required" {
			var matchExpressions []v1.LabelSelectorRequirement
			for k, v := range schedulePolicy.PodAffinity.Selector {
				matchExpressions = append(matchExpressions, v1.LabelSelectorRequirement{
					Key:      k,
					Operator: v1.LabelSelectorOpIn,
					Values:   []string{v},
				})
			}
			podAffinityTerms = append(podAffinityTerms,
				apiv1.PodAffinityTerm{
					LabelSelector: &v1.LabelSelector{
						MatchExpressions: matchExpressions,
					},
					Namespaces:  []string{namespace},
					TopologyKey: "kubernetes.io/hostname",
				})
		} else {
			var matchExpressions []v1.LabelSelectorRequirement
			for k, v := range schedulePolicy.PodAffinity.Selector {
				matchExpressions = append(matchExpressions, v1.LabelSelectorRequirement{
					Key:      k,
					Operator: v1.LabelSelectorOpIn,
					Values:   []string{v},
				})
			}
			WeightedPodAffinityTerms =
				append(WeightedPodAffinityTerms,
					apiv1.WeightedPodAffinityTerm{
						Weight: 50,
						PodAffinityTerm: apiv1.PodAffinityTerm{
							LabelSelector: &v1.LabelSelector{
								MatchExpressions: matchExpressions,
							},
							Namespaces:  []string{namespace},
							TopologyKey: "kubernetes.io/hostname",
						},
					})
		}

		// PodAntiAffinity
		if schedulePolicy.PodAntiAffinity.Type == "required" {
			var matchExpressions []v1.LabelSelectorRequirement
			for k, v := range schedulePolicy.PodAntiAffinity.Selector {
				matchExpressions = append(matchExpressions, v1.LabelSelectorRequirement{
					Key:      k,
					Operator: v1.LabelSelectorOpIn,
					Values:   []string{v},
				})
			}
			podAntiAffinityTerms =
				append(podAntiAffinityTerms,
					apiv1.PodAffinityTerm{
						LabelSelector: &v1.LabelSelector{
							MatchExpressions: matchExpressions,
						},
						Namespaces:  []string{namespace},
						TopologyKey: "kubernetes.io/hostname",
					})
		} else {
			var matchExpressions []v1.LabelSelectorRequirement
			for k, v := range schedulePolicy.PodAntiAffinity.Selector {
				matchExpressions = append(matchExpressions, v1.LabelSelectorRequirement{
					Key:      k,
					Operator: v1.LabelSelectorOpIn,
					Values:   []string{v},
				})
			}
			weightedPodAntiAffinityTerms =
				append(weightedPodAntiAffinityTerms, apiv1.WeightedPodAffinityTerm{
					Weight: 50,
					PodAffinityTerm: apiv1.PodAffinityTerm{
						LabelSelector: &v1.LabelSelector{
							MatchExpressions: matchExpressions,
						},
						Namespaces:  []string{namespace},
						TopologyKey: "kubernetes.io/hostname",
					},
				})
		}

		var nodeRequired *apiv1.NodeSelector
		if nodeSelectorTerms != nil {
			nodeRequired = &apiv1.NodeSelector{NodeSelectorTerms: nodeSelectorTerms}
		}

		if nodeSelectorTerms != nil {

		}

		spec.Affinity = &apiv1.Affinity{
			NodeAffinity: &apiv1.NodeAffinity{
				//RequiredDuringSchedulingIgnoredDuringExecution:  &apiv1.NodeSelector{NodeSelectorTerms:nodeSelectorTerms},
				RequiredDuringSchedulingIgnoredDuringExecution:  nodeRequired,
				PreferredDuringSchedulingIgnoredDuringExecution: nodePreferredSchedulingTerms,
			},
			PodAffinity: &apiv1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution:  podAffinityTerms,
				PreferredDuringSchedulingIgnoredDuringExecution: WeightedPodAffinityTerms,
			},
			PodAntiAffinity: &apiv1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution:  podAntiAffinityTerms,
				PreferredDuringSchedulingIgnoredDuringExecution: weightedPodAntiAffinityTerms,
			},
		}
	}
}
