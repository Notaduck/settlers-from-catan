package db

import (
	"os"
	"testing"
)

func TestInitialize(t *testing.T) {
	tmpFile := "test_catan.db"
	defer os.Remove(tmpFile)

	db, err := Initialize(tmpFile)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	var tableCount int
	err = db.Get(&tableCount, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('games', 'players')")
	if err != nil {
		t.Fatalf("Failed to query tables: %v", err)
	}
	if tableCount != 2 {
		t.Errorf("Expected 2 tables, got %d", tableCount)
	}
}

func TestPlayersTableHasIsHost(t *testing.T) {
	tmpFile := "test_catan_host.db"
	defer os.Remove(tmpFile)

	db, err := Initialize(tmpFile)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO games (id, code, state, status) VALUES ('game1', 'ABC123', '{}', 'waiting')")
	if err != nil {
		t.Fatalf("Failed to insert game: %v", err)
	}

	_, err = db.Exec("INSERT INTO players (id, game_id, name, session_token, is_host) VALUES ('player1', 'game1', 'TestPlayer', 'token123', 1)")
	if err != nil {
		t.Fatalf("Failed to insert player with is_host: %v", err)
	}

	var isHost int
	err = db.Get(&isHost, "SELECT is_host FROM players WHERE id = 'player1'")
	if err != nil {
		t.Fatalf("Failed to query is_host: %v", err)
	}
	if isHost != 1 {
		t.Errorf("Expected is_host to be 1, got %d", isHost)
	}
}
