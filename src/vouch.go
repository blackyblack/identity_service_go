package main

// VouchEvent represents a stored vouch.
type VouchEvent struct {
	From string
	To   string
	// TODO: Add Timestamp field
	// TODO: maybe store proof for external verification?
}

// VouchHandler handles vouch requests
func VouchHandler(state *AppState, from string, _signature string, _nonce string, to string) IdentityError {
	state.AddVouch(VouchEvent{
		From: from,
		To:   to,
	})
	return nil
}
