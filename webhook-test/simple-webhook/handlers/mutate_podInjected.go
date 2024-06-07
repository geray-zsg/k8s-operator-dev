package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

// HandleMutate handles the mutating webhook requests
func PodEnvInjectedHandleMutate(w http.ResponseWriter, r *http.Request) {
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

	if admissionReview.Request.Kind.Kind != "Pod" {
		admissionResponse.Allowed = true
	} else {
		var pod corev1.Pod
		if err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod); err != nil {
			http.Error(w, "could not unmarshal pod", http.StatusBadRequest)
			return
		}

		// Prepare the patch operations
		// 检查是否存在环境变量
		var patch string
		if len(pod.Spec.Containers[0].Env) == 0 {
			patch = `[{"op": "add", "path": "/spec/containers/0/env", "value": [{"name": "INJECTED_ENV", "value": "injected-value"}]}]`
		} else {
			patch = `[{"op": "add", "path": "/spec/containers/0/env/-", "value": {"name": "INJECTED_ENV", "value": "injected-value"}}]`
		}

		admissionResponse.Patch = []byte(patch)
		patchType := admissionv1.PatchTypeJSONPatch
		admissionResponse.PatchType = &patchType
		admissionResponse.Allowed = true
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
