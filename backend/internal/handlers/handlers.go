package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	catanv1 "settlers_from_catan/gen/proto/catan/v1"
	"settlers_from_catan/internal/game"
	"settlers_from_catan/internal/hub"
)

type Handler struct {
	hub      *hub.Hub
	db       *sqlx.DB
	upgrader websocket.Upgrader
}

var marshalOptions = protojson.MarshalOptions{UseProtoNames: false, EmitUnpopulated: true}
var unmarshalOptions = protojson.UnmarshalOptions{DiscardUnknown: true}

// NewHandler builds a handler with DB + hub wired.
func NewHandler(db *sqlx.DB, hub *hub.Hub) *Handler {
	return &Handler{
		db:  db,
		hub: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// sendError sends an error payload to the client and logs it.
func (h *Handler) sendError(client *hub.Client, code, message string) {
	log.Printf("Error [%s]: %s", code, message)
	payload := &catanv1.ErrorPayload{
		Code:    code,
		Message: message,
	}
	h.sendServerMessage(client, "error", payload)
}

func (h *Handler) sendServerMessage(client *hub.Client, kind string, payload proto.Message) {
	data, err := marshalServerMessage(kind, payload)
	if err != nil {
		log.Printf("Failed to marshal server message: %v", err)
		return
	}
	client.Send(data)
}

func (h *Handler) broadcastServerMessage(gameID, kind string, payload proto.Message) {
	data, err := marshalServerMessage(kind, payload)
	if err != nil {
		log.Printf("Failed to marshal broadcast: %v", err)
		return
	}
	h.hub.BroadcastToGame(gameID, data)
}

// broadcastGameStateProto broadcasts the game state.
func (h *Handler) broadcastGameStateProto(gameID string, state *game.GameState) {
	payload := &catanv1.GameStatePayload{State: state}
	h.broadcastServerMessage(gameID, "gameState", payload)
}

func marshalServerMessage(kind string, payload proto.Message) ([]byte, error) {
	var payloadMap map[string]any
	if payload != nil {
		data, err := marshalOptions.Marshal(payload)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &payloadMap); err != nil {
			return nil, err
		}
	}
	message := map[string]any{
		"message": map[string]any{
			"oneofKind": kind,
			kind:        payloadMap,
		},
	}
	return json.Marshal(message)
}

func (h *Handler) loadGameState(gameID string) (*game.GameState, error) {
	var stateJSON string
	if err := h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", gameID); err != nil {
		return nil, err
	}
	var state game.GameState
	if err := unmarshalOptions.Unmarshal([]byte(stateJSON), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (h *Handler) saveGameState(gameID string, state *game.GameState) error {
	updatedJSON, err := marshalOptions.Marshal(state)
	if err != nil {
		return err
	}
	_, err = h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(updatedJSON), gameID)
	return err
}

func (h *Handler) handleClientMessage(client *hub.Client, raw []byte) {
	var msg clientEnvelope
	if err := json.Unmarshal(raw, &msg); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse client message")
		return
	}

	switch msg.Message.OneofKind {
	case "playerReady":
		h.handlePlayerReady(client, msg.payloadBytes(msg.Message.PlayerReady))
	case "startGame":
		h.handleStartGame(client, msg.payloadBytes(msg.Message.StartGame))
	case "rollDice":
		h.handleRollDice(client, msg.payloadBytes(msg.Message.RollDice))
	case "buildStructure":
		h.handleBuildStructure(client, msg.payloadBytes(msg.Message.BuildStructure))
	case "endTurn":
		h.handleEndTurn(client, msg.payloadBytes(msg.Message.EndTurn))
	case "proposeTrade":
		h.handleProposeTrade(client, msg.payloadBytes(msg.Message.ProposeTrade))
	case "respondTrade":
		h.handleRespondTrade(client, msg.payloadBytes(msg.Message.RespondTrade))
	case "discardCards":
		h.handleDiscardCards(client, msg.payloadBytes(msg.Message.DiscardCards))
	case "moveRobber":
		h.handleMoveRobber(client, msg.payloadBytes(msg.Message.MoveRobber))
	case "bankTrade":
		h.handleBankTrade(client, msg.payloadBytes(msg.Message.BankTrade))
	case "buyDevCard":
		h.handleBuyDevCard(client, msg.payloadBytes(msg.Message.BuyDevCard))
	case "playDevCard":
		h.handlePlayDevCard(client, msg.payloadBytes(msg.Message.PlayDevCard))
	case "setTurnPhase":
		h.handleSetTurnPhase(client, msg.payloadBytes(msg.Message.SetTurnPhase))
	default:
		h.sendError(client, "UNSUPPORTED_MESSAGE", "Unsupported message type")
	}
}

