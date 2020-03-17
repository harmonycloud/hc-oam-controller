# MysqlCluster workload

MysqlCluster workload is used to auto create and manage mysql cluster in kubernetes. This is implemented by the HarmonyCloud.

## Installation

To use the MysqlCluster workload, you must install `mysql-operator` in your Kubernetes cluster.

## Workload Settings

| Name | Description | Allowable values | Required | Default |
| :-- | :--| :-- | :-- | :-- |
| `spec` | `Spec` of the custom resource `MysqlCluster` | `object` | &#9745; | 
| `config` | The config information of `MysqlCluster`   | `object` | &#9745; | 

## Example
```shell script
$ kubectl apply -f component-schematics.yaml 
componentschematic.core.oam.dev/mysql-cluster-demo created
$ kubectl apply -f application-configurations.yaml 
applicationconfiguration.core.oam.dev/mysql-app created
$ kubectl get mysqlcluster,sts,po,svc,cm,pvc 
NAME                                                                  AGE
mysqlcluster.mysql.middleware.harmonycloud.cn/mysql-cluster-example   6m8s

NAME                                     READY   AGE
statefulset.apps/mysql-cluster-example   2/2     6m8s

NAME                          READY   STATUS    RESTARTS   AGE
pod/mysql-cluster-example-0   2/2     Running   0          6m8s
pod/mysql-cluster-example-1   2/2     Running   0          4m43s

NAME                                     TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)     AGE
service/mysql-cluster-example            ClusterIP   None         <none>        20001/TCP   6m9s
service/mysql-cluster-example-readonly   ClusterIP   None         <none>        20001/TCP   6m9s

NAME                                  DATA   AGE
configmap/mysql-cluster-demo-config   1      6m9s

NAME                                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/mysql-cluster-example   Bound    pvc-1edc7f3f-ec69-4fa5-bed9-cf0b398ee9a9   1G         RWX            local-storage  6m9s

```
