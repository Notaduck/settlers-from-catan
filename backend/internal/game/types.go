package game

// Re-export all generated types for convenience
// DO NOT add manual types here - update proto/catan/v1/*.proto instead

import (
	pb "settlers_from_catan/gen/proto/catan/v1"
)

// Type aliases for convenience
type (
	Resource      = pb.Resource
	TileResource  = pb.TileResource
	BuildingType  = pb.BuildingType
	StructureType = pb.StructureType
	GameStatus    = pb.GameStatus
	TurnPhase     = pb.TurnPhase
	PlayerColor   = pb.PlayerColor
	DevCardType   = pb.DevCardType
	TradeStatus   = pb.TradeStatus
	HexCoord      = pb.HexCoord
	Hex           = pb.Hex
	Vertex        = pb.Vertex
	Edge          = pb.Edge
	Building      = pb.Building
	Road          = pb.Road
	ResourceCount = pb.ResourceCount
	PlayerState   = pb.PlayerState
	BoardState    = pb.BoardState
	GameState     = pb.GameState
	TradeOffer    = pb.TradeOffer
	SetupPhase    = pb.SetupPhase
	// Messages
	ClientMessage = pb.ClientMessage
	ServerMessage = pb.ServerMessage
)

// Re-export constants
const (
	// Resources
	ResourceUnspecified = pb.Resource_RESOURCE_UNSPECIFIED
	ResourceWood        = pb.Resource_RESOURCE_WOOD
	ResourceBrick       = pb.Resource_RESOURCE_BRICK
	ResourceSheep       = pb.Resource_RESOURCE_SHEEP
	ResourceWheat       = pb.Resource_RESOURCE_WHEAT
	ResourceOre         = pb.Resource_RESOURCE_ORE

	// Tile Resources (includes desert)
	TileResourceUnspecified = pb.TileResource_TILE_RESOURCE_UNSPECIFIED
	TileResourceWood        = pb.TileResource_TILE_RESOURCE_WOOD
	TileResourceBrick       = pb.TileResource_TILE_RESOURCE_BRICK
	TileResourceSheep       = pb.TileResource_TILE_RESOURCE_SHEEP
	TileResourceWheat       = pb.TileResource_TILE_RESOURCE_WHEAT
	TileResourceOre         = pb.TileResource_TILE_RESOURCE_ORE
	TileResourceDesert      = pb.TileResource_TILE_RESOURCE_DESERT

	// Building Types
	BuildingTypeUnspecified = pb.BuildingType_BUILDING_TYPE_UNSPECIFIED
	BuildingTypeSettlement  = pb.BuildingType_BUILDING_TYPE_SETTLEMENT
	BuildingTypeCity        = pb.BuildingType_BUILDING_TYPE_CITY

	// Structure Types
	StructureTypeUnspecified = pb.StructureType_STRUCTURE_TYPE_UNSPECIFIED
	StructureTypeSettlement  = pb.StructureType_STRUCTURE_TYPE_SETTLEMENT
	StructureTypeCity        = pb.StructureType_STRUCTURE_TYPE_CITY
	StructureTypeRoad        = pb.StructureType_STRUCTURE_TYPE_ROAD

	// Game Status
	GameStatusUnspecified = pb.GameStatus_GAME_STATUS_UNSPECIFIED
	GameStatusWaiting     = pb.GameStatus_GAME_STATUS_WAITING
	GameStatusSetup       = pb.GameStatus_GAME_STATUS_SETUP
	GameStatusPlaying     = pb.GameStatus_GAME_STATUS_PLAYING
	GameStatusFinished    = pb.GameStatus_GAME_STATUS_FINISHED

	// Turn Phases
	TurnPhaseUnspecified = pb.TurnPhase_TURN_PHASE_UNSPECIFIED
	TurnPhaseRoll        = pb.TurnPhase_TURN_PHASE_ROLL
	TurnPhaseTrade       = pb.TurnPhase_TURN_PHASE_TRADE
	TurnPhaseBuild       = pb.TurnPhase_TURN_PHASE_BUILD

	// Player Colors
	PlayerColorUnspecified = pb.PlayerColor_PLAYER_COLOR_UNSPECIFIED
	PlayerColorRed         = pb.PlayerColor_PLAYER_COLOR_RED
	PlayerColorBlue        = pb.PlayerColor_PLAYER_COLOR_BLUE
	PlayerColorGreen       = pb.PlayerColor_PLAYER_COLOR_GREEN
	PlayerColorOrange      = pb.PlayerColor_PLAYER_COLOR_ORANGE

	// Dev Card Types
	DevCardTypeUnspecified  = pb.DevCardType_DEV_CARD_TYPE_UNSPECIFIED
	DevCardTypeKnight       = pb.DevCardType_DEV_CARD_TYPE_KNIGHT
	DevCardTypeRoadBuilding = pb.DevCardType_DEV_CARD_TYPE_ROAD_BUILDING
	DevCardTypeYearOfPlenty = pb.DevCardType_DEV_CARD_TYPE_YEAR_OF_PLENTY
	DevCardTypeMonopoly     = pb.DevCardType_DEV_CARD_TYPE_MONOPOLY
	DevCardTypeVictoryPoint = pb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT

	// Trade Status
	TradeStatusUnspecified = pb.TradeStatus_TRADE_STATUS_UNSPECIFIED
	TradeStatusPending     = pb.TradeStatus_TRADE_STATUS_PENDING
	TradeStatusAccepted    = pb.TradeStatus_TRADE_STATUS_ACCEPTED
	TradeStatusRejected    = pb.TradeStatus_TRADE_STATUS_REJECTED
	TradeStatusCancelled   = pb.TradeStatus_TRADE_STATUS_CANCELLED
)
