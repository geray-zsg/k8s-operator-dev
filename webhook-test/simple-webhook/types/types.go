package types

import (
	"github.com/golang/glog"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// WHParameters defines the structure for webhook parameters
type WHParameters struct {
	// server parameters
	TLSKey     string
	TLSCert    string
	TLSPort    string
	HealthPort string

	//	handlers parameters
	// LabelsToCheck    []string
	// DeploymentPrefix string
}

// Enable or disable specific features as needed
type ConfigEnabel struct {
	// mutatingwebhookconfigurations
	MutatePodEnvInjectedHandle bool

	// validatingwebhookconfigurations
	ValidateNamespaceLabelsHandle bool
	ValidateCheckDeploymentPrefix bool
}

type ConfigHandlersParameters struct {
	//	handlers parameters
	LabelsToCheck    []string
	DeploymentPrefix string
}

type Client struct {
	dynamicClient dynamic.Interface
}

func NewClient() (*Client, error) {
	// 1.GET config in k8s cluster（自动获取部署在 Kubernetes 集群内的 Pod 的服务账户和 API 服务器的地址，无需手动提供 kubeconfig 文件。）
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorf("Failed to obtain the internal configuration file of the cluster .")
		return nil, err
	}

	// 2.使用配置创建动态客户端
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		glog.Errorf("Failed to configure dynamic client.")
		return nil, err
	}

	// 3.返回包含动态客户端的 Client 实例
	return &Client{dynamicClient: dynamicClient}, nil
}
