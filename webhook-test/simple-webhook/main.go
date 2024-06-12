package main

import (
	"crypto/tls"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"simple-webhook/registerhandlers"
	"simple-webhook/types"
	"strings"
	"syscall"

	"github.com/golang/glog"
)

func main() {
	var parameters types.WHParameters
	var configEnable types.ConfigEnabel
	var configHandlersParameters types.ConfigHandlersParameters
	var labels string

	// Get command line parameters
	flag.StringVar(&parameters.TLSKey, "tlsKeyFile", "/etc/webhook/certs/tls.key", "File containing the x509 private key to --tlsCertFile.--tlsKeyFile")
	flag.StringVar(&parameters.TLSCert, "tlsCertFile", "/etc/webhook/certs/tls.crt", "File containing the x509 Certificate for HTTPS. --tlsCertFile.")
	flag.StringVar(&labels, "labels", "nci.yunshan.net/vpc,kubesphere.io/workspace,kubesphere.io/namespace", "Comma-separated list of labels to check. --labels")
	flag.StringVar(&parameters.TLSPort, "tlsPort", ":8443", "Webhook Server staring tls port.--tlsPort")
	flag.StringVar(&parameters.HealthPort, "healthPort", ":8080", "Webhook Server staring tls port.--healthPort")

	// 定义命令行参数
	flag.BoolVar(&configEnable.NamespaceLabelsHandleValidate, "enable-namespace-validation", true, "Enable namespace validation.--enable-namespace-validation")
	flag.BoolVar(&configEnable.PodEnvInjectedHandleMutate, "enable-podEnv-Injecte", false, "Enable pod env Injecte. --enable-podEnv-Injecte")

	// Add glog flags
	flag.Set("logtostderr", "true")
	flag.Parse()

	// Ensure flag is correctly initialized
	defer glog.Flush()

	// Check for certificates
	if parameters.TLSCert == "" || parameters.TLSKey == "" {
		glog.Fatalf("No available certificates.")
	}

	// Split the labels into a slice
	configHandlersParameters.LabelsToCheck = strings.Split(labels, ",")
	glog.Infof("Labels to check: %v", configHandlersParameters.LabelsToCheck)
	glog.Infof("Namespace validation enabled: %v", configEnable.NamespaceLabelsHandleValidate)

	// Register webhook handlers
	registerhandlers.RegisterHandlers(configEnable, configHandlersParameters)

	// Load the certificate and key files
	cert, err := tls.LoadX509KeyPair(parameters.TLSCert, parameters.TLSKey)
	if err != nil {
		glog.Fatalf("Failed to load key pair: %v", err)
	}

	server := &http.Server{
		Addr: parameters.TLSPort,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	// Start webhook server in a new goroutine
	go func() {
		glog.Info("Starting webhook server on port ...", parameters.TLSPort)
		if err := server.ListenAndServeTLS("", ""); err != nil {
			glog.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Start health and readiness checks
	go registerhandlers.StartHealthCheckServer(parameters.HealthPort)

	// Listen for OS shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	glog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
	glog.Fatalf("Server shutdown failed: %v", err)

}
