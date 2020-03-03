
# Examples
This is an example of how to use the hc-oam-controller.
```shell script
[root@10 examples]# kubectl create ns oam-test
namespace/oam-test created
[root@10 examples]# kubectl -n oam-test apply -f samples/
applicationconfiguration.core.oam.dev/simple-app created
componentschematic.core.oam.dev/stateless-component created
[root@10 examples]# kubectl -n oam-test get deploy
NAME   READY   UP-TO-DATE   AVAILABLE   AGE
demo   2/2     2            2           39s
[root@10 examples]# kubectl -n oam-test get svc
NAME   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
demo   ClusterIP   10.245.54.200   <none>        80/TCP    44s
[root@10 examples]# kubectl -n oam-test get pods
NAME                    READY   STATUS    RESTARTS   AGE
demo-575577d55f-5d254   1/1     Running   0          50s
demo-575577d55f-8brrz   1/1     Running   0          50s
[root@10 examples]# kubectl -n oam-test delete -f samples/
applicationconfiguration.core.oam.dev "simple-app" deleted
componentschematic.core.oam.dev "stateless-component" deleted
[root@10 examples]# kubectl -n oam-test get deploy
No resources found.
[root@10 examples]# kubectl -n oam-test get svc
No resources found.
[root@10 examples]# kubectl -n oam-test get pods
No resources found.
```