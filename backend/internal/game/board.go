package game

import (
	"fmt"
	"math/rand"

	pb "settlers_from_catan/gen/proto/catan/v1"
)

// Standard Catan board resources (19 hex tiles)
var standardResources = []pb.TileResource{
	// 4 wood
	pb.TileResource_TILE_RESOURCE_WOOD, pb.TileResource_TILE_RESOURCE_WOOD,
	pb.TileResource_TILE_RESOURCE_WOOD, pb.TileResource_TILE_RESOURCE_WOOD,
	// 4 sheep
	pb.TileResource_TILE_RESOURCE_SHEEP, pb.TileResource_TILE_RESOURCE_SHEEP,
	pb.TileResource_TILE_RESOURCE_SHEEP, pb.TileResource_TILE_RESOURCE_SHEEP,
	// 4 wheat
	pb.TileResource_TILE_RESOURCE_WHEAT, pb.TileResource_TILE_RESOURCE_WHEAT,
	pb.TileResource_TILE_RESOURCE_WHEAT, pb.TileResource_TILE_RESOURCE_WHEAT,
	// 3 brick
	pb.TileResource_TILE_RESOURCE_BRICK, pb.TileResource_TILE_RESOURCE_BRICK,
	pb.TileResource_TILE_RESOURCE_BRICK,
	// 3 ore
	pb.TileResource_TILE_RESOURCE_ORE, pb.TileResource_TILE_RESOURCE_ORE,
	pb.TileResource_TILE_RESOURCE_ORE,
	// 1 desert
	pb.TileResource_TILE_RESOURCE_DESERT,
}

// Standard Catan number tokens (excluding desert)
var standardNumbers = []int32{
	2,
	3, 3,
	4, 4,
	5, 5,
	6, 6,
	8, 8,
	9, 9,
	10, 10,
	11, 11,
	12,
}

// Axial coordinates for standard Catan board (spiral layout)
var hexCoords = []struct{ q, r int32 }{
	// Center
	{0, 0},
	// Inner ring (clockwise from top)
	{0, -1}, {1, -1}, {1, 0}, {0, 1}, {-1, 1}, {-1, 0},
	// Outer ring (clockwise from top)
	{0, -2}, {1, -2}, {2, -2}, {2, -1}, {2, 0}, {1, 1}, {0, 2}, {-1, 2}, {-2, 2}, {-2, 1}, {-2, 0}, {-1, -1},
}

// GenerateBoard creates a randomized standard Catan board
func GenerateBoard() *pb.BoardState {
	// Shuffle resources
	resources := make([]pb.TileResource, len(standardResources))
	copy(resources, standardResources)
	rand.Shuffle(len(resources), func(i, j int) {
		resources[i], resources[j] = resources[j], resources[i]
	})

	// Shuffle numbers
	numbers := make([]int32, len(standardNumbers))
	copy(numbers, standardNumbers)
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})

	// Create hexes
	hexes := make([]*pb.Hex, len(hexCoords))
	numberIdx := 0
	var desertCoord *pb.HexCoord

	for i, coord := range hexCoords {
		hex := &pb.Hex{
			Coord: &pb.HexCoord{
				Q: coord.q,
				R: coord.r,
			},
			Resource: resources[i],
		}

		if resources[i] == pb.TileResource_TILE_RESOURCE_DESERT {
			hex.Number = 0
			desertCoord = hex.Coord
		} else {
			hex.Number = numbers[numberIdx]
			numberIdx++
		}

		hexes[i] = hex
	}

	// Generate vertices and edges
	vertices := generateVertices(hexes)
	edges := generateEdges(vertices)

	return &pb.BoardState{
		Hexes:     hexes,
		Vertices:  vertices,
		Edges:     edges,
		RobberHex: desertCoord,
	}
}

// Vertex directions relative to a hex center (6 corners)
// We use a naming convention: each vertex is identified by its position
// relative to the hex that "owns" it
var vertexOffsets = []struct {
	name string
	dq   float64 // offset in q direction (fractional)
	dr   float64 // offset in r direction (fractional)
}{
	{"N", 0, -0.67},
	{"NE", 0.5, -0.33},
	{"SE", 0.5, 0.33},
	{"S", 0, 0.67},
	{"SW", -0.5, 0.33},
	{"NW", -0.5, -0.33},
}

