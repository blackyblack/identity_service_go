package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// Implements Storage using SQLite database.
type SQLiteStorage struct {
	db *sql.DB
}

// Creates a new SQLite storage.
// Use ":memory:" for an in-memory SQLite database or a file path for persistent storage.
func NewSQLiteStorage(dataSourceName string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteStorage{db: db}, nil
}

func createTables(db *sql.DB) error {
	// Create vouches table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS vouches (
			from_user TEXT NOT NULL,
			to_user TEXT NOT NULL,
			PRIMARY KEY (from_user, to_user)
		)
	`)
	if err != nil {
		return err
	}

	// Create index on vouches.from_user for faster lookups
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_vouches_from_user ON vouches(from_user)`)
	if err != nil {
		return err
	}

	// Create index on vouches.to_user for faster reverse lookups
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_vouches_to_user ON vouches(to_user)`)
	if err != nil {
		return err
	}

	// Create proofs table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS proofs (
			user TEXT PRIMARY KEY,
			balance INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create penalties table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS penalties (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user TEXT NOT NULL,
			amount INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create index on penalties.user for faster lookups
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_penalties_user ON penalties(user)`)
	if err != nil {
		return err
	}

	return nil
}

// Returns all users who have vouches, proofs, or penalties recorded.
func (s *SQLiteStorage) Users() ([]string, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT user FROM (
			SELECT from_user AS user FROM vouches
			UNION
			SELECT to_user AS user FROM vouches
			UNION
			SELECT user FROM proofs
			UNION
			SELECT user FROM penalties
		)
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var user string
		if err := rows.Scan(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Records an incoming vouch event.
func (s *SQLiteStorage) AddVouch(vouch VouchEvent) error {
	_, err := s.db.Exec("REPLACE INTO vouches (from_user, to_user) VALUES (?, ?)", vouch.From, vouch.To)
	return err
}

// Returns a copy of all stored outgoing vouches for a specific user.
func (s *SQLiteStorage) UserVouchesFrom(user string) ([]VouchEvent, error) {
	rows, err := s.db.Query("SELECT from_user, to_user FROM vouches WHERE from_user = ?", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vouches []VouchEvent
	for rows.Next() {
		var v VouchEvent
		if err := rows.Scan(&v.From, &v.To); err != nil {
			return nil, err
		}
		vouches = append(vouches, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if vouches == nil {
		vouches = make([]VouchEvent, 0)
	}
	return vouches, nil
}

// Returns a copy of all stored incoming vouches for a specific user.
func (s *SQLiteStorage) UserVouchesTo(user string) ([]VouchEvent, error) {
	rows, err := s.db.Query("SELECT from_user, to_user FROM vouches WHERE to_user = ?", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vouches []VouchEvent
	for rows.Next() {
		var v VouchEvent
		if err := rows.Scan(&v.From, &v.To); err != nil {
			return nil, err
		}
		vouches = append(vouches, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if vouches == nil {
		vouches = make([]VouchEvent, 0)
	}
	return vouches, nil
}

// Stores the latest proof event for a user, replacing any prior record.
func (s *SQLiteStorage) SetProof(proof ProofEvent) error {
	_, err := s.db.Exec(`
		INSERT INTO proofs (user, balance) VALUES (?, ?)
		ON CONFLICT(user) DO UPDATE SET balance = excluded.balance
	`, proof.User, proof.Balance)
	return err
}

// Returns the stored proof event for a user, if any.
func (s *SQLiteStorage) ProofRecord(user string) (ProofEvent, error) {
	var proof ProofEvent
	err := s.db.QueryRow("SELECT user, balance FROM proofs WHERE user = ?", user).Scan(&proof.User, &proof.Balance)
	if err == sql.ErrNoRows {
		return ProofEvent{User: user}, nil
	}
	if err != nil {
		return ProofEvent{User: user}, err
	}
	return proof, nil
}

// Records a penalty event.
func (s *SQLiteStorage) AddPenalty(penalty PenaltyEvent) error {
	_, err := s.db.Exec("INSERT INTO penalties (user, amount) VALUES (?, ?)", penalty.User, penalty.Amount)
	return err
}

// Returns all stored penalties for a user.
func (s *SQLiteStorage) Penalties(user string) ([]PenaltyEvent, error) {
	rows, err := s.db.Query("SELECT user, amount FROM penalties WHERE user = ? ORDER BY id", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var penalties []PenaltyEvent
	for rows.Next() {
		var p PenaltyEvent
		if err := rows.Scan(&p.User, &p.Amount); err != nil {
			return nil, err
		}
		penalties = append(penalties, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if penalties == nil {
		penalties = make([]PenaltyEvent, 0)
	}
	return penalties, nil
}

// Closes the database connection.
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
