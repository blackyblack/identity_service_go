package main

import (
	"os"
	"testing"
)

// storageFactory is a function that creates a new Storage instance
type storageFactory func(t *testing.T) Storage

// testStorageImplementations runs the given test function against both
// in-memory and SQLite storage implementations to ensure identical behavior.
func testStorageImplementations(t *testing.T, name string, testFn func(t *testing.T, storage Storage)) {
	factories := map[string]storageFactory{
		"Memory": func(t *testing.T) Storage {
			return NewMemoryStorage()
		},
		"SQLite": func(t *testing.T) Storage {
			tmpFile, err := os.CreateTemp("", "test_*.db")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			tmpFile.Close()
			t.Cleanup(func() {
				os.Remove(tmpFile.Name())
			})

			storage, err := NewSQLiteStorage(tmpFile.Name())
			if err != nil {
				t.Fatalf("Failed to create SQLite storage: %v", err)
			}
			t.Cleanup(func() {
				storage.Close()
			})
			return storage
		},
	}

	for storageType, factory := range factories {
		t.Run(storageType+"/"+name, func(t *testing.T) {
			storage := factory(t)
			testFn(t, storage)
		})
	}
}

func TestStorageEmpty(t *testing.T) {
	testStorageImplementations(t, "Empty", func(t *testing.T, storage Storage) {
		vouches, err := storage.Vouches()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(vouches) != 0 {
			t.Fatalf("expected 0 vouches, got %d", len(vouches))
		}

		_, ok, err := storage.ProofRecord("alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatal("expected no proof record for alice")
		}

		penalties, err := storage.Penalties("alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(penalties) != 0 {
			t.Fatalf("expected 0 penalties, got %d", len(penalties))
		}
	})
}

func TestStorageAddVouch(t *testing.T) {
	testStorageImplementations(t, "AddVouch", func(t *testing.T, storage Storage) {
		v1 := VouchEvent{From: "alice", To: "bob"}
		v2 := VouchEvent{From: "bob", To: "carol"}

		if err := storage.AddVouch(v1); err != nil {
			t.Fatalf("unexpected error adding vouch: %v", err)
		}
		if err := storage.AddVouch(v2); err != nil {
			t.Fatalf("unexpected error adding vouch: %v", err)
		}

		vouches, err := storage.Vouches()
		if err != nil {
			t.Fatalf("unexpected error getting vouches: %v", err)
		}
		if len(vouches) != 2 {
			t.Fatalf("expected 2 vouches, got %d", len(vouches))
		}
		if vouches[0] != v1 || vouches[1] != v2 {
			t.Fatalf("unexpected vouch order: %#v", vouches)
		}
	})
}

