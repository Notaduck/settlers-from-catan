package main

import (
	"log"
	"net/http"

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
	http.HandleFunc("/ws", handler.HandleWebSocket)
	http.HandleFunc("/api/games", handler.HandleCreateGame)
	http.HandleFunc("/api/games/", handler.HandleGameRoutes)

	// CORS middleware for development
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
