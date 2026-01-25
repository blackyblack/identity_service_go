package main

// VouchGraphEdge represents a vouch event between two users.
type VouchGraphEdge struct {
	Event VouchEvent
	Peer  *VouchGraphNode
}

// VouchGraphNode represents a user in the vouch graph.
type VouchGraphNode struct {
	Outgoing []VouchGraphEdge
	Incoming []VouchGraphEdge
}

// VouchGraph organizes users and their vouch relationships.
type VouchGraph struct {
	Nodes map[string]*VouchGraphNode
}

// BuildVouchGraph constructs a graph where users are nodes and vouches are edges.
func BuildVouchGraph(vouches []VouchEvent) VouchGraph {
	nodes := make(map[string]*VouchGraphNode)

	getNode := func(user string) *VouchGraphNode {
		if node, ok := nodes[user]; ok {
			return node
		}
		node := VouchGraphNode{}
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
