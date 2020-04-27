package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	traits2 "hc-oam-controller/api/core.oam.dev/v1alpha1/traits"
	hcv1beta1 "hc-oam-controller/api/harmonycloud.cn/v1beta1"
	"k8s.io/api/autoscaling/v2beta2"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"

	//"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
)

var (
	traitsConverterLog = ctrl.Log.WithName("traits-converter")
)

func convertHpa(owner v1.OwnerReference, annotations map[string]string, kind string, apiVersion string, instanceName string, traits []v1alpha1.TraitBinding) *v2beta2.HorizontalPodAutoscaler {
	annotations["role"] = "trait"
	var hpa *v2beta2.HorizontalPodAutoscaler
	for _, tr := range traits {
		if tr.Name != "auto-scaler" {
			continue
		}
		values, err := parsePropertiesOfTrait(tr)
		if err != nil {
			continue
		}

		min, ok := values["minimum"]
		var minimum int32
		if !ok {
			minimum = 1
		} else {
			minimum = int32(min.(float64))
		}

		max, ok := values["maximum"]
		var maximum int32
		if !ok {
			maximum = 10
		} else {
			maximum = int32(max.(float64))
		}

		cpu, ok := values["cpu"]
		var cpuMetric v2beta2.MetricSpec
		if ok {
			utilization := int32(cpu.(float64))
			cpuMetric = v2beta2.MetricSpec{
				Type: v2beta2.ResourceMetricSourceType,
				Resource: &v2beta2.ResourceMetricSource{
					Name: "cpu",
					Target: v2beta2.MetricTarget{
						Type:               v2beta2.UtilizationMetricType,
						AverageUtilization: &utilization,
					},
				},
			}
		}

		memory, ok := values["memory"]
		var memoryMetric v2beta2.MetricSpec
		if ok {
			utilization := int32(memory.(float64))
			memoryMetric = v2beta2.MetricSpec{
				Type: v2beta2.ResourceMetricSourceType,
				Resource: &v2beta2.ResourceMetricSource{
					Name: "memory",
					Target: v2beta2.MetricTarget{
						Type:               v2beta2.UtilizationMetricType,
						AverageUtilization: &utilization,
					},
				},
			}
		}

		hpa = &v2beta2.HorizontalPodAutoscaler{
			ObjectMeta: v1.ObjectMeta{
				Name: instanceName,
				OwnerReferences: []v1.OwnerReference{
					owner,
				},
				Annotations: annotations,
			},
			Spec: v2beta2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: v2beta2.CrossVersionObjectReference{
					Kind:       kind,
					Name:       instanceName,
					APIVersion: apiVersion,
				},
				MinReplicas: &minimum,
				MaxReplicas: maximum,
				Metrics: []v2beta2.MetricSpec{
					cpuMetric,
					memoryMetric,
				},
			},
		}
	}
	return hpa
}

