package main

import "sync"

// AppState stores in-memory application data.
type AppState struct {
	mu      sync.RWMutex
	vouches []VouchEvent
	proofs  map[string]ProofEvent
	// Penalties per user
	penalties map[string][]PenaltyEvent

	// TODO: maybe keep vouch graph cached here?
}

// NewAppState initializes an empty application state.
func NewAppState() *AppState {
	return &AppState{
		vouches:   make([]VouchEvent, 0),
		proofs:    make(map[string]ProofEvent),
		penalties: make(map[string][]PenaltyEvent),
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

// SetProof stores the latest proof event for a user, replacing any prior record.
func (s *AppState) SetProof(proof ProofEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proofs[proof.User] = proof
}

// ProofRecord returns the stored proof event for a user, if any.
func (s *AppState) ProofRecord(user string) (ProofEvent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	proof, ok := s.proofs[user]
	return proof, ok
}

// AddPenalty records a penalty event.
func (s *AppState) AddPenalty(penalty PenaltyEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.penalties[penalty.User] = append(s.penalties[penalty.User], penalty)
}

// Penalties returns a copy of all stored penalties per user.
func (s *AppState) Penalties(user string) []PenaltyEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	penaltiesCopy := make([]PenaltyEvent, len(s.penalties[user]))
	copy(penaltiesCopy, s.penalties[user])
	return penaltiesCopy
}

// ModerationBalance computes a user's balance as the proof balance minus penalties.
// If no proof record exists, the base balance is 0.
func (s *AppState) ModerationBalance(user string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	base := int64(0)
	if proof, ok := s.proofs[user]; ok {
		base = int64(proof.Balance)
	}

	penaltySum := int64(0)
	// TODO: add penalty decay based on timestamp
	for _, penalty := range s.penalties[user] {
		penaltySum += int64(penalty.Amount)
	}

	return base - penaltySum
}
