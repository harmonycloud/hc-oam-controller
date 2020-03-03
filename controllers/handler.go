package controllers

import (
	"github.com/oam-dev/oam-go-sdk/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

type Handler struct {
	Name      string
	Oamclient *versioned.Clientset
	K8sclient *kubernetes.Clientset
}
