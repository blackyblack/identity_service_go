package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// VouchRequest represents the request body for the vouch endpoint
type VouchRequest struct {
	From      string `json:"from"`
	Signature string `json:"signature"`
	Nonce     string `json:"nonce"`
	To        string `json:"to"`
}

// VouchResponse represents the response for the vouch endpoint
type VouchResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// IdtResponse represents the response for the idt endpoint
type IdtResponse struct {
	User string `json:"user"`
}

// VouchHandler handles POST requests to /vouch
func VouchHandler(w http.ResponseWriter, r *http.Request) {
	var req VouchRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(VouchResponse{Success: false, Message: "Invalid JSON"})
		return
	}
	
	// Validate required fields
	if req.From == "" || req.Signature == "" || req.Nonce == "" || req.To == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(VouchResponse{Success: false, Message: "Missing required fields"})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(VouchResponse{Success: true, Message: "Vouch accepted"})
}

// IdtHandler handles GET requests to /idt/:user
func IdtHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]
	
	if user == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "User parameter is required"})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IdtResponse{User: user})
}

// SetupRouter creates and configures the HTTP router
func SetupRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/vouch", VouchHandler).Methods("POST")
	router.HandleFunc("/idt/{user}", IdtHandler).Methods("GET")
	return router
}

func main() {
	router := SetupRouter()
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
