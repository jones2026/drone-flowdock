package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestExitWithBadResponseStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	var raw []byte

	origLogFatalf := logFatalf

	// After this test, replace the original fatal function
	defer func() { logFatalf = origLogFatalf }()

	var actualError string
	logFatalf = func(format string, args ...interface{}) {
		if len(args) > 0 {
			actualError = fmt.Sprintf(format, args)
		} else {
			actualError = format
		}
	}
	var origURL = apiURL
	defer func() { apiURL = origURL }()
	apiURL = ts.URL
	client := &http.Client{}
	postMessage(client, raw)
	expectedMessageError := "Failed to post message, flowdock api returned: [503 Service Unavailable]"
	if actualError != expectedMessageError {
		t.Fatalf("Expected error:\n%s\ngot:\n%s", expectedMessageError, actualError)
	}

	var fileName = "some_file"
	f, _ := os.Create(fileName)
	f.Close()
	defer os.Remove(f.Name())
	uploadFile(client, f, "some_thread_id")
	expectedUploadError := "Failed to post file: [503 Service Unavailable]"
	if actualError != expectedUploadError {
		t.Fatalf("Expected error:\n%s\ngot:\n%s", expectedUploadError, actualError)
	}

}

func TestRequiredFlowToken(t *testing.T) {
	os.Clearenv()
	var actualError error
	os.Setenv("PLUGIN_MESSAGE", "some_value")
	actualError = fetchSettings()
	expectedError := errors.New("required key PLUGIN_FLOW_TOKEN missing value")
	if actualError == nil {
		t.Fatalf("Expected error, but got none")
	} else if actualError.Error() != expectedError.Error() {
		t.Fatalf("Expected error:\n%s\ngot:\n%s", expectedError, actualError)
	}
}

func TestRequiredMessage(t *testing.T) {
	os.Clearenv()
	var actualError error
	os.Setenv("PLUGIN_FLOW_TOKEN", "some_value")
	actualError = fetchSettings()
	expectedError := errors.New("required key PLUGIN_MESSAGE missing value")
	if actualError == nil {
		t.Fatalf("Expected error, but got none")
	} else if actualError.Error() != expectedError.Error() {
		t.Fatalf("Expected error:\n%s\ngot:\n%s", expectedError, actualError)
	}
}

func TestFlowdockRequestSetup(t *testing.T) {
	expectedURL := "my_awesome_api_url"
	expectedByteBuffer := bytes.NewBuffer([]byte("some message body"))
	var origURL = apiURL
	defer func() { apiURL = origURL }()
	apiURL = expectedURL
	actualRequest := getFlowdockRequest(expectedByteBuffer)

	if actualRequest.Header.Get("X-flowdock-wait-for-message") != "true" {
		t.Fatalf("Expected header X-flowdock-wait-for-message to be true, but was not")
	}
	if actualRequest.URL.String() != expectedURL {
		t.Fatalf("Expected URL to be %s, instead was %s", expectedURL, actualRequest.URL.String())
	}
}
