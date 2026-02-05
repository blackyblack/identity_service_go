package main

import "testing"

func TestNewAppStateEmpty(t *testing.T) {
	state := NewAppState()
	if state == nil {
		t.Fatal("expected non-nil state")
	}
	if got := len(state.Users()); got != 0 {
		t.Fatalf("expected no users, got %d", got)
	}
	proof, err := state.ProofRecord("alice")
	if err != nil || proof.User != "alice" || proof.Balance != 0 {
		t.Fatal("expected no proof record for alice")
	}
	if got := len(state.Penalties("alice")); got != 0 {
		t.Fatalf("expected no penalties for alice, got %d", got)
	}
}

func TestAppStateVouchesReturnsCopy(t *testing.T) {
	state := NewAppState()
	first := VouchEvent{From: "alice", To: "bob"}
	second := VouchEvent{From: "bob", To: "carol"}
	state.AddVouch(first)
	state.AddVouch(second)

	vouches := state.UserVouchesFrom("alice")
	if got := len(vouches); got != 1 {
		t.Fatalf("expected 1 vouch, got %d", got)
	}
	if vouches[0] != first {
		t.Fatalf("unexpected vouch order: %#v", vouches)
	}

	vouches[0] = VouchEvent{From: "mallory", To: "trent"}
	vouches = append(vouches, VouchEvent{From: "dan", To: "erin"})

	vouchesAfter := state.UserVouchesFrom("alice")
	if got := len(vouchesAfter); got != 1 {
		t.Fatalf("expected 1 vouch after copy mutation, got %d", got)
	}
	if vouchesAfter[0] != first {
		t.Fatalf("state mutated through copy: %#v", vouchesAfter)
	}
}

func TestAppStatePenaltiesReturnsCopy(t *testing.T) {
	state := NewAppState()
	first := PenaltyEvent{User: "alice", Amount: 10}
	second := PenaltyEvent{User: "alice", Amount: 20}
	state.AddPenalty(first)
	state.AddPenalty(second)

	penalties := state.Penalties("alice")
	if got := len(penalties); got != 2 {
		t.Fatalf("expected 2 penalties, got %d", got)
	}
	if penalties[0] != first || penalties[1] != second {
		t.Fatalf("unexpected penalty order: %#v", penalties)
	}

	penalties[0] = PenaltyEvent{User: "alice", Amount: 99}
	penalties = append(penalties, PenaltyEvent{User: "alice", Amount: 50})

	penaltiesAfter := state.Penalties("alice")
	if got := len(penaltiesAfter); got != 2 {
		t.Fatalf("expected 2 penalties after copy mutation, got %d", got)
	}
	if penaltiesAfter[0] != first || penaltiesAfter[1] != second {
		t.Fatalf("state mutated through copy: %#v", penaltiesAfter)
	}
}

func TestAppStateSetProofReplacesExisting(t *testing.T) {
	state := NewAppState()

	state.SetProof(ProofEvent{User: "alice", Balance: 10})
	state.SetProof(ProofEvent{User: "alice", Balance: 25})

	proof, err := state.ProofRecord("alice")
	if err != nil {
		t.Fatal("expected proof record for alice")
	}
	if proof.Balance != 25 {
		t.Fatalf("expected latest balance 25, got %d", proof.Balance)
	}
}

func TestAppStateModerationBalanceSubtractsPenalties(t *testing.T) {
	state := NewAppState()

	// TODO: update test when penalty decay is implemented

	state.SetProof(ProofEvent{User: "alice", Balance: 100})
	state.AddPenalty(PenaltyEvent{User: "alice", Amount: 10})
	state.AddPenalty(PenaltyEvent{User: "alice", Amount: 15})
	state.AddPenalty(PenaltyEvent{User: "bob", Amount: 50})

	if got := state.ModerationBalance("alice"); got != 75 {
		t.Fatalf("expected moderated balance 75, got %d", got)
	}
}

func TestAppStateModerationBalanceNoProofRecord(t *testing.T) {
	state := NewAppState()

	// User with no proof record should have base balance of 0
	if got := state.ModerationBalance("alice"); got != 0 {
		t.Fatalf("expected moderated balance 0, got %d", got)
	}

	// User with no proof record but with penalties should have negative balance
	state.AddPenalty(PenaltyEvent{User: "bob", Amount: 10})
	state.AddPenalty(PenaltyEvent{User: "bob", Amount: 20})

	if got := state.ModerationBalance("bob"); got != -30 {
		t.Fatalf("expected moderated balance -30, got %d", got)
	}
}
