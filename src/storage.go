package main

// Defines the interface for storing vouches, proofs, and penalties.
// Both in-memory and persistent implementations should satisfy this interface.
type Storage interface {
	// Records an incoming vouch event.
	AddVouch(vouch VouchEvent) error

	// Returns all users who have vouches, proofs, or penalties recorded.
	Users() ([]string, error)

	// Returns all stored outgoing vouches for a specific user.
	UserVouchesFrom(user string) ([]VouchEvent, error)

	// Returns all stored incoming vouches for a specific user.
	UserVouchesTo(user string) ([]VouchEvent, error)

	// Stores the latest proof event for a user, replacing any prior record.
	SetProof(proof ProofEvent) error

	// Returns the stored proof event for a user, if any.
	ProofRecord(user string) (ProofEvent, error)

	// Records a penalty event.
	AddPenalty(penalty PenaltyEvent) error

	// Returns all penalties for a user.
	Penalties(user string) ([]PenaltyEvent, error)

	// Releases any resources used by the storage.
	Close() error
}
