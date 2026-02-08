package main

import (
	"testing"
	"time"
)

func TestDecayedAmountNoElapsed(t *testing.T) {
	now := time.Now().UTC()
	got := DecayedAmount(100, now, now)
	if got != 100 {
		t.Fatalf("expected 100, got %d", got)
	}
}

func TestDecayedAmountPartialDecay(t *testing.T) {
	now := time.Now().UTC()
	tenDaysAgo := now.Add(-10 * 24 * time.Hour)
	got := DecayedAmount(100, tenDaysAgo, now)
	if got != 90 {
		t.Fatalf("expected 90, got %d", got)
	}
}

func TestDecayedAmountFullDecay(t *testing.T) {
	now := time.Now().UTC()
	past := now.Add(-200 * 24 * time.Hour)
	got := DecayedAmount(100, past, now)
	if got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestDecayedAmountFutureTimestamp(t *testing.T) {
	now := time.Now().UTC()
	future := now.Add(10 * 24 * time.Hour)
	got := DecayedAmount(100, future, now)
	if got != 100 {
		t.Fatalf("expected 100, got %d", got)
	}
}

func TestDecayedAmountPartialDay(t *testing.T) {
	now := time.Now().UTC()
	// 1.5 days ago should only decay by 1 (truncated to full days)
	past := now.Add(-36 * time.Hour)
	got := DecayedAmount(100, past, now)
	if got != 99 {
		t.Fatalf("expected 99, got %d", got)
	}
}

func TestPenaltyMultipleWithDecay(t *testing.T) {
	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	state := NewAppState()
	state.now = func() time.Time { return now }

	// Penalty of 100, created 10 days ago -> decayed to 90
	state.AddPenalty(PenaltyEvent{
		User:      "alice",
		Amount:    100,
		Timestamp: now.Add(-10 * 24 * time.Hour),
	})
	// Penalty of 50, created 5 days ago -> decayed to 45
	state.AddPenalty(PenaltyEvent{
		User:      "alice",
		Amount:    50,
		Timestamp: now.Add(-5 * 24 * time.Hour),
	})

	got := Penalty(state, "alice", nil)
	// 90 + 45 = 135
	if got != 135 {
		t.Fatalf("expected penalty 135, got %d", got)
	}
}

func TestBalanceWithProofAndPenaltyDecay(t *testing.T) {
	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	state := NewAppState()
	state.now = func() time.Time { return now }

	// Proof balance 100 created 10 days ago -> decayed to 90
	state.SetProof(ProofEvent{
		User:      "alice",
		Balance:   100,
		Timestamp: now.Add(-10 * 24 * time.Hour),
	})
	// Penalty 30 created 5 days ago -> decayed to 25
	state.AddPenalty(PenaltyEvent{
		User:      "alice",
		Amount:    30,
		Timestamp: now.Add(-5 * 24 * time.Hour),
	})

	got := Balance(state, "alice", nil)
	// 90 - 25 = 65
	if got != 65 {
		t.Fatalf("expected balance 65, got %d", got)
	}
}
