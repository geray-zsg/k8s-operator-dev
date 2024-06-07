package handlers

import "net/http"

// If the prefix of deployment is kubesphere-router-,it is not allowed to pass through.
func checkDeployPrefixHandleValidate(w http.ResponseWriter, r *http.Request) http.HandlerFunc {
	return nil
}
