package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExitWithBadResponseStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	// var raw []byte

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

	// postMessage(raw, nil, ts.URL)
	expectedError := "Failed to post message, flowdock api returned: [503 Service Unavailable]"
	if actualError == expectedError {
		return
	}
	// t.Fatalf("Expected error:\n%s\ngot:\n%s", expectedError, actualError)
}
