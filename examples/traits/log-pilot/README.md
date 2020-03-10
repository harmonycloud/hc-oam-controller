# Ingress trait

Ingress trait is used for components with service workloads and provides load balancing, SSL termination and name-based virtual hosting.

## Installation

To successfully use an `log-pilot` trait, you will need to install [AliyunContainerService/log-pilot](https://github.com/AliyunContainerService/log-pilot) in your kubernetes cluster. 

```shell script
$ kubectl apply -f https://raw.githubusercontent.com/AliyunContainerService/log-pilot/master/examples/pilot-elastisearch-kubernetes-2.yml 
daemonset.extensions/log-pilot created
```

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
| `container` | The name of the container. This matches the container name declared in the ComponentSchematic. | `string` | &#9745; |
| `path` | path is the log file path, should end with / | `string` | &#9745; | 
| `name` | name is an identify, can be any string you want. The valid characters in name are 0-9a-zA-Z_- | `string` | &#9745;| |
| `tags` | tags will be appended to log. (e.g. k1=v1,k2=v2) | `string` | | |
| `pilotLogPrefix` | pilotLogPrefix is an attribute named PILOT_LOG_PREFIX declared in log-pilot configuration. | `string` | | `aliyun` |

## Usage
This is usage of an log-pilot trait. You would attach this to a component within the application configuration:

```yaml
# Usage logs-pilot trait entry
- name: log-pilot
  properties:
    container: tomcat
    path: /usr/local/tomcat/logs/
    name: access
    tags: k1=v1,k2=v2
```

## Example
```shell script
$ kubectl apply -f component-schematics.yaml 
componentschematic.core.oam.dev/tomcat created
$ kubectl apply -f application-configurations.yaml 
applicationconfiguration.core.oam.dev/logpilot-example created
$ kubectl get deploy
NAME            READY   UP-TO-DATE   AVAILABLE   AGE
logpilot-demo   1/1     1            1           20s
$ kubectl get deploy logpilot-demo -oyaml
...
    spec:
      containers:
      - env:
        - name: aliyun_logs_access
          value: /usr/local/tomcat/logs/*
        - name: aliyun_logs_access_tags
          value: k1=v1,k2=v2
        image: tomcat:8.0
        name: tomcat
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        volumeMounts:
        - mountPath: /usr/local/tomcat/logs/
          name: tomcat-log
      volumes:
      - emptyDir: {}
        name: tomcat-log
...
```