// handlePlayerReady handles a player's ready toggle in the lobby.
func (h *Handler) handlePlayerReady(client *hub.Client, payload []byte) {
	var req catanv1.PlayerReadyMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse player ready request")
		return
	}
	state, err := h.loadGameState(client.GameID)
	if err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to load game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}
	if err := game.SetPlayerReady(state, client.PlayerID, req.GetReady()); err != nil {
		h.sendError(client, "READY_ERROR", err.Error())
		return
	}
	if err := h.saveGameState(client.GameID, state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to persist game state")
		return
	}
	readyPayload := &catanv1.PlayerReadyChangedPayload{
		PlayerId: client.PlayerID,
		IsReady:  req.GetReady(),
	}
	h.broadcastServerMessage(client.GameID, "playerReadyChanged", readyPayload)
	h.broadcastGameStateProto(client.GameID, state)
}

// handleStartGame handles the host starting the game.
func (h *Handler) handleStartGame(client *hub.Client, payload []byte) {
	var req catanv1.StartGameMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse start game request")
		return
	}
	state, err := h.loadGameState(client.GameID)
	if err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to load game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}
	if err := game.StartGame(state, client.PlayerID); err != nil {
		h.sendError(client, "START_GAME_ERROR", err.Error())
		return
	}
	if err := h.saveGameState(client.GameID, state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to persist game state")
		return
	}
	startedPayload := &catanv1.GameStartedPayload{
		State: state,
	}
	h.broadcastServerMessage(client.GameID, "gameStarted", startedPayload)
	h.broadcastGameStateProto(client.GameID, state)
}

// handleRollDice handles dice rolls during a player's turn.
func (h *Handler) handleRollDice(client *hub.Client, payload []byte) {
	var req catanv1.RollDiceMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse roll dice request")
		return
	}
	state, err := h.loadGameState(client.GameID)
	if err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to load game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}
	result, err := game.PerformDiceRoll(state, client.PlayerID)
	if err != nil {
		h.sendError(client, "ROLL_DICE_ERROR", err.Error())
		return
	}
	if err := h.saveGameState(client.GameID, state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to persist game state")
		return
	}
	distributions := make([]*catanv1.ResourceDistribution, 0, len(result.ResourcesGained))
	for playerID, resources := range result.ResourcesGained {
		distributions = append(distributions, &catanv1.ResourceDistribution{
			PlayerId:  playerID,
			Resources: resources,
		})
	}
	rolledPayload := &catanv1.DiceRolledPayload{
		PlayerId:             client.PlayerID,
		Values:               []int32{int32(result.Die1), int32(result.Die2)},
		ResourcesDistributed: distributions,
	}
	h.broadcastServerMessage(client.GameID, "diceRolled", rolledPayload)
	h.broadcastGameStateProto(client.GameID, state)
}

