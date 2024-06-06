package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"simple-webhook/handlers"
	"simple-webhook/types"
	"strings"
)

// type WHParameters struct {
// 	labelsToCheck []string
// 	tlsKey        string
// 	tlscert       string
// }

func main() {

	var parameters types.WHParameters
	var labels string

	// Get command line parameters
	flag.StringVar(&parameters.TLSKey, "tlsKeyFile", "/etc/webhook/certs/tls.key", "File containing the x509 private key to --tlsCertFile.--tlsKeyFile")
	flag.StringVar(&parameters.TLSCert, "tlsCertFile", "/etc/webhook/certs/tls.crt", "File containing the x509 Certificate for HTTPS. --tlsCertFile.")
	flag.StringVar(&labels, "labels", "kubesphere.io/vpc,other-label", "Comma-separated list of labels to check")
	flag.Parse()

	// Split the labels into a slice
	parameters.LabelsToCheck = strings.Split(labels, ",")

	http.HandleFunc("/mutate", handlers.PodEnvInjectedHandleMutate)
	http.HandleFunc("/validate", handlers.NamespaceLabelsHandleValidate(parameters.LabelsToCheck)) // 添加验证处理函数

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
