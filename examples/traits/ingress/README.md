# Ingress trait

Ingress trait is used for components with service workloads and provides load balancing, SSL termination and name-based virtual hosting.

## Installation

To successfully use an `ingress` trait, you will need to install one of the Kubernetes Ingress controllers. We recommend [nginx-ingress](https://hub.helm.sh/charts/stable/nginx-ingress):

```shell script
$ helm install nginx-ingress stable/nginx-ingress
```

*Note:* You still must manage your DNS configuration as well. Mapping an ingress to `example.com` will not work if you do not also control the domain mapping for `example.com`.

## Supported workload types

- `core.oam.dev/v1alpha1.Server`
- `core.oam.dev/v1alpha1.SingletonServer`

## Properties

| Name | Description | Allowable values | Required | Default |
| :-- | :--| :-- | :-- | :-- |
| `hostname` | Host name for the ingress. | `string` | &#9745; |
| `servicePort` | Port number on the service to bind to the ingress. | `integer`. See notes below. | &#9745; | 
| `path` | Path to expose. | `string` | | `/`|

## Usage
To find your service port, you can do one of two things:

- find the port on the [ComponentSchematic](https://github.com/oam-dev/spec/blob/master/3.component_model.md#port)
- find the port on the desired Kubernetes [Service](https://kubernetes.io/docs/concepts/services-networking/service/) object

For example, here's how to find the port on a `ComponentSchematic`:

```yaml
apiVersion: core.oam.dev/v1alpha1
kind: ComponentSchematic
metadata:
  name: nginx-replicated-v1
spec:
  workloadType: core.oam.dev/v1alpha1.Server
  containers:
  - image: nginx:latest
    name: server
    ports:
      - containerPort: 80                  # <-- this is the service port
        name: http
        protocol: TCP
```

So to use this on an ingress, you would need to add this to your `ApplicationConfiguration`:

```yaml
apiVersion: core.oam.dev/v1alpha1
kind: ApplicationConfiguration
metadata:
  name: ingress-example
spec:
  components:
    - componentName: nginx-replicated-v1
      instanceName: ingress-demo
      traits:
        - name: ingress
          properties:
            hostname: example.com
            path: /
            servicePort: 80            # <-- set this to the value in the component
```

Because each component may have multiple ports, the specific port must be defined in the `ApplicationConfiguration`.

## Example
```shell script
$ kubectl apply -f component-schematics.yaml 
componentschematic.core.oam.dev/nginx-replicated-v1 created
$ kubectl apply -f application-configurations.yaml 
applicationconfiguration.core.oam.dev/ingress-example created
$
$ kubectl get deploy,svc,ing
NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/ingress-demo   1/1     1            1           12m

NAME                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
service/ingress-demo   ClusterIP   10.245.96.196   <none>        80/TCP    12m

NAME                              HOSTS         ADDRESS   PORTS   AGE
ingress.extensions/ingress-demo   example.com             80      97s
$
$ kubectl -n kube-system get pod -owide | grep nginx-ingress-controller
nginx-ingress-controller-dr7nz                1/1     Running   0          14m    10.10.102.34     10.10.102.34-slave    <none>           <none>
$
$ echo '10.10.102.34 example.com' >> /etc/hosts
$
$ curl example.com
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
    body {
        width: 35em;
        margin: 0 auto;
        font-family: Tahoma, Verdana, Arial, sans-serif;
    }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
```