// handleBuildStructure handles settlement, city, and road placement.
func (h *Handler) handleBuildStructure(client *hub.Client, payload []byte) {
	var req catanv1.BuildStructureMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse build request")
		return
	}
	state, err := h.loadGameState(client.GameID)
	if err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to load game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}
	location := req.GetLocation()
	var serverKind string
	var serverPayload proto.Message
	switch state.Status {
	case game.GameStatusSetup:
		switch req.GetStructureType() {
		case catanv1.StructureType_STRUCTURE_TYPE_SETTLEMENT:
			if err := game.PlaceSetupSettlement(state, client.PlayerID, location); err != nil {
				h.sendError(client, "BUILD_ERROR", err.Error())
				return
			}
			serverKind = "buildingPlaced"
			serverPayload = &catanv1.BuildingPlacedPayload{
				PlayerId:     client.PlayerID,
				BuildingType: catanv1.BuildingType_BUILDING_TYPE_SETTLEMENT,
				VertexId:     location,
			}
		case catanv1.StructureType_STRUCTURE_TYPE_ROAD:
			if err := game.PlaceSetupRoad(state, client.PlayerID, location); err != nil {
				h.sendError(client, "BUILD_ERROR", err.Error())
				return
			}
			serverKind = "roadPlaced"
			serverPayload = &catanv1.RoadPlacedPayload{
				PlayerId: client.PlayerID,
				EdgeId:   location,
			}
		default:
			h.sendError(client, "BUILD_ERROR", "Invalid structure for setup phase")
			return
		}
	case game.GameStatusPlaying:
		switch req.GetStructureType() {
		case catanv1.StructureType_STRUCTURE_TYPE_SETTLEMENT:
			if err := game.PlaceSettlement(state, client.PlayerID, location); err != nil {
				h.sendError(client, "BUILD_ERROR", err.Error())
				return
			}
			serverKind = "buildingPlaced"
			serverPayload = &catanv1.BuildingPlacedPayload{
				PlayerId:     client.PlayerID,
				BuildingType: catanv1.BuildingType_BUILDING_TYPE_SETTLEMENT,
				VertexId:     location,
			}
		case catanv1.StructureType_STRUCTURE_TYPE_CITY:
			if err := game.PlaceCity(state, client.PlayerID, location); err != nil {
				h.sendError(client, "BUILD_ERROR", err.Error())
				return
			}
			serverKind = "buildingPlaced"
			serverPayload = &catanv1.BuildingPlacedPayload{
				PlayerId:     client.PlayerID,
				BuildingType: catanv1.BuildingType_BUILDING_TYPE_CITY,
				VertexId:     location,
			}
		case catanv1.StructureType_STRUCTURE_TYPE_ROAD:
			if err := game.PlaceRoad(state, client.PlayerID, location); err != nil {
				h.sendError(client, "BUILD_ERROR", err.Error())
				return
			}
			serverKind = "roadPlaced"
			serverPayload = &catanv1.RoadPlacedPayload{
				PlayerId: client.PlayerID,
				EdgeId:   location,
			}
		default:
			h.sendError(client, "BUILD_ERROR", "Unsupported structure type")
			return
		}
	default:
		h.sendError(client, "BUILD_ERROR", "Game not in buildable state")
		return
	}
	if err := h.saveGameState(client.GameID, state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to persist game state")
		return
	}
	if serverKind != "" && serverPayload != nil {
		h.broadcastServerMessage(client.GameID, serverKind, serverPayload)
	}
	h.broadcastGameStateProto(client.GameID, state)
	if state.Status == game.GameStatusFinished {
		if winnerID, ok := game.DetermineWinner(state); ok {
			gameOverPayload := game.BuildGameOverPayload(state, winnerID)
			h.broadcastServerMessage(client.GameID, "gameOver", gameOverPayload)
		}
	}
}

// handleEndTurn advances to the next player.
func (h *Handler) handleEndTurn(client *hub.Client, payload []byte) {
	var req catanv1.EndTurnMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse end turn request")
		return
	}
	state, err := h.loadGameState(client.GameID)
	if err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to load game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}
	if err := game.EndTurn(state, client.PlayerID); err != nil {
		h.sendError(client, "END_TURN_ERROR", err.Error())
		return
	}
	if err := h.saveGameState(client.GameID, state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to persist game state")
		return
	}
	turnPayload := &catanv1.TurnChangedPayload{
		ActivePlayerId: state.Players[state.CurrentTurn].Id,
		Phase:          state.TurnPhase,
	}
	h.broadcastServerMessage(client.GameID, "turnChanged", turnPayload)
	h.broadcastGameStateProto(client.GameID, state)
}

