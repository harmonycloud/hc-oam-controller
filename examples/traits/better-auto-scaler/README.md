# Better autoscaler trait

Better autoscaler trait is used to autoscale components with replicable workloads. This is implemented by the HarmonyCloud and it is an enhanced version of hpa.

## Installation

To use the better-auto-scaler trait, you must install the `better-autoscaler-controller`.

## Supported workload types

- `core.oam.dev/v1alpha1.Server`
- `core.oam.dev/v1alpha1.Task`
- `core.oam.dev/v1alpha1.Worker`
- `harmonycloud.cn/v1alpha1.MysqlCluster`

## Properties

| Name | Description | Allowable values | Required | Default |
| :-- | :--| :-- | :-- | :-- |
| `minimum` | Lower threshold of replicas to run. | `integer` | | `1`
| `maximum` | Higher threshold of replicas to run.  | `integer`. Cannot be less than `minimum` value. | | `10`
| `memory-up` | Memory consumption threshold (as percent) that will cause a scale-up event. | `integer` ||
| `memory-down` | Memory consumption threshold (as percent) that will cause a scale-down event. | `integer` ||
| `cpu-up` | CPU consumption threshold (as percent) that will cause a scale-up event. | `integer` ||
| `cpu-down` | CPU consumption threshold (as percent) that will cause a scale-down event. | `integer` ||

## Usage
This is usage of an better-auto-scaler trait. You would attach this to a component within the application configuration:

```yaml
# Usage better-auto-scaler trait entry
- name: better-auto-scaler
  properties:
    maximum : 6
    minimum : 2
    cpu-up : 50
    cpu-down : 20
    memory-up : 50
    memory-down: 20
```

## Example
```shell script
$ kubectl apply -f component-schematics.yaml 
componentschematic.core.oam.dev/hc-hpa-example created
$ kubectl apply -f application-configurations.yaml 
applicationconfiguration.core.oam.dev/better-autoscaler-example created
$
$ kubectl get deploy,hchpa,pod
NAME                                     READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/better-autoscaled-demo   2/2     2            2           2m50s

NAME                                                             AGE
horizontalpodautoscaler.harmonycloud.cn/better-autoscaled-demo   2m49s

NAME                                          READY   STATUS    RESTARTS   AGE
pod/better-autoscaled-demo-595dbb46cf-29mzt   1/1     Running   0          2m51s
pod/better-autoscaled-demo-595dbb46cf-jpdj8   1/1     Running   0          2m35s
```