func convertHcHpa(owner v1.OwnerReference, annotations map[string]string, kind string, apiVersion string, instanceName string, traits []v1alpha1.TraitBinding) *hcv1beta1.HorizontalPodAutoscaler {
	annotations["role"] = "trait"
	var hcHpa *hcv1beta1.HorizontalPodAutoscaler
	for _, tr := range traits {
		if tr.Name != "better-auto-scaler" {
			continue
		}
		betterAutoScaler := new(traits2.BetterAutoScaler)
		if err := json.Unmarshal(tr.Properties.Raw, &betterAutoScaler); err != nil {
			traitsConverterLog.Info(err.Error())
		}

		strVarToIntVar(&betterAutoScaler.Maximum)
		strVarToIntVar(&betterAutoScaler.Minimum)
		strVarToIntVar(&betterAutoScaler.CpuDown)
		strVarToIntVar(&betterAutoScaler.CpuUp)
		strVarToIntVar(&betterAutoScaler.MemoryDown)
		strVarToIntVar(&betterAutoScaler.MemoryUp)

		if betterAutoScaler.Minimum.IntVal < 1 {
			betterAutoScaler.Minimum = intstr.IntOrString{
				Type:   0,
				IntVal: 1,
				StrVal: "",
			}
		}

		if betterAutoScaler.Minimum.IntVal < 1 {
			betterAutoScaler.Minimum = intstr.IntOrString{
				Type:   0,
				IntVal: 2,
				StrVal: "",
			}
		}

		var cpuUpMetric hcv1beta1.MetricSpec
		if betterAutoScaler.CpuUp.IntVal > 0 && betterAutoScaler.CpuUp.IntVal < 100 {
			utilization := betterAutoScaler.CpuUp
			cpuUpMetric = hcv1beta1.MetricSpec{
				Type: hcv1beta1.ResourceMetricSourceType,
				Resource: &hcv1beta1.ResourceMetricSource{
					Name: "cpu",
					Target: hcv1beta1.MetricTarget{
						Type:               hcv1beta1.UtilizationMetricType,
						AverageUtilization: &utilization.IntVal,
						ScaleType:          hcv1beta1.ScaleUpMetricsTargetType,
					},
				},
			}
		}

		var cpuDownMetric hcv1beta1.MetricSpec
		if betterAutoScaler.CpuDown.IntVal > 0 && betterAutoScaler.CpuDown.IntVal < 100 {
			utilization := betterAutoScaler.CpuDown
			cpuDownMetric = hcv1beta1.MetricSpec{
				Type: hcv1beta1.ResourceMetricSourceType,
				Resource: &hcv1beta1.ResourceMetricSource{
					Name: "cpu",
					Target: hcv1beta1.MetricTarget{
						Type:               hcv1beta1.UtilizationMetricType,
						AverageUtilization: &utilization.IntVal,
						ScaleType:          hcv1beta1.ScaleDownMetricsTargetType,
					},
				},
			}
		}

		var memoryUpMetric hcv1beta1.MetricSpec
		if betterAutoScaler.MemoryUp.IntVal > 0 && betterAutoScaler.MemoryUp.IntVal < 100 {
			utilization := betterAutoScaler.MemoryUp
			memoryUpMetric = hcv1beta1.MetricSpec{
				Type: hcv1beta1.ResourceMetricSourceType,
				Resource: &hcv1beta1.ResourceMetricSource{
					Name: "memory",
					Target: hcv1beta1.MetricTarget{
						Type:               hcv1beta1.UtilizationMetricType,
						AverageUtilization: &utilization.IntVal,
						ScaleType:          hcv1beta1.ScaleUpMetricsTargetType,
					},
				},
			}
		}

		var memoryDownMetric hcv1beta1.MetricSpec
		if betterAutoScaler.MemoryDown.IntVal > 0 && betterAutoScaler.MemoryDown.IntVal < 100 {
			utilization := betterAutoScaler.MemoryDown
			memoryDownMetric = hcv1beta1.MetricSpec{
				Type: hcv1beta1.ResourceMetricSourceType,
				Resource: &hcv1beta1.ResourceMetricSource{
					Name: "memory",
					Target: hcv1beta1.MetricTarget{
						Type:               hcv1beta1.UtilizationMetricType,
						AverageUtilization: &utilization.IntVal,
						ScaleType:          hcv1beta1.ScaleDownMetricsTargetType,
					},
				},
			}
		}

		hcHpa = &hcv1beta1.HorizontalPodAutoscaler{
			ObjectMeta: v1.ObjectMeta{
				Name: instanceName,
				OwnerReferences: []v1.OwnerReference{
					owner,
				},
				Annotations: annotations,
			},
			Spec: hcv1beta1.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: hcv1beta1.CrossVersionObjectReference{
					Kind:       kind,
					Name:       instanceName,
					APIVersion: apiVersion,
				},
				MinReplicas: &betterAutoScaler.Minimum.IntVal,
				MaxReplicas: betterAutoScaler.Maximum.IntVal,
				Metrics: []hcv1beta1.MetricSpec{
					cpuUpMetric,
					cpuDownMetric,
					memoryUpMetric,
					memoryDownMetric,
				},
			},
		}
	}
	return hcHpa
}

func convertIngress(owner v1.OwnerReference, annotations map[string]string, instanceName string, traits []v1alpha1.TraitBinding) *v1beta1.Ingress {
	annotations["role"] = "trait"
	var ingressRules []v1beta1.IngressRule
	var ingress *v1beta1.Ingress
	for _, tr := range traits {
		if tr.Name != "ingress" {
			continue
		}
		values, err := parsePropertiesOfTrait(tr)
		if err != nil {
			continue
		}
		hostname := values["hostname"].(string)

		path := "/"
		if p, ok := values["path"]; ok {
			path = p.(string)
		}
		port := int32(values["servicePort"].(float64))

		httpIngressPath := v1beta1.HTTPIngressPath{
			Path: path,
			Backend: v1beta1.IngressBackend{
				ServiceName: instanceName,
				ServicePort: intstr.IntOrString{
					Type:   0,
					IntVal: port,
					StrVal: "",
				},
			},
		}
		httpIngressRuleValue := new(v1beta1.HTTPIngressRuleValue)
		httpIngressRuleValue.Paths = append(httpIngressRuleValue.Paths, httpIngressPath)
		ingressRule := v1beta1.IngressRule{
			Host:             hostname,
			IngressRuleValue: v1beta1.IngressRuleValue{HTTP: httpIngressRuleValue},
		}
		ingressRules = append(ingressRules, ingressRule)
	}
	if cap(ingressRules) != 0 {
		ingress = &v1beta1.Ingress{
			ObjectMeta: v1.ObjectMeta{
				Name: instanceName,
				OwnerReferences: []v1.OwnerReference{
					owner,
				},
				Annotations: annotations,
			},
			Spec: v1beta1.IngressSpec{
				Rules: ingressRules,
			},
		}
	}
	return ingress
}