// handleSetTurnPhase switches between TRADE and BUILD during a player's turn.
func (h *Handler) handleSetTurnPhase(client *hub.Client, payload []byte) {
	var req catanv1.SetTurnPhaseMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse set turn phase request")
		return
	}
	state, err := h.loadGameState(client.GameID)
	if err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to load game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}
	if err := game.SetTurnPhase(state, client.PlayerID, req.GetPhase()); err != nil {
		h.sendError(client, "TURN_PHASE_ERROR", err.Error())
		return
	}
	if err := h.saveGameState(client.GameID, state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to persist game state")
		return
	}
	h.broadcastGameStateProto(client.GameID, state)
}

// --- Trading handlers ---
// handleProposeTrade handles a player's trade proposal (player ↔ player).
func (h *Handler) handleProposeTrade(client *hub.Client, payload []byte) {
	var req catanv1.ProposeTradeMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse propose trade request")
		return
	}

	var stateJSON string
	var state game.GameState
	if err := h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to load game state")
		return
	}
	if err := unmarshalOptions.Unmarshal([]byte(stateJSON), &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to parse game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}

	// Call game logic
	_, err := game.ProposeTrade(&state, client.PlayerID, req.TargetId, req.GetOffering(), req.GetRequesting())
	if err != nil {
		h.sendError(client, "TRADE_PROPOSE_ERROR", err.Error())
		return
	}
	// Save updated state
	if err := h.saveGameState(client.GameID, &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to persist game state")
		return
	}
	// Broadcast new trade
	tradePayload := &catanv1.TradeProposedPayload{
		Trade: state.GetPendingTrades()[len(state.GetPendingTrades())-1], // last added trade
	}
	h.broadcastServerMessage(client.GameID, "tradeProposed", tradePayload)
	// Always broadcast updated game state for robust sync
	h.broadcastGameStateProto(client.GameID, &state)
}

// handleRespondTrade handles a player's response (accept/decline) to a trade.
func (h *Handler) handleRespondTrade(client *hub.Client, payload []byte) {
	var req catanv1.RespondTradeMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse respond trade request")
		return
	}

	var stateJSON string
	var state game.GameState
	if err := h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to load game state")
		return
	}
	if err := unmarshalOptions.Unmarshal([]byte(stateJSON), &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to parse game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}
	var err error
	err = game.RespondTrade(&state, req.GetTradeId(), client.PlayerID, req.Accept)

	if err != nil {
		h.sendError(client, "TRADE_RESPOND_ERROR", err.Error())
		return
	}
	if err := h.saveGameState(client.GameID, &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to persist game state")
		return
	}
	resolvedPayload := &catanv1.TradeResolvedPayload{
		TradeId:    req.GetTradeId(),
		Accepted:   req.GetAccept(),
		AcceptedBy: &client.PlayerID,
	}
	h.broadcastServerMessage(client.GameID, "tradeResolved", resolvedPayload)
	h.broadcastGameStateProto(client.GameID, &state)
}

// handleDiscardCards handles discarding cards for the robber phase.
func (h *Handler) handleDiscardCards(client *hub.Client, payload []byte) {
	var req catanv1.DiscardCardsMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse discard cards request")
		return
	}
	var stateJSON string
	var state game.GameState
	if err := h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to load game state")
		return
	}
	if err := unmarshalOptions.Unmarshal([]byte(stateJSON), &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to parse game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}
	if err := game.DiscardCards(&state, client.PlayerID, req.GetResources()); err != nil {
		h.sendError(client, "DISCARD_ERROR", err.Error())
		return
	}
	if err := h.saveGameState(client.GameID, &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to persist game state")
		return
	}
	discardedPayload := &catanv1.DiscardedCardsPayload{
		PlayerId:  client.PlayerID,
		Resources: req.GetResources(),
	}
	h.broadcastServerMessage(client.GameID, "discardedCards", discardedPayload)
	h.broadcastGameStateProto(client.GameID, &state)
}

