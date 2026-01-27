package main

import "maps"

// VouchGraphEdge represents a vouch event between two users.
type VouchGraphEdge struct {
	Event VouchEvent
	Peer  *VouchGraphNode
}

// VouchGraphNode represents a user in the vouch graph.
type VouchGraphNode struct {
	User     string
	Outgoing []VouchGraphEdge
	Incoming []VouchGraphEdge
}

// VouchGraph organizes users and their vouch relationships.
type VouchGraph struct {
	Nodes map[string]*VouchGraphNode
}

// VouchTreeEdge represents a vouch event in the vouch tree.
type VouchTreeEdge struct {
	Event VouchEvent
	Peer  *VouchTreeNode
}

// VouchTreeNode represents a user in the vouch tree (unidirectional).
type VouchTreeNode struct {
	User  string
	Peers []VouchTreeEdge
}

func NewVouchGraphNode(user string) VouchGraphNode {
	return VouchGraphNode{
		User:     user,
		Outgoing: []VouchGraphEdge{},
		Incoming: []VouchGraphEdge{},
	}
}

func NewVouchTreeNode(user string) VouchTreeNode {
	return VouchTreeNode{
		User:  user,
		Peers: []VouchTreeEdge{},
	}
}

// BuildVouchGraph constructs a graph where users are nodes and vouches are edges.
func BuildVouchGraph(vouches []VouchEvent) VouchGraph {
	nodes := make(map[string]*VouchGraphNode)

	getNode := func(user string) *VouchGraphNode {
		if node, ok := nodes[user]; ok {
			return node
		}
		node := NewVouchGraphNode(user)
		nodes[user] = &node
		return &node
	}

	for _, vouch := range vouches {
		fromNode := getNode(vouch.From)
		toNode := getNode(vouch.To)

		// Create edge
		outEdge := VouchGraphEdge{
			Event: vouch,
			Peer:  toNode,
		}
		inEdge := VouchGraphEdge{
			Event: vouch,
			Peer:  fromNode,
		}

		fromNode.Outgoing = append(fromNode.Outgoing, outEdge)
		toNode.Incoming = append(toNode.Incoming, inEdge)
	}

	return VouchGraph{
		Nodes: nodes,
	}
}

// OutgoingTree builds a depth-limited outgoing vouch tree rooted at the user.
func (g VouchGraph) OutgoingTree(user string, depth int) *VouchTreeNode {
	root := NewVouchTreeNode(user)
	root.Peers = buildOutgoingEdges(&g, user, depth, map[string]bool{})
	return &root
}

// IncomingTree builds a depth-limited incoming vouch tree rooted at the user.
func (g VouchGraph) IncomingTree(user string, depth int) *VouchTreeNode {
	root := NewVouchTreeNode(user)
	root.Peers = buildIncomingEdges(&g, user, depth, map[string]bool{})
	return &root
}

func buildOutgoingEdges(g *VouchGraph, user string, depth int, path map[string]bool) []VouchTreeEdge {
	if depth <= 0 {
		return []VouchTreeEdge{}
	}
	node := g.Nodes[user]
	if node == nil {
		return []VouchTreeEdge{}
	}
	path[user] = true
	edges := make([]VouchTreeEdge, 0, len(node.Outgoing))
	for _, edge := range node.Outgoing {
		peerUser := edge.Event.To
		if path[peerUser] {
			continue
		}
		peerNode := NewVouchTreeNode(peerUser)
		if depth > 1 {
			peerNode.Peers = buildOutgoingEdges(g, peerUser, depth-1, maps.Clone(path))
		}
		edges = append(edges, VouchTreeEdge{
			Event: edge.Event,
			Peer:  &peerNode,
		})
	}
	return edges
}

func buildIncomingEdges(g *VouchGraph, user string, depth int, path map[string]bool) []VouchTreeEdge {
	if depth <= 0 {
		return []VouchTreeEdge{}
	}
	node := g.Nodes[user]
	if node == nil {
		return []VouchTreeEdge{}
	}
	path[user] = true
	edges := make([]VouchTreeEdge, 0, len(node.Incoming))
	for _, edge := range node.Incoming {
		peerUser := edge.Event.From
		if path[peerUser] {
			continue
		}
		peerNode := NewVouchTreeNode(peerUser)
		if depth > 1 {
			peerNode.Peers = buildIncomingEdges(g, peerUser, depth-1, maps.Clone(path))
		}
		edges = append(edges, VouchTreeEdge{
			Event: edge.Event,
			Peer:  &peerNode,
		})
	}
	return edges
}
