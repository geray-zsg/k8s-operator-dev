# 1.部署
## 生成证书
```
mkdir certs
openssl req -newkey rsa:2048 -nodes -keyout certs/tls.key -x509 -days 365 -out certs/tls.crt -subj "/CN=webhook.default.svc"
```

- 创建secret
```
kubectl create secret generic webhook-certs --from-file=certs -n default
```

- 部署deployment和service
```
kubectl apply -f deployment.yaml
```

- 部署MutatingWebhookConfiguration和validating-webhook-config
```
kubectl apply -f validating-webhook-config.yaml
kubectl apply -f mutatingwebhookconfigurations.yaml
```

# 问题处理
## 本地运行证书中没有包含IP
> Error from server (InternalError): error when creating "test-pod.yaml": Internal error occurred: failed calling webhook "webhook.default.svc": failed to call webhook: Post "https://192.168.193.11:8443/mutate?timeout=10s": x509: cannot validate certificate for 192.168.193.11 because it doesn't contain any IP SANs

- 生成包含 IP SAN 的自签名证书
- 创建一个配置文件 openssl.cnf：
```
[ req ]
default_bits       = 2048
distinguished_name = req_distinguished_name
req_extensions     = req_ext
x509_extensions    = v3_ca # The extensions to add to the self-signed cert

[ req_distinguished_name ]
countryName                 = Country Name (2 letter code)
countryName_default         = US
stateOrProvinceName         = State or Province Name (full name)
stateOrProvinceName_default = California
localityName                = Locality Name (eg, city)
localityName_default        = San Francisco
organizationName            = Organization Name (eg, company)
organizationName_default    = My Company
commonName                  = Common Name (e.g. server FQDN or YOUR name)
commonName_default          = webhook.default.svc

[ req_ext ]
subjectAltName = @alt_names

[ v3_ca ]
subjectAltName = @alt_names

[ alt_names ]
IP.1   = 192.168.193.11 # 替换为你的本地 IP
DNS.1  = webhook.default.svc
```

- 生成证书和密钥：
```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout certs/tls.key -out certs/tls.crt -config openssl.cnf -extensions v3_ca

```
- 更新 Kubernetes Secret
使用生成的证书和密钥更新 Kubernetes Secret：
```
kubectl delete secret webhook-certs -n default
kubectl create secret generic webhook-certs --from-file=certs -n default

```
- 更新 MutatingWebhookConfiguration
```
caBundle: $(cat certs/tls.crt | base64 | tr -d '\n')

```

