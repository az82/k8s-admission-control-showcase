package main

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/resty.v1"
	"io/ioutil"
	k8sAdmission "k8s.io/api/admission/v1beta1"
	"k8s.io/klog/glog"
	"net/http"
	"testing"
	"time"
)

const opaPort = 8181
const admissionControllerUrl = "http://localhost:8080/validate"
const okAdmissionReviewFile = "test/admission-review/ok.json"

func TestHttpService(t *testing.T) {
	// Before
	startServer()
	startOpaMock()

	// Given
	requestBody, err := ioutil.ReadFile(okAdmissionReviewFile)
	if err != nil {
		glog.Fatal(err)
		t.FailNow()
	}
	requestReview := k8sAdmission.AdmissionReview{}
	err = json.Unmarshal(requestBody, &requestReview)
	if err != nil {
		glog.Fatal(err)
		t.FailNow()
	}

	// When
	response, err := resty.R().
		SetHeader(contentTypeHeader, jsonContentType).
		SetHeader(acceptHeader, jsonContentType).
		SetBody(requestBody).
		Post(admissionControllerUrl)
	if err != nil {
		assert.Fail(t, "Error executing request", err)
	}

	responseReview := k8sAdmission.AdmissionReview{}
	err = json.Unmarshal(response.Body(), &responseReview)
	if err != nil {
		assert.Fail(t, "Error parsing request", err)
	}

	// Then
	assert.Equal(t, response.StatusCode(), http.StatusOK)
	assert.NotNil(t, responseReview.Response)
	assert.Equal(t, responseReview.Response.UID, requestReview.Request.UID)
	assert.NotNil(t, responseReview.Response.Allowed)
	assert.NotNil(t, responseReview.Response.Result)

	// Cleanup
	shutdownServer()
	shutdownOpaMock()
}

func opaMock(responseWriter http.ResponseWriter, request *http.Request) {
	contentType := request.Header.Get(contentTypeHeader)
	if contentType != jsonContentType {
		badRequest(responseWriter, "Unexpected content type \"%s\"", contentType)
		return
	}

	requestBody := getRequestBody(request)

	admissionRequest := OpaAdmissionRequest{}
	err := json.Unmarshal(requestBody, &admissionRequest)
	if err != nil {
		badRequest(responseWriter, "Error deserializing request: %s", err)
		return
	}

	mockResponse := struct {
		Policy AdmissionPolicy `json:"result"`
	}{
		Policy: AdmissionPolicy{
			Allow: true,
		},
	}

	responseBody, err := json.Marshal(mockResponse)
	if err != nil {
		glog.Error(err)
	}

	responseWriter.Write(responseBody)
}

var opaMockServer *http.Server

func startOpaMock() {
	go func() {
		http.HandleFunc("/v1/data/admissioncontrol/deployment/policy", opaMock)

		opaMockServer = &http.Server{Addr: fmt.Sprintf(":%d", opaPort)}
		err := opaMockServer.ListenAndServe()
		if err != nil {
			glog.Fatal(err)
		}

	}()
	time.Sleep(500 * time.Millisecond)
}

func shutdownOpaMock() {
	opaMockServer.Shutdown(nil)
}

func startServer() {
	go func() {
		main()
	}()
	time.Sleep(500 * time.Millisecond)
}

func shutdownServer() {
	httpServer.Shutdown(nil)
	if httpsServer != nil {
		httpsServer.Shutdown(nil)
	}
}
