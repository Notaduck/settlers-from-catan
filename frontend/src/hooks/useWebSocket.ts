import { useCallback, useEffect, useRef, useState } from "react";
import type { ClientMessage, ServerMessage } from "@/types";

const WS_PATH = "/ws";

interface UseWebSocketOptions {
  onMessage?: (message: ServerMessage) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const pendingMessagesRef = useRef<string[]>([]);
  const optionsRef = useRef(options);

  // Keep options ref up to date without causing re-renders
  useEffect(() => {
    optionsRef.current = options;
  }, [options]);

  const getSessionToken = useCallback(() => {
    const sessionToken = sessionStorage.getItem("sessionToken");
    if (sessionToken) {
      return sessionToken;
    }
    const localToken = localStorage.getItem("sessionToken");
    if (localToken) {
      sessionStorage.setItem("sessionToken", localToken);
      localStorage.removeItem("sessionToken");
    }
    return localToken;
  }, []);

  const connect = useCallback(function connectInternal() {
    // Don't connect if already connected or connecting
    if (
      wsRef.current?.readyState === WebSocket.OPEN ||
      wsRef.current?.readyState === WebSocket.CONNECTING
    ) {
      return;
    }

    const token = getSessionToken();
    if (!token) {
      setError("No session token found");
      return;
    }

    try {
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const isLocalhost =
        window.location.hostname === "localhost" ||
        window.location.hostname === "127.0.0.1";
      const host = isLocalhost && window.location.port === "3000"
        ? `${window.location.hostname}:8080`
        : window.location.host;
      const wsUrl = `${protocol}//${host}${WS_PATH}?token=${token}`;

      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        setError(null);
        if (pendingMessagesRef.current.length > 0) {
          for (const payload of pendingMessagesRef.current) {
            ws.send(payload);
          }
          pendingMessagesRef.current = [];
        }
        optionsRef.current.onConnect?.();
      };

      ws.onclose = () => {
        setIsConnected(false);
        wsRef.current = null;
        optionsRef.current.onDisconnect?.();

        // Attempt reconnection after a short delay
        reconnectTimeoutRef.current = window.setTimeout(() => {
          if (getSessionToken()) {
            connectInternal();
          }
        }, 1000);
      };

      ws.onerror = () => {
        setError("WebSocket connection error");
      };

      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as ServerMessage;
          optionsRef.current.onMessage?.(message);
        } catch {
          console.error("Failed to parse WebSocket message:", event.data);
        }
      };
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to connect");
    }
  }, [getSessionToken]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    wsRef.current?.close();
    wsRef.current = null;
  }, []);

  const sendMessage = useCallback((message: ClientMessage) => {
    const payload = JSON.stringify(message);
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(payload);
      return;
    }
    pendingMessagesRef.current.push(payload);
    connect();
  }, [connect]);

  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  return {
    isConnected,
    error,
    connect,
    disconnect,
    sendMessage,
  };
}
