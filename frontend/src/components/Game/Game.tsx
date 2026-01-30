import React from "react";
import { useGame } from "@/context/GameContext";
import Board from "@/components/Board/Board";
import { SetupPhasePanel } from "./SetupPhasePanel";
import { DevelopmentCardsPanel } from "./DevelopmentCardsPanel";
import { GameStatus } from "@/gen/proto/catan/v1/types";

interface GameProps {
  gameCode: string;
  onLeave: () => void;
}

export function Game({ gameCode: _gameCode, onLeave: _onLeave }: GameProps) {
   // Silence unused prop lint:
   void _gameCode;
   void _onLeave;
   const { gameState } = useGame();

   // DEBUG: Log on every render
   // Props supplied from App, not used yet: gameCode =
   console.log("[Game render] gameState:", gameState && {status: gameState.status, players: gameState.players, turn: gameState.currentTurn, setupPhase: gameState.setupPhase});

   // Show E2E-visible waiting room if status is 'waiting'
   if (gameState?.status === "waiting") {
     return (
       <div className="game-waiting-banner" data-cy="game-waiting" style={{ textAlign: 'center', marginTop: '4rem', fontSize: '1.5rem', color: '#888' }}>
         Waiting for players to join...
       </div>
     );
   }

   if (gameState?.status === GameStatus.SETUP) {
    // Render setup phase: game board container + setup panel, with actual Board
    const { placementState, placementMode, build } = useGame();
    // Set handlers depending on mode
    const onBuildSettlement = placementMode === "settlement" ? (vertexId: string) => build("settlement", vertexId) : undefined;
    const onBuildRoad = placementMode === "road" ? (edgeId: string) => build("road", edgeId) : undefined;
    return (
      <div className="game-board-container">
        <div data-cy="game-phase">SETUP</div>
        <SetupPhasePanel />
        <Board
          board={gameState.board}
          players={gameState.players}
          validVertexIds={placementState?.validVertexIds}
          validEdgeIds={placementState?.validEdgeIds}
          onBuildSettlement={onBuildSettlement}
          onBuildRoad={onBuildRoad}
        />
      </div>
    );
  }

  // Main game play/rendering logic
  if (gameState?.status === GameStatus.PLAYING && gameState.players?.length && typeof gameState.currentTurn === 'number') {
    const currentPlayer = gameState.players[gameState.currentTurn];
    // Replace with the real canBuy/canPlay logic if needed
    const canBuy = true; // Placeholder: should check resource/phase logic
    const canPlay = true; // Placeholder: should check phase/other rules
    // Add real handlers or pull from context as needed
    const onBuyCard = () => {};
    const onPlayCard = (cardType: import("@/types").DevCardType) => { void cardType; /* no-op */ };
    return (
      <div className="game-board-container">
        <Board
          board={gameState.board}
          players={gameState.players}
          validVertexIds={placementState?.validVertexIds}
          validEdgeIds={placementState?.validEdgeIds}
          onBuildSettlement={placementMode === "build" ? (vertexId: string) => build("settlement", vertexId) : undefined}
          onBuildRoad={placementMode === "build" ? (edgeId: string) => build("road", edgeId) : undefined}
        />
        <div className="main-game-ui">
          <div data-cy="game-phase">PLAYING</div>
          <DevelopmentCardsPanel
            currentPlayer={currentPlayer}
            canBuy={canBuy}
            canPlay={canPlay}
            onBuyCard={onBuyCard}
            onPlayCard={onPlayCard}
            turnCounter={gameState.turnCounter}
          />
        </div>
      </div>
    );
  }

  // Fallback
   const sessionToken = sessionStorage.getItem('sessionToken');
   return (
     <div className="lobby-actions">
       <div data-cy="main-game-placeholder">
         {sessionToken == null ?
           'ERROR: No session token found in sessionStorage' :
           `Game in progress or not started. Status: ${typeof gameState?.status !== 'undefined' ? String(gameState.status) : 'undefined'}`}
       </div>
     </div>
   );
}
