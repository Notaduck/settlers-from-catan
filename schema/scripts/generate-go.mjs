import { readFileSync, writeFileSync, mkdirSync } from "fs";
import { parse } from "yaml";

// Read and parse the AsyncAPI spec
const spec = parse(readFileSync("./asyncapi.yaml", "utf-8"));
const schemas = spec.components.schemas;

// Go output
let goOutput = `// Code generated from asyncapi.yaml - DO NOT EDIT.

package types

`;

// Generate inline struct for nested objects
function generateInlineStruct(schema, indent = "\t\t") {
  if (!schema.properties) return "map[string]interface{}";

  const required = schema.required || [];
  let result = "struct {\n";

  for (const [propName, propSchema] of Object.entries(schema.properties)) {
    const goName = toGoName(propName);
    const isRequired = required.includes(propName);
    let goType = schemaToGo(propSchema, indent + "\t");

    // Make optional non-primitives pointers
    if (
      !isRequired &&
      !["int", "bool", "float64", "string"].includes(goType) &&
      !goType.startsWith("[]") &&
      !goType.startsWith("map")
    ) {
      goType = "*" + goType;
    }

    result += `${indent}${goName} ${goType} ${jsonTag(propName, isRequired)}\n`;
  }

  result += `${indent.slice(0, -1)}}`;
  return result;
}

// Helper to convert schema type to Go
function schemaToGo(schema, indent = "\t\t") {
  if (!schema) return "interface{}";

  // Handle $ref
  if (schema.$ref) {
    const refName = schema.$ref.split("/").pop();
    return refName;
  }

  // Handle const - Go doesn't have const types, just use string
  if (schema.const !== undefined) {
    return "string";
  }

  // Handle enum - these become type aliases
  if (schema.enum) {
    return "string"; // Will be handled as separate type
  }

  // Handle type
  switch (schema.type) {
    case "string":
      return "string";
    case "integer":
      return "int";
    case "number":
      return "float64";
    case "boolean":
      return "bool";
    case "array":
      const itemType = schemaToGo(schema.items, indent);
      return `[]${itemType}`;
    case "object":
      if (!schema.properties) return "map[string]interface{}";
      // Inline struct for anonymous objects
      return generateInlineStruct(schema, indent);
    default:
      return "interface{}";
  }
}

function toGoName(name) {
  // Convert camelCase to PascalCase
  return name.charAt(0).toUpperCase() + name.slice(1);
}

function jsonTag(name, required) {
  if (required) {
    return `\`json:"${name}"\``;
  }
  return `\`json:"${name},omitempty"\``;
}

// Generate enums as type + constants
const enumTypes = {
  Resource: ["Wood", "Brick", "Sheep", "Wheat", "Ore"],
  TileResource: ["Wood", "Brick", "Sheep", "Wheat", "Ore", "Desert"],
  BuildingType: ["Settlement", "City"],
  StructureType: ["Settlement", "City", "Road"],
  GameStatus: ["Waiting", "Setup", "Playing", "Finished"],
  TurnPhase: ["Roll", "Trade", "Build"],
  PlayerColor: ["Red", "Blue", "Green", "Orange"],
  DevCardType: [
    "Knight",
    "RoadBuilding",
    "YearOfPlenty",
    "Monopoly",
    "VictoryPoint",
  ],
};

const enumValues = {
  Resource: ["wood", "brick", "sheep", "wheat", "ore"],
  TileResource: ["wood", "brick", "sheep", "wheat", "ore", "desert"],
  BuildingType: ["settlement", "city"],
  StructureType: ["settlement", "city", "road"],
  GameStatus: ["waiting", "setup", "playing", "finished"],
  TurnPhase: ["roll", "trade", "build"],
  PlayerColor: ["red", "blue", "green", "orange"],
  DevCardType: [
    "knight",
    "road_building",
    "year_of_plenty",
    "monopoly",
    "victory_point",
  ],
};

for (const [typeName, constants] of Object.entries(enumTypes)) {
  goOutput += `type ${typeName} string\n\n`;
  goOutput += `const (\n`;
  for (let i = 0; i < constants.length; i++) {
    goOutput += `\t${typeName}${constants[i]} ${typeName} = "${enumValues[typeName][i]}"\n`;
  }
  goOutput += `)\n\n`;
}

