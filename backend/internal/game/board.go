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
	edges := generateEdges(hexes)

	// Create preliminary board state
	board := &pb.BoardState{
		Hexes:     hexes,
		Vertices:  vertices,
		Edges:     edges,
		RobberHex: desertCoord,
	}

	// Add standard ports (maritime trading) - needs vertices to be generated first
	board.Ports = GeneratePortsForBoard(board)
	return board
}

// Vertex directions relative to a hex center (6 corners).
// Use a fixed integer scale to keep vertex IDs deterministic and deduplicated.
const vertexCoordScale = 3

var vertexOffsets = []struct {
	name string
	dq   int
	dr   int
}{
	{"N", -1, 2},
	{"NE", 1, 1},
	{"SE", 2, -1},
	{"S", 1, -2},
	{"SW", -1, -1},
	{"NW", -2, 1},
}

func generateVertices(hexes []*pb.Hex) []*pb.Vertex {
	// Use a map to deduplicate vertices (vertices are shared between hexes)
	type vertexKey struct {
		q int
		r int
	}
	vertexMap := make(map[vertexKey]*pb.Vertex)

	for _, hex := range hexes {
		if hex.Coord == nil {
			continue
		}
		for i := range vertexOffsets {
			vq := int(hex.Coord.Q)*vertexCoordScale + vertexOffsets[i].dq
			vr := int(hex.Coord.R)*vertexCoordScale + vertexOffsets[i].dr
			key := vertexKey{q: vq, r: vr}

			if _, exists := vertexMap[key]; !exists {
				vertexMap[key] = &pb.Vertex{
					Id:            formatVertexID(vq, vr),
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

func generateEdges(hexes []*pb.Hex) []*pb.Edge {
	edgeMap := make(map[string]*pb.Edge)

	for _, hex := range hexes {
		if hex.Coord == nil {
			continue
		}
		vertexIDs := make([]string, len(vertexOffsets))
		for i, offset := range vertexOffsets {
			vertexIDs[i] = vertexIDForHex(hex.Coord, offset)
		}

		for i := 0; i < len(vertexIDs); i++ {
			v1 := vertexIDs[i]
			v2 := vertexIDs[(i+1)%len(vertexIDs)]
			edgeID, vertices := orderedEdge(v1, v2)
			if _, exists := edgeMap[edgeID]; !exists {
				edgeMap[edgeID] = &pb.Edge{
					Id:       edgeID,
					Vertices: vertices,
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

func vertexIDForHex(coord *pb.HexCoord, offset struct {
	name string
	dq   int
	dr   int
}) string {
	vq := int(coord.Q)*vertexCoordScale + offset.dq
	vr := int(coord.R)*vertexCoordScale + offset.dr
	return formatVertexID(vq, vr)
}

func formatVertexID(vq, vr int) string {
	return fmt.Sprintf("%.3f,%.3f", float64(vq)/vertexCoordScale, float64(vr)/vertexCoordScale)
}

func orderedEdge(v1, v2 string) (string, []string) {
	if v1 < v2 {
		return fmt.Sprintf("%s-%s", v1, v2), []string{v1, v2}
	}
	return fmt.Sprintf("%s-%s", v2, v1), []string{v2, v1}
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
		DevCardDeck: InitDevCardDeck(),
	}
}
