package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"gopkg.in/resty.v1"
	"io/ioutil"
	k8sAdmission "k8s.io/api/admission/v1beta1"
	"k8s.io/klog/glog"
	"net/http"
	"testing"
	"time"
)

func TestHttpService(t *testing.T) {
	// Before
	startServer()

	// Given
	requestBody, err := ioutil.ReadFile("test-admission-review.json")
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
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetBody(requestBody).
		Post("http://localhost:8080/validate")
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
}

func startServer() {
	go func() {
		main()
	}()
	time.Sleep(1 * time.Second)
}

func shutdownServer() {
	httpServer.Shutdown(nil)
	if httpsServer != nil {
		httpsServer.Shutdown(nil)
	}
}
