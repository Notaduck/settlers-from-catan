package game

import (
	pb "settlers_from_catan/gen/proto/catan/v1"
	"testing"
)

func TestBoardGeneratesNinePorts(t *testing.T) {
	b := GenerateBoard()
	if len(b.Ports) != 9 {
		t.Fatalf("Board should have 9 ports, got %d", len(b.Ports))
	}
}

func TestPortDistribution(t *testing.T) {
	b := GenerateBoard()
	generic, specific := 0, 0
	resources := map[pb.Resource]bool{
		pb.Resource_RESOURCE_WOOD:  false,
		pb.Resource_RESOURCE_BRICK: false,
		pb.Resource_RESOURCE_SHEEP: false,
		pb.Resource_RESOURCE_WHEAT: false,
		pb.Resource_RESOURCE_ORE:   false,
	}
	for _, p := range b.Ports {
		switch p.Type {
		case pb.PortType_PORT_TYPE_GENERIC:
			generic++
		case pb.PortType_PORT_TYPE_SPECIFIC:
			specific++
			resources[p.Resource] = true
		}
	}
	if generic != 4 {
		t.Errorf("Expected 4 generic ports, got %d", generic)
	}
	if specific != 5 {
		t.Errorf("Expected 5 specific ports, got %d", specific)
	}
	for r, seen := range resources {
		if !seen {
			t.Errorf("Missing specific port for %v", r)
		}
	}
}

func TestPlayerGainsPortAccess(t *testing.T) {
	b := GenerateBoard()
	playerID := "p1"
	// Assign player settlement to first port location's first vertex
	for _, v := range b.Vertices {
		if v.Id == b.Ports[0].Location[0] {
			v.Building = &pb.Building{Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT, OwnerId: playerID}
			break
		}
	}
	if !PlayerHasPortAccess(playerID, b.Ports[0], b) {
		t.Errorf("Player should have port access at vertex %s", b.Ports[0].Location[0])
	}
}

func TestGetBestTradeRatio_Default(t *testing.T) {
	b := GenerateBoard()
	if GetBestTradeRatio("nobody", pb.Resource_RESOURCE_WOOD, b) != 4 {
		t.Errorf("Default trade ratio should be 4:1")
	}
}

func TestGetBestTradeRatio_GenericPort(t *testing.T) {
	b := GenerateBoard()
	playerID := "p1"
	var generic *pb.Port
	for _, p := range b.Ports {
		if p.Type == pb.PortType_PORT_TYPE_GENERIC {
			generic = p
			break
		}
	}
	if generic == nil {
		t.Fatalf("No generic port found")
	}
	for _, v := range b.Vertices {
		if v.Id == generic.Location[0] {
			v.Building = &pb.Building{Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT, OwnerId: playerID}
			break
		}
	}
	if GetBestTradeRatio(playerID, pb.Resource_RESOURCE_WOOD, b) != 3 {
		t.Errorf("Trade ratio should be 3:1 for player with generic port")
	}
}

func TestGetBestTradeRatio_SpecificPort(t *testing.T) {
	b := GenerateBoard()
	playerID := "p2"
	var specific *pb.Port
	for _, p := range b.Ports {
		if p.Type == pb.PortType_PORT_TYPE_SPECIFIC && p.Resource == pb.Resource_RESOURCE_WHEAT {
			specific = p
			break
		}
	}
	if specific == nil {
		t.Fatalf("No wheat-specific port found")
	}
	for _, v := range b.Vertices {
		if v.Id == specific.Location[0] {
			v.Building = &pb.Building{Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT, OwnerId: playerID}
			break
		}
	}
	if GetBestTradeRatio(playerID, pb.Resource_RESOURCE_WHEAT, b) != 2 {
		t.Errorf("Trade ratio should be 2:1 for wheat port")
	}
	if GetBestTradeRatio(playerID, pb.Resource_RESOURCE_WOOD, b) < 2 {
		t.Errorf("Trade ratio for other resource should not be better than 2:1")
	}
}
