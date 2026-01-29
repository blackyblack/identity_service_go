package main

// User's identity information
type IdtInfo struct {
	User    string
	Balance int64
	Penalty uint64
}

// IdtHandler handles identity requests
func IdtHandler(state *AppState, user string) (IdtInfo, IdentityError) {
	userBalance := balance(state, user, nil)
	userPenalty := penalty(state, user, nil)
	return IdtInfo{User: user, Balance: userBalance, Penalty: userPenalty}, nil
}