// handleMoveRobber handles moving the robber and optional stealing.
func (h *Handler) handleMoveRobber(client *hub.Client, payload []byte) {
	var req catanv1.MoveRobberMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse move robber request")
		return
	}
	var stateJSON string
	var state game.GameState
	if err := h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to load game state")
		return
	}
	if err := unmarshalOptions.Unmarshal([]byte(stateJSON), &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to parse game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}
	if req.GetHex() == nil {
		victimID := req.GetVictimId()
		if victimID == "" {
			h.sendError(client, "MOVE_ROBBER_ERROR", "victim required for steal")
			return
		}
		stolenResource, err := game.StealFromPlayer(&state, client.PlayerID, victimID)
		if err != nil {
			h.sendError(client, "MOVE_ROBBER_ERROR", err.Error())
			return
		}
		if err := h.saveGameState(client.GameID, &state); err != nil {
			h.sendError(client, "INTERNAL_ERROR", "Failed to persist game state")
			return
		}
		robberPayload := &catanv1.RobberMovedPayload{
			PlayerId:       client.PlayerID,
			Hex:            state.Board.GetRobberHex(),
			VictimId:       &victimID,
			StolenResource: &stolenResource,
		}
		h.broadcastServerMessage(client.GameID, "robberMoved", robberPayload)
		h.broadcastGameStateProto(client.GameID, &state)
		return
	}

	if err := game.MoveRobber(&state, client.PlayerID, req.GetHex()); err != nil {
		h.sendError(client, "MOVE_ROBBER_ERROR", err.Error())
		return
	}
	var (
		stolenResource *catanv1.Resource
		victimID       *string
	)
	if req.VictimId != nil && *req.VictimId != "" {
		id := *req.VictimId
		res, err := game.StealFromPlayer(&state, client.PlayerID, id)
		if err != nil {
			h.sendError(client, "MOVE_ROBBER_ERROR", err.Error())
			return
		}
		victimID = &id
		stolenResource = &res
	}
	if err := h.saveGameState(client.GameID, &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to persist game state")
		return
	}
	robberPayload := &catanv1.RobberMovedPayload{
		PlayerId:       client.PlayerID,
		Hex:            req.GetHex(),
		VictimId:       victimID,
		StolenResource: stolenResource,
	}
	h.broadcastServerMessage(client.GameID, "robberMoved", robberPayload)
	h.broadcastGameStateProto(client.GameID, &state)
}

// handleBankTrade handles a player's trade with the bank (4:1).
func (h *Handler) handleBankTrade(client *hub.Client, payload []byte) {
	var req catanv1.BankTradeMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse bank trade request")
		return
	}

	var stateJSON string
	var state game.GameState
	if err := h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to load game state")
		return
	}
	if err := unmarshalOptions.Unmarshal([]byte(stateJSON), &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to parse game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}

	err := game.BankTrade(&state, client.PlayerID, req.GetOffering(), req.GetResourceRequested())
	if err != nil {
		h.sendError(client, "BANK_TRADE_ERROR", err.Error())
		return
	}
	if err := h.saveGameState(client.GameID, &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to persist game state")
		return
	}
	// No special payload; just broadcast updated state
	h.broadcastGameStateProto(client.GameID, &state)
}

