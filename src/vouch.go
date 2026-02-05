package main

import "time"

// Handles vouch requests
func VouchHandler(state *AppState, from string, _signature string, _nonce string, to string) IdentityError {
	state.AddVouch(VouchEvent{
		From:      from,
		To:        to,
		Timestamp: time.Now().UTC(),
	})
	return nil
}
