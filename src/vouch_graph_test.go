package main

import "testing"

func TestVouchGraphOutgoingTreeDepth(t *testing.T) {
	v1 := VouchEvent{From: "alice", To: "bob"}
	v2 := VouchEvent{From: "bob", To: "carol"}
	v3 := VouchEvent{From: "alice", To: "dan"}
	state := NewAppState()
	state.AddVouch(v1)
	state.AddVouch(v2)
	state.AddVouch(v3)

	tree := OutgoingTree(state, "alice", 1)
	if tree == nil {
		t.Fatal("expected non-nil outgoing tree")
	}
	if tree.User != "alice" {
		t.Fatalf("expected root user alice, got %q", tree.User)
	}
	if tree.Depth != 0 {
		t.Fatalf("expected root depth 0, got %d", tree.Depth)
	}
	if got := len(tree.Peers); got != 2 {
		t.Fatalf("expected 2 outgoing edges at depth 1, got %d", got)
	}
	if tree.Peers[0].Event != v1 {
		t.Fatalf("unexpected outgoing edge 0: %#v", tree.Peers[0])
	}
	if tree.Peers[0].Peer.User != "bob" {
		t.Fatalf("unexpected outgoing edge 0: %#v", tree.Peers[0])
	}
	if tree.Peers[0].Peer.Depth != 1 {
		t.Fatalf("unexpected outgoing edge 0: %#v", tree.Peers[0])
	}
	if got := len(tree.Peers[0].Peer.Peers); got != 0 {
		t.Fatalf("expected bob to have no outgoing edges at depth 1, got %d", got)
	}
	if tree.Peers[1].Event != v3 {
		t.Fatalf("unexpected outgoing edge 1: %#v", tree.Peers[1])
	}
	if tree.Peers[1].Peer.User != "dan" {
		t.Fatalf("unexpected outgoing edge 1: %#v", tree.Peers[1])
	}
	if tree.Peers[1].Peer.Depth != 1 {
		t.Fatalf("unexpected outgoing edge 1: %#v", tree.Peers[1])
	}
	if got := len(tree.Peers[1].Peer.Peers); got != 0 {
		t.Fatalf("expected dan to have no outgoing edges at depth 1, got %d", got)
	}

	treeDepth2 := OutgoingTree(state, "alice", 2)
	if got := len(treeDepth2.Peers); got != 2 {
		t.Fatalf("expected 2 outgoing edges at depth 2, got %d", got)
	}
	if got := len(treeDepth2.Peers[0].Peer.Peers); got != 1 {
		t.Fatalf("expected bob to have 1 outgoing edge at depth 2, got %d", got)
	}
	if treeDepth2.Peers[0].Peer.Peers[0].Event != v2 {
		t.Fatalf("unexpected bob outgoing edge: %#v", treeDepth2.Peers[0].Peer.Peers[0])
	}
	if treeDepth2.Peers[0].Peer.Peers[0].Peer.User != "carol" {
		t.Fatalf("unexpected bob outgoing edge: %#v", treeDepth2.Peers[0].Peer.Peers[0])
	}
	if treeDepth2.Peers[0].Peer.Peers[0].Peer.Depth != 2 {
		t.Fatalf("unexpected bob outgoing edge: %#v", treeDepth2.Peers[0].Peer.Peers[0])
	}
}

func TestVouchGraphOutgoingTreeCycle(t *testing.T) {
	v1 := VouchEvent{From: "alice", To: "bob"}
	v2 := VouchEvent{From: "bob", To: "alice"}
	v3 := VouchEvent{From: "bob", To: "carol"}
	state := NewAppState()
	state.AddVouch(v1)
	state.AddVouch(v2)
	state.AddVouch(v3)

	tree := OutgoingTree(state, "alice", 3)
	if got := len(tree.Peers); got != 1 {
		t.Fatalf("expected 1 outgoing edge for alice, got %d", got)
	}
	if tree.Peers[0].Event != v1 || tree.Peers[0].Peer.User != "bob" {
		t.Fatalf("unexpected alice outgoing edge: %#v", tree.Peers[0])
	}
	bob := tree.Peers[0].Peer
	if got := len(bob.Peers); got != 1 {
		t.Fatalf("expected cycle edge to be removed, got %d outgoing edges", got)
	}
	if bob.Peers[0].Event != v3 || bob.Peers[0].Peer.User != "carol" {
		t.Fatalf("unexpected bob outgoing edge: %#v", bob.Peers[0])
	}
}

func TestVouchGraphOutgoingTreeBranchIndependence(t *testing.T) {
	v1 := VouchEvent{From: "root", To: "alice"}
	v2 := VouchEvent{From: "root", To: "bob"}
	v3 := VouchEvent{From: "bob", To: "alice"}

	state := NewAppState()
	state.AddVouch(v1)
	state.AddVouch(v2)
	state.AddVouch(v3)

	tree := OutgoingTree(state, "root", 2)
	if got := len(tree.Peers); got != 2 {
		t.Fatalf("expected 2 outgoing edges for root, got %d", got)
	}
	if tree.Peers[0].Event != v1 || tree.Peers[0].Peer.User != "alice" {
		t.Fatalf("unexpected root outgoing edge 0: %#v", tree.Peers[0])
	}
	if got := len(tree.Peers[0].Peer.Peers); got != 0 {
		t.Fatalf("expected alice to have no outgoing edges at depth 2, got %d", got)
	}
	if tree.Peers[1].Event != v2 || tree.Peers[1].Peer.User != "bob" {
		t.Fatalf("unexpected root outgoing edge 1: %#v", tree.Peers[1])
	}
	bob := tree.Peers[1].Peer
	if got := len(bob.Peers); got != 1 {
		t.Fatalf("expected bob to have 1 outgoing edge at depth 2, got %d", got)
	}
	if bob.Peers[0].Event != v3 || bob.Peers[0].Peer.User != "alice" {
		t.Fatalf("unexpected bob outgoing edge: %#v", bob.Peers[0])
	}
}

