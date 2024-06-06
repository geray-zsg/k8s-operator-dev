package main

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	watch "k8s.io/apimachinery/pkg/watch"
)

func main() {
	// 初始化 Kubernetes 客户端
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 调用测试函数
	if err := testEndpointLifecycle(clientset); err != nil {
		fmt.Printf("Test failed: %v\n", err)
	} else {
		fmt.Println("Test passed!")
	}
}

func testEndpointLifecycle(clientset *kubernetes.Clientset) error {
	testNamespaceName := "default"
	testEndpointName := "testservice"
	testEndpoints := v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: testEndpointName,
			Labels: map[string]string{
				"test-endpoint-static": "true",
			},
		},
		Subsets: []v1.EndpointSubset{{
			Addresses: []v1.EndpointAddress{{
				IP: "10.0.0.24",
			}},
			Ports: []v1.EndpointPort{{
				Name:     "http",
				Port:     80,
				Protocol: v1.ProtocolTCP,
			}},
		}},
	}
	// w := &cache.ListWatch{
	// 	WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
	// 		options.LabelSelector = "test-endpoint-static=true"
	// 		return clientset.CoreV1().Endpoints(testNamespaceName).Watch(context.TODO(), options)
	// 	},
	// }
	endpointsList, err := clientset.CoreV1().Endpoints("").List(context.TODO(), metav1.ListOptions{LabelSelector: "test-endpoint-static=true"})
	if err != nil {
		return err
	}

	// 创建 Endpoint
	fmt.Println("Creating an Endpoint...")
	if _, err := clientset.CoreV1().Endpoints(testNamespaceName).Create(context.TODO(), &testEndpoints, metav1.CreateOptions{}); err != nil {
		return err
	}

	// 等待 Endpoint 可用
	// fmt.Println("Waiting for available Endpoint...")
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()

	watcher, err := clientset.CoreV1().Endpoints(testEndpointName).Watch(context.TODO(), metav1.ListOptions{LabelSelector: "test-endpoint-static=true"})
	if err != nil {
		return err
	}

	for {
		select {
		case event := <-watcher.ResultChan():
			switch event.Type {
			case watch.Added, watch.Modified:
				if endpoints, ok := event.Object.(*v1.Endpoints); ok {
					if isEndpointAvailable(endpoints) {
						fmt.Println("Endpoint is available!")
						return nil
					}
				}
			}
		}
	}
	// _, err = watch.Until(ctx, endpointsList.ResourceVersion, w, func(event watch.Event) (bool, error) {
	// 	switch event.Type {
	// 	case watch.Added:
	// 		if endpoints, ok := event.Object.(*v1.Endpoints); ok {
	// 			found := endpoints.ObjectMeta.Name == endpoints.Name &&
	// 				endpoints.Labels["test-endpoint-static"] == "true"
	// 			return found, nil
	// 		}
	// 	default:
	// 		fmt.Printf("Observed event type %v\n", event.Type)
	// 	}
	// 	return false, nil
	// })
	// if err != nil {
	// 	return err
	// }

	// 列出所有的 Endpoints
	fmt.Println("Listing all Endpoints...")
	endpointsList, err = clientset.CoreV1().Endpoints("").List(context.TODO(), metav1.ListOptions{LabelSelector: "test-endpoint-static=true"})
	if err != nil {
		return err
	}
	eventFound := false
	for _, endpoint := range endpointsList.Items {
		if endpoint.ObjectMeta.Name == testEndpointName && endpoint.ObjectMeta.Namespace == testNamespaceName {
			eventFound = true
			break
		}
	}
	if !eventFound {
		return fmt.Errorf("unable to find Endpoint Service in list of Endpoints")
	}

	// 更新 Endpoint
	fmt.Println("Updating the Endpoint...")
	foundEndpoint := endpointsList.Items[0]
	foundEndpoint.ObjectMeta.Labels["test-service"] = "updated"
	_, err = clientset.CoreV1().Endpoints(testNamespaceName).Update(context.TODO(), &foundEndpoint, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

// 判断 Endpoint 是否可用
func isEndpointAvailable(endpoints *v1.Endpoints) bool {
	for _, subset := range endpoints.Subsets {
		if len(subset.Addresses) > 0 {
			return true
		}
	}
	return false
}
