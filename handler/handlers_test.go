package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DerRomtester/testproject-golang-webapi-1/handler"
	"github.com/DerRomtester/testproject-golang-webapi-1/model"
)

func TestCheckAuthValidJson_Success(t *testing.T) {
	validCreds := model.UserCredentials{
		Username: "test_user",
		Password: "password123",
	}

	requestBody, err := json.Marshal(&validCreds)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(requestBody))
	creds, apiError, err := handler.CheckAuthValidJson(req)

	// Verify successful decoding
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if apiError.Message != "" {
		t.Errorf("Unexpected API error: %v", apiError)
	}

	// Assert that the decoded credentials match the original
	if creds.Username != validCreds.Username || creds.Password != validCreds.Password {
		t.Errorf("Expected credentials: %v, got: %v", validCreds, creds)
	}
}

func TestCheckAuthValidJson_InvalidJson(t *testing.T) {
	invalidBody := []byte("{\"username\": \"test_user\"")
	req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(invalidBody))

	_, apiError, err := handler.CheckAuthValidJson(req)

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
	if apiError.Message != "structure of request is wrong" {
		t.Errorf("Expected specific API error message, got: %v", apiError.Message)
	}
}

func TestCheckAuthValidJson_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/auth", nil)
	_, _, err := handler.CheckAuthValidJson(req)

	if err == nil {
		t.Error("Expected error for emtpty request body")
	}
}
