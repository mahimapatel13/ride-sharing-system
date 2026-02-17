export const AIRPORT = {
  lat: 23.2875,
  lng: 77.3370,
  name: "BHOPAL AIRPORT",
  code: "BHO",
};

// Change this to your actual API base URL
export const API_BASE = "http://localhost:8081/api/v1";

// WebSocket base (derived from API_BASE or set separately)
export const WS_BASE = API_BASE.replace(/^http/, "ws");
