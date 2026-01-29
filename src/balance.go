package main

import "log"

const penaltyWeightPerLayer = 0.1
const balanceWeightPerLayer = 0.1

type processFunc[T any] func(node *VouchTreeNode, results map[*VouchTreeNode]T) T

// walkTreePostOrder traverses the vouch tree in post-order and applies the process function to each node.
// It returns a map of nodes to their computed values.
// The process function returns a single aggregated value of type T for each node.
func walkTreePostOrder[T any](root *VouchTreeNode, process processFunc[T]) map[*VouchTreeNode]T {
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

// penalty computes the aggregated penalty for a user.
// If an outgoing tree is not provided, it is built from the vouch graph.
func penalty(state *AppState, user string, tree *VouchTreeNode) uint64 {
	// user should be the root of the tree
	if tree != nil && tree.User != user {
		log.Fatalf("Warning: provided tree root user %v does not match target user %v", tree.User, user)
	}

	if tree == nil {
		// Builds a vouch graph and outgoing user's tree of default depth.
		graph := state.VouchGraph()
		tree = graph.OutgoingTree(user, DefaultTreeDepth)
	}

	// NOTE: Do not check for nil state or tree, allow panic in that case.

	// TODO: apply penalty decay based on timestamp

	penaltySums := make(map[string]uint64)
	basePenalty := func(u string) uint64 {
		if sum, ok := penaltySums[u]; ok {
			return sum
		}
		sum := uint64(0)
		userPenalties := state.Penalties(u)
		for _, p := range userPenalties {
			sum += uint64(p.Amount)
		}
		penaltySums[u] = sum
		return sum
	}

	results := walkTreePostOrder(tree, func(node *VouchTreeNode, results map[*VouchTreeNode]uint64) uint64 {
		total := basePenalty(node.User)
		for _, edge := range node.Peers {
			if edge.Peer == nil {
				continue
			}
			total += uint64(penaltyWeightPerLayer * float64(results[edge.Peer]))
		}
		return total
	})

	return results[tree]
}

// balance computes the aggregated balance for a user.
// If incoming tree is not provided, it is built from the vouch graph.
func balance(state *AppState, user string, incomingTree *VouchTreeNode) int64 {
	// user should be the root of the tree
	if incomingTree != nil && incomingTree.User != user {
		log.Fatalf("Warning: provided tree root user %v does not match target user %v", incomingTree.User, user)
	}

	var graph VouchGraph
	if incomingTree == nil {
		graph = state.VouchGraph()
		// Builds a vouch graph and incoming user's tree of default depth.
		incomingTree = graph.IncomingTree(user, DefaultTreeDepth)
	}

	// NOTE: Do not check for nil state or tree, allow panic in that case.

	// TODO: apply balance decay based on timestamp

	balances := make(map[string]int64)
	baseBalance := func(u string) int64 {
		if sum, ok := balances[u]; ok {
			return sum
		}
		sum := int64(0)
		if proof, ok := state.ProofRecord(u); ok {
			sum = int64(proof.Balance)
		}
		// TODO: rebuilds the outgoing tree for u each time; could be optimized by caching
		outgoingTree := graph.OutgoingTree(u, DefaultTreeDepth)
		sum -= int64(penalty(state, u, outgoingTree))
		balances[u] = sum
		return sum
	}

	results := walkTreePostOrder(incomingTree, func(node *VouchTreeNode, results map[*VouchTreeNode]int64) int64 {
		total := baseBalance(node.User)
		for _, edge := range node.Peers {
			if edge.Peer == nil {
				continue
			}
			total += int64(balanceWeightPerLayer * float64(results[edge.Peer]))
		}
		return total
	})

	return results[incomingTree]
}