func generateVertices(hexes []*pb.Hex) []*pb.Vertex {
	// Use a map to deduplicate vertices (vertices are shared between hexes)
	vertexMap := make(map[string]*pb.Vertex)

	for _, hex := range hexes {
		for i := range vertexOffsets {
			// Calculate approximate position for deduplication
			vq := float64(hex.Coord.Q) + vertexOffsets[i].dq
			vr := float64(hex.Coord.R) + vertexOffsets[i].dr

			// Create a key based on approximate position (rounded to handle float imprecision)
			key := fmt.Sprintf("%.1f,%.1f", vq, vr)

			if _, exists := vertexMap[key]; !exists {
				vertexMap[key] = &pb.Vertex{
					Id:            key,
					AdjacentHexes: []*pb.HexCoord{},
				}
			}

			// Add this hex as adjacent to the vertex
			vertexMap[key].AdjacentHexes = append(vertexMap[key].AdjacentHexes, hex.Coord)
		}
	}

	// Convert map to slice
	vertices := make([]*pb.Vertex, 0, len(vertexMap))
	for _, v := range vertexMap {
		vertices = append(vertices, v)
	}

	return vertices
}

func generateEdges(vertices []*pb.Vertex) []*pb.Edge {
	// Build vertex adjacency based on shared hexes
	// Two vertices are connected if they share at least one adjacent hex
	// and are close enough to each other

	type vertexPos struct {
		q, r float64
	}

	// Parse vertex positions
	positions := make(map[string]vertexPos)
	for _, v := range vertices {
		var q, r float64
		fmt.Sscanf(v.Id, "%f,%f", &q, &r)
		positions[v.Id] = vertexPos{q, r}
	}

	// Find edges (pairs of vertices that are adjacent)
	edgeMap := make(map[string]*pb.Edge)

	for i, v1 := range vertices {
		p1 := positions[v1.Id]

		for j := i + 1; j < len(vertices); j++ {
			v2 := vertices[j]
			p2 := positions[v2.Id]

			// Check if vertices are adjacent (distance should be ~0.67 in hex units)
			dist := (p1.q-p2.q)*(p1.q-p2.q) + (p1.r-p2.r)*(p1.r-p2.r)
			if dist > 0.2 && dist < 0.8 { // Adjacent vertices
				// Check if they share at least one hex
				if sharesHex(v1.AdjacentHexes, v2.AdjacentHexes) {
					edgeID := fmt.Sprintf("%s-%s", v1.Id, v2.Id)
					if v1.Id > v2.Id {
						edgeID = fmt.Sprintf("%s-%s", v2.Id, v1.Id)
					}

					edgeMap[edgeID] = &pb.Edge{
						Id:       edgeID,
						Vertices: []string{v1.Id, v2.Id},
					}
				}
			}
		}
	}

	edges := make([]*pb.Edge, 0, len(edgeMap))
	for _, e := range edgeMap {
		edges = append(edges, e)
	}

	return edges
}

func sharesHex(hexes1, hexes2 []*pb.HexCoord) bool {
	for _, h1 := range hexes1 {
		for _, h2 := range hexes2 {
			if h1.Q == h2.Q && h1.R == h2.R {
				return true
			}
		}
	}
	return false
}

// NewGameState creates a new game state with the given players
func NewGameState(gameID, code string, playerNames []string, playerIDs []string) *pb.GameState {
	colors := []pb.PlayerColor{
		pb.PlayerColor_PLAYER_COLOR_RED,
		pb.PlayerColor_PLAYER_COLOR_BLUE,
		pb.PlayerColor_PLAYER_COLOR_GREEN,
		pb.PlayerColor_PLAYER_COLOR_ORANGE,
	}

	players := make([]*pb.PlayerState, len(playerNames))
	for i, name := range playerNames {
		players[i] = &pb.PlayerState{
			Id:            playerIDs[i],
			Name:          name,
			Color:         colors[i%len(colors)],
			Resources:     &pb.ResourceCount{},
			DevCardCount:  0,
			KnightsPlayed: 0,
			VictoryPoints: 0,
			Connected:     false,
			IsReady:       false,
			IsHost:        i == 0, // First player is the host
		}
	}

	return &pb.GameState{
		Id:          gameID,
		Code:        code,
		Board:       GenerateBoard(),
		Players:     players,
		CurrentTurn: 0,
		TurnPhase:   pb.TurnPhase_TURN_PHASE_ROLL,
		Dice:        []int32{0, 0},
		Status:      pb.GameStatus_GAME_STATUS_WAITING,
	}
}
