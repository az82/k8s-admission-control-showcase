package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/golang/glog"
	"gopkg.in/resty.v1"
	k8sAdmission "k8s.io/api/admission/v1beta1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const httpPort = 8080
const httpsPort = 8443
const opaUrl = "http://localhost:8181/v1/data/admissioncontrol/deployment/policy"

const certFile = ".certs/tls.crt"
const keyFile = ".certs/tls.key"

const contentTypeHeader = "Content-Type"
const acceptHeader = "Accept"
const jsonContentType = "application/json"

// AdmissionPolicy as returned by OPA
type AdmissionPolicy struct {
	// Indicates whether or not the admission request should be permitted
	Allow bool `json:"allow"`

	// Human-readable message detailing the reasons admission was denied
	// Ignored if Allow is true
	Reason string `json:"reason,omitempty"`

	// List of URLs deployment processors that should be called with this deployment
	// Each processor will be called with the deployments AdmissionReview as request body
	// Ignored if Allow is false
	Processors []url.URL `json:"processors,omitempty"`
}

// OpaAdmissionRequest is a wrapper for sending k8sAdmission.AdmissionReview to OPA
type OpaAdmissionRequest struct {
	Review k8sAdmission.AdmissionReview `json:"input"`
}

// OpaAdmissionResponse is a wrapper for sending AdmissionPolicy to OPA
type OpaAdmissionResponse struct {
	Policy AdmissionPolicy `json:"result"`
}

// Get the admission policy (from OPA)
func getAdmissionPolicy(admissionReview k8sAdmission.AdmissionReview) (*AdmissionPolicy, error) {
	opaRequest := OpaAdmissionRequest{
		Review: admissionReview,
	}

	requestBody, err := json.Marshal(opaRequest)
	if err != nil {
		return nil, err
	}

	response, err := resty.R().
		SetHeader(contentTypeHeader, jsonContentType).
		SetHeader(acceptHeader, jsonContentType).
		SetBody(requestBody).
		Post(opaUrl)
	if err != nil {
		return nil, err
	}

	opaResponse := &OpaAdmissionResponse{}
	err = json.Unmarshal(response.Body(), opaResponse)
	if err != nil {
		return nil, err
	}

	return &opaResponse.Policy, nil
}

func addAdmissionResponse(admissionReview *k8sAdmission.AdmissionReview, admissionPolicy AdmissionPolicy) {
	var status = &k8sMeta.Status{}

	if admissionPolicy.Allow {
		status.Status = "Success" // As prescribed by the k8sMeta.Status spec
	} else {
		status.Status = "Failure" // As prescribed by the k8sMeta.Status spec
		status.Message = admissionPolicy.Reason
		// HTTP Status is 403 Forbidden
		status.Reason = k8sMeta.StatusReasonForbidden
		status.Code = http.StatusForbidden
	}

	admissionReview.Response = &k8sAdmission.AdmissionResponse{
		Allowed: admissionPolicy.Allow,
		Result:  status,
		// Copy UID as required by the spec
		UID: admissionReview.Request.UID,
	}
}

func getRequestBody(req *http.Request) []byte {
	if req.Body != nil {
		data, err := ioutil.ReadAll(req.Body)
		if err == nil {
			return data
		}
	}

	return []byte{}
}

func badRequest(response http.ResponseWriter, msg string, args ...interface{}) {
	glog.Errorf(msg, args...)
	http.Error(response, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func internalError(response http.ResponseWriter, args ...interface{}) {
	glog.Error(args...)
	http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func validate(responseWriter http.ResponseWriter, request *http.Request) {
	// Check content type
	contentType := request.Header.Get(contentTypeHeader)
	if contentType != jsonContentType {
		badRequest(responseWriter, "Unexpected content type \"%s\"", contentType)
		return
	}

	requestBody := getRequestBody(request)
	glog.Info("Review request ", string(requestBody))

	// Deserialize request
	admissionReview := k8sAdmission.AdmissionReview{}
	err := json.Unmarshal(requestBody, &admissionReview)
	if err != nil {
		badRequest(responseWriter, "Error deserializing request: %s", err)
		return
	}

	admissionPolicy, err := getAdmissionPolicy(admissionReview)
	if err != nil {
		internalError(responseWriter, "Is OPA running?", err)
		return
	}

	addAdmissionResponse(&admissionReview, *admissionPolicy)

	responseBody, err := json.Marshal(admissionReview)
	if err != nil {
		internalError(responseWriter, err)
		return
	}
	glog.Info("Review response ", string(responseBody))

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

	// Serve HTTPs
	// Skip if no certificate are provided for easy debugging
	if _, err := os.Stat(".certs/tls.crt"); !os.IsNotExist(err) {
		glog.Infof("Listening on HTTPs port %d", httpsPort)
		httpsServer = &http.Server{Addr: fmt.Sprintf(":%d", httpsPort)}
		go func() {
			err := httpsServer.ListenAndServeTLS(certFile, keyFile)
			if err != nil {
				glog.Fatal(err)
			}
		}()
	} else {
		glog.Error("No certificate found. No HTTPs.")
	}

	// Serve HTTP for debugging
	glog.Infof("Listening on HTTP port %d", httpPort)
	httpServer = &http.Server{Addr: fmt.Sprintf(":%d", httpPort)}
	err := httpServer.ListenAndServe()
	if err != nil {
		glog.Fatal(err)
	}

}