func TestVouchGraphIncomingTreeDepth(t *testing.T) {
	v1 := VouchEvent{From: "alice", To: "bob"}
	v2 := VouchEvent{From: "carol", To: "bob"}
	v3 := VouchEvent{From: "dan", To: "carol"}

	state := NewAppState()
	state.AddVouch(v1)
	state.AddVouch(v2)
	state.AddVouch(v3)

	tree := IncomingTree(state, "bob", 2)
	if tree == nil {
		t.Fatal("expected non-nil incoming tree")
	}
	if tree.User != "bob" {
		t.Fatalf("expected root user bob, got %q", tree.User)
	}
	if tree.Depth != 0 {
		t.Fatalf("expected root depth 0, got %d", tree.Depth)
	}
	if got := len(tree.Peers); got != 2 {
		t.Fatalf("expected 2 incoming edges at depth 2, got %d", got)
	}
	if tree.Peers[0].Event != v1 {
		t.Fatalf("unexpected incoming edge 0: %#v", tree.Peers[0])
	}
	if tree.Peers[0].Peer.User != "alice" {
		t.Fatalf("unexpected incoming edge 0: %#v", tree.Peers[0])
	}
	if tree.Peers[0].Peer.Depth != 1 {
		t.Fatalf("unexpected incoming edge 0: %#v", tree.Peers[0])
	}
	if got := len(tree.Peers[0].Peer.Peers); got != 0 {
		t.Fatalf("expected alice to have no incoming edges at depth 2, got %d", got)
	}
	if tree.Peers[1].Event != v2 {
		t.Fatalf("unexpected incoming edge 1: %#v", tree.Peers[1])
	}
	if tree.Peers[1].Peer.User != "carol" {
		t.Fatalf("unexpected incoming edge 1: %#v", tree.Peers[1])
	}
	if tree.Peers[1].Peer.Depth != 1 {
		t.Fatalf("unexpected incoming edge 1: %#v", tree.Peers[1])
	}
	if got := len(tree.Peers[1].Peer.Peers); got != 1 {
		t.Fatalf("expected carol to have 1 incoming edge at depth 2, got %d", got)
	}
	if tree.Peers[1].Peer.Peers[0].Event != v3 {
		t.Fatalf("unexpected carol incoming edge: %#v", tree.Peers[1].Peer.Peers[0])
	}
	if tree.Peers[1].Peer.Peers[0].Peer.User != "dan" {
		t.Fatalf("unexpected carol incoming edge: %#v", tree.Peers[1].Peer.Peers[0])
	}
	if tree.Peers[1].Peer.Peers[0].Peer.Depth != 2 {
		t.Fatalf("unexpected carol incoming edge: %#v", tree.Peers[1].Peer.Peers[0])
	}
}

func TestVouchGraphIncomingTreeCycle(t *testing.T) {
	v1 := VouchEvent{From: "alice", To: "bob"}
	v2 := VouchEvent{From: "bob", To: "alice"}
	v3 := VouchEvent{From: "carol", To: "alice"}

	state := NewAppState()
	state.AddVouch(v1)
	state.AddVouch(v2)
	state.AddVouch(v3)

	tree := IncomingTree(state, "bob", 3)
	if got := len(tree.Peers); got != 1 {
		t.Fatalf("expected 1 incoming edge for bob, got %d", got)
	}
	if tree.Peers[0].Event != v1 || tree.Peers[0].Peer.User != "alice" {
		t.Fatalf("unexpected bob incoming edge: %#v", tree.Peers[0])
	}
	alice := tree.Peers[0].Peer
	if got := len(alice.Peers); got != 1 {
		t.Fatalf("expected cycle edge to be removed, got %d incoming edges", got)
	}
	if alice.Peers[0].Event != v3 || alice.Peers[0].Peer.User != "carol" {
		t.Fatalf("unexpected alice incoming edge: %#v", alice.Peers[0])
	}
}

func TestVouchGraphIncomingTreeBranchIndependence(t *testing.T) {
	v1 := VouchEvent{From: "alice", To: "root"}
	v2 := VouchEvent{From: "carol", To: "root"}
	v3 := VouchEvent{From: "alice", To: "carol"}

	state := NewAppState()
	state.AddVouch(v1)
	state.AddVouch(v2)
	state.AddVouch(v3)

	tree := IncomingTree(state, "root", 2)
	if got := len(tree.Peers); got != 2 {
		t.Fatalf("expected 2 incoming edges for root, got %d", got)
	}
	if tree.Peers[0].Event != v1 || tree.Peers[0].Peer.User != "alice" {
		t.Fatalf("unexpected root incoming edge 0: %#v", tree.Peers[0])
	}
	if got := len(tree.Peers[0].Peer.Peers); got != 0 {
		t.Fatalf("expected alice to have no incoming edges at depth 2, got %d", got)
	}
	if tree.Peers[1].Event != v2 || tree.Peers[1].Peer.User != "carol" {
		t.Fatalf("unexpected root incoming edge 1: %#v", tree.Peers[1])
	}
	carol := tree.Peers[1].Peer
	if got := len(carol.Peers); got != 1 {
		t.Fatalf("expected carol to have 1 incoming edge at depth 2, got %d", got)
	}
	if carol.Peers[0].Event != v3 || carol.Peers[0].Peer.User != "alice" {
		t.Fatalf("unexpected carol incoming edge: %#v", carol.Peers[0])
	}
}
