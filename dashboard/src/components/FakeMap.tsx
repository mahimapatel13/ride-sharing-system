import { useState } from "react";
import { AIRPORT } from "@/config/constants";

interface FakeMapProps {
  onSelectPickup: (lat: number, lng: number) => void;
  pickup: { lat: number; lng: number } | null;
  disabled?: boolean;
}

const FakeMap = ({ onSelectPickup, pickup, disabled }: FakeMapProps) => {
  const [hoverPos, setHoverPos] = useState<{ x: number; y: number } | null>(null);

  const handleClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (disabled) return;
    const rect = e.currentTarget.getBoundingClientRect();
    const x = (e.clientX - rect.left) / rect.width;
    const y = (e.clientY - rect.top) / rect.height;

    // Map click position to coordinates near airport (±0.05 degrees)
    const lat = AIRPORT.lat + (0.5 - y) * 0.1;
    const lng = AIRPORT.lng + (x - 0.5) * 0.1;
    onSelectPickup(parseFloat(lat.toFixed(6)), parseFloat(lng.toFixed(6)));
  };

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    if (disabled) return;
    const rect = e.currentTarget.getBoundingClientRect();
    setHoverPos({ x: e.clientX - rect.left, y: e.clientY - rect.top });
  };

  // Convert pickup back to pixel position
  const pickupPos = pickup
    ? {
        x: ((pickup.lng - AIRPORT.lng) / 0.1 + 0.5),
        y: (0.5 - (pickup.lat - AIRPORT.lat) / 0.1),
      }
    : null;

  return (
    <div className="brutal-card p-0 overflow-hidden">
      <div className="bg-foreground px-4 py-2 flex items-center justify-between">
        <span className="airport-label text-primary-foreground">
          ✈ SELECT PICKUP
        </span>
        <span className="airport-badge bg-accent text-accent-foreground">
          {AIRPORT.code}
        </span>
      </div>

      <div
        className={`relative h-[300px] bg-taxiway ${disabled ? "opacity-50 cursor-not-allowed" : "cursor-crosshair"}`}
        onClick={handleClick}
        onMouseMove={handleMouseMove}
        onMouseLeave={() => setHoverPos(null)}
      >
        {/* Grid lines */}
        {Array.from({ length: 8 }).map((_, i) => (
          <div
            key={`h-${i}`}
            className="absolute left-0 right-0 border-t border-runway/40"
            style={{ top: `${(i + 1) * 11.11}%` }}
          />
        ))}
        {Array.from({ length: 8 }).map((_, i) => (
          <div
            key={`v-${i}`}
            className="absolute top-0 bottom-0 border-l border-runway/40"
            style={{ left: `${(i + 1) * 11.11}%` }}
          />
        ))}

        {/* Runway line */}
        <div className="absolute top-1/2 left-[20%] right-[20%] h-[6px] bg-runway -translate-y-1/2">
          <div className="absolute inset-0 flex items-center justify-center gap-3">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="w-6 h-[2px] bg-card" />
            ))}
          </div>
        </div>

        {/* Airport marker */}
        <div
          className="absolute w-4 h-4 bg-foreground border-2 border-card"
          style={{ left: "50%", top: "50%", transform: "translate(-50%, -50%)" }}
        />
        <span
          className="absolute font-mono text-[10px] font-bold text-foreground"
          style={{ left: "50%", top: "50%", transform: "translate(12px, -50%)" }}
        >
          {AIRPORT.code}
        </span>

        {/* Pickup marker */}
        {pickupPos && (
          <>
            <div
              className="absolute w-5 h-5 bg-accent border-[3px] border-foreground"
              style={{
                left: `${pickupPos.x * 100}%`,
                top: `${pickupPos.y * 100}%`,
                transform: "translate(-50%, -50%)",
              }}
            />
            <div
              className="absolute font-mono text-[10px] font-bold text-foreground bg-accent px-1"
              style={{
                left: `${pickupPos.x * 100}%`,
                top: `${pickupPos.y * 100}%`,
                transform: "translate(12px, -50%)",
              }}
            >
              PICKUP
            </div>
          </>
        )}

        {/* Crosshair cursor indicator */}
        {hoverPos && !disabled && (
          <>
            <div
              className="absolute w-px h-full bg-foreground/20 top-0 pointer-events-none"
              style={{ left: hoverPos.x }}
            />
            <div
              className="absolute h-px w-full bg-foreground/20 left-0 pointer-events-none"
              style={{ top: hoverPos.y }}
            />
          </>
        )}

        {!pickup && !disabled && (
          <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
            <span className="airport-label text-muted-foreground bg-card px-3 py-1 brutal-border">
              CLICK TO SET PICKUP
            </span>
          </div>
        )}
      </div>

      {pickup && (
        <div className="px-4 py-2 border-t-[3px] border-foreground bg-card font-mono text-xs flex gap-6">
          <span>
            <span className="text-muted-foreground">LAT </span>
            <span className="font-bold">{pickup.lat}</span>
          </span>
          <span>
            <span className="text-muted-foreground">LNG </span>
            <span className="font-bold">{pickup.lng}</span>
          </span>
        </div>
      )}
    </div>
  );
};

export default FakeMap;
