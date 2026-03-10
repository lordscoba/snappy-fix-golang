package testhealth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/snappy-fix-golang/tests"
)

func TestGetHealth(t *testing.T) {
	// Set up the Gin router and controller
	router, _ := SetupTestRouter()

	// Create the test HTTP GET request
	req, err := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	if err != nil {
		t.Fatalf("Failed to create GET request: %v", err)
	}

	// Create a response recorder to capture the result
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert the HTTP status code
	tests.AssertStatusCode(t, w.Code, http.StatusOK)

	// Parse the JSON response body
	response := tests.ParseResponse(w)

	// Extract and assert the response message
	message, ok := response["message"].(string)
	if !ok {
		t.Fatalf("Expected 'message' field in response to be string, got %T", response["message"])
	}
	tests.AssertResponseMessage(t, message, "ping successful")
}
