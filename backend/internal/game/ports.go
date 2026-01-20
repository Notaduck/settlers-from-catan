package game

import (
	"math/rand"
	"settlers_from_catan/gen/proto/catan/v1"
)

// Port generation constants
var (
	portTypes = []catanv1.PortType{
		catanv1.PortType_PORT_TYPE_GENERIC, catanv1.PortType_PORT_TYPE_GENERIC,
		catanv1.PortType_PORT_TYPE_GENERIC, catanv1.PortType_PORT_TYPE_GENERIC,
		catanv1.PortType_PORT_TYPE_SPECIFIC, catanv1.PortType_PORT_TYPE_SPECIFIC,
		catanv1.PortType_PORT_TYPE_SPECIFIC, catanv1.PortType_PORT_TYPE_SPECIFIC, catanv1.PortType_PORT_TYPE_SPECIFIC,
	}

	portResources = []catanv1.Resource{
		catanv1.Resource_RESOURCE_UNSPECIFIED,
		catanv1.Resource_RESOURCE_UNSPECIFIED,
		catanv1.Resource_RESOURCE_UNSPECIFIED,
		catanv1.Resource_RESOURCE_UNSPECIFIED,
		catanv1.Resource_RESOURCE_WOOD,
		catanv1.Resource_RESOURCE_BRICK,
		catanv1.Resource_RESOURCE_SHEEP,
		catanv1.Resource_RESOURCE_WHEAT,
		catanv1.Resource_RESOURCE_ORE,
	}
)

// GeneratePortsForBoard returns the 9 standard ports, selecting coastal vertices from the board.
func GeneratePortsForBoard(board *catanv1.BoardState) []*catanv1.Port {
	// Find coastal vertices (vertices with fewer than 3 adjacent hexes)
	coastalVertices := []string{}
	for _, v := range board.Vertices {
		if len(v.AdjacentHexes) < 3 {
			coastalVertices = append(coastalVertices, v.Id)
		}
	}

	// Select 18 coastal vertices for 9 ports (2 vertices per port)
	// We'll use every 2nd coastal vertex to spread ports around the coast
	portVertices := [][]string{}
	step := len(coastalVertices) / 9
	if step < 2 {
		step = 2
	}

	for i := 0; i < 9 && i*step+1 < len(coastalVertices); i++ {
		portVertices = append(portVertices, []string{
			coastalVertices[i*step],
			coastalVertices[i*step+1],
		})
	}

	// If we don't have enough coastal vertices, fill in with remaining ones
	for len(portVertices) < 9 && len(portVertices)*2 < len(coastalVertices) {
		idx := len(portVertices) * 2
		if idx+1 < len(coastalVertices) {
			portVertices = append(portVertices, []string{
				coastalVertices[idx],
				coastalVertices[idx+1],
			})
		}
	}

	// Randomize port types and resources
	ports := make([]*catanv1.Port, len(portVertices))
	order := rand.Perm(len(portTypes))
	for i := range portVertices {
		if i >= len(portVertices) {
			break
		}
		idx := order[i%len(order)]
		ports[i] = &catanv1.Port{
			Location: portVertices[i],
			Type:     portTypes[idx],
			Resource: portResources[idx],
		}
	}

	return ports
}

// PlayerHasPortAccess returns true if player has a settlement/city on any vertex touching a port.
func PlayerHasPortAccess(playerID string, port *catanv1.Port, board *catanv1.BoardState) bool {
	for _, vid := range port.Location {
		for _, v := range board.Vertices {
			if v.Id != vid || v.Building == nil {
				continue
			}
			if v.Building.OwnerId == playerID {
				return true
			}
		}
	}
	return false
}

// GetBestTradeRatio returns best available trade ratio for player for the resource (default 4 for bank).
func GetBestTradeRatio(playerID string, resource catanv1.Resource, board *catanv1.BoardState) int {
	best := 4
	for _, port := range board.Ports {
		if PlayerHasPortAccess(playerID, port, board) {
			if port.Type == catanv1.PortType_PORT_TYPE_GENERIC {
				if best > 3 {
					best = 3
				}
			} else if port.Type == catanv1.PortType_PORT_TYPE_SPECIFIC && port.Resource == resource {
				if best > 2 {
					best = 2
				}
			}
		}
	}
	return best
}
