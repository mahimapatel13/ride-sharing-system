import { useState, useCallback } from "react";
import { AIRPORT } from "@/config/constants";
import FakeMap from "@/components/FakeMap";
import FarePanel from "@/components/FarePanel";
import BookingPanel from "@/components/BookingPanel";
import StatusDisplay from "@/components/StatusDisplay";

const Index = () => {
  const [pickup, setPickup] = useState<{ lat: number; lng: number } | null>(null);
  const [rideActive, setRideActive] = useState(false);
  const [matched, setMatched] = useState(false);

  const handleSelectPickup = useCallback((lat: number, lng: number) => {
    setPickup({ lat, lng });
  }, []);

  const handleBooked = useCallback(() => {
    setRideActive(true);
  }, []);

  const handleCancel = useCallback(() => {
    setRideActive(false);
  }, []);

  const handleReset = useCallback(() => {
    setRideActive(false);
    setMatched(false);
    setPickup(null);
  }, []);

  const isLocked = rideActive || matched;

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b-[3px] border-foreground bg-card">
        <div className="max-w-5xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="bg-foreground text-primary-foreground px-3 py-1">
              <span className="font-mono text-lg font-bold">✈</span>
            </div>
            <div>
              <h1 className="font-mono text-lg font-bold tracking-tight">
                {AIRPORT.name}
              </h1>
              <p className="font-mono text-[10px] text-muted-foreground tracking-widest">
                CAB BOOKING TERMINAL
              </p>
            </div>
          </div>
          <div className="airport-badge bg-accent text-accent-foreground">
            {AIRPORT.code} • {AIRPORT.lat}°N {AIRPORT.lng}°E
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-5xl mx-auto px-4 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
          {/* Map - takes 3 columns */}
          <div className="lg:col-span-3">
            <FakeMap
              onSelectPickup={handleSelectPickup}
              pickup={pickup}
              disabled={isLocked}
            />
          </div>

          {/* Panels - takes 2 columns */}
          <div className="lg:col-span-2 space-y-6">
            {!rideActive && (
              <>
                <FarePanel
                  pickup={pickup}
                  onFareCalculated={() => {}}
                  disabled={isLocked}
                />
                <BookingPanel
                  pickup={pickup}
                  onBooked={handleBooked}
                  disabled={isLocked}
                />
              </>
            )}

            <StatusDisplay
              active={rideActive}
              onCancel={handleCancel}
              onReset={handleReset}
            />
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t-[3px] border-foreground mt-12">
        <div className="max-w-5xl mx-auto px-4 py-3 flex items-center justify-between">
          <span className="font-mono text-[10px] text-muted-foreground">
            GATE A1 • TERMINAL 1
          </span>
          <span className="font-mono text-[10px] text-muted-foreground">
            SYS OPERATIONAL
          </span>
        </div>
      </footer>
    </div>
  );
};

export default Index;
