package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/adapters/db/postgresql"
	"github.com/snappy-fix-golang/internal/config"
	"github.com/snappy-fix-golang/internal/domain/migrations"
	logutil "github.com/snappy-fix-golang/pkg/logger"
)

func Setup() *logutil.Logger {

	//Warning !!!!! Do not recreate this action anywhere on the app
	logger, err := logutil.NewLogger()
	if err != nil {
		// If logger initialization fails, we can't use it.
		// Panic here so the application fails fast.
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	config := config.Setup(logger, "../../app")

	postgresql.ConnectToDatabase(logger, config.TestDatabase)
	db := db.Connection()
	if config.TestDatabase.Migrate {
		migrations.RunAllMigrations(context.Background(), db)
		// fix correct seed call
		// seed.SeedDatabase(db.Postgresql)
	}
	return logger
}

func ParseResponse(w *httptest.ResponseRecorder) map[string]interface{} {
	res := make(map[string]interface{})
	json.NewDecoder(w.Body).Decode(&res)
	return res
}

func AssertStatusCode(t *testing.T, got, expected int) {
	if got != expected {
		t.Errorf("handler returned wrong status code: got status %d expected status %d", got, expected)
	}
}

func AssertResponseMessage(t *testing.T, got, expected string) {
	if got != expected {
		t.Errorf("handler returned wrong message: got message: %q expected: %q", got, expected)
	}
}
func AssertBool(t *testing.T, got, expected bool) {
	if got != expected {
		t.Errorf("handler returned wrong boolean: got %v expected %v", got, expected)
	}
}

func AssertValidationError(t *testing.T, response map[string]interface{}, field string, expectedMessage string) {
	errors, ok := response["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'error' field in response")
	}

	errorMessage, exists := errors[field]
	if !exists {
		t.Fatalf("expected validation error message for field '%s'", field)
	}

	if errorMessage != expectedMessage {
		t.Errorf("unexpected error message for field '%s': got %v, want %v", field, errorMessage, expectedMessage)
	}
}
