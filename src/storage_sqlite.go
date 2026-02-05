package main

import (
	"database/sql"
	"time"

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
			timestamp INTEGER NOT NULL,
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
			balance INTEGER NOT NULL,
			timestamp INTEGER NOT NULL
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
			amount INTEGER NOT NULL,
			timestamp INTEGER NOT NULL
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

	if err := ensureColumn(db, "vouches", "timestamp",
		"ALTER TABLE vouches ADD COLUMN timestamp INTEGER NOT NULL DEFAULT (strftime('%s','now'))"); err != nil {
		return err
	}
	if err := ensureColumn(db, "proofs", "timestamp",
		"ALTER TABLE proofs ADD COLUMN timestamp INTEGER NOT NULL DEFAULT (strftime('%s','now'))"); err != nil {
		return err
	}
	if err := ensureColumn(db, "penalties", "timestamp",
		"ALTER TABLE penalties ADD COLUMN timestamp INTEGER NOT NULL DEFAULT (strftime('%s','now'))"); err != nil {
		return err
	}

	return nil
}

func ensureColumn(db *sql.DB, tableName string, columnName string, alterStatement string) error {
	exists, err := columnExists(db, tableName, columnName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = db.Exec(alterStatement)
	return err
}

func columnExists(db *sql.DB, tableName string, columnName string) (bool, error) {
	rows, err := db.Query("PRAGMA table_info(" + tableName + ")")
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var primaryKey int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return false, err
		}
		if name == columnName {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	return false, nil
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
	_, err := s.db.Exec(
		"REPLACE INTO vouches (from_user, to_user, timestamp) VALUES (?, ?, ?)",
		vouch.From,
		vouch.To,
		vouch.Timestamp.Unix(),
	)
	return err
}

// Returns a copy of all stored outgoing vouches for a specific user.
func (s *SQLiteStorage) UserVouchesFrom(user string) ([]VouchEvent, error) {
	rows, err := s.db.Query("SELECT from_user, to_user, timestamp FROM vouches WHERE from_user = ?", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vouches []VouchEvent
	for rows.Next() {
		var v VouchEvent
		var timestamp int64
		if err := rows.Scan(&v.From, &v.To, &timestamp); err != nil {
			return nil, err
		}
		v.Timestamp = time.Unix(timestamp, 0).UTC()
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
	rows, err := s.db.Query("SELECT from_user, to_user, timestamp FROM vouches WHERE to_user = ?", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vouches []VouchEvent
	for rows.Next() {
		var v VouchEvent
		var timestamp int64
		if err := rows.Scan(&v.From, &v.To, &timestamp); err != nil {
			return nil, err
		}
		v.Timestamp = time.Unix(timestamp, 0).UTC()
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
		INSERT INTO proofs (user, balance, timestamp) VALUES (?, ?, ?)
		ON CONFLICT(user) DO UPDATE SET balance = excluded.balance, timestamp = excluded.timestamp
	`, proof.User, proof.Balance, proof.Timestamp.Unix())
	return err
}

// Returns the stored proof event for a user, if any.
func (s *SQLiteStorage) ProofRecord(user string) (ProofEvent, error) {
	var proof ProofEvent
	var timestamp int64
	err := s.db.QueryRow("SELECT user, balance, timestamp FROM proofs WHERE user = ?", user).Scan(
		&proof.User,
		&proof.Balance,
		&timestamp,
	)
	if err == sql.ErrNoRows {
		return ProofEvent{User: user}, nil
	}
	if err != nil {
		return ProofEvent{User: user}, err
	}
	proof.Timestamp = time.Unix(timestamp, 0).UTC()
	return proof, nil
}

// Records a penalty event.
func (s *SQLiteStorage) AddPenalty(penalty PenaltyEvent) error {
	_, err := s.db.Exec(
		"INSERT INTO penalties (user, amount, timestamp) VALUES (?, ?, ?)",
		penalty.User,
		penalty.Amount,
		penalty.Timestamp.Unix(),
	)
	return err
}

// Returns all stored penalties for a user.
func (s *SQLiteStorage) Penalties(user string) ([]PenaltyEvent, error) {
	rows, err := s.db.Query("SELECT user, amount, timestamp FROM penalties WHERE user = ? ORDER BY id", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var penalties []PenaltyEvent
	for rows.Next() {
		var p PenaltyEvent
		var timestamp int64
		if err := rows.Scan(&p.User, &p.Amount, &timestamp); err != nil {
			return nil, err
		}
		p.Timestamp = time.Unix(timestamp, 0).UTC()
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
