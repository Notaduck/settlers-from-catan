import { useState } from "react";
import { GameProvider } from "@/context";
import { Lobby } from "@/components/Lobby";
import { Game } from "@/components/Game";
import "./App.css";

function App() {
  const [gameCode, setGameCode] = useState<string | null>(() =>
    localStorage.getItem("gameCode")
  );
  const [sessionToken, setSessionToken] = useState<string | null>(() =>
    localStorage.getItem("sessionToken")
  );
  const [playerId, setPlayerId] = useState<string | null>(() =>
    localStorage.getItem("playerId")
  );

  const handleGameJoined = (code: string, token: string, id: string) => {
    localStorage.setItem("sessionToken", token);
    localStorage.setItem("playerId", id);
    localStorage.setItem("gameCode", code);
    setSessionToken(token);
    setPlayerId(id);
    setGameCode(code);
  };

  const handleLeaveGame = () => {
    localStorage.removeItem("sessionToken");
    localStorage.removeItem("playerId");
    localStorage.removeItem("gameCode");
    setSessionToken(null);
    setPlayerId(null);
    setGameCode(null);
  };

  return (
    <GameProvider playerId={playerId}>
      <div className="app">
        <header className="app-header">
          <h1>Settlers from Catan</h1>
        </header>
        <main className="app-main">
          {sessionToken && gameCode ? (
            <Game gameCode={gameCode} onLeave={handleLeaveGame} />
          ) : (
            <Lobby onGameJoined={handleGameJoined} />
          )}
        </main>
      </div>
    </GameProvider>
  );
}

export default App;
