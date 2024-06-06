package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WHParameters struct {
	labelsToCheck []string
	tlsKey        string
	tlscert       string
}

func main() {

	var parameters WHParameters
	var labels string
	// get command line parameters
	flag.StringVar(&parameters.tlsKey, "tlsKeyFile", "etc/webhook/certs/tls.key", "File containing the x509 private key to --tlsCertFile.--tlsKeyFile")
	flag.StringVar(&parameters.tlsKey, "tlsCertFile", "etc/webhook/certs/tls.cert", "File containing the x509 Certificate for HTTPS. --tlsCertFile.")
	flag.StringVar(&labels, "labels", "kubesphere.io/vpc,other-label", "Comma-separated list of labels to check")

	flag.Parse()

	// Split the labels into a slice
	parameters.labelsToCheck = strings.Split(labels, ",")

	http.HandleFunc("/mutate", handleMutate)
	http.HandleFunc("/validate", handleValidate) // 添加验证处理函数

	// Load the certificate and key files
	cert, err := tls.LoadX509KeyPair("/etc/webhook/certs/tls.crt", "/etc/webhook/certs/tls.key")
	if err != nil {
		log.Fatalf("Failed to load key pair: %v", err)
	}

	server := &http.Server{
		Addr: ":8443",
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	log.Println("Starting webhook server...")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleMutate(w http.ResponseWriter, r *http.Request) {
	var admissionReview admissionv1.AdmissionReview
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &admissionReview); err != nil {
		http.Error(w, "could not unmarshal request", http.StatusBadRequest)
		return
	}

	admissionResponse := admissionv1.AdmissionResponse{
		UID: admissionReview.Request.UID,
	}

	if admissionReview.Request.Kind.Kind != "Pod" {
		admissionResponse.Allowed = true
	} else {
		var pod corev1.Pod
		if err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod); err != nil {
			http.Error(w, "could not unmarshal pod", http.StatusBadRequest)
			return
		}

		// Prepare the patch operations
		// 检查是否存在环境变量
		var patch string
		if len(pod.Spec.Containers[0].Env) == 0 {
			patch = `[{"op": "add", "path": "/spec/containers/0/env", "value": [{"name": "INJECTED_ENV", "value": "injected-value"}]}]`
		} else {
			patch = `[{"op": "add", "path": "/spec/containers/0/env/-", "value": {"name": "INJECTED_ENV", "value": "injected-value"}}]`
		}

		admissionResponse.Patch = []byte(patch)
		patchType := admissionv1.PatchTypeJSONPatch
		admissionResponse.PatchType = &patchType
		admissionResponse.Allowed = true
	}

	admissionReview.Response = &admissionResponse
	respBytes, err := json.Marshal(admissionReview)
	if err != nil {
		http.Error(w, "could not marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	var admissionReview admissionv1.AdmissionReview
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &admissionReview); err != nil {
		http.Error(w, "could not unmarshal request", http.StatusBadRequest)
		return
	}

	admissionResponse := admissionv1.AdmissionResponse{
		UID: admissionReview.Request.UID,
	}

	if admissionReview.Request.Kind.Kind == "Namespace" {
		var namespace corev1.Namespace
		if err := json.Unmarshal(admissionReview.Request.Object.Raw, &namespace); err != nil {
			http.Error(w, "could not unmarshal namespace", http.StatusBadRequest)
			return
		}

		oldNamespace := corev1.Namespace{}
		if err := json.Unmarshal(admissionReview.Request.OldObject.Raw, &oldNamespace); err != nil {
			http.Error(w, "could not unmarshal old namespace", http.StatusBadRequest)
			return
		}

		fmt.Println("namespace:", namespace)
		fmt.Println("oldNamespace:", oldNamespace)

		// Define labels to check
		labelsToCheck := []string{"kubesphere.io/vpc", "other-label"}

		// Check if oldNamespace.Labels is nil or specified labels are modified
		if oldNamespace.Labels == nil {
			admissionResponse.Allowed = true
		} else {
			for _, label := range labelsToCheck {
				oldValue, oldExists := oldNamespace.Labels[label]
				newValue, newExists := namespace.Labels[label]
				fmt.Printf("oldValue中%v的值是： %v\n", label, oldValue)
				fmt.Printf("newValue中%v的值是： %v\n", label, newValue)

				// Check if label is modified
				if oldExists && newExists && newValue != oldValue {
					admissionResponse.Allowed = false
					admissionResponse.Result = &metav1.Status{
						Message: fmt.Sprintf("Modifying the %s label is not allowed", label),
					}
					break
				}

				// Check if label is added
				if !oldExists && newExists {
					admissionResponse.Allowed = true
					fmt.Printf("The label %v is added, allow the creation\n", label)
					break
				}

				// Check if label is not modified
				if oldExists && newExists && newValue == oldValue {
					admissionResponse.Allowed = true
					fmt.Printf("The label %v is not modified, allow the creation\n", label)
					break
				}

				// Check if label is deleted
				if oldExists && !newExists {
					admissionResponse.Allowed = false
					admissionResponse.Result = &metav1.Status{
						Message: fmt.Sprintf("Deleting the %s label is not allowed", label),
					}
					break
				}

				// break
			}
		}
	} else {
		admissionResponse.Allowed = true
	}

	// Ensure a default message is set if the request is denied without a specific message
	if !admissionResponse.Allowed && admissionResponse.Result == nil {
		admissionResponse.Result = &metav1.Status{
			Message: "Request denied",
		}
	}

	admissionReview.Response = &admissionResponse
	respBytes, err := json.Marshal(admissionReview)
	if err != nil {
		http.Error(w, "could not marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}
