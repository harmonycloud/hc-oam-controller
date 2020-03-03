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
## Get started
```shell script
[root@10 hc-oam-controller]# kubectl apply -f config/oam/crds/
customresourcedefinition.apiextensions.k8s.io/applicationconfigurations.core.oam.dev created
customresourcedefinition.apiextensions.k8s.io/applicationscopes.core.oam.dev created
customresourcedefinition.apiextensions.k8s.io/componentschematics.core.oam.dev created
customresourcedefinition.apiextensions.k8s.io/traits.core.oam.dev created
customresourcedefinition.apiextensions.k8s.io/workloadtypes.core.oam.dev created
[root@10 hc-oam-controller]# kubectl apply -f config/hc-oam-controller/
namespace/oam-system created
deployment.apps/hc-oam-controller created
clusterrole.rbac.authorization.k8s.io/hc-oam-controller-role created
clusterrolebinding.rbac.authorization.k8s.io/hc-oam-controller-rolebinding created
[root@10 hc-oam-controller]# kubectl -n oam-system get pods
NAME                                 READY   STATUS    RESTARTS   AGE
hc-oam-controller-666457bc6f-hsthf   1/1     Running   0          3m

```
## Examples
There are some simple examples of how to use the hc-oam-controller in [examples](examples/README.md).
