package main

import (
	"container/heap"
	"log"
)

const penaltyWeightPerLayer = 0.1
const balanceWeightPerLayer = 0.1
const maxBalanceVouchers = 5

// An Heap is a min-heap struct.
type Heap []int64

func (h Heap) Len() int           { return len(h) }
func (h Heap) Less(i, j int) bool { return h[i] < h[j] }
func (h Heap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *Heap) Push(x any) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(int64))
}

func (h *Heap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// Computes the aggregated penalty for a user.
// If an outgoing tree is not provided, it is built from the vouch graph.
func Penalty(state *AppState, user string, tree *VouchTreeNode) uint64 {
	// user should be the root of the tree
	if tree != nil && tree.User != user {
		log.Fatalf("Warning: provided tree root user %v does not match target user %v", tree.User, user)
	}

	if tree == nil {
		// Builds a vouch graph and outgoing user's tree of default depth.
		tree = OutgoingTree(state, user, DefaultTreeDepth)
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

	results := WalkTreePostOrder(tree, func(node *VouchTreeNode, results map[*VouchTreeNode]uint64) uint64 {
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

// Computes the aggregated balance for a user.
// If incoming tree is not provided, it is built from the vouch graph.
func Balance(state *AppState, user string, incomingTree *VouchTreeNode) int64 {
	// user should be the root of the tree
	if incomingTree != nil && incomingTree.User != user {
		log.Fatalf("Warning: provided tree root user %v does not match target user %v", incomingTree.User, user)
	}

	if incomingTree == nil {
		// Builds a vouch graph and incoming user's tree of default depth.
		incomingTree = IncomingTree(state, user, DefaultTreeDepth)
	}

	// NOTE: Do not check for nil state or tree, allow panic in that case.

	// TODO: apply balance decay based on timestamp

	balances := make(map[string]int64)
	baseBalance := func(u string) int64 {
		if sum, ok := balances[u]; ok {
			return sum
		}
		sum := int64(0)
		proof, err := state.ProofRecord(u)
		if err != nil {
			log.Printf("Error getting proof record for user %s: %v", u, err)
		} else {
			sum = int64(proof.Balance)
		}

		// TODO: rebuilds the outgoing tree for u each time; could be optimized by caching
		outgoingTree := OutgoingTree(state, u, DefaultTreeDepth)
		sum -= int64(Penalty(state, u, outgoingTree))
		balances[u] = sum
		return sum
	}

	results := WalkTreePostOrder(incomingTree, func(node *VouchTreeNode, results map[*VouchTreeNode]int64) int64 {
		total := baseBalance(node.User)
		peerBalances := make(Heap, 0, len(node.Peers))
		heap.Init(&peerBalances)
		for _, edge := range node.Peers {
			if edge.Peer == nil {
				continue
			}
			// Only consider positive balances from peers
			if results[edge.Peer] <= 0 {
				continue
			}
			heap.Push(&peerBalances, results[edge.Peer])
		}
		limit := min(peerBalances.Len(), maxBalanceVouchers)
		for i := 0; i < limit; i++ {
			total += int64(balanceWeightPerLayer * float64(peerBalances.Pop().(int64)))
		}
		return total
	})

	return results[incomingTree]
}
