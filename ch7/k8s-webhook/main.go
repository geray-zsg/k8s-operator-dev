package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// 处理Admission请求的函数
func handleAdmission(w http.ResponseWriter, r *http.Request) {
	var admissionReviewReq admissionv1.AdmissionReview
	var admissionReviewResp admissionv1.AdmissionReview

	// 1.读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Cloud not read request body: %v", err)
		http.Error(w, "cloud not read request body: %v", http.StatusBadRequest)
		return
	}

	// 2. 解析请求体
	if err := json.Unmarshal(body, &admissionReviewReq); err != nil {
		log.Printf("Could not unmarshal request: %v", err)
		http.Error(w, "could not unmarshal request", http.StatusBadRequest)
		return
	}

	// 3. 创建响应对象
	admissionReviewResp.Response = &admissionv1.AdmissionResponse{
		UID: admissionReviewReq.Request.UID,
	}

	// 4. 检查删除操作和资源类型(首先检查操作类型是否为删除(DELETE)且资源类型是否为Deployment。如果是，则进一步检查部署名称是否以kubesphere-router开头。如果满足条件，拒绝删除请求并返回相应的错误信息。否则，允许删除请求。)
	if admissionReviewReq.Request.Operation == admissionv1.Delete &&
		admissionReviewReq.Request.Kind.Kind == "Deployment" {
		deploymentName := admissionReviewReq.Request.Name
		if strings.HasPrefix(deploymentName, "kubesphere-router") {
			admissionReviewResp.Response.Allowed = false
			admissionReviewResp.Response.Result = &metav1.Status{
				Message: fmt.Sprintf("Deletion of deployment '%s' is not allowed", deploymentName),
			}
		} else {
			admissionReviewResp.Response.Allowed = true
		}
	} else {
		admissionReviewResp.Response.Allowed = true
	}

	// 5. 发送响应
	respBytes, err := json.Marshal(admissionReviewResp)
	if err != nil {
		log.Printf("Could not marshal response: %v", err)
		http.Error(w, "could not marshal response", http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(respBytes); err != nil {
		log.Printf("Could not write response: %v", err)
		http.Error(w, "could not write response", http.StatusInternalServerError)
		return
	}

}

func main() {
	// 1. 配置HTTP处理函数
	http.HandleFunc("/validate", handleAdmission)

	// 2. 启动服务器
	log.Println("Starting server...")
	if err := http.ListenAndServeTLS(":8443", "/etc/ssl/tls.crt", "/etc/ssl/tls.key", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
