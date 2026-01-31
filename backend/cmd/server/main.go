package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"settlers_from_catan/internal/db"
	"settlers_from_catan/internal/handlers"
	"settlers_from_catan/internal/hub"
)

func main() {
	// Initialize database
	database, err := db.Initialize("./catan.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	// Initialize WebSocket hub
	h := hub.NewHub()
	go h.Run()

	// Initialize handlers
	handler := handlers.NewHandler(database, h)

	// Routes
	http.HandleFunc("/health", handler.HandleHealth)
	http.HandleFunc("/ws", handler.HandleWebSocket)
	http.HandleFunc("/api/games", handler.HandleCreateGame)
	http.HandleFunc("/api/games/", func(w http.ResponseWriter, r *http.Request) {
		// If path ends with /join and is POST, delegate to HandleJoinGame
		if r.Method == "POST" && len(r.URL.Path) > len("/api/games/") && strings.HasSuffix(r.URL.Path, "/join") {
			handler.HandleJoinGame(w, r)
			return
		}
		handler.HandleGameRoutes(w, r)
	})

	// Test endpoints (only available when DEV_MODE=true)
	if os.Getenv("DEV_MODE") == "true" {
		log.Println("DEV_MODE enabled - test endpoints available")
		http.HandleFunc("/test/grant-resources", handler.HandleGrantResources)
		http.HandleFunc("/test/grant-dev-card", handler.HandleGrantDevCard)
		http.HandleFunc("/test/force-dice-roll", handler.HandleForceDiceRoll)
		http.HandleFunc("/test/set-game-state", handler.HandleSetGameState)
	}

	// CORS middleware for development
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
