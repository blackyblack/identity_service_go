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
