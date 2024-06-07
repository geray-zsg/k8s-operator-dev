package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	admissionv1 "k8s.io/api/admission/v1"
)

func AllowedHandlers() http.HandlerFunc {
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
			UID:     admissionReview.Request.UID,
			Allowed: true,
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
