# 1. in-cluster-configuration
```
# Create ClusterRoleBinding
kubectl create clusterrolebinding default-view --clusterrole=view --serviceaccount=default:default
```

- 构建镜像并运行
```
cat Dockerfile
FROM busybox

COPY ./in-cluster /in-cluster
USER 65532:65532
CMD ["/bin/sh","-c","/app"]
```

```
GOOS=linux go build -o ./in-cluster
docker build -t geray/in-cluster:v1 .
docker push geray/in-cluster:v1

kubectl run -i in-cluster --image=geray/in-cluster:v1
```

# 2. VMware挂载windows目录
1. 编辑虚拟机设置》选项》共享文件夹》总是启用》选择目录并设置名称
2. 启动虚拟机》`vmware-hgfsclient` 命令查看那改在的名称（例如：k8soperator）
```
[root@master1 ~]# vmware-hgfsclient
k8soperator
```
3. 挂载
```
mkdir -p /data/k8soperator
# 注意：这样虽然映射成功，并且，使用 ll /www 查看里面的内容，权限都是最高权限，但是，其他用户却还是无法访问的
vmhgfs-fuse .host:/k8soperator /data/k8soperator

# 普通用户权限挂载方式
vmhgfs-fuse .host:/k8soperator /data/k8soperator -o subtype=vmhgfs-fuse,allow_other,nonempty

```
    1. k8soperator 是设置的共享目录名称
    2.`/data/k8soperator` ：是挂载路径，没有需要提前创建
4. 开机挂载
```
vim  /etc/rc.d/rc.local
vmhgfs-fuse .host:/k8soperator /data/k8soperator -o subtype=vmhgfs-fuse,allow_other,nonempty

chmod +x /etc/rc.d/rc.local
```

## 卸载
```
[root@master1 ~]# df -h | grep k8soperator
vmhgfs-fuse                                                                                                586G  469G  118G   80% /data/k8soperator
[root@master1 ~]# umount /data/k8soperator
[root@master1 ~]# df -h | grep k8soperator
[root@master1 ~]# 
```