// handleBuyDevCard handles purchasing a development card
func (h *Handler) handleBuyDevCard(client *hub.Client, payload []byte) {
	var req catanv1.BuyDevCardMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse buy dev card request")
		return
	}

	var stateJSON string
	var state game.GameState
	if err := h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to load game state")
		return
	}
	if err := unmarshalOptions.Unmarshal([]byte(stateJSON), &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to parse game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}

	// Buy dev card
	cardType, err := game.BuyDevCard(&state, client.PlayerID)
	if err != nil {
		h.sendError(client, "BUY_DEV_CARD_ERROR", err.Error())
		return
	}

	if err := h.saveGameState(client.GameID, &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to persist game state")
		return
	}

	// Notify the player what they drew (only to that player)
	boughtPayload := &catanv1.DevCardBoughtPayload{
		PlayerId: client.PlayerID,
		CardType: cardType,
	}
	h.sendServerMessage(client, "devCardBought", boughtPayload)

	// Broadcast updated game state to all
	h.broadcastGameStateProto(client.GameID, &state)
}

// handlePlayDevCard handles playing a development card
func (h *Handler) handlePlayDevCard(client *hub.Client, payload []byte) {
	var req catanv1.PlayDevCardMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse play dev card request")
		return
	}

	var stateJSON string
	var state game.GameState
	if err := h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to load game state")
		return
	}
	if err := unmarshalOptions.Unmarshal([]byte(stateJSON), &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to parse game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}

	// Play dev card
	err := game.PlayDevCard(&state, client.PlayerID, req.CardType, req.TargetResource, req.Resources)
	if err != nil {
		h.sendError(client, "PLAY_DEV_CARD_ERROR", err.Error())
		return
	}

	// Knight card triggers robber phase (move + optional steal)
	if req.CardType == catanv1.DevCardType_DEV_CARD_TYPE_KNIGHT {
		if state.RobberPhase == nil {
			state.RobberPhase = &catanv1.RobberPhase{}
		}
		state.RobberPhase.MovePendingPlayerId = &client.PlayerID
	}

	if err := h.saveGameState(client.GameID, &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to persist game state")
		return
	}

	// Broadcast updated game state to all
	h.broadcastGameStateProto(client.GameID, &state)
	if state.Status == game.GameStatusFinished {
		if winnerID, ok := game.DetermineWinner(&state); ok {
			gameOverPayload := game.BuildGameOverPayload(&state, winnerID)
			h.broadcastServerMessage(client.GameID, "gameOver", gameOverPayload)
		}
	}
}

// HandleCreateGame handles POST /api/games to create a new game.
func (h *Handler) HandleCreateGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var req catanv1.CreateGameRequest
	if err := unmarshalOptions.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	playerName := strings.TrimSpace(req.GetPlayerName())
	if playerName == "" {
		http.Error(w, "Player name is required", http.StatusBadRequest)
		return
	}

	gameID := uuid.NewString()
	playerID := uuid.NewString()
	sessionToken := uuid.NewString()

	code, err := h.generateUniqueCode(6)
	if err != nil {
		http.Error(w, "Failed to generate game code", http.StatusInternalServerError)
		return
	}

	state := game.NewGameState(gameID, code, []string{playerName}, []string{playerID})
	stateJSON, err := marshalOptions.Marshal(state)
	if err != nil {
		http.Error(w, "Failed to create game state", http.StatusInternalServerError)
		return
	}

	if _, err := h.db.Exec(
		"INSERT INTO games (id, code, state, status) VALUES (?, ?, ?, ?)",
		gameID,
		code,
		string(stateJSON),
		state.Status.String(),
	); err != nil {
		http.Error(w, "Failed to create game", http.StatusInternalServerError)
		return
	}

	if _, err := h.db.Exec(
		"INSERT INTO players (id, game_id, name, color, session_token, is_host, connected) VALUES (?, ?, ?, ?, ?, ?, ?)",
		playerID,
		gameID,
		playerName,
		state.Players[0].Color.String(),
		sessionToken,
		1,
		0,
	); err != nil {
		http.Error(w, "Failed to create player", http.StatusInternalServerError)
		return
	}

	resp := &catanv1.CreateGameResponse{
		GameId:       gameID,
		Code:         code,
		SessionToken: sessionToken,
		PlayerId:     playerID,
	}

	data, err := marshalOptions.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// HandleGameRoutes handles /api/games/{code} and /api/games/{code}/join.
