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
