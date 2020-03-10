# hc-oam-controller

hc-oam-controller is an [Open Application Model (OAM)](https://github.com/oam-dev/spec) implementation based on [oam-dev/oam-go-sdk](https://github.com/oam-dev/oam-go-sdk) for Kubernetes in [HarmonyCloud](http://harmonycloud.cn/).

## Workloads

The *workload type* (`workloadType`) is a field in the `ComponentSchematic` used by the developer to direct the runtime to properly execute the given component. 

OAM have two kinds of [workload types](https://github.com/oam-dev/spec/blob/master/3.component_model.md#workload-types).

* Core workload types
* Extended workload types

### Core Workloads

Hc-oam-controller supports all of the Open Application Model [Core Workload Types](https://github.com/oam-dev/spec/blob/master/3.component_model.md#core-workload-types):

|Name|Type|Service endpoint|Replicable|Daemonized|
|-|-|-|-|-|
|Server|core.oam.dev/v1alpha1.Server|Yes|Yes|Yes
|Singleton Server|core.oam.dev/v1alpha1.SingletonServer|Yes|No|Yes
|Task|core.oam.dev/v1alpha1.Task|No|Yes|No
|Singleton Task|core.oam.dev/v1alpha1.SingletonTask|No|No|No
|Worker|core.oam.dev/v1alpha1.Worker|No|Yes|Yes
|Singleton Worker|core.oam.dev/v1alpha1.SingletonWorker|No|No|Yes

## Traits

A [trait](https://github.com/oam-dev/spec/blob/master/5.traits.md) represents a piece of add-on functionality that attaches to a component instance. Traits augment components with additional operational features such as traffic routing rules (including load balancing policy, network ingress routing, circuit breaking, rate limiting), auto-scaling policies, upgrade strategies, and more. As such, traits represent features of the system that are operational concerns, as opposed to developer concerns.               

Currently, hc-oam-controller supports the following traits:

- [Manual Scaler](examples/traits/manual-scaler/README.md)
- [Autoscaler](examples/traits/auto-scaler/README.md)
- [Ingress](examples/traits/ingress/README.md)
- [Volume Mounter](examples/traits/volume-mounter/README.md)
- [Log-pilot](examples/traits/log-pilot/README.md)

## Get started

```shell script
$ kubectl apply -f config/oam/crds/
customresourcedefinition.apiextensions.k8s.io/applicationconfigurations.core.oam.dev created
customresourcedefinition.apiextensions.k8s.io/applicationscopes.core.oam.dev created
customresourcedefinition.apiextensions.k8s.io/componentschematics.core.oam.dev created
customresourcedefinition.apiextensions.k8s.io/traits.core.oam.dev created
customresourcedefinition.apiextensions.k8s.io/workloadtypes.core.oam.dev created
$ kubectl apply -f config/hc-oam-controller/
namespace/oam-system created
deployment.apps/hc-oam-controller created
clusterrole.rbac.authorization.k8s.io/hc-oam-controller-role created
clusterrolebinding.rbac.authorization.k8s.io/hc-oam-controller-rolebinding created
$ kubectl -n oam-system get pod
NAME                                 READY   STATUS    RESTARTS   AGE
hc-oam-controller-666457bc6f-hsthf   1/1     Running   0          3m
```

## Examples

This is a simple example of how to use the hc-oam-controller.

````shell script
$ kubectl apply -f examples/samples/simple-example 
applicationconfiguration.core.oam.dev/simple-app created
componentschematic.core.oam.dev/stateless-component created
$
$ kubectl get deploy,svc,pod
NAME                         READY   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/demo   2/2     2            2           48s

NAME           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
service/demo   ClusterIP   10.245.101.21   <none>        80/TCP    48s

NAME                        READY   STATUS    RESTARTS   AGE
pod/demo-575577d55f-48lz6   1/1     Running   0          48s
pod/demo-575577d55f-5p5dc   1/1     Running   0          48s
$
$ kubectl delete -f examples/samples/simple-example 
applicationconfiguration.core.oam.dev "simple-app" deleted
componentschematic.core.oam.dev "stateless-component" deleted
$ kubectl get deploy,pod,svc                       
No resources found.
````

There are more examples in [examples/](examples/README.md).
