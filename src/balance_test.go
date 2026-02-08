package main

import (
	"testing"
	"time"
)

func TestPenaltyBuildsOutgoingTree(t *testing.T) {
	state := NewAppState()
	timestamp := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	state.now = func() time.Time { return timestamp }

	v1 := VouchEvent{From: "alice", To: "bob", Timestamp: timestamp}
	v2 := VouchEvent{From: "bob", To: "carol", Timestamp: timestamp}

	state.AddVouch(v1)
	state.AddVouch(v2)

	state.AddPenalty(PenaltyEvent{User: "alice", Amount: 5, Timestamp: timestamp})
	state.AddPenalty(PenaltyEvent{User: "bob", Amount: 100, Timestamp: timestamp})
	state.AddPenalty(PenaltyEvent{User: "carol", Amount: 1000, Timestamp: timestamp})

	got := Penalty(state, "alice", nil)
	if got != 25 {
		t.Fatalf("expected penalty 25, got %d", got)
	}
}

func TestPenaltyUsesProvidedTree(t *testing.T) {
	state := NewAppState()
	timestamp := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	state.now = func() time.Time { return timestamp }

	v1 := VouchEvent{From: "alice", To: "bob", Timestamp: timestamp}
	v2 := VouchEvent{From: "bob", To: "carol", Timestamp: timestamp}

	state.AddVouch(v1)
	state.AddVouch(v2)

	state.AddPenalty(PenaltyEvent{User: "alice", Amount: 5, Timestamp: timestamp})
	state.AddPenalty(PenaltyEvent{User: "bob", Amount: 100, Timestamp: timestamp})
	state.AddPenalty(PenaltyEvent{User: "carol", Amount: 1000, Timestamp: timestamp})

	tree := OutgoingTree(state, "alice", 1)

	got := Penalty(state, "alice", tree)
	if got != 15 {
		t.Fatalf("expected penalty 15, got %d", got)
	}

	// Without provided tree, full depth is used
	got = Penalty(state, "alice", nil)
	if got != 25 {
		t.Fatalf("expected penalty 25, got %d", got)
	}
}

func TestPenaltySumsUserPenalties(t *testing.T) {
	state := NewAppState()
	timestamp := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	state.now = func() time.Time { return timestamp }

	state.AddPenalty(PenaltyEvent{User: "alice", Amount: 10, Timestamp: timestamp})
	state.AddPenalty(PenaltyEvent{User: "alice", Amount: 7, Timestamp: timestamp.Add(1 * time.Hour)})

	got := Penalty(state, "alice", nil)
	if got != 17 {
		t.Fatalf("expected penalty 17, got %d", got)
	}
}

func TestPenaltyUserNotInGraphUsesDirectPenalties(t *testing.T) {
	state := NewAppState()
	timestamp := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	state.now = func() time.Time { return timestamp }

	state.AddVouch(VouchEvent{From: "alice", To: "bob", Timestamp: timestamp})
	state.AddPenalty(PenaltyEvent{User: "mallory", Amount: 12, Timestamp: timestamp})
	state.AddPenalty(PenaltyEvent{User: "alice", Amount: 50, Timestamp: timestamp})

	tree := OutgoingTree(state, "mallory", DefaultTreeDepth)
	if got := len(tree.Peers); got != 0 {
		t.Fatalf("expected no outgoing peers for mallory, got %d", got)
	}

	got := Penalty(state, "mallory", nil)
	if got != 12 {
		t.Fatalf("expected penalty 12, got %d", got)
	}
}

func TestPenaltyEmptyStateNoPenalties(t *testing.T) {
	state := NewAppState()

	got := Penalty(state, "alice", nil)
	if got != 0 {
		t.Fatalf("expected penalty 0, got %d", got)
	}
}

func TestPenaltyInheritsFromVouchedUsers(t *testing.T) {
	state := NewAppState()
	timestamp := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	state.now = func() time.Time { return timestamp }

	state.AddVouch(VouchEvent{From: "alice", To: "bob", Timestamp: timestamp})
	state.AddVouch(VouchEvent{From: "alice", To: "carol", Timestamp: timestamp})
	state.AddPenalty(PenaltyEvent{User: "bob", Amount: 50, Timestamp: timestamp})
	state.AddPenalty(PenaltyEvent{User: "carol", Amount: 70, Timestamp: timestamp})
	got := Penalty(state, "alice", nil)
	if got != 12 {
		t.Fatalf("expected penalty 12, got %d", got)
	}
}

func TestBalanceEmptyStateNoProof(t *testing.T) {
	state := NewAppState()

	got := Balance(state, "alice", nil)
	if got != 0 {
		t.Fatalf("expected balance 0, got %d", got)
	}
}

func TestBalanceUserNotInGraphUsesDirectPenalties(t *testing.T) {
	state := NewAppState()
	timestamp := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	state.now = func() time.Time { return timestamp }

	state.AddVouch(VouchEvent{From: "alice", To: "bob", Timestamp: timestamp})
	state.SetProof(ProofEvent{User: "mallory", Balance: 40, Timestamp: timestamp})
	state.AddPenalty(PenaltyEvent{User: "mallory", Amount: 12, Timestamp: timestamp})
	state.AddPenalty(PenaltyEvent{User: "alice", Amount: 50, Timestamp: timestamp})
	tree := IncomingTree(state, "mallory", DefaultTreeDepth)
	if got := len(tree.Peers); got != 0 {
		t.Fatalf("expected no incoming peers for mallory, got %d", got)
	}

	got := Balance(state, "mallory", nil)
	if got != 28 {
		t.Fatalf("expected balance 28, got %d", got)
	}
}

