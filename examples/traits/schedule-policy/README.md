# Schedule Policy  trait

The schedule policy trait is used to schedule instance's pods to expect nodes.

## Installation

None. *The schedule policy trait has no external dependencies.*

## Supported workload types

- `core.oam.dev/v1alpha1.Server`
- `core.oam.dev/v1alpha1.SingletonServer`
- `core.oam.dev/v1alpha1.Task`
- `core.oam.dev/v1alpha1.SingletonTask`
- `core.oam.dev/v1alpha1.Worker`
- `core.oam.dev/v1alpha1.SingletonWorker`

## Properties

| Name | Description | Allowable values | Required | Default |
| :-- | :--| :-- | :-- | :-- |
| `nodeAffinity` | Describes node affinity scheduling rules for the instance's pod. | object | |
| `podAffinity` | Describes pod affinity scheduling rules. | object | |
| `podAntiAffinity` | Describes pod anti-affinity scheduling rules. | |

## Usage
This is usage of how to use the schedule policy trait:

```yaml
apiVersion: core.oam.dev/v1alpha1
kind: ApplicationConfiguration
metadata:
  name: schedule-example
spec:
  components:
    - componentName: nginx-replicated
      instanceName: schedule-demo
      traits:
        - name: schedule-policy
          properties:
            nodeAffinity:
              type: required
              selector:
                k-node-1: v-node-1
                k-node-2: v-node-2
            podAffinity:
              type: required
              selector:
                k-pod-1: v-pod-1
                k-pod-2: v-pod-2
            podAntiAffinity:
              type: required
              selector:
                k-pod-anti-1: v-pod-anti-1
                k-pod-anti-2: v-pod-anti-2
```

## Example
```shell script
chenbilong@chenbilongdeMBP schedule-policy % kubectl create -f component-schematics.yaml 
componentschematic.core.oam.dev/nginx-replicated created
chenbilong@chenbilongdeMBP schedule-policy % kubectl create -f application-configurations.yaml 
applicationconfiguration.core.oam.dev/schedule-example created
chenbilong@chenbilongdeMBP schedule-policy % kubectl get deploy schedule-demo -oyaml
apiVersion: apps/v1
kind: Deployment
metadata:
...
  name: schedule-demo
...
spec:
...
  template:
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: k-node-1
                operator: In
                values:
                - v-node-1
              - key: k-node-2
                operator: In
                values:
                - v-node-2
        podAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: k-pod-1
                operator: In
                values:
                - v-pod-1
              - key: k-pod-2
                operator: In
                values:
                - v-pod-2
            namespaces:
            - default
            topologyKey: kubernetes.io/hostname
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: k-pod-anti-1
                operator: In
                values:
                - v-pod-anti-1
              - key: k-pod-anti-2
                operator: In
                values:
                - v-pod-anti-2
            namespaces:
            - default
            topologyKey: kubernetes.io/hostname
...

```
