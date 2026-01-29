package main

// ProofEvent represents a moderation action that sets a user's balance.
// Only one proof record is stored per user; newer proofs replace older ones.
type ProofEvent struct {
	User    string
	Balance uint64
	// TODO: Add Timestamp field
}

// PenaltyEvent represents a moderation action that penalizes a user.
type PenaltyEvent struct {
	User   string
	Amount uint64
	// TODO: Add Timestamp field
}

// ProveHandler sets a user's balance by storing the latest proof record.
func ProveHandler(state *AppState, user string, balance uint64) IdentityError {
	state.SetProof(ProofEvent{
		User:    user,
		Balance: balance,
	})
	return nil
}

// PunishHandler records a penalty for the user.
func PunishHandler(state *AppState, user string, amount uint64) IdentityError {
	state.AddPenalty(PenaltyEvent{
		User:   user,
		Amount: amount,
	})
	return nil
}