func TestBalanceNegativeWhenPenaltiesExceedProof(t *testing.T) {
	state := NewAppState()
	timestamp := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	state.now = func() time.Time { return timestamp }

	state.AddVouch(VouchEvent{From: "bob", To: "alice", Timestamp: timestamp})
	state.SetProof(ProofEvent{User: "bob", Balance: 10, Timestamp: timestamp})
	state.AddPenalty(PenaltyEvent{User: "bob", Amount: 50, Timestamp: timestamp})

	got := Balance(state, "bob", nil)
	if got != -40 {
		t.Fatalf("expected balance -40, got %d", got)
	}
}

func TestBalanceBuildsIncomingTree(t *testing.T) {
	state := NewAppState()
	timestamp := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	state.now = func() time.Time { return timestamp }

	state.AddVouch(VouchEvent{From: "alice", To: "bob", Timestamp: timestamp})
	state.AddVouch(VouchEvent{From: "carol", To: "bob", Timestamp: timestamp})

	state.SetProof(ProofEvent{User: "alice", Balance: 100, Timestamp: timestamp})
	state.SetProof(ProofEvent{User: "carol", Balance: 50, Timestamp: timestamp})
	state.SetProof(ProofEvent{User: "bob", Balance: 10, Timestamp: timestamp})

	got := Balance(state, "bob", nil)
	if got != 25 {
		t.Fatalf("expected balance 25, got %d", got)
	}
}

func TestBalanceUsesProvidedTree(t *testing.T) {
	state := NewAppState()
	timestamp := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	state.now = func() time.Time { return timestamp }

	v1 := VouchEvent{From: "dan", To: "carol", Timestamp: timestamp}
	v2 := VouchEvent{From: "carol", To: "bob", Timestamp: timestamp}

	state.AddVouch(v1)
	state.AddVouch(v2)

	state.SetProof(ProofEvent{User: "dan", Balance: 1000, Timestamp: timestamp})
	state.SetProof(ProofEvent{User: "carol", Balance: 100, Timestamp: timestamp})
	state.SetProof(ProofEvent{User: "bob", Balance: 10, Timestamp: timestamp})

	tree := IncomingTree(state, "bob", 1)

	got := Balance(state, "bob", tree)
	if got != 20 {
		t.Fatalf("expected balance 20, got %d", got)
	}

	// Without provided tree, full depth is used
	got = Balance(state, "bob", nil)
	if got != 30 {
		t.Fatalf("expected balance 30, got %d", got)
	}
}

func TestBalanceAppliesPenaltyAggregation(t *testing.T) {
	state := NewAppState()
	timestamp := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	state.now = func() time.Time { return timestamp }

	state.AddVouch(VouchEvent{From: "carol", To: "bob", Timestamp: timestamp})

	state.SetProof(ProofEvent{User: "bob", Balance: 100, Timestamp: timestamp})
	state.SetProof(ProofEvent{User: "carol", Balance: 100, Timestamp: timestamp})

	state.AddPenalty(PenaltyEvent{User: "bob", Amount: 50, Timestamp: timestamp})

	got := Balance(state, "bob", nil)
	if got != 59 {
		t.Fatalf("expected balance 59, got %d", got)
	}
}

func TestBalanceAppliesTransientPenaltyFromOutgoingPeers(t *testing.T) {
	state := NewAppState()
	timestamp := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	state.now = func() time.Time { return timestamp }

	state.AddVouch(VouchEvent{From: "alice", To: "bob", Timestamp: timestamp})
	state.AddVouch(VouchEvent{From: "alice", To: "mallory", Timestamp: timestamp})

	state.SetProof(ProofEvent{User: "alice", Balance: 100, Timestamp: timestamp})
	state.AddPenalty(PenaltyEvent{User: "mallory", Amount: 100, Timestamp: timestamp})
	got := Balance(state, "bob", nil)
	// alice's balance = 100 - 10% of mallory's 100 penalty = 90
	// bob's balance = 10% of alice's balance = 9
	if got != 9 {
		t.Fatalf("expected balance 9, got %d", got)
	}
}

func TestBalanceUsesTopVoucherBalances(t *testing.T) {
	state := NewAppState()
	timestamp := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	state.now = func() time.Time { return timestamp }

	vouchers := []struct {
		user    string
		balance uint64
	}{
		{user: "alice", balance: 10},
		{user: "bruce", balance: 20},
		{user: "carol", balance: 30},
		{user: "dana", balance: 40},
		{user: "erin", balance: 50},
		{user: "frank", balance: 100},
	}

	for _, v := range vouchers {
		state.AddVouch(VouchEvent{From: v.user, To: "bob", Timestamp: timestamp})
		state.SetProof(ProofEvent{User: v.user, Balance: v.balance, Timestamp: timestamp})
	}

	got := Balance(state, "bob", nil)
	// Top 5 voucher balances: 100 + 50 + 40 + 30 + 20 = 240; weighted by 10% = 24
	if got != 24 {
		t.Fatalf("expected balance 24, got %d", got)
	}
}

func TestBalanceIgnoresNegativeVoucherBalance(t *testing.T) {
	state := NewAppState()
	timestamp := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	state.now = func() time.Time { return timestamp }

	state.AddVouch(VouchEvent{From: "mallory", To: "bob", Timestamp: timestamp})
	state.SetProof(ProofEvent{User: "mallory", Balance: 10, Timestamp: timestamp})
	state.AddPenalty(PenaltyEvent{User: "mallory", Amount: 50, Timestamp: timestamp})
	got := Balance(state, "bob", nil)
	if got != 0 {
		t.Fatalf("expected balance 0, got %d", got)
	}
}
