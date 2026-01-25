package main

// User's identity information
type IdtInfo struct {
	User    string
	Balance int
}

// IdtHandler handles identity requests
func IdtHandler(state *AppState, _user string) (IdtInfo, IdentityError) {
	return IdtInfo{}, nil
}
