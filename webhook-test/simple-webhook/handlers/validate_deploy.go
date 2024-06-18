package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/glog"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		glog.Infof("Request body: %s", string(body))
		if err := json.Unmarshal(body, &admissionReview); err != nil {
			http.Error(w, "could not unmarshal request", http.StatusBadRequest)
			glog.Errorf("Error unmarshalling request body: %v", err)
			return
		}
		admissionResponse := admissionv1.AdmissionResponse{
			UID: admissionReview.Request.UID,
		}

		// 填写代码
		glog.Info("下面是检查deployment的逻辑代码")
		// if admissionReview.Request.Kind.Kind == "Deployment" && admissionReview.Request.Operation == admissionv1.Delete {
		if admissionReview.Request.Kind.Kind == "Deployment" {
			var deploy appsv1.Deployment
			if err := json.Unmarshal(admissionReview.Request.OldObject.Raw, &deploy); err != nil {
				http.Error(w, "could not unmarshal deployment", http.StatusBadRequest)
				glog.Errorf("Error unmarshalling deployment: %v", err)
				return
			}

			// glog.Infof("deployment: %s", deploy)
			glog.Infof("Deployment.Name===========================================》: %s", deploy)

			if strings.HasPrefix(deploy.Name, deployPrefix) {
				admissionResponse.Allowed = true
				glog.Infof("Deployment %s cannot be deleted because its prefix is %s on %s.", deploy.Name, deployPrefix, &deploy.Namespace)
				admissionResponse.Result = &metav1.Status{
					Message: fmt.Sprintf("Deployment %s cannot be deleted because its prefix is %s.", deploy.Name, deployPrefix),
				}
			} else {
				glog.Infof("Deployment %s can be deleted.", deploy.Name)
				admissionResponse.Allowed = true
			}

		} else {
			admissionResponse.Allowed = true
		}

		glog.Infof("END~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")

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

// func (c *types.Client) getKSGateway(gatewayName string) bool {
// 	// set request GC
// 	gvr := schema.GroupVersionResource{
// 		Group:    "gateway.kubesphere.io",
// 		Version:  "v1alpha1",
// 		Resource: "gateways",
// 	}

// 	c.dynamicClient.Resource(gvr).Get()

// 	return true
// }
