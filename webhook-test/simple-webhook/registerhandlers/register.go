package registerhandlers

import (
	"net/http"
	"simple-webhook/handlers"
	"simple-webhook/types"
)

func RegisterHandlers(configEnable types.ConfigEnabel, labels []string) {
	if configEnable.PodEnvInjectedHandleMutate {
		http.HandleFunc("/mutate", handlers.PodEnvInjectedHandleMutate)
	}
	if configEnable.NamespaceLabelsHandleValidate {
		http.HandleFunc("/validate", handlers.NamespaceLabelsHandleValidate(labels))
	} else {
		http.HandleFunc("/validate", handlers.AllowedHandlers())
	}
}
