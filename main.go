package main

import (
	"encoding/json"
	"flag"
	"github.com/golang/glog"
	"io/ioutil"
	k8sAdmission "k8s.io/api/admission/v1beta1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"
)

func getBody(req *http.Request) []byte {
	if req.Body != nil {
		data, err := ioutil.ReadAll(req.Body)
		if err == nil {
			return data
		}
	}

	return []byte{}
}

func badRequest(response http.ResponseWriter, msg string, args ...interface{}) {
	glog.Error(msg, args)
	http.Error(response, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func internalError(response http.ResponseWriter, args ...interface{}) {
	glog.Error(args...)
	http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func validate(responseWriter http.ResponseWriter, request *http.Request) {
	// Check content type
	contentType := request.Header.Get("Content-Type")
	if contentType != "application/json" {
		badRequest(responseWriter, "Unexpected content type \"%s\"", contentType)
		return
	}

	// Deserialize request
	requestReview := k8sAdmission.AdmissionReview{}
	err := json.Unmarshal(getBody(request), &requestReview)
	if err != nil {
		badRequest(responseWriter, "Error deserializing request: %s", err)
		return
	}

	// Create response
	var admissionResponse = &k8sAdmission.AdmissionResponse{
		Allowed: false,
		Result: &k8sMeta.Status{
			Status:  "Failure",
			Message: "You're not getting in with these shoes.",
			Reason:  k8sMeta.StatusReasonForbidden,
			Code:    http.StatusForbidden,
		},
		// Copy UID as required by the spec
		// TODO: May cause nil pointer deref
		UID: requestReview.Request.UID,
	}

	responseReview := k8sAdmission.AdmissionReview{
		Response: admissionResponse,
	}

	responseBody, err := json.Marshal(responseReview)
	if err != nil {
		internalError(responseWriter, err)
		return
	}

	_, err = responseWriter.Write(responseBody)
	if err != nil {
		internalError(responseWriter, err)
	}

}

var httpServer *http.Server
var httpsServer *http.Server

func main() {
	flag.Parse()

	http.HandleFunc("/validate", validate)

	// Listen on 8443 / TLS
	// Skip if no certificate are provided for easy debugging
	if _, err := os.Stat(".certs/tls.crt"); !os.IsNotExist(err) {
		glog.Info("Listening on HTTPs port 8443")
		httpsServer = &http.Server{Addr: ":8443"}
		go func() {
			err := httpsServer.ListenAndServeTLS(".certs/tls.crt", ".certs/tls.key")
			if err != nil {
				glog.Fatal(err)
			}
		}()
	} else {
		glog.Error("No certificate found. No HTTPs.")
	}

	// Listen on 8080 for debugging
	glog.Info("Listening on HTTP port 8080")
	httpServer = &http.Server{Addr: ":8080"}
	err := httpServer.ListenAndServe()
	if err != nil {
		glog.Fatal(err)
	}

}
