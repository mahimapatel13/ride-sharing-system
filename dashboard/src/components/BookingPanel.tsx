import { useState } from "react";
import { API_BASE } from "@/config/constants";

interface BookingPanelProps {
  pickup: { lat: number; lng: number } | null;
  onBooked: () => void;
  disabled?: boolean;
}

const BookingPanel = ({ pickup, onBooked, disabled }: BookingPanelProps) => {
  const [riderId, setRiderId] = useState("1");
  const [luggage, setLuggage] = useState("1");
  const [tolerance, setTolerance] = useState("5");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

const handleBook = () => {
  if (!pickup) return;

  setLoading(true);
  setError(null);

  try {
    const ws = new WebSocket(`${API_BASE.replace("http", "ws")}/ride/request`);

    ws.onopen = () => {
      console.log("WebSocket connected");

      // Send booking request over WS
      ws.send(
        JSON.stringify({
          RiderID: parseInt(riderId),
          Luggage: parseInt(luggage),
          Lat: pickup.lat,
          Lng: pickup.lng,
          Tolerance: parseInt(tolerance),
        })
      );
    };

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      console.log("WS message:", msg);

      if (msg.status === "MATCHED") {
        onBooked();
        ws.close();
      } else if (msg.status === "CANCELLED") {
        setError("Ride cancelled");
        ws.close();
      } else if (msg.type === "error") {
        setError(msg.message || "Booking failed");
        ws.close();
      }
    };

    ws.onerror = (err) => {
      console.error("WS error", err);
      setError("WebSocket error");
      setLoading(false);
    };

    ws.onclose = () => {
      console.log("WebSocket closed");
      setLoading(false);
    };
  } catch (err) {
    setError(err instanceof Error ? err.message : "Booking failed");
    setLoading(false);
  }
};


  return (
    <div className="brutal-card">
      <div className="bg-foreground px-4 py-2">
        <span className="airport-label text-primary-foreground">
          🚖 BOOK CAB
        </span>
      </div>

      <div className="p-4 space-y-3 min-h-[200px] ">
        <div className="grid grid-cols-3 gap-3">
          <div>
            <label className="airport-label text-muted-foreground block mb-1">
              Rider ID
            </label>
            <input
              type="number"
              value={riderId}
              onChange={(e) => setRiderId(e.target.value)}
              className="brutal-input w-full px-3 py-2 text-sm font-mono"
              disabled={disabled}
            />
          </div>
          <div>
            <label className="airport-label text-muted-foreground block mb-1">
              Luggage
            </label>
            <input
              type="number"
              value={luggage}
              onChange={(e) => setLuggage(e.target.value)}
              className="brutal-input w-full px-3 py-2 text-sm font-mono"
              disabled={disabled}
            />
          </div>
          <div>
            <label className="airport-label text-muted-foreground block mb-1">
              Tolerance
            </label>
            <input
              type="number"
              value={tolerance}
              onChange={(e) => setTolerance(e.target.value)}
              className="brutal-input w-full px-3 py-2 text-sm font-mono"
              disabled={disabled}
            />
          </div>
        </div>

        {!pickup && (
          <p className="font-mono text-xs text-muted-foreground">
            SELECT PICKUP LOCATION ON MAP FIRST
          </p>
        )}

        <button
          onClick={handleBook}
          disabled={!pickup || loading || disabled}
          className={`brutal-btn brutal-btn-primary w-full px-6 py-3 text-sm ${
            !pickup || loading || disabled ? "brutal-btn-disabled" : ""
          }`}
        >
          {loading ? "REQUESTING..." : "BOOK CAB"}
        </button>

        {error && (
          <div className="brutal-border bg-destructive/10 p-3">
            <span className="font-mono text-xs text-destructive font-bold">
              ERR: {error}
            </span>
          </div>
        )}
      </div>
    </div>
  );
};

export default BookingPanel;
