import { useEffect, useMemo, useState } from "react";
import { DiscardModal } from "./DiscardModal";
import { StealModal } from "./StealModal";
import { BankTradeModal } from "./BankTradeModal";
import { ProposeTradeModal } from "./ProposeTradeModal";
import { IncomingTradeModal } from "./IncomingTradeModal";
import { useGame } from "@/context";
import { Board } from "@/components/Board";
import { PlayerPanel } from "@/components/PlayerPanel";
import { GameStatus, PlayerColor, StructureType } from "@/types";
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

interface GameProps {
  gameCode: string;
  onLeave: () => void;
}

export function Game({ gameCode, onLeave }: GameProps) {
  // Trading modals
  const [showBankTrade, setShowBankTrade] = useState(false);
  const [showProposeTrade, setShowProposeTrade] = useState(false);
  const [showIncomingTrade, setShowIncomingTrade] = useState(false);
  // Simulated offer for stub incoming modal demo:
  const [incomingTrade, setIncomingTrade] = useState<{from: string; offer: Record<string,number>; request: Record<string,number>}|null>(null);
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
    if (isRobberStealRequired) {
      setStealClosed(false);
    }
  }, [isRobberStealRequired]);


  useEffect(() => {
    connect();
    return () => disconnect();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Only run once on mount/unmount

  const players = gameState?.players ?? [];

  // Find current player's state
  const currentPlayer = useMemo(
    () => players.find((p) => p.id === currentPlayerId),
    [players, currentPlayerId]
  );

  // Check if all players are ready
  const allPlayersReady = useMemo(
    () => players.length >= 2 && players.every((p) => p.isReady),
    [players]
  );

  // Check if current player is host/ready
  const isHost = currentPlayer?.isHost ?? false;
  const isReady = currentPlayer?.isReady ?? false;
  // Handle both enum value and JSON string representation
  const isWaiting =
    gameState?.status === GameStatus.WAITING ||
    gameState?.status === ("GAME_STATUS_WAITING" as unknown as GameStatus);
  const isSetup =
    gameState?.status === GameStatus.SETUP ||
    gameState?.status === ("GAME_STATUS_SETUP" as unknown as GameStatus);

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

  // Game Over overlay
  const isGameOver = !!gameOver || gameState?.status === GameStatus.FINISHED || (gameState?.status as unknown as string) === 'GAME_STATUS_FINISHED';
  const interactionsDisabled = isGameOver;

  return (
    <div className="game" data-cy="game">
      {/* BANK TRADE MODAL */}
      {!interactionsDisabled && (
        <BankTradeModal
          open={showBankTrade}
          onClose={() => setShowBankTrade(false)}
          onSubmit={() => {
            // TODO: Hook up to trade send
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
          onSubmit={(offer, request) => {
            // TODO: Hook up to trade send
            setShowProposeTrade(false);
            // For demo: simulate incoming trade right away
            setIncomingTrade({from:currentPlayer?.name||'Player',offer,request});
            setShowIncomingTrade(true);
          }}
          players={players}
          myResources={(currentPlayer?.resources ?? {wood:0,brick:0,sheep:0,wheat:0,ore:0}) as unknown as Record<string, number>}
        />
      )}
      {/* INCOMING TRADE MODAL (stub/demo, always shows same test offer) */}
      {!interactionsDisabled && (
        <IncomingTradeModal
          open={showIncomingTrade}
          onAccept={() => { setShowIncomingTrade(false); setIncomingTrade(null); }}
          onDecline={() => { setShowIncomingTrade(false); setIncomingTrade(null); }}
          fromPlayer={incomingTrade?.from || 'Other'}
          offer={incomingTrade?.offer || {wood:1}}
          request={incomingTrade?.request || {brick:1}}
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
           {(isSetup && gameState?.status === GameStatus.SETUP) && (
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
           {/* Trading/Build Toggle and Trade UI (only in PLAYING+TRADE phase) */}
           {!interactionsDisabled && gameState?.status === GameStatus.PLAYING && gameState?.turnPhase === ("TURN_PHASE_TRADE" as unknown as typeof gameState.turnPhase) && currentPlayer?.id === currentPlayerId && (
             <div className="trade-phase-control">
               <button
                 className="btn btn-secondary"
                 data-cy="trade-phase-btn"
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
           {resourceGainText && gameState?.status === GameStatus.SETUP && (
            <div
              className="setup-resource-toast"
              data-cy="setup-resource-toast"
            >
              {resourceGainText}
            </div>
          )}
           {!interactionsDisabled && placementModeLabel && ((isSetup && gameState?.status === GameStatus.SETUP) || (gameState?.status === GameStatus.PLAYING)) && (
            <div className="placement-mode" data-cy="placement-mode">
              {/* (Placement and trading UI are independent; both can be present) */}
              {placementModeLabel}
            </div>
          )}
           <div className="game-board-content">
             {/* ---- END TRADING ---- */}
             {gameState?.board && (
               <Board
                 board={gameState.board}
                 players={gameState.players}
                 validVertexIds={placementState.validVertexIds}
                 validEdgeIds={placementState.validEdgeIds}
                 onBuildSettlement={interactionsDisabled ? undefined : (vertexId) => build(StructureType.SETTLEMENT, vertexId)}
                  onBuildRoad={interactionsDisabled ? undefined : (edgeId) => build(StructureType.ROAD, edgeId)}
                  isRobberMoveMode={isRobberMoveRequired}
                  onSelectRobberHex={interactionsDisabled ? undefined : isRobberMoveRequired ? (hex) => {
                    if (hex.coord) {
                      sendRobberMove({ q: hex.coord.q, r: hex.coord.r, s: -(hex.coord.q + hex.coord.r) });
                    }
                  } : undefined}
               />
             )}
            {gameState && (
              <PlayerPanel
                players={gameState.players}
                 currentTurn={gameState.currentTurn}
                 turnPhase={gameState.turnPhase}
                 dice={gameState.dice}
                 gameStatus={gameState.status as unknown as string}
                 isGameOver={isGameOver}
              />
            )}
          </div>
        </div>
      )}
    </div>
  );
}
