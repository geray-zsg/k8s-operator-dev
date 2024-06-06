package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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

			oldNamespace := corev1.Namespace{}
			if err := json.Unmarshal(admissionReview.Request.OldObject.Raw, &oldNamespace); err != nil {
				http.Error(w, "could not unmarshal old namespace", http.StatusBadRequest)
				return
			}

			fmt.Println("namespace:", namespace)
			fmt.Println("oldNamespace:", oldNamespace)

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
}
