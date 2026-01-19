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

// Standard port locations by coastal edge (TO BE MATCHED to vertex IDs after generateVertices)
var standardPortEdges = [][]string{
	{"7.000,2.000", "8.000,1.000"}, // real edge IDs arranged clockwise by board geometry
	{"8.000,1.000", "8.000,-1.000"},
	{"8.000,-1.000", "7.000,-2.000"},
	{"7.000,-2.000", "5.000,-3.000"},
	{"3.000,-3.000", "1.000,-3.000"},
	{"-1.000,-3.000", "-3.000,-2.000"},
	{"-4.000,-1.000", "-4.000,1.000"},
	{"-3.000,2.000", "-1.000,3.000"},
	{"1.000,3.000", "3.000,3.000"},
}

// GeneratePorts returns the 9 standard ports. Assumes standard board layout/scale.
func GeneratePorts() []*catanv1.Port {
	// For standard map, ports are fixed.
	ports := make([]*catanv1.Port, 9)
	order := rand.Perm(9)
	for i, idx := range order {
		typ := portTypes[idx]
		res := portResources[idx]
		ports[i] = &catanv1.Port{
			Location: append([]string{}, standardPortEdges[i]...),
			Type:     typ,
			Resource: res,
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
