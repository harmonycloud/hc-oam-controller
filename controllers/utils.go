package controllers

import "github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"

func parseParameters(parameterValues []v1alpha1.ParameterValue) map[string]string {
	parameterMap := map[string]string{}
	for _, p := range parameterValues {
		parameterMap[p.Name] = p.Value
	}
	return parameterMap
}
