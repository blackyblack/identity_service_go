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
	
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/vouch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	VouchHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var resp VouchResponse
	json.NewDecoder(w.Body).Decode(&resp)
	
	if !resp.Success {
		t.Errorf("Expected success to be true, got false")
	}
	
	if resp.Message != "Vouch accepted" {
		t.Errorf("Expected message 'Vouch accepted', got '%s'", resp.Message)
	}
}

// TestVouchHandler_MissingFields tests the vouch endpoint with missing fields
func TestVouchHandler_MissingFields(t *testing.T) {
	reqBody := VouchRequest{
		From:      "user1",
		Signature: "sig123",
		// Missing nonce and to
	}
	
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/vouch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	VouchHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	var resp VouchResponse
	json.NewDecoder(w.Body).Decode(&resp)
	
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
	
	VouchHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	var resp VouchResponse
	json.NewDecoder(w.Body).Decode(&resp)
	
	if resp.Success {
		t.Errorf("Expected success to be false, got true")
	}
}

// TestIdtHandler_Success tests the idt endpoint with a valid user parameter
func TestIdtHandler_Success(t *testing.T) {
	router := SetupRouter()
	
	req := httptest.NewRequest("GET", "/idt/testuser", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var resp IdtResponse
	json.NewDecoder(w.Body).Decode(&resp)
	
	if resp.User != "testuser" {
		t.Errorf("Expected user 'testuser', got '%s'", resp.User)
	}
}

// TestIdtHandler_MultipleUsers tests the idt endpoint with different user parameters
func TestIdtHandler_MultipleUsers(t *testing.T) {
	router := SetupRouter()
	
	users := []string{"alice", "bob", "charlie123"}
	
	for _, user := range users {
		req := httptest.NewRequest("GET", "/idt/"+user, nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d for user %s, got %d", http.StatusOK, user, w.Code)
		}
		
		var resp IdtResponse
		json.NewDecoder(w.Body).Decode(&resp)
		
		if resp.User != user {
			t.Errorf("Expected user '%s', got '%s'", user, resp.User)
		}
	}
}

// TestVouchHandler_AllFields tests vouch endpoint with all fields populated
func TestVouchHandler_AllFields(t *testing.T) {
	reqBody := VouchRequest{
		From:      "alice@example.com",
		Signature: "0x123456789abcdef",
		Nonce:     "random-nonce-12345",
		To:        "bob@example.com",
	}
	
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/vouch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	VouchHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var resp VouchResponse
	json.NewDecoder(w.Body).Decode(&resp)
	
	if !resp.Success {
		t.Errorf("Expected success to be true")
	}
}

// TestSetupRouter tests that the router is configured correctly
func TestSetupRouter(t *testing.T) {
	router := SetupRouter()
	
	if router == nil {
		t.Error("Expected router to be non-nil")
	}
	
	// Test that POST to /vouch works
	vouchReq := VouchRequest{
		From:      "test",
		Signature: "test",
		Nonce:     "test",
		To:        "test",
	}
	body, _ := json.Marshal(vouchReq)
	req := httptest.NewRequest("POST", "/vouch", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected vouch endpoint to work, got status %d", w.Code)
	}
	
	// Test that GET to /idt/:user works
	req2 := httptest.NewRequest("GET", "/idt/testuser", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	
	if w2.Code != http.StatusOK {
		t.Errorf("Expected idt endpoint to work, got status %d", w2.Code)
	}
}
