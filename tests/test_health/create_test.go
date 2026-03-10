package testhealth

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/snappy-fix-golang/tests"
)

func TestPostHealth(t *testing.T) {
	// Set up the router and controller
	router, _ := SetupTestRouter()

	// Prepare a dummy request body (adjust this if your endpoint expects actual data)
	body := bytes.NewBuffer([]byte(`{}`))

	// Create POST request
	req, err := http.NewRequest(http.MethodPost, "/api/v1/health", body)
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Record the response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert the response status
	tests.AssertStatusCode(t, w.Code, http.StatusOK)

	// Parse the response
	response := tests.ParseResponse(w)

	// Extract and assert the message
	message, ok := response["message"].(string)
	if !ok {
		t.Fatalf("Expected 'message' field in response to be string, got %T", response["message"])
	}
	tests.AssertResponseMessage(t, message, "ping successful")
}
