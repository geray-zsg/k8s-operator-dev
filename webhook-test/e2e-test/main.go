package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// go run endpoint_lifecycle.go -kubeconfig=$HOME/.kube/config -namespace=default

func main() {
	// 解析kubeconfig文件路径
	kubeconfig := flag.String("kubeconfig", os.Getenv("HOME")+"/.kube/config", "绝对路径到kubeconfig文件")
	namespace := flag.String("namespace", "default", "Kubernetes命名空间")
	flag.Parse()

	// 加载kubeconfig文件并创建clientset
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("加载kubeconfig失败: %v\n", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("创建clientset失败: %v\n", err)
		os.Exit(1)
	}

	endpointName := "test-endpoint"

	// 步骤1：创建Endpoint
	endpoint := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: endpointName,
			Labels: map[string]string{
				"test": "initial",
			},
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{IP: "10.0.0.1"},
				},
				Ports: []v1.EndpointPort{
					{Port: 80},
				},
			},
		},
	}

	_, err = clientset.CoreV1().Endpoints(*namespace).Create(context.TODO(), endpoint, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("创建Endpoint失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("已创建EndPoint")

	// 验证创建:获取ep
	createdEndpoint, err := clientset.CoreV1().Endpoints(*namespace).Get(context.TODO(), endpointName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("获取已创建的Endpoint失败: %v\n", err)
		os.Exit(1)
	}
	if createdEndpoint.Name != endpointName {
		fmt.Printf("Endpoint名称不匹配: 期望 %s, 实际 %s\n", endpointName, createdEndpoint.Name)
		os.Exit(1)
	}

	fmt.Printf("NAMESPACE: %v, NAME: %v , ENDPOINTS: %v, LABELS: %v", createdEndpoint.Namespace, createdEndpoint.Name, createdEndpoint.Subsets, createdEndpoint.Labels)

	// 步骤2：更新Endpoint的标签
	createdEndpoint.Labels["test"] = "updated"
	_, err = clientset.CoreV1().Endpoints(*namespace).Update(context.TODO(), createdEndpoint, metav1.UpdateOptions{})
	if err != nil {
		fmt.Printf("更新Endpoint标签失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Endpoint标签已更新")

	// 验证标签更新
	updatedEndpoint, err := clientset.CoreV1().Endpoints(*namespace).Get(context.TODO(), endpointName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("获取更新后的Endpoint失败: %v\n", err)
		os.Exit(1)
	}
	if updatedEndpoint.Labels["test"] != "updated" {
		fmt.Printf("Endpoint标签更新不匹配: 期望 'updated', 实际 '%s'\n", updatedEndpoint.Labels["test"])
		os.Exit(1)
	}
	fmt.Printf("NAMESPACE: %v, NAME: %v , ENDPOINTS: %v, LABELS: %v", updatedEndpoint.Namespace, updatedEndpoint.Name, updatedEndpoint.Subsets, updatedEndpoint.Labels)
	fmt.Println("Endpoint标签更新验证通过")

	// 步骤3：通过补丁更新Endpoint的IPv4地址和端口
	patch := []byte(`{"subsets": [{"addresses": [{"ip": "10.0.0.2"}], "ports": [{"port": 8080}]}]}`)
	_, err = clientset.CoreV1().Endpoints(*namespace).Patch(context.TODO(), endpointName, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		fmt.Printf("补丁更新Endpoint失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Endpoint已补丁更新ip、port")

	// 验证补丁更新
	patchedEndpoint, err := clientset.CoreV1().Endpoints(*namespace).Get(context.TODO(), endpointName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("获取补丁更新后的Endpoint失败: %v\n", err)
		os.Exit(1)
	}
	if patchedEndpoint.Subsets[0].Addresses[0].IP != "10.0.0.2" {
		fmt.Printf("EndpointIP补丁更新不匹配: 期望 '10.0.0.2', 实际 '%s'\n", patchedEndpoint.Subsets[0].Addresses[0].IP)
		os.Exit(1)
	}
	if patchedEndpoint.Subsets[0].Ports[0].Port != 8080 {
		fmt.Printf("Endpoint端口补丁更新不匹配: 期望 '8080', 实际 '%d'\n", patchedEndpoint.Subsets[0].Ports[0].Port)
		os.Exit(1)
	}

	fmt.Printf("NAMESPACE: %v, NAME: %v , ENDPOINTS: %v, LABELS: %v", patchedEndpoint.Namespace, patchedEndpoint.Name, patchedEndpoint.Subsets, patchedEndpoint.Labels)

	// 步骤4：监听Endpoint删除事件
	watcher, err := clientset.CoreV1().Endpoints(*namespace).Watch(context.TODO(), metav1.ListOptions{
		LabelSelector: "test=updated",
	})
	if err != nil {
		fmt.Printf("设置监听失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("监听已设置")

	// 通过标签删除Endpoint
	err = clientset.CoreV1().Endpoints(*namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: "test=updated",
	})
	if err != nil {
		fmt.Printf("通过标签删除Endpoint失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Endpoint已删除")

	// 验证删除事件
	timeout := time.After(30 * time.Second)
	for {
		select {
		case event := <-watcher.ResultChan():
			if event.Type == watch.Deleted {
				deletedEndpoint := event.Object.(*v1.Endpoints)
				if deletedEndpoint.Name == endpointName {
					fmt.Printf("成功接收到Endpoint %s 的删除事件\n", endpointName)
					return
				}
			}
		case <-timeout:
			fmt.Println("等待Endpoint删除事件超时")
			os.Exit(1)
		}
	}

}
