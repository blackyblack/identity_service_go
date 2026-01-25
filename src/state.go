package main

import "sync"

// AppState stores in-memory application data.
type AppState struct {
	mu      sync.RWMutex
	vouches []VouchEvent
}

// NewAppState initializes an empty application state.
func NewAppState() *AppState {
	return &AppState{
		vouches: make([]VouchEvent, 0),
	}
}

// AddVouch records an incoming vouch event.
func (s *AppState) AddVouch(vouch VouchEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vouches = append(s.vouches, vouch)
}

// Vouches returns a copy of all stored vouches.
func (s *AppState) Vouches() []VouchEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vouchesCopy := make([]VouchEvent, len(s.vouches))
	copy(vouchesCopy, s.vouches)
	return vouchesCopy
}

// VouchGraph builds a vouch graph from the stored vouches.
func (s *AppState) VouchGraph() VouchGraph {
	return BuildVouchGraph(s.Vouches())
}
