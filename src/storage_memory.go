package main

import "sync"

// MemoryStorage implements Storage using in-memory data structures.
type MemoryStorage struct {
	mu        sync.RWMutex
	vouches   []VouchEvent
	proofs    map[string]ProofEvent
	penalties map[string][]PenaltyEvent
}

// NewMemoryStorage initializes an empty in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		vouches:   make([]VouchEvent, 0),
		proofs:    make(map[string]ProofEvent),
		penalties: make(map[string][]PenaltyEvent),
	}
}

// AddVouch records an incoming vouch event.
func (s *MemoryStorage) AddVouch(vouch VouchEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vouches = append(s.vouches, vouch)
	return nil
}

// Vouches returns a copy of all stored vouches.
func (s *MemoryStorage) Vouches() ([]VouchEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vouchesCopy := make([]VouchEvent, len(s.vouches))
	copy(vouchesCopy, s.vouches)
	return vouchesCopy, nil
}

// SetProof stores the latest proof event for a user, replacing any prior record.
func (s *MemoryStorage) SetProof(proof ProofEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proofs[proof.User] = proof
	return nil
}

// ProofRecord returns the stored proof event for a user, if any.
func (s *MemoryStorage) ProofRecord(user string) (ProofEvent, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	proof, ok := s.proofs[user]
	return proof, ok, nil
}

// AddPenalty records a penalty event.
func (s *MemoryStorage) AddPenalty(penalty PenaltyEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.penalties[penalty.User] = append(s.penalties[penalty.User], penalty)
	return nil
}

// Penalties returns a copy of all stored penalties for a user.
func (s *MemoryStorage) Penalties(user string) ([]PenaltyEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	penaltiesCopy := make([]PenaltyEvent, len(s.penalties[user]))
	copy(penaltiesCopy, s.penalties[user])
	return penaltiesCopy, nil
}

// Close is a no-op for in-memory storage.
func (s *MemoryStorage) Close() error {
	return nil
}
