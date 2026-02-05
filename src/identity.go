package main

// User's identity information
type IdtInfo struct {
	User    string
	Balance int64
	Penalty uint64
}

// Handles identity requests
func IdtHandler(state *AppState, user string) (IdtInfo, IdentityError) {
	userBalance := Balance(state, user, nil)
	userPenalty := Penalty(state, user, nil)
	return IdtInfo{User: user, Balance: userBalance, Penalty: userPenalty}, nil
}
