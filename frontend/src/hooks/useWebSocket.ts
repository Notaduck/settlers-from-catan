import { useCallback, useEffect, useRef, useState } from "react";
import type { ClientMessage, ServerMessage } from "@/types";

const WS_URL = "/ws";

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
  const optionsRef = useRef(options);

  // Keep options ref up to date without causing re-renders
  useEffect(() => {
    optionsRef.current = options;
  }, [options]);

  const getSessionToken = useCallback(() => {
    return localStorage.getItem("sessionToken");
  }, []);

  const connect = useCallback(() => {
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
      const wsUrl = `${protocol}//${window.location.host}${WS_URL}?token=${token}`;

      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        setError(null);
        optionsRef.current.onConnect?.();
      };

      ws.onclose = () => {
        setIsConnected(false);
        wsRef.current = null;
        optionsRef.current.onDisconnect?.();

        // Attempt reconnection after 3 seconds
        reconnectTimeoutRef.current = window.setTimeout(() => {
          if (getSessionToken()) {
            connect();
          }
        }, 3000);
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
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    } else {
      console.error("WebSocket is not connected");
    }
  }, []);

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
