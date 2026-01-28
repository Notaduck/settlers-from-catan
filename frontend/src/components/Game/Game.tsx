import React from "react";
import { useGame } from "@/context/GameContext";
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

  if (gameState?.status === GameStatus.SETUP) {
    // Render setup phase: game board container + setup panel
    return (
       <div data-cy="game-board-container" className="game-board-container">
         <div data-cy="game-phase">SETUP</div>
         <SetupPhasePanel />
         {/* Place board rendering here if exists, or placeholder for now */}
         <div data-cy="board-placeholder">[Board will be rendered here during setup]</div>
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
      <div data-cy="game-board-container" className="game-board-container">
        {/* Add board/game UI as needed here */}
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
