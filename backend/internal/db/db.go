package db

import (
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func Initialize(dbPath string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite", dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}

	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *sqlx.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS games (
		id TEXT PRIMARY KEY,
		code TEXT UNIQUE NOT NULL,
		state TEXT NOT NULL,
		status TEXT DEFAULT 'waiting',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_games_code ON games(code);
	CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);

	CREATE TABLE IF NOT EXISTS players (
		id TEXT PRIMARY KEY,
		game_id TEXT REFERENCES games(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		color TEXT,
		session_token TEXT UNIQUE,
		is_host INTEGER DEFAULT 0,
		connected INTEGER DEFAULT 0,
		last_seen TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_players_game_id ON players(game_id);
	CREATE INDEX IF NOT EXISTS idx_players_session_token ON players(session_token);
	`

	_, err := db.Exec(schema)
	return err
}
