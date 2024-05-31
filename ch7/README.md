
```
mkdir application-operator && cd application-operator 
kubebuilder init --domain geray.cn --owner "Geray" --repo github.com/geray/application-operator
kubebuilder create api --group apps --version v1 --kind Application

```

