import { useEffect, useRef, useState } from "react";
import { WS_BASE } from "@/config/constants";

type RideStatus = "IDLE" | "SEARCHING" | "MATCHED" | "CANCELLED";

interface MatchData {
  cab_id: number;
  msg: string;
}

interface StatusDisplayProps {
  active: boolean;
  onCancel: () => void;
  onReset: () => void;
}

const StatusDisplay = ({ active, onCancel, onReset }: StatusDisplayProps) => {
  const [status, setStatus] = useState<RideStatus>("IDLE");
  const [matchData, setMatchData] = useState<MatchData | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!active) {
      setStatus("IDLE");
      setMatchData(null);
      return;
    }

    setStatus("SEARCHING");

    // Open WebSocket connection
    const ws = new WebSocket(`${WS_BASE}/api/v1/ride/ws`);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === "status" && data.status === "MATCHED") {
          setStatus("MATCHED");
          setMatchData({ cab_id: data.cab_id, msg: data.msg });
        }
      } catch {
        // Ignore malformed messages
      }
    };

    ws.onerror = () => {
      // WebSocket errors are silent in this minimal UI
    };

    ws.onclose = () => {
      wsRef.current = null;
    };

    return () => {
      if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
        ws.close();
      }
      wsRef.current = null;
    };
  }, [active]);

  const handleCancel = () => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: "CANCEL_RIDE" }));
      wsRef.current.close();
    }
    wsRef.current = null;
    setStatus("CANCELLED");
    onCancel();
  };

  if (status === "IDLE") return null;

  return (
    <div className="space-y-4">
      {status === "SEARCHING" && (
        <div className="brutal-card">
          <div className="bg-accent px-4 py-2">
            <span className="airport-label text-accent-foreground">
              ◎ RIDE STATUS
            </span>
          </div>
          <div className="p-6 text-center space-y-4">
            <div className="flex items-center justify-center gap-2">
              <span className="w-2 h-2 bg-foreground animate-pulse-dot" />
              <span className="w-2 h-2 bg-foreground animate-pulse-dot [animation-delay:0.3s]" />
              <span className="w-2 h-2 bg-foreground animate-pulse-dot [animation-delay:0.6s]" />
            </div>
            <p className="font-mono text-sm font-bold">SEARCHING FOR DRIVER...</p>
            <p className="font-mono text-xs text-muted-foreground">
              WEBSOCKET CONNECTION ACTIVE
            </p>
            <button
              onClick={handleCancel}
              className="brutal-btn brutal-btn-danger px-6 py-2 text-sm"
            >
              ✕ CANCEL RIDE
            </button>
          </div>
        </div>
      )}

      {status === "MATCHED" && matchData && (
        <div className="brutal-card">
          <div className="bg-accent px-4 py-2">
            <span className="airport-label text-accent-foreground">
              ✓ DRIVER MATCHED
            </span>
          </div>
          <div className="p-6 space-y-4">
            <div className="border-[3px] border-foreground bg-accent/20 p-4 text-center">
              <p className="airport-label text-muted-foreground mb-1">CAB ID</p>
              <p className="font-mono text-4xl font-bold">{matchData.cab_id}</p>
            </div>
            <div className="bg-foreground text-primary-foreground p-3 text-center">
              <p className="font-mono text-sm font-bold">{matchData.msg}</p>
            </div>
            <button
              onClick={onReset}
              className="brutal-btn brutal-btn-primary w-full px-6 py-3 text-sm"
            >
              NEW BOOKING
            </button>
          </div>
        </div>
      )}

      {status === "CANCELLED" && (
        <div className="brutal-card">
          <div className="bg-secondary px-4 py-2">
            <span className="airport-label text-secondary-foreground">
              ✕ RIDE CANCELLED
            </span>
          </div>
          <div className="p-4 text-center space-y-3">
            <p className="font-mono text-sm text-muted-foreground">
              RIDE REQUEST CANCELLED
            </p>
            <button
              onClick={onReset}
              className="brutal-btn brutal-btn-primary px-6 py-2 text-sm"
            >
              START OVER
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default StatusDisplay;
