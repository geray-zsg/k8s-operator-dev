package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
)

// If the prefix of deployment is kubesphere-router-,it is not allowed to pass through.
func CheckDeployPrefixHandleValidate(deployPrefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var admissionReview admissionv1.AdmissionReview
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "could not read request", http.StatusBadRequest)
			glog.Errorf("Error reading request body: %v", err)
			return
		}

		if err := json.Unmarshal(body, &admissionReview); err != nil {
			http.Error(w, "could not unmarshal request", http.StatusBadRequest)
			glog.Errorf("Error unmarshalling request body: %v", err)
			return
		}
		admissionResponse := admissionv1.AdmissionResponse{
			UID: admissionReview.Request.UID,
			// Allowed: true,
		}

		// 填写代码
		if admissionReview.Request.Kind.Kind == "Deployment" {
			var deploy appsv1.Deployment
			if json.Unmarshal(admissionReview.Request.Object.Raw, &deploy); err != nil {
				http.Error(w, "could not unmarshal deployment", http.StatusBadRequest)
			}

			glog.Info("deployment:", deploy)

			admissionResponse.Allowed = true
			glog.Info("Deployment.Name:", deploy.Name)

		}

		admissionReview.Response = &admissionResponse
		respBytes, err := json.Marshal(admissionReview)
		if err != nil {
			http.Error(w, "could not marshal response", http.StatusInternalServerError)
			glog.Errorf("Error marshalling response body: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(respBytes)
	}
}
