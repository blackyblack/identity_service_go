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

// ProofRequest represents the request body for the prove endpoint
type ProofRequest struct {
	User    string `json:"user"`
	Balance uint64 `json:"balance"`

	// TODO: add proof field
	// TODO: add moderator's credentials
}

// PunishRequest represents the request body for the punish endpoint
type PunishRequest struct {
	User   string `json:"user"`
	Amount uint64 `json:"amount"`

	// TODO: add punish reason field
	// TODO: add moderator's credentials
}

// AnyResponse represents the common response
type AnyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// IdtResponse represents the response for the idt endpoint
type IdtResponse struct {
	User    string `json:"user"`
	Balance int64  `json:"balance"`
	Penalty uint64 `json:"penalty"`
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

// proveHandler handles POST requests to /prove
func proveHandler(state *AppState, w http.ResponseWriter, r *http.Request) {
	var req ProofRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.User == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	res := ProveHandler(state, req.User, req.Balance)
	if res != nil {
		sendErrorResponse(w, http.StatusBadRequest, res.Error())
		return
	}

	data, err := json.Marshal(AnyResponse{Success: true, Message: "Proof accepted"})
	if err != nil {
		log.Printf("Failed to encode proof response to JSON: %v", err)
		sendInternalError(w)
		return
	}
	w.Write(data)
}

// punishHandler handles POST requests to /punish
func punishHandler(state *AppState, w http.ResponseWriter, r *http.Request) {
	var req PunishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.User == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	res := PunishHandler(state, req.User, req.Amount)
	if res != nil {
		sendErrorResponse(w, http.StatusBadRequest, res.Error())
		return
	}

	data, err := json.Marshal(AnyResponse{Success: true, Message: "Punish accepted"})
	if err != nil {
		log.Printf("Failed to encode punish response to JSON: %v", err)
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
	response := IdtResponse{User: res.User, Balance: res.Balance, Penalty: res.Penalty}
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
	router.HandleFunc("/prove", func(w http.ResponseWriter, r *http.Request) {
		proveHandler(appState, w, r)
	}).Methods("POST")
	router.HandleFunc("/punish", func(w http.ResponseWriter, r *http.Request) {
		punishHandler(appState, w, r)
	}).Methods("POST")
	router.HandleFunc("/idt/{user}", func(w http.ResponseWriter, r *http.Request) {
		idtHandler(appState, w, r)
	}).Methods("GET")
	return router
}
