import { Suspense, useEffect, useMemo, useState, useCallback } from "react";
import { DiscardModal } from "./DiscardModal";
import { StealModal } from "./StealModal";
import { BankTradeModal } from "./BankTradeModal";
import { ProposeTradeModal } from "./ProposeTradeModal";
import { IncomingTradeModal } from "./IncomingTradeModal";
import { DevelopmentCardsPanel } from "./DevelopmentCardsPanel";
import { YearOfPlentyModal } from "./YearOfPlentyModal";
import { MonopolyModal } from "./MonopolyModal";
import { useGame } from "@/context";
import { Board } from "@/components/Board";
import { PlayerPanel } from "@/components/PlayerPanel";
import {
  BuildingType,
  GameStatus,
  PlayerColor,
  StructureType,
  DevCardType,
  ResourceCount,
  TurnPhase,
} from "@/types";
import { GameOver } from "./GameOver";
import "./GameOver.css";
import "./Game.css";

// Map PlayerColor enum to CSS color strings
const PLAYER_COLORS: Record<PlayerColor, string> = {
  [PlayerColor.UNSPECIFIED]: "#808080",
  [PlayerColor.RED]: "#e74c3c",
  [PlayerColor.BLUE]: "#3498db",
  [PlayerColor.GREEN]: "#2ecc71",
  [PlayerColor.ORANGE]: "#e67e22",
};

const BUILDING_TYPE_ALIASES: Record<string, BuildingType> = {
  BUILDING_TYPE_UNSPECIFIED: BuildingType.UNSPECIFIED,
  BUILDING_TYPE_SETTLEMENT: BuildingType.SETTLEMENT,
  BUILDING_TYPE_CITY: BuildingType.CITY,
  SETTLEMENT: BuildingType.SETTLEMENT,
  CITY: BuildingType.CITY,
};

function normalizeBuildingType(
  type: BuildingType | string | undefined
): BuildingType {
  if (type === undefined || type === null) {
    return BuildingType.UNSPECIFIED;
  }
  if (typeof type === "string") {
    return BUILDING_TYPE_ALIASES[type] ?? BuildingType.UNSPECIFIED;
  }
  return type;
}

function isStatus(
  status: GameStatus | string | undefined,
  expected: GameStatus,
  expectedString: string
): boolean {
  return (
    status === expected ||
    status === (expectedString as unknown as GameStatus)
  );
}

function isTurnPhase(
  phase: TurnPhase | string | undefined,
  expected: TurnPhase,
  expectedString: string
): boolean {
  return (
    phase === expected ||
    phase === (expectedString as unknown as TurnPhase)
  );
}

interface GameProps {
  gameCode: string;
  onLeave: () => void;
}

