package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	pb "settlers_from_catan/gen/proto/catan/v1"
	"settlers_from_catan/internal/game"
)

// TestHandlers are only available when DEV_MODE=true
// These endpoints allow E2E tests to manipulate game state for testing

// isDevMode checks if DEV_MODE environment variable is set to "true"
func isDevMode() bool {
	return os.Getenv("DEV_MODE") == "true"
}

// HandleGrantResources grants resources to a player (test endpoint only)
// POST /test/grant-resources
// Body: { gameCode: string, playerId: string, resources: { wood?: number, brick?: number, ... } }
func (h *Handler) HandleGrantResources(w http.ResponseWriter, r *http.Request) {
	if !isDevMode() {
		http.Error(w, "Test endpoints not available", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		GameCode  string `json:"gameCode"`
		PlayerID  string `json:"playerId"`
		Resources struct {
			Wood  int32 `json:"wood"`
			Brick int32 `json:"brick"`
			Sheep int32 `json:"sheep"`
			Wheat int32 `json:"wheat"`
			Ore   int32 `json:"ore"`
		} `json:"resources"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Load game state and ID
	var gameRow struct {
		ID    string `db:"id"`
		State string `db:"state"`
	}
	err := h.db.Get(&gameRow, "SELECT id, state FROM games WHERE code = ?", req.GameCode)
	if err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	var state game.GameState
	if err := unmarshalOptions.Unmarshal([]byte(gameRow.State), &state); err != nil {
		http.Error(w, "Failed to deserialize game state", http.StatusInternalServerError)
		return
	}

	// Find player and grant resources
	var player *game.PlayerState
	for _, p := range state.Players {
		if p.Id == req.PlayerID {
			player = p
			break
		}
	}

	if player == nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	// Grant resources
	if player.Resources == nil {
		player.Resources = &game.ResourceCount{}
	}
	player.Resources.Wood += req.Resources.Wood
	player.Resources.Brick += req.Resources.Brick
	player.Resources.Sheep += req.Resources.Sheep
	player.Resources.Wheat += req.Resources.Wheat
	player.Resources.Ore += req.Resources.Ore

	// Save updated state
	stateBytes, err := marshalOptions.Marshal(&state)
	if err != nil {
		http.Error(w, "Failed to serialize game state", http.StatusInternalServerError)
		return
	}

	_, err = h.db.Exec("UPDATE games SET state = ? WHERE code = ?", string(stateBytes), req.GameCode)
	if err != nil {
		http.Error(w, "Failed to update game state", http.StatusInternalServerError)
		return
	}

	// Broadcast updated state to all players
	h.broadcastGameStateProto(gameRow.ID, &state)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// HandleForceDiceRoll forces the next dice roll to a specific value (test endpoint only)
// POST /test/force-dice-roll
// Body: { gameCode: string, diceValue: number }
func (h *Handler) HandleForceDiceRoll(w http.ResponseWriter, r *http.Request) {
	if !isDevMode() {
		http.Error(w, "Test endpoints not available", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		GameCode  string `json:"gameCode"`
		DiceValue int32  `json:"diceValue"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.DiceValue < 2 || req.DiceValue > 12 {
		http.Error(w, "Dice value must be between 2 and 12", http.StatusBadRequest)
		return
	}

	// Load game state and ID
	var gameRow struct {
		ID    string `db:"id"`
		State string `db:"state"`
	}
	err := h.db.Get(&gameRow, "SELECT id, state FROM games WHERE code = ?", req.GameCode)
	if err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	var state game.GameState
	if err := unmarshalOptions.Unmarshal([]byte(gameRow.State), &state); err != nil {
		http.Error(w, "Failed to deserialize game state", http.StatusInternalServerError)
		return
	}

	// Convert total dice value to individual dice values
	// For testing purposes, we'll split the value as evenly as possible
	die1, die2 := splitDiceValue(req.DiceValue)

	// Perform the dice roll with forced values
	result, err := game.PerformDiceRollWithValues(&state, state.Players[state.CurrentTurn].Id, die1, die2)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to perform dice roll: %v", err), http.StatusBadRequest)
		return
	}

	// Save updated state
	stateBytes, err := marshalOptions.Marshal(&state)
	if err != nil {
		http.Error(w, "Failed to serialize game state", http.StatusInternalServerError)
		return
	}

	_, err = h.db.Exec("UPDATE games SET state = ? WHERE code = ?", string(stateBytes), req.GameCode)
	if err != nil {
		http.Error(w, "Failed to update game state", http.StatusInternalServerError)
		return
	}

	// Broadcast the dice roll result (same as normal dice roll handler)
	distributions := make([]*pb.ResourceDistribution, 0, len(result.ResourcesGained))
	for playerID, resources := range result.ResourcesGained {
		distributions = append(distributions, &pb.ResourceDistribution{
			PlayerId:  playerID,
			Resources: resources,
		})
	}
	rolledPayload := &pb.DiceRolledPayload{
		PlayerId:             state.Players[state.CurrentTurn].Id,
		Values:               []int32{int32(die1), int32(die2)},
		ResourcesDistributed: distributions,
	}
	h.broadcastServerMessage(req.GameCode, "diceRolled", rolledPayload)
	h.broadcastGameStateProto(gameRow.ID, &state)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"result": map[string]interface{}{
			"die1":  die1,
			"die2":  die2,
			"total": req.DiceValue,
		},
	})

	log.Printf("[TEST] Forced dice roll %d (%d+%d) for game %s", req.DiceValue, die1, die2, req.GameCode)
}

