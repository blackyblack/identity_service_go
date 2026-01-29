package main

import "testing"

func TestPenaltyBuildsOutgoingTree(t *testing.T) {
	state := NewAppState()

	v1 := VouchEvent{From: "alice", To: "bob"}
	v2 := VouchEvent{From: "bob", To: "carol"}

	state.AddVouch(v1)
	state.AddVouch(v2)

	state.AddPenalty(PenaltyEvent{User: "alice", Amount: 5})
	state.AddPenalty(PenaltyEvent{User: "bob", Amount: 100})
	state.AddPenalty(PenaltyEvent{User: "carol", Amount: 1000})

	got := penalty(state, "alice", nil)
	if got != 25 {
		t.Fatalf("expected penalty 25, got %d", got)
	}
}

func TestPenaltyUsesProvidedTree(t *testing.T) {
	state := NewAppState()

	v1 := VouchEvent{From: "alice", To: "bob"}
	v2 := VouchEvent{From: "bob", To: "carol"}

	state.AddVouch(v1)
	state.AddVouch(v2)

	state.AddPenalty(PenaltyEvent{User: "alice", Amount: 5})
	state.AddPenalty(PenaltyEvent{User: "bob", Amount: 100})
	state.AddPenalty(PenaltyEvent{User: "carol", Amount: 1000})

	graph := BuildVouchGraph([]VouchEvent{v1, v2})
	tree := graph.OutgoingTree("alice", 1)

	got := penalty(state, "alice", tree)
	if got != 15 {
		t.Fatalf("expected penalty 15, got %d", got)
	}

	// Without provided tree, full depth is used
	got = penalty(state, "alice", nil)
	if got != 25 {
		t.Fatalf("expected penalty 25, got %d", got)
	}
}

func TestPenaltySumsUserPenalties(t *testing.T) {
	state := NewAppState()
	state.AddPenalty(PenaltyEvent{User: "alice", Amount: 10})
	state.AddPenalty(PenaltyEvent{User: "alice", Amount: 7})

	got := penalty(state, "alice", nil)
	if got != 17 {
		t.Fatalf("expected penalty 17, got %d", got)
	}
}

func TestPenaltyUserNotInGraphUsesDirectPenalties(t *testing.T) {
	state := NewAppState()
	state.AddVouch(VouchEvent{From: "alice", To: "bob"})
	state.AddPenalty(PenaltyEvent{User: "mallory", Amount: 12})
	state.AddPenalty(PenaltyEvent{User: "alice", Amount: 50})

	graph := state.VouchGraph()
	if _, ok := graph.Nodes["mallory"]; ok {
		t.Fatal("expected mallory to be absent from vouch graph")
	}
	tree := graph.OutgoingTree("mallory", DefaultTreeDepth)
	if got := len(tree.Peers); got != 0 {
		t.Fatalf("expected no outgoing peers for mallory, got %d", got)
	}

	got := penalty(state, "mallory", nil)
	if got != 12 {
		t.Fatalf("expected penalty 12, got %d", got)
	}
}

func TestPenaltyEmptyStateNoPenalties(t *testing.T) {
	state := NewAppState()

	got := penalty(state, "alice", nil)
	if got != 0 {
		t.Fatalf("expected penalty 0, got %d", got)
	}
}

func TestPenaltyInheritsFromVouchedUsers(t *testing.T) {
	state := NewAppState()
	state.AddVouch(VouchEvent{From: "alice", To: "bob"})
	state.AddVouch(VouchEvent{From: "alice", To: "carol"})
	state.AddPenalty(PenaltyEvent{User: "bob", Amount: 50})
	state.AddPenalty(PenaltyEvent{User: "carol", Amount: 70})

	got := penalty(state, "alice", nil)
	if got != 12 {
		t.Fatalf("expected penalty 12, got %d", got)
	}
}

func TestBalanceEmptyStateNoProof(t *testing.T) {
	state := NewAppState()

	got := balance(state, "alice", nil)
	if got != 0 {
		t.Fatalf("expected balance 0, got %d", got)
	}
}

