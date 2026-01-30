package main

import "log"

// AppState wraps a Storage implementation to provide application-level operations.
type AppState struct {
	storage Storage
}

// NewAppState initializes an application state with in-memory storage.
func NewAppState() *AppState {
	return &AppState{
		storage: NewMemoryStorage(),
	}
}

// NewAppStateWithStorage initializes an application state with a specific storage implementation.
func NewAppStateWithStorage(storage Storage) *AppState {
	return &AppState{
		storage: storage,
	}
}

// AddVouch records an incoming vouch event.
func (s *AppState) AddVouch(vouch VouchEvent) {
	if err := s.storage.AddVouch(vouch); err != nil {
		log.Printf("Error adding vouch: %v", err)
	}
}

// Vouches returns all stored vouches.
func (s *AppState) Vouches() []VouchEvent {
	vouches, err := s.storage.Vouches()
	if err != nil {
		log.Printf("Error getting vouches: %v", err)
		return []VouchEvent{}
	}
	return vouches
}

// VouchGraph builds a vouch graph from the stored vouches.
func (s *AppState) VouchGraph() VouchGraph {
	return BuildVouchGraph(s.Vouches())
}

// SetProof stores the latest proof event for a user, replacing any prior record.
func (s *AppState) SetProof(proof ProofEvent) {
	if err := s.storage.SetProof(proof); err != nil {
		log.Printf("Error setting proof: %v", err)
	}
}

// ProofRecord returns the stored proof event for a user, if any.
func (s *AppState) ProofRecord(user string) (ProofEvent, bool) {
	proof, ok, err := s.storage.ProofRecord(user)
	if err != nil {
		log.Printf("Error getting proof record: %v", err)
		return ProofEvent{}, false
	}
	return proof, ok
}

// AddPenalty records a penalty event.
func (s *AppState) AddPenalty(penalty PenaltyEvent) {
	if err := s.storage.AddPenalty(penalty); err != nil {
		log.Printf("Error adding penalty: %v", err)
	}
}

// Penalties returns all stored penalties for a user.
func (s *AppState) Penalties(user string) []PenaltyEvent {
	penalties, err := s.storage.Penalties(user)
	if err != nil {
		log.Printf("Error getting penalties: %v", err)
		return []PenaltyEvent{}
	}
	return penalties
}

// ModerationBalance computes a user's balance as the proof balance minus penalties.
// If no proof record exists, the base balance is 0.
func (s *AppState) ModerationBalance(user string) int64 {
	base := int64(0)
	if proof, ok := s.ProofRecord(user); ok {
		base = int64(proof.Balance)
	}

	penaltySum := int64(0)
	// TODO: add penalty decay based on timestamp
	for _, penalty := range s.Penalties(user) {
		penaltySum += int64(penalty.Amount)
	}

	return base - penaltySum
}

// Close releases any resources used by the storage.
func (s *AppState) Close() error {
	return s.storage.Close()
}
