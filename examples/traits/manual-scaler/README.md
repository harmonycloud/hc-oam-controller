# Manual Scaler trait

Manual Scaler trait is used to manually scale components with replicable workload types.

## Installation

None. *The manual scaler trait has no external dependencies.*

## Supported workload types

- `core.oam.dev/v1alpha1.Server`
- `core.oam.dev/v1alpha1.Task`
- `core.oam.dev/v1alpha1.Worker`

## Properties

| Name | Description | Allowable values | Required | Default |
| :-- | :--| :-- | :-- | :-- |
| `replicaCount` | The target number of replicas to create for a given component instance. | `integer` | &#9745; | |

## Usage
This is usage of an autoscaler trait. You would attach this to a component within the application configuration.       
The following snippet from an application configuration shows how the manual scaler trait is applied and configured for a component. 

```yaml
# Usage manual scaler trait entry
traits:
  - name: manual-scaler
    properties:
      replicaCount: 3
```
This example illustrates using the manual scaler trait to manually scale a component by specifying the number of replicas the runtime should create. This trait has only one attribute in its properties: `replicaCount`. This takes an integer, and a value is required for this trait to be successfully applied. If, for example, an application configuration used this trait but did not provide the `replicaCount`, the system would reject the application configuration.

## Example
```shell script
$ kubectl apply -f component-schematics.yaml 
componentschematic.core.oam.dev/nginx-component created
$ kubectl apply -f application-configurations.yaml 
applicationconfiguration.core.oam.dev/manual-scaler-example created
$
$ kubectl get deploy,pod
NAME                                       READY   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/manual-scaler-demo   3/3     3            3           34s

NAME                                      READY   STATUS              RESTARTS   AGE
pod/manual-scaler-demo-8587fdb6ff-8d9pw   1/1     Running             0          35s
pod/manual-scaler-demo-8587fdb6ff-g4rgs   1/1     Running             0          35s
pod/manual-scaler-demo-8587fdb6ff-qrdfr   1/1     Running             0          35s
```
