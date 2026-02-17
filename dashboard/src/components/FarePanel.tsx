import { useState } from "react";
import { API_BASE } from "@/config/constants";

interface FarePanelProps {
  pickup: { lat: number; lng: number } | null;
  onFareCalculated: (fareId: number, fare: number) => void;
  disabled?: boolean;
}

const FarePanel = ({ pickup, onFareCalculated, disabled }: FarePanelProps) => {
  const [loading, setLoading] = useState(false);
  const [fare, setFare] = useState<{ fare_id: number; fare: number } | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleCheckFare = async () => {
    if (!pickup) return;
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/ride/fare`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({latitude: pickup.lat, langitude: pickup.lng }),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setFare(data);
      onFareCalculated(data.fare_id, data.fare);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to calculate fare");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="brutal-card">
      <div className="bg-foreground px-4 py-2">
        <span className="airport-label text-primary-foreground">
          ₹ CHECK FARE
        </span>
      </div>

      <div className="p-4 space-y-4 min-h-[140px]">
        {!pickup && (
          <p className="font-mono text-xs text-muted-foreground">
            SELECT PICKUP LOCATION ON MAP FIRST
          </p>
        )}

        <button
          onClick={handleCheckFare}
          disabled={!pickup || loading || disabled}
          className={`brutal-btn brutal-btn-accent w-full px-6 py-3 text-sm ${
            !pickup || loading || disabled ? "brutal-btn-disabled" : ""
          }`}
        >
          {loading ? "CALCULATING..." : "CHECK FARE"}
        </button>

        {error && (
          <div className="brutal-border bg-destructive/10 p-3">
            <span className="font-mono text-xs text-destructive font-bold">
              ERR: {error}
            </span>
          </div>
        )}

        {fare && (
          <div className="border-[3px] border-foreground bg-accent p-4 space-y-2">
            <div className="flex justify-between items-baseline">
              <span className="airport-label text-accent-foreground">ESTIMATED FARE</span>
              <span className="font-mono text-3xl font-bold text-accent-foreground">
                ₹{fare.fare}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="airport-label text-accent-foreground/70">FARE ID</span>
              <span className="font-mono text-sm font-bold text-accent-foreground">
                #{fare.fare_id}
              </span>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default FarePanel;
