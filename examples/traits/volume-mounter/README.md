# Volume Mounter trait

The volume mounter trait is responsible for attaching a Kubernetes [PersistentVolume Claim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims) (PVC) to a component.

## Installation

None. *The volume mounter trait has no external dependencies.*

## Supported workload types

- `core.oam.dev/v1alpha1.Server`
- `core.oam.dev/v1alpha1.SingletonServer`
- `core.oam.dev/v1alpha1.Task`
- `core.oam.dev/v1alpha1.SingletonTask`
- `core.oam.dev/v1alpha1.Worker`
- `core.oam.dev/v1alpha1.SingletonWorker`
- `harmonycloud.cn/v1alpha1.MysqlCluster`

## Properties

| Name | Description | Allowable values | Required | Default |
| :-- | :--| :-- | :-- | :-- |
| `volumeName` | The name of the volume this backs. | string. Matches the [volume](https://github.com/oam-dev/spec/blob/master/3.component_model.md#volume) name declared in ComponentSchematic. | &#9745; |
| `storageClass` | The storage class that a PVC requires. | string. According to the available StorageClasses(s) (`kubectl get storageclass`) in your cluster and/or `default` | &#9745; |

## Usage
This is usage of how to attach a storage volume to your container:

```yaml
apiVersion: core.oam.dev/v1alpha1
kind: ComponentSchematic
metadata:
  name: server-with-volume
spec:
  workloadType: core.oam.dev/v1alpha1.Server
  containers:
    - name: server
      image: nginx:latest
      resources:
        volumes:
          - name: myvol
            mountPath: /myvol
            disk:
              required: "50M"
              ephemeral: true
```

In the component schematic [`volumes`](https://github.com/oam-dev/spec/blob/master/3.component_model.md#volume) section, one volume is specified. It must be at least `50M` in size. It is `ephemeral`, which means that the component author does not expect the data to persist if the pod is destroyed.

Sometimes, components need to persist data. In such cases, the `ephemeral` flag should be set to `false`:

```yaml
apiVersion: core.oam.dev/v1alpha1
kind: ComponentSchematic
metadata:
  name: server-with-volume
spec:
  workloadType: core.oam.dev/v1alpha1.Server
  containers:
    - name: server
      image: nginx:latest
      resources:
        volumes:
          - name: myvol
            mountPath: /myvol
            disk:
              required: "50M"
              ephemeral: false
```

In the Kubernetes implementation of OAM, a Persistent Volume Claim (PVC) is used to satisfy the non-ephemeral case. However. A trait must be applied that will indicate how the PVC is created:

```yaml
apiVersion: core.oam.dev/v1alpha1
kind: ApplicationConfiguration
metadata:
  name: example-server-with-volume
spec:
  components:
    - componentName: server-with-volume-v1
      instanceName: example-server-with-volume
      traits:
        - name: volume-mounter
          properties:
            volumename: myvol
            storageClass: default
```

The `volume-mounter` trait ensures that a PVC is created with the given name (`myvol`) using the given storage class (`default`). Typically, the `volumeName` should match the `resources.volumes[].name` field from the `ComponentSchematic`. Thus `myvol` above will match the volume declared in the `volumes` section of `server-with-volume-v1`.

When this request is processed by oam-controller, it will first create the Kubernetes PVC named `myvol` and then create a Kubernetes pod that attaches that PVC as a `volumeMount`.

Attaching PVCs to Pods _may take extra time_, as the underlying system must first provision storage.

## Example
```shell script
$ kubectl get storageclass
NAME                       PROVISIONER                                       AGE
default                    nfs-client-provisioner-default                    124d
$
$ kubectl apply -f component-schematics.yaml 
componentschematic.core.oam.dev/server-with-volume created
$ kubectl apply -f application-configurations.yaml 
applicationconfiguration.core.oam.dev/example-server-with-volume created
$
$ kubectl get pvc,deploy,pod
NAME                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/myvol   Bound    pvc-0966159f-6202-11ea-81d4-005056b71aad   50M        RWO            default        55s

NAME                                               DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/example-server-with-volume   1         1         1            1           55s

NAME                                              READY   STATUS   RESTARTS   AGE
pod/example-server-with-volume-57c488c79b-259rb   1/1     Running    0          55s
$ kubectl get pv pvc-0966159f-6202-11ea-81d4-005056b71aad
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM            STORAGECLASS   REASON   AGE
pvc-0966159f-6202-11ea-81d4-005056b71aad   50M        RWO            Delete           Bound    oam-test/myvol   default                 55s
```
