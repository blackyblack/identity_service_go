package main

import "maps"

// Use it for reasonably deep tree traversals.
// We assume that vouch weight reduces 10x per hop, and each user has 10 peers on average.
// So depth 8 means weight 10^-8 at the leaves.
// IDT token is estimated to have 6 decimal digits of precision, and each user should have no
// more than 100 IDT, so 8th level gives smallest contributable weight of 0.000001 IDT.
// NOTE: penalties are not limited with 100 IDT, so one may consider another depth for penalty calculations.
const DefaultTreeDepth = 8

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
	Depth int
	Peers []VouchTreeEdge
}

func NewVouchGraphNode(user string) VouchGraphNode {
	return VouchGraphNode{
		User:     user,
		Outgoing: []VouchGraphEdge{},
		Incoming: []VouchGraphEdge{},
	}
}

func NewVouchTreeNode(user string, depth int) VouchTreeNode {
	return VouchTreeNode{
		User:  user,
		Depth: depth,
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

// buildTree is a helper function to build a depth-limited tree in a specified direction.
// If depth is negative, the search is unlimited.
func (g VouchGraph) buildTree(user string, depth int, isOutgoing bool) *VouchTreeNode {
	if depth == 0 {
		return &VouchTreeNode{User: user, Depth: 0, Peers: []VouchTreeEdge{}}
	}
	currDepth := 0

	root := NewVouchTreeNode(user, currDepth)

	type stackItem struct {
		node *VouchTreeNode
		path map[string]bool
	}

	stack := []stackItem{{
		node: &root,
		path: map[string]bool{user: true},
	}}

	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Negative depth means unlimited search
		if depth > 0 && curr.node.Depth >= depth {
			continue
		}

		graphNode, exists := g.Nodes[curr.node.User]
		if !exists {
			continue
		}
		if graphNode == nil {
			continue
		}

		edges := graphNode.Incoming
		if isOutgoing {
			edges = graphNode.Outgoing
		}

		for _, edge := range edges {
			peerUser := edge.Event.From
			if isOutgoing {
				peerUser = edge.Event.To
			}

			if curr.path[peerUser] {
				// Cycle detected based on current path
				continue
			}

			// Create the tree node for the peer
			peerTreeNode := NewVouchTreeNode(peerUser, curr.node.Depth+1)

			// Link it to the current tree node
			curr.node.Peers = append(curr.node.Peers, VouchTreeEdge{
				Event: edge.Event,
				Peer:  &peerTreeNode,
			})

			// Copy path for the next iteration to maintain context isolation
			newPath := maps.Clone(curr.path)
			newPath[peerUser] = true

			stack = append(stack, stackItem{
				node: &peerTreeNode,
				path: newPath,
			})
		}
		currDepth++
	}

	return &root
}

// OutgoingTree builds a depth-limited outgoing vouch tree rooted at the user iteratively.
func (g VouchGraph) OutgoingTree(user string, depth int) *VouchTreeNode {
	return g.buildTree(user, depth, true)
}

// IncomingTree builds a depth-limited incoming vouch tree rooted at the user iteratively.
func (g VouchGraph) IncomingTree(user string, depth int) *VouchTreeNode {
	return g.buildTree(user, depth, false)
}