func TestBalanceUserNotInGraphUsesDirectPenalties(t *testing.T) {
	state := NewAppState()
	state.AddVouch(VouchEvent{From: "alice", To: "bob"})
	state.SetProof(ProofEvent{User: "mallory", Balance: 40})
	state.AddPenalty(PenaltyEvent{User: "mallory", Amount: 12})
	state.AddPenalty(PenaltyEvent{User: "alice", Amount: 50})

	graph := state.VouchGraph()
	if _, ok := graph.Nodes["mallory"]; ok {
		t.Fatal("expected mallory to be absent from vouch graph")
	}
	tree := graph.IncomingTree("mallory", DefaultTreeDepth)
	if got := len(tree.Peers); got != 0 {
		t.Fatalf("expected no incoming peers for mallory, got %d", got)
	}

	got := balance(state, "mallory", nil)
	if got != 28 {
		t.Fatalf("expected balance 28, got %d", got)
	}
}

func TestBalanceNegativeWhenPenaltiesExceedProof(t *testing.T) {
	state := NewAppState()
	state.AddVouch(VouchEvent{From: "bob", To: "alice"})

	state.SetProof(ProofEvent{User: "bob", Balance: 10})
	state.AddPenalty(PenaltyEvent{User: "bob", Amount: 50})

	got := balance(state, "bob", nil)
	if got != -40 {
		t.Fatalf("expected balance -40, got %d", got)
	}
}

func TestBalanceBuildsIncomingTree(t *testing.T) {
	state := NewAppState()

	state.AddVouch(VouchEvent{From: "alice", To: "bob"})
	state.AddVouch(VouchEvent{From: "carol", To: "bob"})

	state.SetProof(ProofEvent{User: "alice", Balance: 100})
	state.SetProof(ProofEvent{User: "carol", Balance: 50})
	state.SetProof(ProofEvent{User: "bob", Balance: 10})

	got := balance(state, "bob", nil)
	if got != 25 {
		t.Fatalf("expected balance 25, got %d", got)
	}
}

func TestBalanceUsesProvidedTree(t *testing.T) {
	state := NewAppState()

	v1 := VouchEvent{From: "dan", To: "carol"}
	v2 := VouchEvent{From: "carol", To: "bob"}

	state.AddVouch(v1)
	state.AddVouch(v2)

	state.SetProof(ProofEvent{User: "dan", Balance: 1000})
	state.SetProof(ProofEvent{User: "carol", Balance: 100})
	state.SetProof(ProofEvent{User: "bob", Balance: 10})

	graph := state.VouchGraph()
	tree := graph.IncomingTree("bob", 1)

	got := balance(state, "bob", tree)
	if got != 20 {
		t.Fatalf("expected balance 20, got %d", got)
	}

	// Without provided tree, full depth is used
	got = balance(state, "bob", nil)
	if got != 30 {
		t.Fatalf("expected balance 30, got %d", got)
	}
}

func TestBalanceAppliesPenaltyAggregation(t *testing.T) {
	state := NewAppState()
	state.AddVouch(VouchEvent{From: "carol", To: "bob"})

	state.SetProof(ProofEvent{User: "bob", Balance: 100})
	state.SetProof(ProofEvent{User: "carol", Balance: 100})

	state.AddPenalty(PenaltyEvent{User: "bob", Amount: 50})

	got := balance(state, "bob", nil)
	if got != 59 {
		t.Fatalf("expected balance 59, got %d", got)
	}
}

func TestBalanceAppliesTransientPenaltyFromOutgoingPeers(t *testing.T) {
	state := NewAppState()

	state.AddVouch(VouchEvent{From: "alice", To: "bob"})
	state.AddVouch(VouchEvent{From: "alice", To: "mallory"})

	state.SetProof(ProofEvent{User: "alice", Balance: 100})
	state.AddPenalty(PenaltyEvent{User: "mallory", Amount: 100})

	got := balance(state, "bob", nil)
	// alice's balance = 100 - 10% of mallory's 100 penalty = 90
	// bob's balance = 10% of alice's balance = 9
	if got != 9 {
		t.Fatalf("expected balance 9, got %d", got)
	}
}
