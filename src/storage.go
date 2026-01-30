package main

// Storage defines the interface for storing vouches, proofs, and penalties.
// Both in-memory and persistent implementations should satisfy this interface.
type Storage interface {
	// AddVouch records an incoming vouch event.
	AddVouch(vouch VouchEvent) error

	// Vouches returns all stored vouches.
	Vouches() ([]VouchEvent, error)

	// SetProof stores the latest proof event for a user, replacing any prior record.
	SetProof(proof ProofEvent) error

	// ProofRecord returns the stored proof event for a user, if any.
	ProofRecord(user string) (ProofEvent, bool, error)

	// AddPenalty records a penalty event.
	AddPenalty(penalty PenaltyEvent) error

	// Penalties returns all penalties for a user.
	Penalties(user string) ([]PenaltyEvent, error)

	// Close releases any resources used by the storage.
	Close() error
}