func TestStorageVouchesReturnsCopy(t *testing.T) {
	testStorageImplementations(t, "VouchesReturnsCopy", func(t *testing.T, storage Storage) {
		v1 := VouchEvent{From: "alice", To: "bob"}
		v2 := VouchEvent{From: "bob", To: "carol"}

		if err := storage.AddVouch(v1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := storage.AddVouch(v2); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		vouches, err := storage.Vouches()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Mutate the returned slice
		vouches[0] = VouchEvent{From: "mallory", To: "trent"}
		vouches = append(vouches, VouchEvent{From: "dan", To: "erin"})

		// Verify storage is unchanged
		vouchesAfter, err := storage.Vouches()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(vouchesAfter) != 2 {
			t.Fatalf("expected 2 vouches after mutation, got %d", len(vouchesAfter))
		}
		if vouchesAfter[0] != v1 || vouchesAfter[1] != v2 {
			t.Fatalf("storage mutated through copy: %#v", vouchesAfter)
		}
	})
}

func TestStorageSetProof(t *testing.T) {
	testStorageImplementations(t, "SetProof", func(t *testing.T, storage Storage) {
		proof := ProofEvent{User: "alice", Balance: 100}
		if err := storage.SetProof(proof); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok, err := storage.ProofRecord("alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatal("expected proof record for alice")
		}
		if got.User != "alice" || got.Balance != 100 {
			t.Fatalf("unexpected proof: %#v", got)
		}
	})
}

func TestStorageSetProofReplacesExisting(t *testing.T) {
	testStorageImplementations(t, "SetProofReplacesExisting", func(t *testing.T, storage Storage) {
		if err := storage.SetProof(ProofEvent{User: "alice", Balance: 10}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := storage.SetProof(ProofEvent{User: "alice", Balance: 25}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		proof, ok, err := storage.ProofRecord("alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatal("expected proof record for alice")
		}
		if proof.Balance != 25 {
			t.Fatalf("expected latest balance 25, got %d", proof.Balance)
		}
	})
}

func TestStorageAddPenalty(t *testing.T) {
	testStorageImplementations(t, "AddPenalty", func(t *testing.T, storage Storage) {
		p1 := PenaltyEvent{User: "alice", Amount: 10}
		p2 := PenaltyEvent{User: "alice", Amount: 20}

		if err := storage.AddPenalty(p1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := storage.AddPenalty(p2); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		penalties, err := storage.Penalties("alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(penalties) != 2 {
			t.Fatalf("expected 2 penalties, got %d", len(penalties))
		}
		if penalties[0] != p1 || penalties[1] != p2 {
			t.Fatalf("unexpected penalty order: %#v", penalties)
		}
	})
}

func TestStoragePenaltiesReturnsCopy(t *testing.T) {
	testStorageImplementations(t, "PenaltiesReturnsCopy", func(t *testing.T, storage Storage) {
		p1 := PenaltyEvent{User: "alice", Amount: 10}
		p2 := PenaltyEvent{User: "alice", Amount: 20}

		if err := storage.AddPenalty(p1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := storage.AddPenalty(p2); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		penalties, err := storage.Penalties("alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Mutate the returned slice
		penalties[0] = PenaltyEvent{User: "alice", Amount: 99}
		penalties = append(penalties, PenaltyEvent{User: "alice", Amount: 50})

		// Verify storage is unchanged
		penaltiesAfter, err := storage.Penalties("alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(penaltiesAfter) != 2 {
			t.Fatalf("expected 2 penalties after mutation, got %d", len(penaltiesAfter))
		}
		if penaltiesAfter[0] != p1 || penaltiesAfter[1] != p2 {
			t.Fatalf("storage mutated through copy: %#v", penaltiesAfter)
		}
	})
}

func TestStoragePenaltiesPerUser(t *testing.T) {
	testStorageImplementations(t, "PenaltiesPerUser", func(t *testing.T, storage Storage) {
		p1 := PenaltyEvent{User: "alice", Amount: 10}
		p2 := PenaltyEvent{User: "bob", Amount: 20}
		p3 := PenaltyEvent{User: "alice", Amount: 30}

		if err := storage.AddPenalty(p1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := storage.AddPenalty(p2); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := storage.AddPenalty(p3); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		alicePenalties, err := storage.Penalties("alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(alicePenalties) != 2 {
			t.Fatalf("expected 2 penalties for alice, got %d", len(alicePenalties))
		}

		bobPenalties, err := storage.Penalties("bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(bobPenalties) != 1 {
			t.Fatalf("expected 1 penalty for bob, got %d", len(bobPenalties))
		}

		carolPenalties, err := storage.Penalties("carol")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(carolPenalties) != 0 {
			t.Fatalf("expected 0 penalties for carol, got %d", len(carolPenalties))
		}
	})
}

func TestStorageMultipleUsers(t *testing.T) {
	testStorageImplementations(t, "MultipleUsers", func(t *testing.T, storage Storage) {
		// Add vouches
		if err := storage.AddVouch(VouchEvent{From: "alice", To: "bob"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := storage.AddVouch(VouchEvent{From: "bob", To: "carol"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Add proofs
		if err := storage.SetProof(ProofEvent{User: "alice", Balance: 100}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := storage.SetProof(ProofEvent{User: "bob", Balance: 200}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Add penalties
		if err := storage.AddPenalty(PenaltyEvent{User: "alice", Amount: 10}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := storage.AddPenalty(PenaltyEvent{User: "bob", Amount: 50}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify vouches
		vouches, err := storage.Vouches()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(vouches) != 2 {
			t.Fatalf("expected 2 vouches, got %d", len(vouches))
		}

		// Verify proofs
		aliceProof, ok, err := storage.ProofRecord("alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok || aliceProof.Balance != 100 {
			t.Fatalf("unexpected alice proof: %#v", aliceProof)
		}

		bobProof, ok, err := storage.ProofRecord("bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok || bobProof.Balance != 200 {
			t.Fatalf("unexpected bob proof: %#v", bobProof)
		}

		_, ok, err = storage.ProofRecord("carol")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatal("expected no proof record for carol")
		}

		// Verify penalties
		alicePenalties, err := storage.Penalties("alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(alicePenalties) != 1 || alicePenalties[0].Amount != 10 {
			t.Fatalf("unexpected alice penalties: %#v", alicePenalties)
		}

		bobPenalties, err := storage.Penalties("bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(bobPenalties) != 1 || bobPenalties[0].Amount != 50 {
			t.Fatalf("unexpected bob penalties: %#v", bobPenalties)
		}
	})
}

func TestStorageClose(t *testing.T) {
	testStorageImplementations(t, "Close", func(t *testing.T, storage Storage) {
		// Add some data
		if err := storage.AddVouch(VouchEvent{From: "alice", To: "bob"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Close should not error
		if err := storage.Close(); err != nil {
			t.Fatalf("unexpected error closing storage: %v", err)
		}
	})
}

// TestSQLiteStoragePersistence verifies that SQLite data persists across connections.
func TestSQLiteStoragePersistence(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_persistence_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Create storage and add data
	storage1, err := NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}

	if err := storage1.AddVouch(VouchEvent{From: "alice", To: "bob"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := storage1.SetProof(ProofEvent{User: "alice", Balance: 100}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := storage1.AddPenalty(PenaltyEvent{User: "alice", Amount: 10}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Close and reopen
	storage1.Close()

	storage2, err := NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to reopen SQLite storage: %v", err)
	}
	defer storage2.Close()

	// Verify data persisted
	vouches, err := storage2.Vouches()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vouches) != 1 || vouches[0].From != "alice" || vouches[0].To != "bob" {
		t.Fatalf("vouches did not persist: %#v", vouches)
	}

	proof, ok, err := storage2.ProofRecord("alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || proof.Balance != 100 {
		t.Fatalf("proof did not persist: %#v", proof)
	}

	penalties, err := storage2.Penalties("alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(penalties) != 1 || penalties[0].Amount != 10 {
		t.Fatalf("penalties did not persist: %#v", penalties)
	}
}

// TestAppStateWithSQLiteStorage verifies that AppState works correctly with SQLite storage.
func TestAppStateWithSQLiteStorage(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_appstate_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}
	defer storage.Close()

	state := NewAppStateWithStorage(storage)

	// Test the same operations as with memory storage
	state.AddVouch(VouchEvent{From: "alice", To: "bob"})
	state.AddVouch(VouchEvent{From: "bob", To: "carol"})

	vouches := state.Vouches()
	if len(vouches) != 2 {
		t.Fatalf("expected 2 vouches, got %d", len(vouches))
	}

	state.SetProof(ProofEvent{User: "alice", Balance: 100})
	proof, ok := state.ProofRecord("alice")
	if !ok || proof.Balance != 100 {
		t.Fatalf("unexpected proof: %#v", proof)
	}

	state.AddPenalty(PenaltyEvent{User: "alice", Amount: 10})
	penalties := state.Penalties("alice")
	if len(penalties) != 1 || penalties[0].Amount != 10 {
		t.Fatalf("unexpected penalties: %#v", penalties)
	}

	balance := state.ModerationBalance("alice")
	if balance != 90 {
		t.Fatalf("expected balance 90, got %d", balance)
	}
}

// TestAppStateWithBothStorages verifies that AppState behaves identically with both storage types.
func TestAppStateWithBothStorages(t *testing.T) {
	// Test with memory storage
	memState := NewAppState()

	// Test with SQLite storage
	tmpFile, err := os.CreateTemp("", "test_both_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	sqliteStorage, err := NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create SQLite storage: %v", err)
	}
	defer sqliteStorage.Close()
	sqliteState := NewAppStateWithStorage(sqliteStorage)

	// Perform the same operations on both
	states := []*AppState{memState, sqliteState}
	for _, state := range states {
		state.AddVouch(VouchEvent{From: "alice", To: "bob"})
		state.AddVouch(VouchEvent{From: "bob", To: "carol"})
		state.SetProof(ProofEvent{User: "alice", Balance: 100})
		state.SetProof(ProofEvent{User: "bob", Balance: 50})
		state.AddPenalty(PenaltyEvent{User: "alice", Amount: 10})
		state.AddPenalty(PenaltyEvent{User: "alice", Amount: 15})
	}

	// Verify both states have identical results
	memVouches := memState.Vouches()
	sqliteVouches := sqliteState.Vouches()
	if len(memVouches) != len(sqliteVouches) {
		t.Fatalf("vouch count mismatch: memory=%d, sqlite=%d", len(memVouches), len(sqliteVouches))
	}
	for i := range memVouches {
		if memVouches[i] != sqliteVouches[i] {
			t.Fatalf("vouch mismatch at index %d: memory=%#v, sqlite=%#v", i, memVouches[i], sqliteVouches[i])
		}
	}

	memProof, memOk := memState.ProofRecord("alice")
	sqliteProof, sqliteOk := sqliteState.ProofRecord("alice")
	if memOk != sqliteOk || memProof != sqliteProof {
		t.Fatalf("proof mismatch: memory=%#v/%v, sqlite=%#v/%v", memProof, memOk, sqliteProof, sqliteOk)
	}

	memPenalties := memState.Penalties("alice")
	sqlitePenalties := sqliteState.Penalties("alice")
	if len(memPenalties) != len(sqlitePenalties) {
		t.Fatalf("penalty count mismatch: memory=%d, sqlite=%d", len(memPenalties), len(sqlitePenalties))
	}
	for i := range memPenalties {
		if memPenalties[i] != sqlitePenalties[i] {
			t.Fatalf("penalty mismatch at index %d: memory=%#v, sqlite=%#v", i, memPenalties[i], sqlitePenalties[i])
		}
	}

	memBalance := memState.ModerationBalance("alice")
	sqliteBalance := sqliteState.ModerationBalance("alice")
	if memBalance != sqliteBalance {
		t.Fatalf("balance mismatch: memory=%d, sqlite=%d", memBalance, sqliteBalance)
	}
}
