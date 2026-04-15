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
		role TEXT NOT NULL DEFAULT 'player' CHECK (role IN ('admin', 'leader', 'vice_leader', 'player')),
		avatar_url TEXT,
		skill_level TEXT NOT NULL DEFAULT 'BG1' CHECK (skill_level IN ('BG1', 'BG2', 'S', 'N', 'P-', 'P', 'P+')),
		skill_stars INTEGER NOT NULL DEFAULT 1 CHECK (skill_stars >= 1 AND skill_stars <= 5),
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

	CREATE TABLE IF NOT EXISTS courts (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		location TEXT DEFAULT '',
		max_rooms INTEGER DEFAULT 10,
		bonus_exp_percent INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed', 'maintenance')),
		created_by TEXT NOT NULL REFERENCES users(id),
		created_at TEXT DEFAULT (datetime('now')),
		updated_at TEXT DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS court_leaders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		court_id TEXT NOT NULL REFERENCES courts(id) ON DELETE CASCADE,
		user_id TEXT NOT NULL REFERENCES users(id),
		assigned_at TEXT DEFAULT (datetime('now')),
		UNIQUE(court_id, user_id)
	);

	CREATE INDEX IF NOT EXISTS idx_court_leaders_court ON court_leaders(court_id);
	CREATE INDEX IF NOT EXISTS idx_court_leaders_user ON court_leaders(user_id);

	CREATE TABLE IF NOT EXISTS matches (
		id TEXT PRIMARY KEY,
		court_id TEXT REFERENCES courts(id),
		room_name TEXT NOT NULL,
		match_type TEXT NOT NULL CHECK (match_type IN ('singles', 'doubles')),
		match_mode TEXT NOT NULL CHECK (match_mode IN ('casual', 'ranked', 'skill_test')),
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
	CREATE INDEX IF NOT EXISTS idx_courts_status ON courts(status);
	`

	if _, err := sqlDB.Exec(schema); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Add role column if it doesn't exist (for existing databases)
	sqlDB.Exec(`ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'player' CHECK (role IN ('admin', 'leader', 'vice_leader', 'player'))`)

	// Add court_id column to matches if it doesn't exist
	sqlDB.Exec(`ALTER TABLE matches ADD COLUMN court_id TEXT REFERENCES courts(id)`)
	sqlDB.Exec(`CREATE INDEX IF NOT EXISTS idx_matches_court ON matches(court_id)`)

	// Add skill columns to users if they don't exist
	sqlDB.Exec(`ALTER TABLE users ADD COLUMN skill_level TEXT NOT NULL DEFAULT 'BG1' CHECK (skill_level IN ('BG1', 'BG2', 'S', 'N', 'P-', 'P', 'P+'))`)
	sqlDB.Exec(`ALTER TABLE users ADD COLUMN skill_stars INTEGER NOT NULL DEFAULT 1 CHECK (skill_stars >= 1 AND skill_stars <= 5)`)

	// Add bonus_exp_percent to courts if it doesn't exist
	sqlDB.Exec(`ALTER TABLE courts ADD COLUMN bonus_exp_percent INTEGER NOT NULL DEFAULT 0`)

	// Create court_leaders table if it doesn't exist
	sqlDB.Exec(`CREATE TABLE IF NOT EXISTS court_leaders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		court_id TEXT NOT NULL REFERENCES courts(id) ON DELETE CASCADE,
		user_id TEXT NOT NULL REFERENCES users(id),
		assigned_at TEXT DEFAULT (datetime('now')),
		UNIQUE(court_id, user_id)
	)`)
	sqlDB.Exec(`CREATE INDEX IF NOT EXISTS idx_court_leaders_court ON court_leaders(court_id)`)
	sqlDB.Exec(`CREATE INDEX IF NOT EXISTS idx_court_leaders_user ON court_leaders(user_id)`)

	// Seed default admin account
	seedAdmin()

	fmt.Println("✅ Database migrated successfully")
}

func seedAdmin() {
	var exists string
	err := sqlDB.QueryRow("SELECT id FROM users WHERE username = 'admin'").Scan(&exists)
	if err == nil {
		return // admin already exists
	}

	// bcrypt hash of the admin password (cost 10)
	const adminHash = "$2b$10$3f0Dt./DGgZv4qBaILhP8ufUU5Lp1s5ySGjEWT4h8h5yUH.Zkewsy"

	id := "admin-" + fmt.Sprintf("%s", "00000000-0000-0000-0000-000000000001")
	_, err = sqlDB.Exec(`INSERT INTO users (id, username, email, password_hash, full_name, role, exp_to_next_level)
		VALUES (?, 'admin', 'admin@badmintonhub.local', ?, 'Administrator', 'admin', 100)`,
		id, adminHash)
	if err != nil {
		log.Printf("⚠️ Failed to seed admin: %v", err)
		return
	}
	fmt.Println("✅ Default admin account created")
}
