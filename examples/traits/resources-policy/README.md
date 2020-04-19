# Resources Policy trait

The resources policy trait is used for components to limit container's resources.

## Installation

None. *The resources policy trait has no external dependencies.*

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
| `container` | The container name. | string. Matches the container name declared in ComponentSchematic. | &#9745; |
| `limits` | Limits describes the maximum amount of compute resources allowed. | map | &#9745; |

## Usage
This is usage of how to use the resources policy:

```yaml
apiVersion: core.oam.dev/v1alpha1
kind: ApplicationConfiguration
metadata:
  name: resource-example
spec:
  components:
    - componentName: nginx-replicated
      instanceName: resource-demo
      traits:
        - name: resources-policy
          properties:
            container: server
            limits:
              cpu: 200m
              memory: 256Mi
```

## Example
```shell script
$ resourecs-policy % kubectl create -f component-schematics.yaml 
componentschematic.core.oam.dev/nginx-replicated created
$ resourecs-policy % kubectl create -f application-configurations.yaml 
applicationconfiguration.core.oam.dev/resource-example created
$ resourecs-policy % kubectl get deploy resource-demo -oyaml
apiVersion: apps/v1
kind: Deployment
...
spec:
  ...
  template:
    spec:
      containers:
      - image: nginx:latest
        imagePullPolicy: Always
        name: server
        ports:
        - containerPort: 80
          name: http
          protocol: TCP
        resources:
          limits:
            cpu: 200m
            memory: 256Mi
          requests:
            cpu: 100m
            memory: 128Mi
...
```
