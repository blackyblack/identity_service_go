package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestVouchHandler_Success tests the vouch endpoint with valid input
func TestVouchHandler_Success(t *testing.T) {
	reqBody := VouchRequest{
		From:      "user1",
		Signature: "sig123",
		Nonce:     "nonce456",
		To:        "user2",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	req := httptest.NewRequest("POST", "/vouch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	vouchHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp AnyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("Expected success to be true, got false")
	}

	if resp.Message != "Vouch accepted" {
		t.Fatalf("Expected message 'Vouch accepted', got '%s'", resp.Message)
	}
}

// TestVouchHandler_MissingFields tests the vouch endpoint with missing fields
func TestVouchHandler_MissingFields(t *testing.T) {
	reqBody := VouchRequest{
		From:      "user1",
		Signature: "sig123",
		// Missing nonce and to
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	req := httptest.NewRequest("POST", "/vouch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	vouchHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp AnyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Success {
		t.Errorf("Expected success to be false, got true")
	}

	if resp.Message != "Missing required fields" {
		t.Errorf("Expected message 'Missing required fields', got '%s'", resp.Message)
	}
}

// TestVouchHandler_InvalidJSON tests the vouch endpoint with invalid JSON
func TestVouchHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/vouch", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	vouchHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp AnyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Success {
		t.Errorf("Expected success to be false, got true")
	}
}

// TestIdtHandler_Success tests the idt endpoint with a valid user parameter
func TestIdtHandler_Success(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("GET", "/idt/testuser", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp IdtResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.User != "testuser" {
		t.Errorf("Expected user 'testuser', got '%s'", resp.User)
	}
}
