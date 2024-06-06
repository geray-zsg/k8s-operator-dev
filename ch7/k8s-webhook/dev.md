# 创建过程
```
# 初始化项目
mkdir k8s-webhook
cd k8s-webhook
kubebuilder init --domain geray.cn --owner "Geray" --repo github.com/geray/k8s-webhook

# 创建API和webhook
kubebuilder create api --group webhook.geray.cn --version v1beta1 --kind Gwebhook
kubebuilder create webhook --group webhook.geray.cn --version v1beta1 --kind Gwebhook --defaulting --programmatic-validation
```

# 业务逻辑代码


# 使用Go编写一个简单的webhook
> 如果删除的Deployment是kubesphere-router前缀开头则不能删除
> Admission Webhook分为两种类型：MutatingAdmissionWebhook和ValidatingAdmissionWebhook。我的需求属于ValidatingAdmissionWebhook。
- 步骤
编写一个简单的Go应用程序作为Webhook服务器。
部署Webhook服务器，并确保它能够被Kubernetes API Server访问。
配置Kubernetes API Server以使用这个Webhook。

```
go mod init github.io/geray/k8s-webhook
go mod tidy

```

- 编写逻辑代码

## 配置证书
```
# 配置证书
openssl genrsa -out ca.key 2048  
openssl req -x509 -new -nodes -key ca.key -sha256 -days 365 -out ca.crt

openssl genrsa -out tls.key 2048  

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
CN = webhook-server.default
  
[ req_ext ]  
subjectAltName = @alt_names  
  
[ alt_names ]  
DNS.1 = webhook-server
DNS.2 = webhook-server.default
DNS.3 = webhook-server.default.svc
DNS.4 = webhook-server.default.svc.cluster
DNS.4 = webhook-server.default.svc.cluster.local
# 如果需要IP地址，添加类似以下行（替换为实际IP地址）  
# IP.1 = 192.168.1.1
EOF

openssl req -new -key tls.key -out tls.csr -config csr.conf

openssl x509 -req -in tls.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out tls.crt -days 365 -sha256

# 创建secret
#openssl req -newkey rsa:2048 -nodes -keyout tls.key -x509 -days 365 -out tls.crt -subj "/CN=webhook-server.default.svc"
#kubectl create secret tls webhook-server-tls --cert=tls.crt --key=tls.key -n default
#kubectl get secret webhook-server-tls -o jsonpath='{.data.tls\.crt}' -n default        # caBundle
kubectl  create secret generic webhook-server-tls --from-file=tls.crt=tls.crt --from-file=tls.key=tls.key --from-file=ca.crt=ca.crt

# 构建镜像
docker build -t k8s-webhook:v1beat1 .

```

## 部署服务
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webhook-server
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: webhook-server
  template:
    metadata:
      labels:
        app: webhook-server
    spec:
      containers:
        - name: webhook-server
          image: your-webhook-server-image
          ports:
            - containerPort: 8443
          volumeMounts:
            - name: webhook-certs
              mountPath: "/tmp/tls"
              readOnly: true
      volumes:
        - name: webhook-certs
          secret:
            secretName: webhook-server-tls
---
apiVersion: v1
kind: Service
metadata:
  name: webhook-server
  namespace: default
spec:
  ports:
    - port: 443
      targetPort: 8443
  selector:
    app: webhook-server
```

## 配置Admission Webhook
```
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: validating-webhook-configuration
webhooks:
  - name: deployment-validation.webhook.k8s.io
    rules:
      - apiGroups: ["apps"]
        apiVersions: ["v1"]
        operations: ["DELETE"]
        resources: ["deployments"]
    clientConfig:
      service:
        name: webhook-server
        namespace: default
        path: "/validate"
      caBundle: <base64-encoded-ca-cert>
    admissionReviewVersions: ["v1"]
    sideEffects: None
```