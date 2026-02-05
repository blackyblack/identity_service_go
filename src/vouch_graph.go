package main

import (
	"maps"
	"time"
)

// Use it for reasonably deep tree traversals.
// We assume that vouch weight reduces 10x per hop, and each user has 10 peers on average.
// So depth 8 means weight 10^-8 at the leaves.
// IDT token is estimated to have 6 decimal digits of precision, and each user should have no
// more than 100 IDT, so 8th level gives smallest contributable weight of 0.000001 IDT.
// NOTE: penalties are not limited with 100 IDT, so one may consider another depth for penalty calculations.
const DefaultTreeDepth = 8

// Represents a stored vouch.
type VouchEvent struct {
	From      string
	To        string
	Timestamp time.Time
	// TODO: maybe store proof for external verification?
}

// Represents a vouch event in the vouch tree.
type VouchTreeEdge struct {
	Event VouchEvent
	Peer  *VouchTreeNode
}

// Represents a user in the vouch tree (unidirectional).
type VouchTreeNode struct {
	User  string
	Depth int
	Peers []VouchTreeEdge
}

// Defines the type for the processing function used in tree traversal.
type ProcessFunc[T any] func(node *VouchTreeNode, results map[*VouchTreeNode]T) T

func NewVouchTreeNode(user string, depth int) VouchTreeNode {
	return VouchTreeNode{
		User:  user,
		Depth: depth,
		Peers: []VouchTreeEdge{},
	}
}

// Builds a depth-limited tree in a specified direction.
// If depth is negative, the search is unlimited.
func buildTree(state *AppState, user string, depth int, isOutgoing bool) *VouchTreeNode {
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

		var events []VouchEvent

		if isOutgoing {
			events = state.UserVouchesFrom(curr.node.User)
		} else {
			events = state.UserVouchesTo(curr.node.User)
		}

		for _, event := range events {
			peerUser := event.From
			if isOutgoing {
				peerUser = event.To
			}

			if curr.path[peerUser] {
				// Cycle detected based on current path
				continue
			}

			// Create the tree node for the peer
			peerTreeNode := NewVouchTreeNode(peerUser, curr.node.Depth+1)

			// Link it to the current tree node
			curr.node.Peers = append(curr.node.Peers, VouchTreeEdge{
				Event: event,
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

// Builds a depth-limited outgoing vouch tree rooted at the user iteratively.
func OutgoingTree(state *AppState, user string, depth int) *VouchTreeNode {
	return buildTree(state, user, depth, true)
}

// Builds a depth-limited incoming vouch tree rooted at the user iteratively.
func IncomingTree(state *AppState, user string, depth int) *VouchTreeNode {
	return buildTree(state, user, depth, false)
}

// Traverses the vouch tree in post-order and applies the process function to each node.
// It returns a map of nodes to their computed values.
func WalkTreePostOrder[T any](root *VouchTreeNode, process ProcessFunc[T]) map[*VouchTreeNode]T {
	results := make(map[*VouchTreeNode]T)
	if root == nil {
		return results
	}

	type stackItem struct {
		node    *VouchTreeNode
		visited bool
	}

	stack := []stackItem{{node: root, visited: false}}

	for len(stack) > 0 {
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if item.node == nil {
			continue
		}

		if item.visited {
			results[item.node] = process(item.node, results)
			continue
		}

		stack = append(stack, stackItem{node: item.node, visited: true})
		// NOTE: Traversal order here does not matter
		for i := range item.node.Peers {
			peer := item.node.Peers[i].Peer
			if peer == nil {
				continue
			}
			stack = append(stack, stackItem{node: peer, visited: false})
		}
	}

	return results
}
