package game

import (
	pb "settlers_from_catan/gen/proto/catan/v1"
)

// LongestRoadResult holds the results per player
type LongestRoadResult struct {
	PlayerId string
	Length   int
}

// GetLongestRoadLengths returns road lengths for all players.
func GetLongestRoadLengths(board *pb.BoardState, vertices []*pb.Vertex) map[string]int {
	lengths := map[string]int{}
	roadEdges := map[string]*pb.Edge{}
	for _, e := range board.Edges {
		if e.Road != nil {
			roadEdges[e.Id] = e
		}
	}
	// For each player, run DFS from each road edge
	playerRoadEdges := map[string][]*pb.Edge{}
	for _, e := range roadEdges {
		owner := e.Road.OwnerId
		playerRoadEdges[owner] = append(playerRoadEdges[owner], e)
	}
	for playerId, edges := range playerRoadEdges {
		maxLen := 0
		for _, edge := range edges {
			visited := map[string]bool{}
			length := dfsLongestRoad(board, vertices, edge, visited, playerId)
			if length > maxLen {
				maxLen = length
			}
		}
		lengths[playerId] = maxLen
	}
	return lengths
}

// dfsLongestRoad performs DFS and returns length of path
func dfsLongestRoad(board *pb.BoardState, vertices []*pb.Vertex, current *pb.Edge, visited map[string]bool, playerId string) int {
	visited[current.Id] = true
	maxPath := 1
	for _, vId := range current.Vertices {
		v := getVertexById(vertices, vId)
		if v == nil {
			continue
		}
		if isBlockedByOpponentBuilding(v, playerId) {
			continue
		}
		adjEdges := getEdgesConnectedToVertex(board, v.Id)
		for _, adj := range adjEdges {
			if adj.Id == current.Id || visited[adj.Id] {
				continue
			}
			if adj.Road == nil || adj.Road.OwnerId != playerId {
				continue
			}
			path := 1 + dfsLongestRoad(board, vertices, adj, visited, playerId)
			if path > maxPath {
				maxPath = path
			}
		}
	}
	visited[current.Id] = false // Backtrack
	return maxPath
}

func isBlockedByOpponentBuilding(vertex *pb.Vertex, playerId string) bool {
	if vertex.Building != nil && vertex.Building.OwnerId != playerId {
		return true
	}
	return false
}

func getVertexById(vertices []*pb.Vertex, id string) *pb.Vertex {
	for _, v := range vertices {
		if v.Id == id {
			return v
		}
	}
	return nil
}

func getEdgesConnectedToVertex(board *pb.BoardState, vertexId string) []*pb.Edge {
	var edges []*pb.Edge
	for _, e := range board.Edges {
		for _, vId := range e.Vertices {
			if vId == vertexId {
				edges = append(edges, e)
			}
		}
	}
	return edges
}

// GetLongestRoadPlayerId returns the player who currently has longest road (>=5), ties go to current holder.
func GetLongestRoadPlayerId(state *pb.GameState) string {
	lengths := GetLongestRoadLengths(state.Board, state.Board.Vertices)
	maxLen := 0
	claimers := []string{}
	for pid, l := range lengths {
		if l > maxLen {
			maxLen = l
			claimers = []string{pid}
		} else if l == maxLen && maxLen > 0 {
			claimers = append(claimers, pid)
		}
	}
	if maxLen >= 5 {
		// Award to sole maxLen owner or tie-break to current holder
		current := ""
		if state.LongestRoadPlayerId != nil {
			current = *state.LongestRoadPlayerId
		}
		for _, c := range claimers {
			if c == current {
				return c
			}
		}
		// Otherwise, first in claimers
		return claimers[0]
	}
	return "" // No one qualifies
}

// UpdateLongestRoadBonus recalculates and updates the longest road bonus holder
func UpdateLongestRoadBonus(state *pb.GameState) {
	newHolderID := GetLongestRoadPlayerId(state)
	currentHolderID := ""
	if state.LongestRoadPlayerId != nil {
		currentHolderID = *state.LongestRoadPlayerId
	}

	if newHolderID != currentHolderID {
		// Bonus is changing hands
		if newHolderID == "" {
			// No one qualifies anymore
			state.LongestRoadPlayerId = nil
		} else {
			// Award to new holder
			state.LongestRoadPlayerId = &newHolderID
		}
	}
}
