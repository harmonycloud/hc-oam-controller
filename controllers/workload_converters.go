package controllers

import (
	"encoding/json"
	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	hcv1alpha1 "hc-oam-controller/api/harmonycloud.cn/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func convertMysqlCluster(owner v1.OwnerReference, compConf v1alpha1.ComponentConfiguration, comp v1alpha1.ComponentSchematic, parameterMap map[string]string) (*hcv1alpha1.MysqlCluster, *corev1.ConfigMap, *corev1.PersistentVolumeClaim, error) {
	type value struct {
		Name        string               `json:"name"`
		Description string               `json:"description,omitempty"`
		Type        string               `json:"type"`
		Required    bool                 `json:"required,omitempty"`
		Default     string               `json:"default,omitempty"`
		Value       runtime.RawExtension `json:"value,omitempty"`
		FromParam   string               `json:"fromParam,omitempty"`
	}

	values := new([]value)

	if err := json.Unmarshal(comp.Spec.WorkloadSettings.Raw, &values); err != nil {
		handlerLog.Info("workloadsetting value spec error", "Error", err)
		return nil, nil, nil, err
	}

	var mysqlClusterSpec *hcv1alpha1.MysqlClusterSpec
	var configContent string
	for _, v := range *values {
		if v.Name == "spec" {
			mysqlClusterSpec = new(hcv1alpha1.MysqlClusterSpec)
			if err := json.Unmarshal(v.Value.Raw, &mysqlClusterSpec); err != nil {
				handlerLog.Info("MysqlCluster value spec error", "Error", err)
				return nil, nil, nil, err
			}
		}
		if v.Name == "config" {
			for _, p := range comp.Spec.Parameters {
				if p.Name == v.FromParam {
					configContent = p.Default
					break
				}
			}
		}
	}

	mysqlCluster := &hcv1alpha1.MysqlCluster{
		ObjectMeta: v1.ObjectMeta{
			Name: compConf.InstanceName,
			OwnerReferences: []v1.OwnerReference{
				owner,
			},
			Labels: map[string]string{
				"operatorname": "mysql-operator",
			},
		},
		Spec: *mysqlClusterSpec,
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name: mysqlClusterSpec.CmName,
			OwnerReferences: []v1.OwnerReference{
				owner,
			},
		},
		Data: map[string]string{
			"my.cnf.tmpl": configContent,
		},
	}

	var storageClass string
	for _, tr := range compConf.Traits {
		if tr.Name != "volume-mounter" {
			continue
		}
		vs, err := parsePropertiesOfTrait(tr)
		if err != nil {
			continue
		}
		storageClass = vs["storageClass"].(string)
	}

	var required resource.Quantity
	required, _ = resource.ParseQuantity(mysqlClusterSpec.VolumeQuota + "G")

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			//Name: mysqlClusterSpec.PvcName,
			Name: mysqlCluster.Name,
			OwnerReferences: []v1.OwnerReference{
				owner,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			Selector: nil,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": required,
				},
			},
			StorageClassName: &storageClass,
		},
	}

	return mysqlCluster, configMap, pvc, nil
}
