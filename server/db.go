package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var sqlDB *sql.DB

func initDB() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = filepath.Join(".", "data", "badminton.db")
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	var err error
	sqlDB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	sqlDB.Exec("PRAGMA journal_mode = WAL")
	sqlDB.Exec("PRAGMA foreign_keys = ON")

	if err = sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Database connected")
}

func migrateDB() {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		full_name TEXT NOT NULL,
		avatar_url TEXT,
		level INTEGER DEFAULT 1,
		exp INTEGER DEFAULT 0,
		exp_to_next_level INTEGER DEFAULT 100,
		rank TEXT DEFAULT 'Bronze',
		wins INTEGER DEFAULT 0,
		losses INTEGER DEFAULT 0,
		draws INTEGER DEFAULT 0,
		total_matches INTEGER DEFAULT 0,
		win_rate REAL DEFAULT 0,
		points INTEGER DEFAULT 0,
		rank_points INTEGER DEFAULT 0,
		created_at TEXT DEFAULT (datetime('now')),
		updated_at TEXT DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS matches (
		id TEXT PRIMARY KEY,
		room_name TEXT NOT NULL,
		match_type TEXT NOT NULL CHECK (match_type IN ('singles', 'doubles')),
		match_mode TEXT NOT NULL CHECK (match_mode IN ('casual', 'ranked')),
		max_sets INTEGER DEFAULT 3,
		status TEXT NOT NULL DEFAULT 'waiting' CHECK (status IN ('waiting', 'playing', 'scoring', 'finished', 'cancelled')),
		winner_team TEXT CHECK (winner_team IN ('A', 'B', 'draw', NULL)),
		referee_id TEXT REFERENCES users(id),
		created_by TEXT NOT NULL REFERENCES users(id),
		started_at TEXT,
		ended_at TEXT,
		created_at TEXT DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS match_players (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		match_id TEXT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
		user_id TEXT NOT NULL REFERENCES users(id),
		team TEXT NOT NULL CHECK (team IN ('A', 'B')),
		exp_gained INTEGER DEFAULT 0,
		rank_points_gained INTEGER DEFAULT 0,
		joined_at TEXT DEFAULT (datetime('now')),
		UNIQUE(match_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS match_sets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		match_id TEXT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
		set_number INTEGER NOT NULL,
		team_a_score INTEGER NOT NULL DEFAULT 0,
		team_b_score INTEGER NOT NULL DEFAULT 0,
		UNIQUE(match_id, set_number)
	);

	CREATE INDEX IF NOT EXISTS idx_match_players_match ON match_players(match_id);
	CREATE INDEX IF NOT EXISTS idx_match_players_user ON match_players(user_id);
	CREATE INDEX IF NOT EXISTS idx_match_sets_match ON match_sets(match_id);
	CREATE INDEX IF NOT EXISTS idx_users_rank_points ON users(rank_points DESC);
	CREATE INDEX IF NOT EXISTS idx_users_level ON users(level DESC);
	`

	if _, err := sqlDB.Exec(schema); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Println("✅ Database migrated successfully")
}