// Generate core structs
function generateStruct(name, schema) {
  if (!schema || schema.enum || schema.type !== "object") return "";

  let output = `type ${name} struct {\n`;
  const required = schema.required || [];

  if (schema.properties) {
    for (const [propName, propSchema] of Object.entries(schema.properties)) {
      const goName = toGoName(propName);
      const isRequired = required.includes(propName);
      let goType = schemaToGo(propSchema);

      // Handle inline objects
      if (goType === "struct" && propSchema.properties) {
        const inlineRequired = propSchema.required || [];
        let inline = "struct {\n";
        for (const [iProp, iSchema] of Object.entries(propSchema.properties)) {
          const iGoName = toGoName(iProp);
          const iIsRequired = inlineRequired.includes(iProp);
          let iGoType = schemaToGo(iSchema);
          if (
            !iIsRequired &&
            !["int", "bool", "float64", "string"].includes(iGoType) &&
            !iGoType.startsWith("[]")
          ) {
            iGoType = "*" + iGoType;
          }
          inline += `\t\t${iGoName} ${iGoType} ${jsonTag(
            iProp,
            iIsRequired
          )}\n`;
        }
        inline += "\t}";
        goType = inline;
      } else if (
        !isRequired &&
        !["int", "bool", "float64", "string"].includes(goType) &&
        !goType.startsWith("[]") &&
        !goType.startsWith("map")
      ) {
        goType = "*" + goType;
      }

      output += `\t${goName} ${goType} ${jsonTag(propName, isRequired)}\n`;
    }
  }

  output += `}\n\n`;
  return output;
}

// Generate structs in order
const structOrder = [
  "HexCoord",
  "Hex",
  "Building",
  "Road",
  "Vertex",
  "Edge",
  "ResourceCount",
  "PlayerState",
  "BoardState",
  "GameState",
  "TradeOffer",
  // Client messages
  "JoinGameMessage",
  "StartGameMessage",
  "RollDiceMessage",
  "BuildStructureMessage",
  "ProposeTradeMessage",
  "RespondTradeMessage",
  "MoveRobberMessage",
  "EndTurnMessage",
  "PlayDevCardMessage",
  // Server messages
  "GameStateMessage",
  "PlayerJoinedMessage",
  "PlayerLeftMessage",
  "DiceRolledMessage",
  "BuildingPlacedMessage",
  "RoadPlacedMessage",
  "TradeProposedMessage",
  "TradeResolvedMessage",
  "RobberMovedMessage",
  "TurnChangedMessage",
  "GameStartedMessage",
  "GameOverMessage",
  "ErrorMessage",
  // REST
  "CreateGameRequest",
  "CreateGameResponse",
  "JoinGameRequest",
  "JoinGameResponse",
  "GameInfoResponse",
];

for (const name of structOrder) {
  goOutput += generateStruct(name, schemas[name]);
}

// Generate message type constants
goOutput += `// Message type constants
const (
\t// Client -> Server
\tMsgTypeJoinGame    = "JOIN_GAME"
\tMsgTypeStartGame   = "START_GAME"
\tMsgTypeRollDice    = "ROLL_DICE"
\tMsgTypeBuild       = "BUILD"
\tMsgTypeProposeTrade = "PROPOSE_TRADE"
\tMsgTypeRespondTrade = "RESPOND_TRADE"
\tMsgTypeMoveRobber  = "MOVE_ROBBER"
\tMsgTypeEndTurn     = "END_TURN"
\tMsgTypePlayDevCard = "PLAY_DEV_CARD"
\t// Server -> Client
\tMsgTypeGameState      = "GAME_STATE"
\tMsgTypePlayerJoined   = "PLAYER_JOINED"
\tMsgTypePlayerLeft     = "PLAYER_LEFT"
\tMsgTypeDiceRolled     = "DICE_ROLLED"
\tMsgTypeBuildingPlaced = "BUILDING_PLACED"
\tMsgTypeRoadPlaced     = "ROAD_PLACED"
\tMsgTypeTradeProposed  = "TRADE_PROPOSED"
\tMsgTypeTradeResolved  = "TRADE_RESOLVED"
\tMsgTypeRobberMoved    = "ROBBER_MOVED"
\tMsgTypeTurnChanged    = "TURN_CHANGED"
\tMsgTypeGameStarted    = "GAME_STARTED"
\tMsgTypeGameOver       = "GAME_OVER"
\tMsgTypeError          = "ERROR"
)

// BaseMessage is used for initial message type detection
type BaseMessage struct {
\tType string \`json:"type"\`
}
`;

// Write Go file
const goDir = "../backend/gen/types";
mkdirSync(goDir, { recursive: true });
writeFileSync(`${goDir}/types.go`, goOutput);

console.log("✅ Generated Go types at backend/gen/types/types.go");
