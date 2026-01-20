import { useEffect, useState } from "react";
import { GameProvider } from "@/context";
import { Lobby } from "@/components/Lobby";
import { Game } from "@/components/Game";
import "./App.css";

const STORAGE_KEYS = ["gameCode", "sessionToken", "playerId"] as const;

function readSessionValue(key: string) {
  const sessionValue = sessionStorage.getItem(key);
  if (sessionValue) {
    return sessionValue;
  }
  const localValue = localStorage.getItem(key);
  if (localValue) {
    sessionStorage.setItem(key, localValue);
    localStorage.removeItem(key);
  }
  return localValue;
}

function App() {
  const [gameCode, setGameCode] = useState<string | null>(() =>
    readSessionValue("gameCode")
  );
  const [sessionToken, setSessionToken] = useState<string | null>(() =>
    readSessionValue("sessionToken")
  );
  const [playerId, setPlayerId] = useState<string | null>(() =>
    readSessionValue("playerId")
  );

  const handleGameJoined = (code: string, token: string, id: string) => {
    sessionStorage.setItem("sessionToken", token);
    sessionStorage.setItem("playerId", id);
    sessionStorage.setItem("gameCode", code);
    STORAGE_KEYS.forEach((key) => localStorage.removeItem(key));
    setSessionToken(token);
    setPlayerId(id);
    setGameCode(code);
  };

  const handleLeaveGame = () => {
    STORAGE_KEYS.forEach((key) => {
      sessionStorage.removeItem(key);
      localStorage.removeItem(key);
    });
    setSessionToken(null);
    setPlayerId(null);
    setGameCode(null);
  };

  useEffect(() => {
    STORAGE_KEYS.forEach((key) => {
      const value = readSessionValue(key);
      if (!value) {
        sessionStorage.removeItem(key);
      }
    });
  }, []);

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
