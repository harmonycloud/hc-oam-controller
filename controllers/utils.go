package controllers

import (
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

func addModuleStatus(statusList *[]v1alpha1.ModuleStatus, name string, kind string, group string, status string) {
	moduleStatus := v1alpha1.ModuleStatus{
		NamespacedName: name,
		Kind:           kind,
		GroupVersion:   group,
		Status:         status,
	}
	flag := false
	for _, s := range *statusList {
		if s.Kind == kind && s.NamespacedName == name && s.GroupVersion == group {
			flag = true
			break
		}
	}
	if !flag {
		*statusList = append(*statusList, moduleStatus)
	}
}
