package main

import (
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

	client := &http.Client{}
	postMessage(client, raw, ts.URL)
	expectedError := "Failed to post message, flowdock api returned: [503 Service Unavailable]"
	if actualError == expectedError {
		return
	}
	t.Fatalf("Expected error:\n%s\ngot:\n%s", expectedError, actualError)
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
