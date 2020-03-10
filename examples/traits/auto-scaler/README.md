# Autoscaler trait

Autoscaler trait is used to autoscale components with replicable workloads. This is implemented by the Kubernetes [Horizontal Pod Autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/).

## Installation

To use the autoscaler trait, you must install a controller for Kubernetes `HorizontalPodAutoscaler`.

## Supported workload types

- `core.oam.dev/v1alpha1.Server`
- `core.oam.dev/v1alpha1.Task`
- `core.oam.dev/v1alpha1.Worker`

## Properties

| Name | Description | Allowable values | Required | Default |
| :-- | :--| :-- | :-- | :-- |
| `minimum` | Lower threshold of replicas to run. | `integer` | | `1`
| `maximum` | Higher threshold of replicas to run.  | `integer`. Cannot be less than `minimum` value. | | `10`
| `memory` | Memory consumption threshold (as percent) that will cause a scale event. | `integer` ||
| `cpu` | CPU consumption threshold (as percent) that will cause a scale event. | `integer` ||

## Usage
This is usage of an autoscaler trait. You would attach this to a component within the application configuration:

```yaml
# Usage autoscaler trait entry
- name: auto-scaler
  properties:
    maximum: 6
    minimim: 2
    cpu: 50
    memory: 50
```

## Example
```shell script
$ kubectl apply -f component-schematics.yaml 
componentschematic.core.oam.dev/hpa-example created
$ kubectl apply -f application-configurations.yaml 
applicationconfiguration.core.oam.dev/autoscaler-example created
$
$ kubectl get deploy,hpa,pod
NAME                                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/autoscaled-demo   2/2     2            2           9m17s

NAME                                                  REFERENCE                    TARGETS          MINPODS   MAXPODS   REPLICAS   AGE
horizontalpodautoscaler.autoscaling/autoscaled-demo   Deployment/autoscaled-demo   3%/50%, 1%/50%   2         6         2          9m16s

NAME                                   READY   STATUS    RESTARTS   AGE
pod/autoscaled-demo-55d9d74886-mzxnh   1/1     Running   0          9m1s
pod/autoscaled-demo-55d9d74886-rjmgc   1/1     Running   0          9m17s
```
