package controllers

import (
	"fmt"
	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	"strings"
)

func parseParameters(parameterValues []v1alpha1.ParameterValue, variables []v1alpha1.Variable) map[string]string {
	parameterMap := map[string]string{}
	variablesMap := parseVariables(variables)
	for _, p := range parameterValues {
		if strings.HasPrefix(p.Value, "[fromVariable(") && strings.HasSuffix(p.Value, ")]") {
			key := MiddleString(p.Value, "[fromVariable(", ")]")
			p.Value = variablesMap[key]
		}
		parameterMap[p.Name] = p.Value
	}
	return parameterMap
}

func parseVariables(variables []v1alpha1.Variable) map[string]string {
	variablesMap := map[string]string{}
	for _, v := range variables {
		variablesMap[v.Name] = v.Value
	}
	return variablesMap
}

//get the middle string
func MiddleString(str, starting, ending string) string {
	s := strings.Index(str, starting)
	if s < 0 {
		return ""
	}
	s += len(starting)
	e := strings.Index(str[s:], ending)
	if e < 0 {
		return ""
	}
	return str[s : s+e]
}

func addResourceStatus(statusList *[]v1alpha1.ResourceStatus, name string, apiVersion, kind string, component string, role string, status string) {
	resourceStatus := v1alpha1.ResourceStatus{
		NamespacedName: name,
		ApiVersion:     apiVersion,
		Kind:           kind,
		Component:      component,
		Role:           role,
		Status:         status,
	}
	if statusList == nil {
		statusList = new([]v1alpha1.ResourceStatus)
	}

	flag := false
	for i, s := range *statusList {
		if s.ApiVersion == s.ApiVersion && s.Kind == kind && s.NamespacedName == name && s.Component == component && s.Role == role {
			(*statusList)[i] = resourceStatus
			flag = true
			break
		}
	}
	if !flag {
		*statusList = append(*statusList, resourceStatus)
		fmt.Println("tmp")
	}
}

func addModuleStatus(statusList *[]v1alpha1.ModuleStatus, name string, kind string, groupVersion, status string) {
	moduleStatus := v1alpha1.ModuleStatus{
		NamespacedName: name,
		Kind:           kind,
		GroupVersion:   groupVersion,
		Status:         status,
	}
	if statusList == nil {
		statusList = new([]v1alpha1.ModuleStatus)
		//(*statusList)[0] = moduleStatus
	}

	for i, s := range *statusList {
		if s.Kind == kind && s.NamespacedName == name && s.GroupVersion == groupVersion {
			(*statusList)[i] = moduleStatus
			return
		}
	}
	*statusList = append(*statusList, moduleStatus)
	return
}
