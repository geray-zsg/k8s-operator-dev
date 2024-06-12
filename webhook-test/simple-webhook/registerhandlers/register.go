package registerhandlers

import (
	"net/http"
	"simple-webhook/handlers"
	"simple-webhook/types"
)

// func RegisterHandlers(configEnable types.ConfigEnabel, labels []string) {
func RegisterHandlers(configEnable types.ConfigEnabel, configHandlersParameters types.ConfigHandlersParameters) {
	if configEnable.PodEnvInjectedHandleMutate {
		http.HandleFunc("/mutate", handlers.PodEnvInjectedHandleMutate)
	}
	if configEnable.NamespaceLabelsHandleValidate {
		// glog.Info("configHandlersParameters.LabelsToCheck:", labels)
		// glog.Info("configHandlersParameters.LabelsToCheck:", configHandlersParameters.LabelsToCheck)
		http.HandleFunc("/validate", handlers.NamespaceLabelsHandleValidate(configHandlersParameters.LabelsToCheck))
		// http.HandleFunc("/validate", handlers.NamespaceLabelsHandleValidate(configHandlersParameters.LabelsToCheck))
	} else {
		http.HandleFunc("/validate", handlers.AllowedHandlers())
	}
	if configEnable.CheckDeploymentPrefix {
		http.HandleFunc("/validateDeploy", handlers.CheckDeployPrefixHandleValidate(configHandlersParameters.DeploymentPrefix))
	}
}
