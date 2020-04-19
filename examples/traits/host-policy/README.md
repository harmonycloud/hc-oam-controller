# Host Policy trait

The host policy Trait is used for components to use host namespace..

## Installation

None. *The host policy trait has no external dependencies.*

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
| `hostNetwork` | hostNetwork switch. | boolean | `false` |
| `hostPid` | hostPid switch. | boolean | `false` |
| `hostIpc` | hostIpc switch. | boolean | `false` |

## Usage
This is usage of how to use the host policy trait:

```yaml
apiVersion: core.oam.dev/v1alpha1
kind: ApplicationConfiguration
metadata:
  name: host-example
spec:
  components:
    - componentName: nginx-replicated
      instanceName: host-demo
      traits:
        - name: host-policy
          properties:
            hostNetwork: true
            hostPid: true
            hostIpc: true
```

## Example
```shell script
$ host-policy % kubectl apply -f component-schematics.yaml 
componentschematic.core.oam.dev/nginx-replicated created
$ host-policy % kubectl apply -f application-configurations.yaml 
applicationconfiguration.core.oam.dev/host-example created
$ host-policy % kubectl get deploy host-demo -oyaml
apiVersion: apps/v1
kind: Deployment
...
spec:
  ...
    spec:
      containers:
      - image: nginx:latest
        imagePullPolicy: Always
        name: server
        ports:
        ...
      hostIPC: true
      hostNetwork: true
      hostPID: true
...
```