// HandleSetGameState advances game to a specific phase (test endpoint only)
// POST /test/set-game-state
// Body: { gameCode: string, phase?: string, status?: string }
func (h *Handler) HandleSetGameState(w http.ResponseWriter, r *http.Request) {
	if !isDevMode() {
		http.Error(w, "Test endpoints not available", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		GameCode string `json:"gameCode"`
		Phase    string `json:"phase,omitempty"`
		Status   string `json:"status,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Load game state and ID
	var gameRow struct {
		ID    string `db:"id"`
		State string `db:"state"`
	}
	err := h.db.Get(&gameRow, "SELECT id, state FROM games WHERE code = ?", req.GameCode)
	if err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	var state game.GameState
	if err := unmarshalOptions.Unmarshal([]byte(gameRow.State), &state); err != nil {
		http.Error(w, "Failed to deserialize game state", http.StatusInternalServerError)
		return
	}

	// Update phase if provided
	if req.Phase != "" {
		switch req.Phase {
		case "ROLL":
			state.TurnPhase = game.TurnPhaseRoll
		case "TRADE":
			state.TurnPhase = game.TurnPhaseTrade
		case "BUILD":
			state.TurnPhase = game.TurnPhaseBuild
		default:
			http.Error(w, "Invalid phase", http.StatusBadRequest)
			return
		}
	}

	// Update status if provided
	if req.Status != "" {
		switch req.Status {
		case "WAITING":
			state.Status = game.GameStatusWaiting
		case "SETUP":
			state.Status = game.GameStatusSetup
		case "PLAYING":
			state.Status = game.GameStatusPlaying
		case "FINISHED":
			state.Status = game.GameStatusFinished
		default:
			http.Error(w, "Invalid status", http.StatusBadRequest)
			return
		}
	}

	// Save updated state
	stateBytes, err := marshalOptions.Marshal(&state)
	if err != nil {
		http.Error(w, "Failed to serialize game state", http.StatusInternalServerError)
		return
	}

	_, err = h.db.Exec("UPDATE games SET state = ? WHERE code = ?", string(stateBytes), req.GameCode)
	if err != nil {
		http.Error(w, "Failed to update game state", http.StatusInternalServerError)
		return
	}

	// Broadcast updated state to all players
	h.broadcastGameStateProto(gameRow.ID, &state)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	log.Printf("[TEST] Set game %s to phase=%s status=%s", req.GameCode, req.Phase, req.Status)
}

// splitDiceValue splits a total dice value (2-12) into two individual dice (1-6 each)
// For testing purposes, distributes as evenly as possible
func splitDiceValue(total int32) (int, int) {
	// Ensure valid range
	if total < 2 {
		total = 2
	}
	if total > 12 {
		total = 12
	}

	// Split evenly when possible, otherwise distribute with die1 getting the extra
	if total <= 7 {
		// For 2-7, distribute evenly or give extra to die1
		die2 := int(total) / 2
		die1 := int(total) - die2
		return die1, die2
	} else {
		// For 8-12, start with 6 on one die and distribute remainder
		die1 := 6
		die2 := int(total) - 6
		if die2 > 6 {
			// If die2 would be > 6, redistribute
			die1 = int(total) - 6
			die2 = 6
		}
		return die1, die2
	}
}
