package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
)

type WebhookServer struct {
	server *admissionv1.AdmissionReview
}

func (whsvr *WebhookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// 验证请求
	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// 解析请求
	var review admissionv1.AdmissionReview
	_, _, err := codecs.UniversalDeserializer().Decode(body, nil, &review)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not decode body: %v", err), http.StatusInternalServerError)
		return
	}

	// 处理请求
	whsvr.server = &review
	whsvr.mutate()

	// 序列化响应
	resp, err := json.Marshal(whsvr.server)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
		return
	}

	// 发送响应
	if _, err := w.Write(resp); err != nil {
		klog.Error(err, "write response")
	}
}

func (whsvr *WebhookServer) mutate() {
	// 获取请求中的 Deployment 对象
	deployment := whsvr.server.Request.Object.Raw
	obj, _, err := codecs.UniversalDeserializer().Decode(deployment, nil, &appsv1.Deployment{})
	if err != nil {
		klog.Error(err, "decode deployment")
		whsvr.admit(false, err.Error())
		return
	}

	// 检查 Deployment 的副本数是否超过5
	deploy, ok := obj.(*appsv1.Deployment)
	if !ok {
		klog.Error("failed to cast to *appsv1.Deployment")
		whsvr.admit(false, "failed to cast to *appsv1.Deployment")
		return
	}

	if *deploy.Spec.Replicas > 5 {
		whsvr.admit(false, "Deployment replica count should not exceed 5")
		return
	}

	whsvr.admit(true, "")
}

func (whsvr *WebhookServer) admit(allowed bool, msg string) {
	whsvr.server.Response.Allowed = allowed
	whsvr.server.Response.Result = &metav1.Status{
		Message: msg,
	}
}

func main() {
	whsvr := &WebhookServer{}
	http.HandleFunc("/", whsvr.ServeHTTP)
	klog.Info("Starting webhook server...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		klog.Fatal(err)
	}
}
