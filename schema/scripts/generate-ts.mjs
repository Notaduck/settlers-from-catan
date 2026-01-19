import { readFileSync, writeFileSync, mkdirSync } from "fs";
import { parse } from "yaml";

// Read and parse the AsyncAPI spec
const spec = parse(readFileSync("./asyncapi.yaml", "utf-8"));
const schemas = spec.components.schemas;

// TypeScript output
let tsOutput = `// AUTO-GENERATED FILE - DO NOT EDIT
// Generated from asyncapi.yaml

`;

// Helper to convert schema type to TypeScript
function schemaToTS(schema, indent = "") {
  if (!schema) return "unknown";

  // Handle $ref
  if (schema.$ref) {
    const refName = schema.$ref.split("/").pop();
    return refName;
  }

  // Handle const
  if (schema.const !== undefined) {
    return JSON.stringify(schema.const);
  }

  // Handle enum
  if (schema.enum) {
    return schema.enum.map((v) => JSON.stringify(v)).join(" | ");
  }

  // Handle type
  switch (schema.type) {
    case "string":
      return "string";
    case "integer":
    case "number":
      return "number";
    case "boolean":
      return "boolean";
    case "array":
      return `${schemaToTS(schema.items)}[]`;
    case "object":
      if (!schema.properties) return "Record<string, unknown>";
      const required = schema.required || [];
      const props = Object.entries(schema.properties)
        .map(([key, val]) => {
          const opt = required.includes(key) ? "" : "?";
          return `${indent}  ${key}${opt}: ${schemaToTS(val, indent + "  ")};`;
        })
        .join("\n");
      return `{\n${props}\n${indent}}`;
    default:
      return "unknown";
  }
}

// Generate enums first
const enumTypes = [
  "Resource",
  "TileResource",
  "BuildingType",
  "StructureType",
  "GameStatus",
  "TurnPhase",
  "PlayerColor",
  "DevCardType",
];

for (const name of enumTypes) {
  const schema = schemas[name];
  if (schema?.enum) {
    tsOutput += `export type ${name} = ${schema.enum
      .map((v) => JSON.stringify(v))
      .join(" | ")};\n`;
    tsOutput += `export const ${name}Values = [${schema.enum
      .map((v) => JSON.stringify(v))
      .join(", ")}] as const;\n\n`;
  }
}

// Generate interfaces
const interfaceOrder = [
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

for (const name of interfaceOrder) {
  const schema = schemas[name];
  if (!schema || schema.enum) continue;

  tsOutput += `export interface ${name} ${schemaToTS(schema)};\n\n`;
}

// Generate message type unions
tsOutput += `// Message type unions
export type ClientMessage =
  | JoinGameMessage
  | StartGameMessage
  | RollDiceMessage
  | BuildStructureMessage
  | ProposeTradeMessage
  | RespondTradeMessage
  | MoveRobberMessage
  | EndTurnMessage
  | PlayDevCardMessage;

export type ServerMessage =
  | GameStateMessage
  | PlayerJoinedMessage
  | PlayerLeftMessage
  | DiceRolledMessage
  | BuildingPlacedMessage
  | RoadPlacedMessage
  | TradeProposedMessage
  | TradeResolvedMessage
  | RobberMovedMessage
  | TurnChangedMessage
  | GameStartedMessage
  | GameOverMessage
  | ErrorMessage;

// Message type constants
export const MessageTypes = {
  // Client -> Server
  JOIN_GAME: 'JOIN_GAME',
  START_GAME: 'START_GAME',
  ROLL_DICE: 'ROLL_DICE',
  BUILD: 'BUILD',
  PROPOSE_TRADE: 'PROPOSE_TRADE',
  RESPOND_TRADE: 'RESPOND_TRADE',
  MOVE_ROBBER: 'MOVE_ROBBER',
  END_TURN: 'END_TURN',
  PLAY_DEV_CARD: 'PLAY_DEV_CARD',
  // Server -> Client
  GAME_STATE: 'GAME_STATE',
  PLAYER_JOINED: 'PLAYER_JOINED',
  PLAYER_LEFT: 'PLAYER_LEFT',
  DICE_ROLLED: 'DICE_ROLLED',
  BUILDING_PLACED: 'BUILDING_PLACED',
  ROAD_PLACED: 'ROAD_PLACED',
  TRADE_PROPOSED: 'TRADE_PROPOSED',
  TRADE_RESOLVED: 'TRADE_RESOLVED',
  ROBBER_MOVED: 'ROBBER_MOVED',
  TURN_CHANGED: 'TURN_CHANGED',
  GAME_STARTED: 'GAME_STARTED',
  GAME_OVER: 'GAME_OVER',
  ERROR: 'ERROR',
} as const;
`;

// Write TypeScript file
const tsDir = "../frontend/src/gen";
mkdirSync(tsDir, { recursive: true });
writeFileSync(`${tsDir}/types.ts`, tsOutput);

console.log("✅ Generated TypeScript types at frontend/src/gen/types.ts");