func getManuelScale(traits []v1alpha1.TraitBinding) *int32 {
	var def int32 = 1
	for _, tr := range traits {
		if tr.Name != "manual-scaler" {
			continue
		}
		values, err := parsePropertiesOfTrait(tr)
		if err != nil {
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

func getVolumesFromVolumeMounters(traits []v1alpha1.TraitBinding) []apiv1.Volume {
	var volumes []apiv1.Volume
	for _, tr := range traits {
		if tr.Name != "volume-mounter" {
			continue
		}
		values, err := parsePropertiesOfTrait(tr)
		if err != nil {
			continue
		}
		volume := apiv1.Volume{
			Name: values["volumeName"].(string),
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: values["volumeName"].(string),
				},
			},
		}
		volumes = append(volumes, volume)
	}

	return volumes
}

func convertPvcsFromVolumeMounters(owner v1.OwnerReference, annotations map[string]string, comp v1alpha1.ComponentSchematic, traits []v1alpha1.TraitBinding) []apiv1.PersistentVolumeClaim {
	annotations["role"] = "trait"
	var pvcs []apiv1.PersistentVolumeClaim
	for _, tr := range traits {
		if tr.Name != "volume-mounter" {
			continue
		}
		values, err := parsePropertiesOfTrait(tr)
		if err != nil {
			continue
		}
		volumeName := values["volumeName"].(string)
		storageClass := values["storageClass"].(string)

		var oamVolume v1alpha1.Volume
		if comp.Spec.WorkloadType == WorkloadTypeMysqlCluster {
			continue
		}
		for _, c := range comp.Spec.Containers {
			for _, v := range c.Resources.Volumes {
				if v.Name == volumeName {
					oamVolume = v
				}
			}
		}
		if &oamVolume == nil || oamVolume.Disk == nil {
			msg := fmt.Sprintf("Volume can not found in componentSchematic.", "ComponentSchematic", comp.Name, "volume", volumeName)
			handlerLog.Info(msg)
			continue
		}

		var required resource.Quantity
		required, _ = resource.ParseQuantity(oamVolume.Disk.Required)

		pvc := apiv1.PersistentVolumeClaim{
			TypeMeta: v1.TypeMeta{},
			ObjectMeta: v1.ObjectMeta{
				Name: values["volumeName"].(string),
				OwnerReferences: []v1.OwnerReference{
					owner,
				},
				Annotations: annotations,
			},
			Spec: apiv1.PersistentVolumeClaimSpec{
				AccessModes: []apiv1.PersistentVolumeAccessMode{
					apiv1.PersistentVolumeAccessMode(getPvcAccessMode(&oamVolume.AccessMode)),
				},
				Resources: apiv1.ResourceRequirements{
					Requests: apiv1.ResourceList{
						"storage": required,
					},
				},
				StorageClassName: &storageClass,
				VolumeMode:       nil,
			},
		}
		if !oamVolume.Disk.Ephemeral {
			pvc.OwnerReferences = append(pvc.OwnerReferences, owner)
		}
		pvcs = append(pvcs, pvc)
	}
	return pvcs
}

func getPvcAccessMode(mode *v1alpha1.AccessMode) string {
	if mode == nil {
		return "ReadWriteOnce"
	} else if *mode == v1alpha1.RW || *mode == "ReadWriteMany" {
		return "ReadWriteMany"
	} else if *mode == v1alpha1.RO || *mode == "ReadWriteOnce" {
		return "ReadOnlyMany"
	}
	return "ReadWriteOnce"
}

func parsePropertiesOfTrait(trait v1alpha1.TraitBinding) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	err := json.Unmarshal(trait.Properties.Raw, &values)
	if err != nil {
		handlerLog.Info("traits value spec error", "Error", err)
		return nil, err
	}
	return values, nil
}
