
# 1.初始化项目
```
mkdir application-operator && cd application-operator 
kubebuilder init --domain geray.cn --owner "Geray" --repo github.com/geray/application-operator
kubebuilder create api --group apps --version v1 --kind Application
kubebuilder create webhook --group apps --version v1 --kind Application  --defaulting --programmatic-validation
```

# 2.生成证书
## 1.生成自签名的根证书和私钥
```
openssl genrsa -out ca.key 2048  
openssl req -x509 -new -nodes -key ca.key -sha256 -days 365 -out ca.crt
```

## 2.创建Webhook服务的证书签名请求（CSR）
```
openssl genrsa -out webhook-service.key 2048  

cat << EOF > csr.conf
[ req ]  
default_bits = 2048  
prompt = no  
default_md = sha256  
req_extensions = req_ext  
distinguished_name = dn  
  
[ dn ]  
C = CN  
ST = HN  
L = CS  
O = JK  
OU = Geray  
CN = geray.cn
  
[ req_ext ]  
subjectAltName = @alt_names  
  
[ alt_names ]  
DNS.1 = webhook-service
DNS.2 = webhook-service.kube-system
DNS.3 = webhook-service.default
# 如果需要IP地址，添加类似以下行（替换为实际IP地址）  
# IP.1 = 192.168.1.1
EOF

# openssl req -new -key webhook-service.key -out webhook-service.csr
openssl req -new -key webhook-service.key -out webhook-service.csr -config csr.conf
```

## 3.使用根证书签署Webhook服务的证书
```
openssl x509 -req -in webhook-service.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out webhook-service.crt -days 365 -sha256
```

## 4.提取CA Bundle
```
cat ca.crt | base64 -w 0
```

## 5.配置webhook的CA Bundle
```
apiVersion: admissionregistration.k8s.io/v1  
kind: MutatingWebhookConfiguration  
metadata:  
  name: my-webhook-config  
webhooks:  
- name: my-webhook.example.com  
  clientConfig:  
    service:  
      name: webhook-service  
      namespace: webhook-namespace  
      path: /mutate  
    caBundle: <base64 encoded ca.crt>  
  # ... 其他配置 ...
```
- 部署cert-manager
```
helm repo add jetstack https://charts.jetstack.io
"jetstack" has been added to your repositories

helm repo update
helm search repo jetstack
NAME                                    CHART VERSION   APP VERSION     DESCRIPTION                                       
jetstack/cert-manager                   v1.14.5         v1.14.5         A Helm chart for cert-manager                     
jetstack/cert-manager-approver-policy   v0.14.1         v0.14.1         approver-policy is a CertificateRequest approve...
jetstack/cert-manager-csi-driver        v0.8.1          v0.8.1          cert-manager csi-driver enables issuing secretl...
jetstack/cert-manager-csi-driver-spiffe v0.6.0          v0.6.0          csi-driver-spiffe is a Kubernetes CSI plugin wh...
jetstack/cert-manager-google-cas-issuer v0.8.0          v0.8.0          A Helm chart for jetstack/google-cas-issuer       
jetstack/cert-manager-istio-csr         v0.9.0          v0.9.0          istio-csr enables the use of cert-manager for i...
jetstack/cert-manager-trust             v0.2.1          v0.2.0          DEPRECATED: The old name for trust-manager. Use...
jetstack/trust-manager                  v0.10.1         v0.10.1         trust-manager is the easiest way to manage TLS ...
jetstack/version-checker                v0.5.5          v0.5.5          A Helm chart for version-checker   

helm install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --version v1.14.5 --set installCRDs=true

```

