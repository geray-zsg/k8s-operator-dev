package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HandleValidate handles the validating webhook requests
func NamespaceLabelsHandleValidate(labelsToCheck []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

			var oldNamespace corev1.Namespace
			if err := json.Unmarshal(admissionReview.Request.OldObject.Raw, &oldNamespace); err != nil {
				http.Error(w, "could not unmarshal old namespace", http.StatusBadRequest)
				return
			}

			glog.Info("namespace:", namespace)
			glog.Info("oldNamespace:", oldNamespace)

			admissionResponse.Allowed = true

			for _, label := range labelsToCheck {
				oldValue, oldExists := oldNamespace.Labels[label]
				newValue, newExists := namespace.Labels[label]

				// glog.Info("Checking label: %s, oldValue: %s, oldExists: %t, newValue: %s, newExists: %t\n", label, oldValue, oldExists, newValue, newExists)

				if oldExists && newExists && newValue != oldValue {
					admissionResponse.Allowed = false
					// glog.Info("Modifying label: %s, oldValue: %s, newValue: %s \n", label, oldValue, newValue)
					admissionResponse.Result = &metav1.Status{
						Message: fmt.Sprintf("Modifying the %s label is not allowed", label),
					}
					break
				}

				if !oldExists && newExists {
					admissionResponse.Allowed = true
					// glog.Info("Adding the label %v is denied\n", label)
					admissionResponse.Result = &metav1.Status{
						Message: fmt.Sprintf("Adding the %s label is not allowed", label),
					}
					break
				}

				if oldExists && !newExists {
					admissionResponse.Allowed = false
					// glog.Info("Deleting label: %s, oldValue: %s\n", label, oldValue)
					admissionResponse.Result = &metav1.Status{
						Message: fmt.Sprintf("Deleting the %s label is not allowed", label),
					}
					break
				}

				// glog.Info("Label check passed for label: %s, oldValue: %s, newValue: %s \n", label, oldValue, newValue)
			}
		} else {
			glog.Info("Not a Namespace, allowing the request\n")
			admissionResponse.Allowed = true
		}

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
}
