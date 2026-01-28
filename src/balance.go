package main

const penaltyWeightPerLayer = 0.1

type processFunc func(node *VouchTreeNode, results map[*VouchTreeNode]uint64) uint64

// walkVouchTreePostOrder traverses the vouch tree in post-order and applies the process function to each node.
// It returns a map of nodes to their computed values.
// Process function returns a single value for each node, i.e. balance.
func walkTreePostOrder(root *VouchTreeNode, process processFunc) map[*VouchTreeNode]uint64 {
	results := make(map[*VouchTreeNode]uint64)
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
