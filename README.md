appservice-operator
====================

### 场景

当部署一个应用到 Kubernetes 集群中的时候，每次都需要先编写一个 `Deployment` 对象，然后再创建一个 `Service` 对象，通过 Pod 的 label 标签进行关联，设置 `type: NodePort` 来暴露应用服务，每次都需要这样操作，繁琐.


创建一个自定义资源对象 `AppService`，来描述要部署的应用信息.  
每当创建 `AppService` 对象的时候，通过控制器去自动创建对应的 `Deployment` 和 `Service` 对象.


### 如何使用