func (h *Handler) HandleGameRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/games/")
	path = strings.Trim(path, "/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	parts := strings.Split(path, "/")
	code := strings.ToUpper(parts[0])

	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.handleGameInfo(w, r, code)
		return
	}

	if len(parts) == 2 && parts[1] == "join" {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.handleJoinGame(w, r, code)
		return
	}

	http.NotFound(w, r)
}

// HandleWebSocket handles websocket connections for a game.
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing session token", http.StatusUnauthorized)
		return
	}

	var playerRow struct {
		ID     string `db:"id"`
		GameID string `db:"game_id"`
	}
	if err := h.db.Get(&playerRow, "SELECT id, game_id FROM players WHERE session_token = ?", token); err != nil {
		http.Error(w, "Invalid session token", http.StatusUnauthorized)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade websocket: %v", err)
		return
	}

	state, err := h.loadGameState(playerRow.GameID)
	if err != nil {
		conn.Close()
		log.Printf("Failed to load game state: %v", err)
		return
	}
	playerFound := false
	for _, player := range state.Players {
		if player.Id == playerRow.ID {
			playerFound = true
			player.Connected = true
			break
		}
	}
	if !playerFound {
		conn.Close()
		log.Printf("Player not found in game state")
		return
	}
	if err := h.saveGameState(playerRow.GameID, state); err != nil {
		conn.Close()
		log.Printf("Failed to persist game state: %v", err)
		return
	}
	_, _ = h.db.Exec("UPDATE players SET connected = 1, last_seen = CURRENT_TIMESTAMP WHERE id = ?", playerRow.ID)

	client := hub.NewClient(h.hub, conn, playerRow.ID, playerRow.GameID)
	client.OnMessage = func(message []byte) {
		h.handleClientMessage(client, message)
	}
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()

	h.broadcastGameStateProto(playerRow.GameID, state)
}

func (h *Handler) handleJoinGame(w http.ResponseWriter, r *http.Request, code string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	var req catanv1.JoinGameRequest
	if err := unmarshalOptions.Unmarshal(body, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	playerName := strings.TrimSpace(req.GetPlayerName())
	if playerName == "" {
		writeJSONError(w, http.StatusBadRequest, "Player name is required")
		return
	}

	var gameRow struct {
		ID    string `db:"id"`
		State string `db:"state"`
	}
	if err := h.db.Get(&gameRow, "SELECT id, state FROM games WHERE code = ?", code); err != nil {
		writeJSONError(w, http.StatusNotFound, "Game not found")
		return
	}

	var state game.GameState
	if err := unmarshalOptions.Unmarshal([]byte(gameRow.State), &state); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to load game state")
		return
	}
	if len(state.Players) >= game.GetMaxPlayers() {
		writeJSONError(w, http.StatusBadRequest, "Game is full")
		return
	}

	playerID := uuid.NewString()
	sessionToken := uuid.NewString()
	color := nextPlayerColor(len(state.Players))
	state.Players = append(state.Players, &catanv1.PlayerState{
		Id:            playerID,
		Name:          playerName,
		Color:         color,
		Resources:     &catanv1.ResourceCount{},
		DevCardCount:  0,
		KnightsPlayed: 0,
		VictoryPoints: 0,
		Connected:     false,
		IsReady:       false,
		IsHost:        false,
	})

	stateJSON, err := marshalOptions.Marshal(&state)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to update game")
		return
	}
	if _, err := h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(stateJSON), gameRow.ID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to update game")
		return
	}
	if _, err := h.db.Exec(
		"INSERT INTO players (id, game_id, name, color, session_token, is_host, connected) VALUES (?, ?, ?, ?, ?, ?, ?)",
		playerID,
		gameRow.ID,
		playerName,
		color.String(),
		sessionToken,
		0,
		0,
	); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to create player")
		return
	}

	resp := &catanv1.JoinGameResponse{
		GameId:       gameRow.ID,
		SessionToken: sessionToken,
		PlayerId:     playerID,
		Players:      toPlayerInfos(state.Players),
	}
	data, err := marshalOptions.Marshal(resp)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to encode response")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)

	h.broadcastGameStateProto(gameRow.ID, &state)
}

