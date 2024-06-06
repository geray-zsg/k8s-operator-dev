package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"simple-webhook/handlers"
	"simple-webhook/types"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
)

// type WHParameters struct {
// 	labelsToCheck []string
// 	tlsKey        string
// 	tlscert       string
// }

func main() {

	var parameters types.WHParameters
	var configEnable types.ConfigEnabel
	var labels string

	// Get command line parameters
	flag.StringVar(&parameters.TLSKey, "tlsKeyFile", "/etc/webhook/certs/tls.key", "File containing the x509 private key to --tlsCertFile.--tlsKeyFile")
	flag.StringVar(&parameters.TLSCert, "tlsCertFile", "/etc/webhook/certs/tls.crt", "File containing the x509 Certificate for HTTPS. --tlsCertFile.")
	flag.StringVar(&labels, "labels", "nci.yunshan.net/vpc,kubesphere.io/workspace,kubesphere.io/namespace", "Comma-separated list of labels to check. --labels")

	// 定义命令行参数
	flag.BoolVar(&configEnable.NamespaceLabelsHandleValidate, "enable-namespace-validation", true, "Enable namespace validation.--enable-namespace-validation")
	flag.BoolVar(&configEnable.PodEnvInjectedHandleMutate, "enable-podEnv-Injecte", false, "Enable pod env Injecte. --enable-podEnv-Injecte")

	flag.Parse()

	// Split the labels into a slice
	parameters.LabelsToCheck = strings.Split(labels, ",")
	fmt.Println("parameters.LabelsToCheck:", parameters.LabelsToCheck)
	fmt.Println("configEnable.NamespaceLabelsHandleValidate:", configEnable.NamespaceLabelsHandleValidate)
	if configEnable.PodEnvInjectedHandleMutate {
		http.HandleFunc("/mutate", handlers.PodEnvInjectedHandleMutate)
	}

	if configEnable.NamespaceLabelsHandleValidate {
		http.HandleFunc("/validate", handlers.NamespaceLabelsHandleValidate(parameters.LabelsToCheck))
	} else {
		// 注册一个总是返回允许的处理函数（如果关闭了该功能则进行放行）
		http.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
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
				UID:     admissionReview.Request.UID,
				Allowed: true,
			}

			admissionReview.Response = &admissionResponse
			respBytes, err := json.Marshal(admissionReview)
			if err != nil {
				http.Error(w, "could not marshal response", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(respBytes)
		})
	}

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