- 部署
```
# 修改namespace
sed -i 's/namespace: system/namespace: default/g' config/default/*.yaml
sed -i 's/namespace: system/namespace: default/g' config/manager/*.yaml
sed -i 's/namespace: system/namespace: default/g' config/rbac/*.yaml
sed -i 's/namespace: system/namespace: default/g' config/webhook/*.yaml

# 使用cert-manager管理的证书生成secret存在ca.crt，这里应该也要把ca证书带上
kubectl -n default-system create secret generic webhook-server-cert2 --from-file=tls.crt=webhook-service.crt --from-file=tls.key=webhook-service.key


docker build -t geray/application-operator:v0.1 .

# 修改镜像：geray/kube-rbac-proxy:v0.11.0
# 部署

make deploy IMG=geray/application-operator:v0.1

kubectl apply -f config/samples/apps_v1_application.yaml


# deployment挂载证书
kubectl edit deployments.apps application-operator-controller-manager 
        volumeMounts:
        - mountPath: /tmp/k8s-webhook-server/serving-certs/
          name: webhook-secrets
          readOnly: true
      volumes:
      - name: webhook-secrets
        secret:
          defaultMode: 420
          secretName: webhook-server-cert

```

# 问题解决
## 权限
> failed to list *v1.Service: services is forbidden: User "system:serviceaccount:default:application-operator-controller-manager" cannot list resource "services" in API group "" at the cluster scope
- 修改clusterrole和role的权限
```
kubectl get role application-operator-leader-election-role 

kubectl get clusterrole application-operator-manager-role -o yaml
- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - apps
  resources:
  - deployments/status
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - services/status
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
```



## 证书不匹配
> kubectl apply -f config/samples/apps_v1_application.yaml 
> Error from server (InternalError): error when creating "config/samples/apps_v1_application.yaml": Internal error occurred: failed calling webhook "mapplication.kb.io": failed to call webhook: Post "https://webhook-service.default.svc:443/mutate-apps-geray-cn-v1-application?timeout=10s": x509: certificate is not valid for any names, but wanted to match webhook-service.default.svc
- 重新签发证书，使能够解析到webhook-service.default.svc
```
cat << EOF > csr.conf
[ req ]  
default_bits = 2048  
prompt = no  
default_md = sha256  
req_extensions = req_ext  
distinguished_name = dn  
  
[ dn ]  
C = CN  
ST = HN  
L = CS  
O = JK  
OU = Geray  
CN = geray.cn
  
[ req_ext ]  
subjectAltName = @alt_names  
  
[ alt_names ]  
DNS.1 = webhook-service
DNS.2 = webhook-service.default
DNS.3 = webhook-service.default.svc
DNS.4 = webhook-service.default.svc.cluster.local
# 如果需要IP地址，添加类似以下行（替换为实际IP地址）  
# IP.1 = 192.168.1.1
EOF
```

## 请求的地址不对
> 部署在default-system命名空间下，但是请求的地址实在kubectl apply -f config/samples/apps_v1_application.yaml 
> Error from server (InternalError): error when creating "config/samples/apps_v1_application.yaml": Internal error occurred: failed calling webhook "mapplication.kb.io": failed to call webhook: Post "https://webhook-service.default.svc:443/mutate-apps-geray-cn-v1-application?timeout=10s": dial tcp 10.233.4.168:443: connect: connection refused
- 解决办法
```
# 修改 ValidatingWebhookConfiguration 和mutatingwebhookconfigurations配置并重启服务
webhooks:
- admissionReviewVersions:
  - v1
  clientConfig:
    service:
      name: application-operator-controller-manager-metrics-service # service名称
      namespace: default-system                                     # 所在命名空间
      path: /mutate-apps-geray-cn-v1-application
      port: 443                                                     # svc端口

```

## 请求端口错误
> Error from server (InternalError): error when creating "config/samples/apps_v1_application.yaml": Internal error occurred: failed calling webhook "mapplication.kb.io": failed to call webhook: Post "https://application-operator-controller-manager-metrics-service.default-system.svc:443/mutate-apps-geray-cn-v1-application?timeout=10s": no service port 443 found for service "application-operator-controller-manager-metrics-service"

- 查看服务svc使用的是8443，需要修改ValidatingWebhookConfiguration 和mutatingwebhookconfigurations配置都为8443