func (h *Handler) handleGameInfo(w http.ResponseWriter, r *http.Request, code string) {
	var gameRow struct {
		ID     string `db:"id"`
		Status string `db:"status"`
		State  string `db:"state"`
	}
	if err := h.db.Get(&gameRow, "SELECT id, status, state FROM games WHERE code = ?", code); err != nil {
		writeJSONError(w, http.StatusNotFound, "Game not found")
		return
	}
	var state game.GameState
	if err := unmarshalOptions.Unmarshal([]byte(gameRow.State), &state); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to load game state")
		return
	}

	resp := &catanv1.GameInfoResponse{
		Code:        code,
		Status:      state.Status,
		PlayerCount: int32(len(state.Players)),
		Players:     toPlayerInfos(state.Players),
	}
	data, err := marshalOptions.Marshal(resp)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to encode response")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *Handler) generateUniqueCode(length int) (string, error) {
	for i := 0; i < 5; i++ {
		code, err := generateGameCode(length)
		if err != nil {
			return "", err
		}
		var count int
		if err := h.db.Get(&count, "SELECT COUNT(*) FROM games WHERE code = ?", code); err != nil {
			return "", err
		}
		if count == 0 {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique code")
}

func generateGameCode(length int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	buffer := make([]byte, length)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	for i, b := range buffer {
		buffer[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(buffer), nil
}

func nextPlayerColor(index int) catanv1.PlayerColor {
	colors := []catanv1.PlayerColor{
		catanv1.PlayerColor_PLAYER_COLOR_RED,
		catanv1.PlayerColor_PLAYER_COLOR_BLUE,
		catanv1.PlayerColor_PLAYER_COLOR_GREEN,
		catanv1.PlayerColor_PLAYER_COLOR_ORANGE,
	}
	return colors[index%len(colors)]
}

func toPlayerInfos(players []*catanv1.PlayerState) []*catanv1.PlayerInfo {
	infos := make([]*catanv1.PlayerInfo, 0, len(players))
	for _, player := range players {
		infos = append(infos, &catanv1.PlayerInfo{
			Id:    player.Id,
			Name:  player.Name,
			Color: player.Color,
		})
	}
	return infos
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// HandleHealth handles GET /health to check if the server is ready.
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check database connectivity
	if err := h.db.Ping(); err != nil {
		http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "settlers-backend",
	})
}

type clientEnvelope struct {
	Message struct {
		OneofKind      string          `json:"oneofKind"`
		StartGame      json.RawMessage `json:"startGame,omitempty"`
		RollDice       json.RawMessage `json:"rollDice,omitempty"`
		BuildStructure json.RawMessage `json:"buildStructure,omitempty"`
		EndTurn        json.RawMessage `json:"endTurn,omitempty"`
		PlayerReady    json.RawMessage `json:"playerReady,omitempty"`
		ProposeTrade   json.RawMessage `json:"proposeTrade,omitempty"`
		RespondTrade   json.RawMessage `json:"respondTrade,omitempty"`
		DiscardCards   json.RawMessage `json:"discardCards,omitempty"`
		MoveRobber     json.RawMessage `json:"moveRobber,omitempty"`
		BankTrade      json.RawMessage `json:"bankTrade,omitempty"`
		SetTurnPhase   json.RawMessage `json:"setTurnPhase,omitempty"`
		BuyDevCard     json.RawMessage `json:"buyDevCard,omitempty"`
		PlayDevCard    json.RawMessage `json:"playDevCard,omitempty"`
	} `json:"message"`
}

func (e clientEnvelope) payloadBytes(payload json.RawMessage) []byte {
	if len(payload) == 0 {
		return []byte("{}")
	}
	return payload
}
