package main

import "testing"

func TestBuildVouchGraphEmpty(t *testing.T) {
	graph := BuildVouchGraph(nil)
	if got := len(graph.Nodes); got != 0 {
		t.Fatalf("expected empty graph, got %d nodes", got)
	}
}

func TestBuildVouchGraphEdges(t *testing.T) {
	v1 := VouchEvent{From: "alice", To: "bob"}
	v2 := VouchEvent{From: "bob", To: "carol"}
	v3 := VouchEvent{From: "alice", To: "carol"}

	graph := BuildVouchGraph([]VouchEvent{v1, v2, v3})

	if got := len(graph.Nodes); got != 3 {
		t.Fatalf("expected 3 nodes, got %d", got)
	}

	alice := graph.Nodes["alice"]
	bob := graph.Nodes["bob"]
	carol := graph.Nodes["carol"]
	if alice == nil || bob == nil || carol == nil {
		t.Fatal("expected nodes for alice, bob, and carol")
	}

	if got := len(alice.Outgoing); got != 2 {
		t.Fatalf("expected 2 outgoing edges for alice, got %d", got)
	}
	if got := len(alice.Incoming); got != 0 {
		t.Fatalf("expected 0 incoming edges for alice, got %d", got)
	}
	if alice.Outgoing[0].Event != v1 || alice.Outgoing[0].Peer != bob {
		t.Fatalf("unexpected alice outgoing edge 0: %#v", alice.Outgoing[0])
	}
	if alice.Outgoing[1].Event != v3 || alice.Outgoing[1].Peer != carol {
		t.Fatalf("unexpected alice outgoing edge 1: %#v", alice.Outgoing[1])
	}

	if got := len(bob.Incoming); got != 1 {
		t.Fatalf("expected 1 incoming edge for bob, got %d", got)
	}
	if got := len(bob.Outgoing); got != 1 {
		t.Fatalf("expected 1 outgoing edge for bob, got %d", got)
	}
	if bob.Incoming[0].Event != v1 || bob.Incoming[0].Peer != alice {
		t.Fatalf("unexpected bob incoming edge: %#v", bob.Incoming[0])
	}
	if bob.Outgoing[0].Event != v2 || bob.Outgoing[0].Peer != carol {
		t.Fatalf("unexpected bob outgoing edge: %#v", bob.Outgoing[0])
	}

	if got := len(carol.Incoming); got != 2 {
		t.Fatalf("expected 2 incoming edges for carol, got %d", got)
	}
	if got := len(carol.Outgoing); got != 0 {
		t.Fatalf("expected 0 outgoing edges for carol, got %d", got)
	}
	if carol.Incoming[0].Event != v2 || carol.Incoming[0].Peer != bob {
		t.Fatalf("unexpected carol incoming edge 0: %#v", carol.Incoming[0])
	}
	if carol.Incoming[1].Event != v3 || carol.Incoming[1].Peer != alice {
		t.Fatalf("unexpected carol incoming edge 1: %#v", carol.Incoming[1])
	}
}

func TestVouchGraphOutgoingTreeDepth(t *testing.T) {
	v1 := VouchEvent{From: "alice", To: "bob"}
	v2 := VouchEvent{From: "bob", To: "carol"}
	v3 := VouchEvent{From: "alice", To: "dan"}

	graph := BuildVouchGraph([]VouchEvent{v1, v2, v3})

	tree := graph.OutgoingTree("alice", 1)
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

	treeDepth2 := graph.OutgoingTree("alice", 2)
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

	graph := BuildVouchGraph([]VouchEvent{v1, v2, v3})

	tree := graph.OutgoingTree("alice", 3)
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

	graph := BuildVouchGraph([]VouchEvent{v1, v2, v3})

	tree := graph.OutgoingTree("root", 2)
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

	graph := BuildVouchGraph([]VouchEvent{v1, v2, v3})

	tree := graph.IncomingTree("bob", 2)
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

	graph := BuildVouchGraph([]VouchEvent{v1, v2, v3})

	tree := graph.IncomingTree("bob", 3)
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

	graph := BuildVouchGraph([]VouchEvent{v1, v2, v3})

	tree := graph.IncomingTree("root", 2)
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
