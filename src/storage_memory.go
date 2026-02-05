package main

import (
	"sync"
)

// Implements Storage using in-memory data structures.
type MemoryStorage struct {
	mu sync.RWMutex
	// maps user to vouched users
	vouchesFrom map[string]map[string]VouchEvent
	// maps user to users who vouched for them
	vouchesTo map[string]map[string]VouchEvent
	proofs    map[string]ProofEvent
	penalties map[string][]PenaltyEvent
}

// Initializes an empty in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		vouchesFrom: make(map[string]map[string]VouchEvent),
		vouchesTo:   make(map[string]map[string]VouchEvent),
		proofs:      make(map[string]ProofEvent),
		penalties:   make(map[string][]PenaltyEvent),
	}
}

// Returns all users who have vouches, proofs, or penalties recorded.
func (s *MemoryStorage) Users() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userSet := make(map[string]struct{})
	for from := range s.vouchesFrom {
		userSet[from] = struct{}{}
		for to := range s.vouchesFrom[from] {
			userSet[to] = struct{}{}
		}
	}
	for user := range s.proofs {
		userSet[user] = struct{}{}
	}
	for user := range s.penalties {
		userSet[user] = struct{}{}
	}
	users := make([]string, 0, len(userSet))
	for user := range userSet {
		users = append(users, user)
	}
	return users, nil
}

// Records an incoming vouch event.
func (s *MemoryStorage) AddVouch(vouch VouchEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.vouchesFrom[vouch.From] == nil {
		s.vouchesFrom[vouch.From] = make(map[string]VouchEvent)
	}
	s.vouchesFrom[vouch.From][vouch.To] = vouch
	// Also update the reverse mapping
	if s.vouchesTo[vouch.To] == nil {
		s.vouchesTo[vouch.To] = make(map[string]VouchEvent)
	}
	s.vouchesTo[vouch.To][vouch.From] = vouch
	return nil
}

// Returns a copy of all stored outgoing vouches for a specific user.
func (s *MemoryStorage) UserVouchesFrom(user string) ([]VouchEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vouchesCopy := make([]VouchEvent, 0, len(s.vouchesFrom[user]))
	for _, vouch := range s.vouchesFrom[user] {
		vouchesCopy = append(vouchesCopy, vouch)
	}
	return vouchesCopy, nil
}

// Returns a copy of all stored incoming vouches for a specific user.
func (s *MemoryStorage) UserVouchesTo(user string) ([]VouchEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vouchesCopy := make([]VouchEvent, 0, len(s.vouchesTo[user]))
	for _, vouch := range s.vouchesTo[user] {
		vouchesCopy = append(vouchesCopy, vouch)
	}
	return vouchesCopy, nil
}

// Stores the latest proof event for a user, replacing any prior record.
func (s *MemoryStorage) SetProof(proof ProofEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proofs[proof.User] = proof
	return nil
}

// Returns the stored proof event for a user, if any.
func (s *MemoryStorage) ProofRecord(user string) (ProofEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	proof, ok := s.proofs[user]
	// Defaults to zero balance if no proof exists.
	if !ok {
		proof = ProofEvent{User: user}
	}
	return proof, nil
}

// Records a penalty event.
func (s *MemoryStorage) AddPenalty(penalty PenaltyEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.penalties[penalty.User] = append(s.penalties[penalty.User], penalty)
	return nil
}

// Returns a copy of all stored penalties for a user.
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
