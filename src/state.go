package main

import (
	"log"
	"time"
)

// Wraps a Storage implementation to provide application-level operations.
type AppState struct {
	storage Storage
	now     func() time.Time
}

// Returns the current time. Uses the overridable now function if set,
// otherwise defaults to time.Now().UTC().
func (s *AppState) currentTime() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now().UTC()
}

// Initializes an application state with in-memory storage.
func NewAppState() *AppState {
	return &AppState{
		storage: NewMemoryStorage(),
	}
}

// Initializes an application state with a specific storage implementation.
func NewAppStateWithStorage(storage Storage) *AppState {
	return &AppState{
		storage: storage,
	}
}

// Returns all users.
func (s *AppState) Users() []string {
	users, err := s.storage.Users()
	if err != nil {
		log.Printf("Error getting users: %v", err)
		return []string{}
	}
	return users
}

// Records an incoming vouch event.
func (s *AppState) AddVouch(vouch VouchEvent) {
	if err := s.storage.AddVouch(vouch); err != nil {
		log.Printf("Error adding vouch: %v", err)
	}
}

func (s *AppState) UserVouchesFrom(user string) []VouchEvent {
	vouches, err := s.storage.UserVouchesFrom(user)
	if err != nil {
		log.Printf("Error getting vouches from user %s: %v", user, err)
		return []VouchEvent{}
	}
	return vouches
}

func (s *AppState) UserVouchesTo(user string) []VouchEvent {
	vouches, err := s.storage.UserVouchesTo(user)
	if err != nil {
		log.Printf("Error getting vouches to user %s: %v", user, err)
		return []VouchEvent{}
	}
	return vouches
}

// Stores the latest proof event for a user, replacing any prior record.
func (s *AppState) SetProof(proof ProofEvent) {
	if err := s.storage.SetProof(proof); err != nil {
		log.Printf("Error setting proof: %v", err)
	}
}

// Returns the stored proof event for a user, if any.
func (s *AppState) ProofRecord(user string) (ProofEvent, error) {
	proof, err := s.storage.ProofRecord(user)
	if err != nil {
		return ProofEvent{}, err
	}
	return proof, nil
}

// Records a penalty event.
func (s *AppState) AddPenalty(penalty PenaltyEvent) {
	if err := s.storage.AddPenalty(penalty); err != nil {
		log.Printf("Error adding penalty: %v", err)
	}
}

// Returns all stored penalties for a user.
func (s *AppState) Penalties(user string) []PenaltyEvent {
	penalties, err := s.storage.Penalties(user)
	if err != nil {
		log.Printf("Error getting penalties: %v", err)
		return []PenaltyEvent{}
	}
	return penalties
}

// Releases any resources used by the storage.
func (s *AppState) Close() error {
	return s.storage.Close()
}