export function Game({ gameCode, onLeave }: GameProps) {
  // Trading modals
  const [showBankTrade, setShowBankTrade] = useState(false);
  const [showProposeTrade, setShowProposeTrade] = useState(false);
  const [showIncomingTrade, setShowIncomingTrade] = useState(false);
  const [buildSelection, setBuildSelection] = useState<
    "settlement" | "road" | "city" | null
  >(null);
  const [pendingBuildSelection, setPendingBuildSelection] = useState<
    "settlement" | "road" | "city" | null
  >(null);
  
  // Dev card modals
  const [showYearOfPlenty, setShowYearOfPlenty] = useState(false);
  const [showMonopoly, setShowMonopoly] = useState(false);
  const {
    connect,
    disconnect,
    isConnected,
    gameState,
    error,
    startGame,
    setReady,
    currentPlayerId,
    placementMode,
    placementState,
    gameOver,
    build,
    resourceGain,
    clearResourceGain,
    rollDice,
    // Robber UI
    isRobberDiscardRequired,
    robberDiscardAmount,
    robberDiscardMax,
    sendRobberDiscard,
    isRobberMoveRequired,
    sendRobberMove,
    isRobberStealRequired,
    sendRobberSteal,
    robberStealCandidates,
    // Dev cards
    buyDevCard,
    playDevCard,
    // Trading
    proposeTrade,
    respondTrade,
    bankTrade,
    setTurnPhase,
  } = useGame();

  // UI state (for modal closing)
  const [discardClosed, setDiscardClosed] = useState(false);
  const [stealClosed, setStealClosed] = useState(false);

  useEffect(() => {
    if (isRobberDiscardRequired) {
      setDiscardClosed(false);
    }
  }, [isRobberDiscardRequired]);

  useEffect(() => {
    connect();
    return () => disconnect();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Only run once on mount/unmount

  // Memoize players array to prevent unnecessary re-renders
  const players = useMemo(() => gameState?.players ?? [], [gameState?.players]);

  // Find current player's state
  const currentPlayer = useMemo(
    () => players.find((p) => p.id === currentPlayerId),
    [players, currentPlayerId]
  );

  // Check for incoming trades that target the current player
  const incomingTradeOffer = useMemo(() => {
    if (!gameState?.pendingTrades || !currentPlayerId) return null;
    
    // Find a pending trade where:
    // 1. The current player is the target (or it's open to all)
    // 2. The trade is still pending/active
    // 3. The current player is not the proposer
    return gameState.pendingTrades.find(trade => 
      trade.proposerId !== currentPlayerId &&
      (trade.targetId === currentPlayerId || !trade.targetId) &&
      trade.status === 1 // Assuming 1 is PENDING status
    ) || null;
  }, [gameState?.pendingTrades, currentPlayerId]);

  const incomingTradeProposer = useMemo(() => {
    if (!incomingTradeOffer) return null;
    return players.find(p => p.id === incomingTradeOffer.proposerId) || null;
  }, [incomingTradeOffer, players]);

  // Auto-show incoming trade modal when a trade is pending
  useEffect(() => {
    if (incomingTradeOffer && !showIncomingTrade) {
      setShowIncomingTrade(true);
    }
  }, [incomingTradeOffer, showIncomingTrade]);

  // Check if all players are ready
  const allPlayersReady = useMemo(
    () => players.length >= 2 && players.every((p) => p.isReady),
    [players]
  );

  // Check if current player is host/ready
  const isHost = currentPlayer?.isHost ?? false;
  const isReady = currentPlayer?.isReady ?? false;
  const isWaiting = isStatus(
    gameState?.status,
    GameStatus.WAITING,
    "GAME_STATUS_WAITING"
  );
  const isSetup = isStatus(
    gameState?.status,
    GameStatus.SETUP,
    "GAME_STATUS_SETUP"
  );

  const placementModeLabel = useMemo(() => {
    switch (placementMode) {
      case "settlement":
        return "Place Settlement";
      case "road":
        return "Place Road";
      case "build":
        return "Build";
      default:
        return null;
    }
  }, [placementMode]);

  const resourceGainText = useMemo(() => {
    if (!resourceGain || resourceGain.playerId !== currentPlayerId) {
      return null;
    }

    const resourceLabels: Record<string, string> = {
      wood: "Wood",
      brick: "Brick",
      sheep: "Sheep",
      wheat: "Wheat",
      ore: "Ore",
    };

    const parts = Object.entries(resourceGain.resources)
      .filter(([, value]) => (value ?? 0) > 0)
      .map(([key, value]) => `${value} ${resourceLabels[key] ?? key}`);

    if (parts.length === 0) {
      return null;
    }

    return `You received: ${parts.join(", ")}`;
  }, [resourceGain, currentPlayerId]);

  useEffect(() => {
    if (!resourceGainText) {
      return undefined;
    }

    const timeoutId = window.setTimeout(() => {
      clearResourceGain();
    }, 5000);

    return () => window.clearTimeout(timeoutId);
  }, [resourceGainText, clearResourceGain]);

  const setupRound = gameState?.setupPhase?.round ?? 1;
  const currentTurnPlayer = players[gameState?.currentTurn ?? 0];
  const setupInstruction = useMemo(() => {
    if (!isSetup || !gameState?.setupPhase) {
      return null;
    }

    const placementsInTurn = gameState.setupPhase.placementsInTurn ?? 0;
    const placementCount = setupRound >= 2 ? 2 : 1;
    const label = placementsInTurn === 0 ? "Settlement" : "Road";

    return `Place ${label} (${placementCount}/2)`;
  }, [isSetup, gameState?.setupPhase, setupRound]);

  // Dev card logic (must be before early returns to satisfy rules of hooks)
  const isGameOver =
    !!gameOver ||
    isStatus(gameState?.status, GameStatus.FINISHED, "GAME_STATUS_FINISHED");
  const interactionsDisabled = isGameOver;
  const isMyTurn = gameState?.players?.[gameState.currentTurn ?? 0]?.id === currentPlayerId;
  const isPlaying = isStatus(gameState?.status, GameStatus.PLAYING, "GAME_STATUS_PLAYING");
  const isRollPhase = isTurnPhase(gameState?.turnPhase, TurnPhase.ROLL, "TURN_PHASE_ROLL");
  const isTradePhase = isTurnPhase(gameState?.turnPhase, TurnPhase.TRADE, "TURN_PHASE_TRADE");
  const isBuildPhase = isTurnPhase(gameState?.turnPhase, TurnPhase.BUILD, "TURN_PHASE_BUILD");

  useEffect(() => {
    if (!isPlaying || !isMyTurn) {
      setBuildSelection(null);
      setPendingBuildSelection(null);
    }
  }, [isPlaying, isMyTurn]);

  useEffect(() => {
    if (!pendingBuildSelection) {
      return;
    }
    if (!isPlaying || !isMyTurn) {
      setPendingBuildSelection(null);
      return;
    }
    if (isBuildPhase) {
      setPendingBuildSelection(null);
      return;
    }
    if (isTradePhase) {
      setTurnPhase(TurnPhase.BUILD);
      return;
    }
  }, [pendingBuildSelection, isPlaying, isMyTurn, isBuildPhase, isTradePhase, setTurnPhase]);

  const handleSelectBuild = useCallback(
    (selection: "settlement" | "road" | "city") => {
      setBuildSelection(selection);
      if (isRollPhase) {
        setPendingBuildSelection(selection);
        rollDice();
        return;
      }
      if (isTradePhase) {
        setTurnPhase(TurnPhase.BUILD);
      }
    },
    [isRollPhase, isTradePhase, rollDice, setTurnPhase]
  );

  const validCityVertexIds = useMemo(() => {
    const ids = new Set<string>();
    if (!gameState?.board || !currentPlayerId) {
      return ids;
    }
    const player = players.find((p) => p.id === currentPlayerId);
    if (!player) {
      return ids;
    }
    const hasResources =
      (player.resources?.ore ?? 0) >= 3 &&
      (player.resources?.wheat ?? 0) >= 2;
    if (!hasResources) {
      return ids;
    }
    for (const vertex of gameState.board.vertices ?? []) {
      const building = vertex.building;
      if (!building || building.ownerId !== currentPlayerId) {
        continue;
      }
      if (normalizeBuildingType(building.type) === BuildingType.SETTLEMENT) {
        ids.add(vertex.id);
      }
    }
    return ids;
  }, [gameState?.board, currentPlayerId, players]);

  const effectiveValidVertexIds = useMemo(() => {
    if (placementMode !== "build") {
      return placementState.validVertexIds;
    }
    if (buildSelection === "road") {
      return new Set<string>();
    }
    if (buildSelection === "city") {
      return validCityVertexIds;
    }
    return placementState.validVertexIds;
  }, [placementMode, placementState.validVertexIds, buildSelection, validCityVertexIds]);

  const effectiveValidEdgeIds = useMemo(() => {
    if (placementMode !== "build") {
      return placementState.validEdgeIds;
    }
    if (buildSelection === "settlement" || buildSelection === "city") {
      return new Set<string>();
    }
    return placementState.validEdgeIds;
  }, [placementMode, placementState.validEdgeIds, buildSelection]);

  const handleBuildVertex = useCallback(
    (vertexId: string) => {
      const isCityBuild =
        placementMode === "build" && buildSelection === "city";
      build(isCityBuild ? StructureType.CITY : StructureType.SETTLEMENT, vertexId);
    },
    [build, placementMode, buildSelection]
  );
  
  const canBuyDevCard = useMemo(() => {
    if (interactionsDisabled || !currentPlayer || !gameState) return false;
    const isMyTurn = gameState.players[gameState.currentTurn]?.id === currentPlayerId;
    const isTradeOrBuildPhase = 
      isTurnPhase(gameState.turnPhase, TurnPhase.TRADE, "TURN_PHASE_TRADE") ||
      isTurnPhase(gameState.turnPhase, TurnPhase.BUILD, "TURN_PHASE_BUILD");
    const hasResources = 
      (currentPlayer.resources?.ore ?? 0) >= 1 &&
      (currentPlayer.resources?.wheat ?? 0) >= 1 &&
      (currentPlayer.resources?.sheep ?? 0) >= 1;
    return isMyTurn && isTradeOrBuildPhase && hasResources;
  }, [interactionsDisabled, currentPlayer, gameState, currentPlayerId]);

  const canPlayDevCard = useMemo(() => {
    if (interactionsDisabled || !currentPlayer || !gameState) return false;
    const isMyTurn = gameState.players[gameState.currentTurn]?.id === currentPlayerId;
    const isPlaying = isStatus(
      gameState.status,
      GameStatus.PLAYING,
      "GAME_STATUS_PLAYING"
    );
    return isMyTurn && isPlaying;
  }, [interactionsDisabled, currentPlayer, gameState, currentPlayerId]);

  const handlePlayDevCard = useCallback((cardType: DevCardType) => {
    if (cardType === DevCardType.YEAR_OF_PLENTY) {
      setShowYearOfPlenty(true);
    } else if (cardType === DevCardType.MONOPOLY) {
      setShowMonopoly(true);
    } else {
      // Knight and Road Building can be played directly
      playDevCard(cardType);
    }
  }, [playDevCard]);

  if (error) {
    return (
      <div className="game-error" data-cy="game-error">
        <h2>Connection Error</h2>
        <p>{error}</p>
        <button
          onClick={onLeave}
          className="btn btn-secondary"
          data-cy="back-to-lobby-btn"
        >
          Back to Lobby
        </button>
      </div>
    );
  }

  if (!isConnected) {
    return (
      <div className="game-loading" data-cy="game-loading">
        <p>Connecting to game...</p>
      </div>
    );
  }

  if (!gameState) {
    return (
      <div className="game-loading" data-cy="game-loading">
        <p>Loading game state...</p>
      </div>
    );
  }

  return (
    <div className="game" data-cy="game">
      {/* BANK TRADE MODAL */}
      {!interactionsDisabled && (
        <BankTradeModal
          open={showBankTrade}
          onClose={() => setShowBankTrade(false)}
          onSubmit={(offering, offeringAmount, requested) => {
            // Create the proper ResourceCount for offering based on the resource type and amount
            const offeringResources: ResourceCount = {
              wood: offering === 1 ? offeringAmount : 0,
              brick: offering === 2 ? offeringAmount : 0,
              sheep: offering === 3 ? offeringAmount : 0,
              wheat: offering === 4 ? offeringAmount : 0,
              ore: offering === 5 ? offeringAmount : 0,
            };
            bankTrade(offeringResources, requested);
            setShowBankTrade(false);
          }}
          resources={currentPlayer?.resources ?? {wood:0,brick:0,sheep:0,wheat:0,ore:0}}
          board={gameState?.board}
          playerId={currentPlayerId ?? ''}
        />
      )}
      {/* PROPOSE TRADE MODAL */}
      {!interactionsDisabled && (
        <ProposeTradeModal
          open={showProposeTrade}
          onClose={() => setShowProposeTrade(false)}
          onSubmit={(offer, request, targetPlayerId) => {
            proposeTrade(offer, request, targetPlayerId);
            setShowProposeTrade(false);
          }}
          players={players}
          myResources={currentPlayer?.resources ?? {wood:0,brick:0,sheep:0,wheat:0,ore:0}}
        />
      )}
      {/* INCOMING TRADE MODAL */}
      {!interactionsDisabled && incomingTradeOffer && incomingTradeProposer && (
        <IncomingTradeModal
          open={showIncomingTrade}
          onAccept={() => { 
            if (incomingTradeOffer) {
              respondTrade(incomingTradeOffer.id, true);
            }
            setShowIncomingTrade(false); 
          }}
          onDecline={() => { 
            if (incomingTradeOffer) {
              respondTrade(incomingTradeOffer.id, false);
            }
            setShowIncomingTrade(false); 
          }}
          fromPlayer={incomingTradeProposer.name}
          offer={incomingTradeOffer.offering ? {
            wood: incomingTradeOffer.offering.wood,
            brick: incomingTradeOffer.offering.brick,
            sheep: incomingTradeOffer.offering.sheep,
            wheat: incomingTradeOffer.offering.wheat,
            ore: incomingTradeOffer.offering.ore
          } : {wood:0,brick:0,sheep:0,wheat:0,ore:0}}
          request={incomingTradeOffer.requesting ? {
            wood: incomingTradeOffer.requesting.wood,
            brick: incomingTradeOffer.requesting.brick,
            sheep: incomingTradeOffer.requesting.sheep,
            wheat: incomingTradeOffer.requesting.wheat,
            ore: incomingTradeOffer.requesting.ore
          } : {wood:0,brick:0,sheep:0,wheat:0,ore:0}}
        />
      )}
      {/* Robber Discard Modal */}
      {!interactionsDisabled && isRobberDiscardRequired && !discardClosed && (
        <DiscardModal
          requiredCount={robberDiscardAmount}
          maxAvailable={robberDiscardMax || { wood: 0, brick: 0, sheep: 0, wheat: 0, ore: 0 }}
          onDiscard={toDiscard => { sendRobberDiscard(toDiscard); setDiscardClosed(true); }}
          onClose={() => setDiscardClosed(true)}
        />
      )}
      {/* Robber Steal Modal */}
      {!interactionsDisabled && isRobberStealRequired && !stealClosed && (
        <StealModal
          candidates={robberStealCandidates}
          onSteal={victimId => { sendRobberSteal(victimId); setStealClosed(true); }}
          onCancel={() => setStealClosed(true)}
        />
      )}
      {/* Dev Card Modals */}
      {!interactionsDisabled && (
        <>
          <YearOfPlentyModal
            open={showYearOfPlenty}
            onClose={() => setShowYearOfPlenty(false)}
            onSubmit={(resources) => {
              playDevCard(DevCardType.YEAR_OF_PLENTY, undefined, resources);
              setShowYearOfPlenty(false);
            }}
          />
          <MonopolyModal
            open={showMonopoly}
            onClose={() => setShowMonopoly(false)}
            onSubmit={(resource) => {
              playDevCard(DevCardType.MONOPOLY, resource);
              setShowMonopoly(false);
            }}
          />
        </>
      )}

      {isGameOver && (
        <GameOver
          gameState={gameState}
          gameOver={gameOver}
          onNewGame={onLeave}
        />
      )}
      <div className="game-header">
        <div className="game-code" data-cy="game-code">
          Game Code: <strong>{gameCode}</strong>
        </div>
        {/* Game Phase Indicator */}
        {gameState?.status && (
          <div className="game-phase" data-cy="game-phase">
            {isStatus(gameState.status, GameStatus.PLAYING, "GAME_STATUS_PLAYING")
              ? "PLAYING"
              : isStatus(gameState.status, GameStatus.SETUP, "GAME_STATUS_SETUP")
                ? "SETUP"
                : isStatus(
                      gameState.status,
                      GameStatus.WAITING,
                      "GAME_STATUS_WAITING"
                    )
                  ? "WAITING"
                  : isStatus(
                        gameState.status,
                        GameStatus.FINISHED,
                        "GAME_STATUS_FINISHED"
                      )
                    ? "FINISHED"
                    : String(gameState.status)}
          </div>
        )}
        <button
          onClick={onLeave}
          className="btn btn-small"
          data-cy="leave-game-btn"
        >
          Leave Game
        </button>
      </div>

      {isWaiting ? (
        <div className="game-waiting" data-cy="game-waiting">
          <h2>Game Lobby</h2>
          <p>Share the game code with your friends to join!</p>
          <div className="players-list" data-cy="players-list">
            <h3>Players ({gameState.players.length}/4)</h3>
            {gameState.players.map((player) => (
              <div
                key={player.id}
                className={`player-item ${player.isReady ? "ready" : ""}`}
                data-cy={`player-item-${player.id}`}
              >
                <span
                  className="player-color"
                  style={{ backgroundColor: PLAYER_COLORS[player.color] }}
                />
                <span className="player-name">
                  {player.name}
                  {player.isHost && (
                    <span className="host-badge" data-cy="host-badge">
                      HOST
                    </span>
                  )}
                </span>
                <span
                  className={`ready-status ${
                    player.isReady ? "ready" : "not-ready"
                  }`}
                  data-cy={`player-ready-status-${player.id}`}
                >
                  {player.isReady ? "✓ Ready" : "Not Ready"}
                </span>
              </div>
            ))}
          </div>

          <div className="lobby-actions">
            {!isReady ? (
              <button
                onClick={() => setReady(true)}
                className="btn btn-ready"
                data-cy="ready-btn"
              >
                I'm Ready
              </button>
            ) : (
              <button
                onClick={() => setReady(false)}
                className="btn btn-secondary"
                data-cy="cancel-ready-btn"
              >
                Cancel Ready
              </button>
            )}

            {isHost && (
              <button
                onClick={startGame}
                className="btn btn-primary"
                disabled={!allPlayersReady}
                title={
                  !allPlayersReady
                    ? "All players must be ready to start"
                    : "Start the game"
                }
                data-cy="start-game-btn"
              >
                Start Game
              </button>
            )}
          </div>

          {gameState.players.length < 2 && (
            <p className="waiting-hint">Waiting for at least 2 players...</p>
          )}
          {gameState.players.length >= 2 && !allPlayersReady && (
            <p className="waiting-hint">
              Waiting for all players to be ready...
            </p>
          )}
        </div>
      ) : (
        <div className="game-board-container" data-cy="game-board-container">
           {(isSetup &&
             isStatus(gameState?.status, GameStatus.SETUP, "GAME_STATUS_SETUP")) && (
            <div className="setup-phase-panel">
              <div
                className="setup-phase-banner"
                data-cy="setup-phase-banner"
              >
                Setup Phase - Round {setupRound}
              </div>
              <div
                className="setup-turn-indicator"
                data-cy="setup-turn-indicator"
              >
                {currentTurnPlayer?.name
                  ? `Current Turn: ${currentTurnPlayer.name}`
                  : "Current Turn: --"}
              </div>
              {setupInstruction && (
                <div
                  className="setup-instruction"
                  data-cy="setup-instruction"
                >
                  {setupInstruction}
                </div>
              )}
            </div>
          )}
           {/* Turn phase switching (TRADE <-> BUILD) */}
           {!interactionsDisabled &&
             isPlaying &&
             isMyTurn &&
             (isTradePhase || isBuildPhase) && (
             <div className="turn-phase-toggle">
               <button
                 className="btn btn-secondary"
                 data-cy="trade-phase-btn"
                 onClick={() => setTurnPhase(TurnPhase.TRADE)}
                 disabled={isTradePhase}
               >
                 Trade
               </button>
               <button
                 className="btn btn-secondary"
                 data-cy="build-phase-btn"
                 onClick={() => setTurnPhase(TurnPhase.BUILD)}
                 disabled={isBuildPhase}
               >
                 Build
               </button>
             </div>
           )}
           {/* Trading UI (only in PLAYING+TRADE phase) */}
           {!interactionsDisabled &&
             isPlaying &&
             isTradePhase &&
             isMyTurn && (
             <div className="trade-phase-control">
                <button
                  className="btn btn-secondary"
                  data-cy="bank-trade-btn"
                  onClick={() => setShowBankTrade(true)}
                >
                  Trade with Bank
                </button>
               <button
                 className="btn btn-secondary"
                 data-cy="propose-trade-btn"
                 onClick={() => setShowProposeTrade(true)}
               >
                 Propose Trade
               </button>
             </div>
           )}
           {resourceGainText &&
             isStatus(gameState?.status, GameStatus.SETUP, "GAME_STATUS_SETUP") && (
            <div
              className="setup-resource-toast"
              data-cy="setup-resource-toast"
            >
              {resourceGainText}
            </div>
          )}
           {!interactionsDisabled &&
             placementModeLabel &&
             ((isSetup &&
               isStatus(gameState?.status, GameStatus.SETUP, "GAME_STATUS_SETUP")) ||
               isStatus(gameState?.status, GameStatus.PLAYING, "GAME_STATUS_PLAYING")) && (
            <div className="placement-mode" data-cy="placement-mode">
              {/* (Placement and trading UI are independent; both can be present) */}
              {placementModeLabel}
            </div>
          )}
           {!interactionsDisabled && isPlaying && isMyTurn && (
            <div className="build-controls" data-cy="build-controls">
              <button
                className={`btn btn-secondary ${buildSelection === "settlement" ? "is-active" : ""}`}
                data-cy="build-settlement-btn"
                onClick={() => handleSelectBuild("settlement")}
              >
                Build Settlement
              </button>
              <button
                className={`btn btn-secondary ${buildSelection === "road" ? "is-active" : ""}`}
                data-cy="build-road-btn"
                onClick={() => handleSelectBuild("road")}
              >
                Build Road
              </button>
              <button
                className={`btn btn-secondary ${buildSelection === "city" ? "is-active" : ""}`}
                data-cy="build-city-btn"
                onClick={() => handleSelectBuild("city")}
              >
                Build City
              </button>
            </div>
          )}
           <div className="game-board-content">
             {/* ---- END TRADING ---- */}
             {gameState?.board && (
               <Suspense
                 fallback={
                   <div className="board-container" data-cy="board-loading">
                     <p>Loading board...</p>
                   </div>
                 }
               >
                 <Board
                   board={gameState.board}
                   players={gameState.players}
                   validVertexIds={effectiveValidVertexIds}
                   validEdgeIds={effectiveValidEdgeIds}
                   onBuildSettlement={interactionsDisabled ? undefined : handleBuildVertex}
                    onBuildRoad={interactionsDisabled ? undefined : (edgeId) => build(StructureType.ROAD, edgeId)}
                    isRobberMoveMode={isRobberMoveRequired}
                    onSelectRobberHex={interactionsDisabled ? undefined : isRobberMoveRequired ? (hex) => {
                      if (hex.coord) {
                        sendRobberMove({ q: hex.coord.q, r: hex.coord.r, s: -(hex.coord.q + hex.coord.r) });
                      }
                    } : undefined}
                 />
               </Suspense>
             )}
             {gameState && (
              <PlayerPanel
                players={gameState.players}
                board={gameState.board}
                  currentTurn={gameState.currentTurn}
                  turnPhase={gameState.turnPhase}
                  dice={gameState.dice}
                  gameStatus={gameState.status as unknown as string}
                  isGameOver={isGameOver}
                  longestRoadPlayerId={gameState.longestRoadPlayerId ?? null}
                  largestArmyPlayerId={gameState.largestArmyPlayerId ?? null}
              />
             )}
             {!interactionsDisabled &&
               isStatus(gameState?.status, GameStatus.PLAYING, "GAME_STATUS_PLAYING") && (
               <DevelopmentCardsPanel
                 currentPlayer={currentPlayer ?? null}
                 canBuy={canBuyDevCard}
                 canPlay={canPlayDevCard}
                 onBuyCard={buyDevCard}
                 onPlayCard={handlePlayDevCard}
               />
             )}
           </div>
        </div>
      )}
    </div>
  );
}
