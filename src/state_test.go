package main

import "testing"

func TestNewAppStateEmpty(t *testing.T) {
	state := NewAppState()
	if state == nil {
		t.Fatal("expected non-nil state")
	}
	if got := len(state.Vouches()); got != 0 {
		t.Fatalf("expected no vouches, got %d", got)
	}
	if got := len(state.VouchGraph().Nodes); got != 0 {
		t.Fatalf("expected no graph nodes, got %d", got)
	}
}

func TestAppStateVouchesReturnsCopy(t *testing.T) {
	state := NewAppState()
	first := VouchEvent{From: "alice", To: "bob"}
	second := VouchEvent{From: "bob", To: "carol"}
	state.AddVouch(first)
	state.AddVouch(second)

	vouches := state.Vouches()
	if got := len(vouches); got != 2 {
		t.Fatalf("expected 2 vouches, got %d", got)
	}
	if vouches[0] != first || vouches[1] != second {
		t.Fatalf("unexpected vouch order: %#v", vouches)
	}

	vouches[0] = VouchEvent{From: "mallory", To: "trent"}
	vouches = append(vouches, VouchEvent{From: "dan", To: "erin"})

	vouchesAfter := state.Vouches()
	if got := len(vouchesAfter); got != 2 {
		t.Fatalf("expected 2 vouches after copy mutation, got %d", got)
	}
	if vouchesAfter[0] != first || vouchesAfter[1] != second {
		t.Fatalf("state mutated through copy: %#v", vouchesAfter)
	}
}

func TestAppStateVouchGraphUsesCurrentState(t *testing.T) {
	state := NewAppState()
	vouch := VouchEvent{From: "alice", To: "bob"}
	state.AddVouch(vouch)

	graph := state.VouchGraph()
	if got := len(graph.Nodes); got != 2 {
		t.Fatalf("expected 2 graph nodes, got %d", got)
	}

	alice := graph.Nodes["alice"]
	bob := graph.Nodes["bob"]
	if alice == nil || bob == nil {
		t.Fatal("expected nodes for alice and bob")
	}

	// Other graph properties are tested in vouch_graph_test.go
}
