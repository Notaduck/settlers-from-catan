import React, { useEffect, useState } from "react";
import { useGame } from "@/context/GameContext";


function pluralize(word: string, count: number) {
  return count === 1 ? word : `${word}s`;
}

function resourceList(resources: Record<string, number>) {
  return Object.entries(resources)
    .filter(([, v]) => v > 0)
    .map(([k, v]) => `${v} ${pluralize(k.charAt(0).toUpperCase() + k.slice(1), v)}`)
    .join(", ");
}

// ResourceCount field extractor
import type { ResourceCount } from "@/gen/proto/catan/v1/types";
function resourceCountToRecord(rc: ResourceCount): Record<string, number> {
  return rc ? {
    wood: rc.wood ?? 0,
    wheat: rc.wheat ?? 0,
    ore: rc.ore ?? 0,
    sheep: rc.sheep ?? 0,
    brick: rc.brick ?? 0
  } : {};
}

export function SetupPhasePanel() {
  const {
     gameState,
     currentPlayerId,
     resourceGain,
     clearResourceGain,
  } = useGame();

  // DEBUG: Log phase status
  useEffect(() => {
     console.log("[SetupPhasePanel] mount; gameState.status:", gameState && gameState.status);
  }, [gameState]);
  const [showToast, setShowToast] = useState(false);

   useEffect(() => {
     if (resourceGain) {
       // Schedule on next tick to avoid cascading render warning
       setTimeout(() => setShowToast(true), 0);
     }
   }, [resourceGain]);

  if (!gameState || !gameState.setupPhase) {
    // Instead of null, show a loading placeholder so E2E test always has selector.
    return (
      <div className="setup-phase-panel" data-cy="setup-phase-banner">
        <div>Waiting for setup phase&hellip;</div>
      </div>
    );
  }
  const { setupPhase } = gameState;

  // Get round and player turn
   const round = setupPhase.round;
   const placementsInTurn = setupPhase.placementsInTurn || 0;
   // We no longer have currentPlayerIndex on setupPhase; fallback to gameState.currentTurn
   const turnPlayerIndex = gameState.currentTurn ?? 0;
   const turnPlayer = gameState.players[turnPlayerIndex];
   
  // Determine if it's current user's turn for highlighting
  const isMyTurn = turnPlayer && currentPlayerId === turnPlayer.id;

  // Banner text
  const bannerText = `Setup Phase - Round ${round}`;

  // Placement Instruction
  let instruction = "";
  let stepNum = 1;
  if (setupPhase.round === 1) {
    stepNum = 1;
  } else if (setupPhase.round === 2) {
    stepNum = 2;
  }
  if (placementsInTurn === 0) {
    instruction = `Place Settlement (${stepNum}/2)`;
  } else {
    instruction = `Place Road (${stepNum}/2)`;
  }

  return (
    <div className="setup-phase-panel">
      <div
        className="setup-phase-banner"
        data-cy="setup-phase-banner"
      >
        {bannerText}
      </div>
      <div
        className="setup-turn-indicator"
        data-cy="setup-turn-indicator"
      >
        Turn: <b>{turnPlayer?.name || "Player"}</b>
        {isMyTurn && <span style={{ marginLeft: 8, color: "#2ecc71" }}>(Your Turn)</span>}
      </div>
      <div
        className="setup-instruction"
        data-cy="setup-instruction"
      >
        {instruction}
      </div>
      <div className="placement-mode" data-cy="placement-mode">
        {instruction}
      </div>
      {showToast && resourceGain && (
         <div className="setup-resource-toast" data-cy="setup-resource-toast">
           You received: {resourceList(resourceCountToRecord(resourceGain.resources))}
          <button
            className="btn btn-small"
            style={{ marginLeft: 12 }}
            onClick={() => {
              setShowToast(false);
              clearResourceGain();
            }}
          >
            ×
          </button>
        </div>
      )}
    </div>
  );
}
