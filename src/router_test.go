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
	appState := NewAppState()
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

	vouchHandler(appState, w, req)

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
	appState := NewAppState()
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

	vouchHandler(appState, w, req)

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
	appState := NewAppState()
	req := httptest.NewRequest("POST", "/vouch", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	vouchHandler(appState, w, req)

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

func TestProveHandler_Success(t *testing.T) {
	appState := NewAppState()
	reqBody := ProofRequest{
		User:    "user1",
		Balance: 42,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	req := httptest.NewRequest("POST", "/prove", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	proveHandler(appState, w, req)

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
	if resp.Message != "Proof accepted" {
		t.Fatalf("Expected message 'Proof accepted', got '%s'", resp.Message)
	}

	proof, ok := appState.ProofRecord("user1")
	if !ok {
		t.Fatal("expected proof record for user1")
	}
	if proof.Balance != 42 {
		t.Fatalf("expected balance 42, got %d", proof.Balance)
	}
}

func TestPunishHandler_Success(t *testing.T) {
	appState := NewAppState()
	appState.SetProof(ProofEvent{User: "user1", Balance: 100})

	reqBody := PunishRequest{
		User:   "user1",
		Amount: 30,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	req := httptest.NewRequest("POST", "/punish", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	punishHandler(appState, w, req)

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
	if resp.Message != "Punish accepted" {
		t.Fatalf("Expected message 'Punish accepted', got '%s'", resp.Message)
	}

	if got := appState.ModerationBalance("user1"); got != 70 {
		t.Fatalf("expected moderated balance 70, got %d", got)
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
	if resp.Balance != 0 {
		t.Errorf("Expected balance 0, got %d", resp.Balance)
	}
}
