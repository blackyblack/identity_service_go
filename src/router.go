package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const PORT int = 8080

// VouchRequest represents the request body for the vouch endpoint
type VouchRequest struct {
	From      string `json:"from"`
	Signature string `json:"signature"`
	Nonce     string `json:"nonce"`
	To        string `json:"to"`
}

// AnyResponse represents the common response
type AnyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// IdtResponse represents the response for the idt endpoint
type IdtResponse struct {
	User string `json:"user"`
}

func contentTypeApplicationJsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// sendErrorResponse sends a JSON error response with the given status code and message.
func sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	data, err := json.Marshal(AnyResponse{Success: false, Message: message})
	if err != nil {
		log.Printf("Failed to encode error response to JSON: %v", err)
		sendInternalError(w)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(data)
}

// sendInternalError sends a generic internal server error response.
func sendInternalError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
}

// vouchHandler handles POST requests to /vouch
func vouchHandler(state *AppState, w http.ResponseWriter, r *http.Request) {
	var req VouchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate required fields
	if req.From == "" || req.Signature == "" || req.Nonce == "" || req.To == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	res := VouchHandler(state, req.From, req.Signature, req.Nonce, req.To)
	if res != nil {
		sendErrorResponse(w, http.StatusBadRequest, res.Error())
		return
	}

	data, err := json.Marshal(AnyResponse{Success: true, Message: "Vouch accepted"})
	if err != nil {
		log.Printf("Failed to encode vouch response to JSON: %v", err)
		sendInternalError(w)
		return
	}
	w.Write(data)
}

// idtHandler handles GET requests to /idt/:user
func idtHandler(state *AppState, w http.ResponseWriter, r *http.Request) {
	user := mux.Vars(r)["user"]
	res, err := IdtHandler(state, user)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	response := IdtResponse{User: res.User}
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to encode idt response to JSON: %v", err)
		sendInternalError(w)
		return
	}
	w.Write(data)
}

// setupRouter creates and configures the HTTP router
func setupRouter() *mux.Router {
	appState := NewAppState()
	router := mux.NewRouter()
	router.Use(contentTypeApplicationJsonMiddleware)
	router.HandleFunc("/vouch", func(w http.ResponseWriter, r *http.Request) {
		vouchHandler(appState, w, r)
	}).Methods("POST")
	router.HandleFunc("/idt/{user}", func(w http.ResponseWriter, r *http.Request) {
		idtHandler(appState, w, r)
	}).Methods("GET")
	return router
}
