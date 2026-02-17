package handlers

import (
	"context"
	// "sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mahimapatel13/ride-sharing-system/internal/api/rest/request"
	"github.com/mahimapatel13/ride-sharing-system/internal/domain/ride"
	"github.com/mmcloughlin/geohash"

	"log"
	"net/http"
)

type RideHandler struct {
    service ride.Service
}

func NewRideHandler(service ride.Service) *RideHandler{
    return &RideHandler{
        service: service,
    }
}

func (h *RideHandler) RequestRide(c *gin.Context) {

	log.Printf("Handling Request Ride  request....")
	// riderReq := request.GetReqBody[request.RideRequest](c)
    

	log.Println("!!!")
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer ws.Close()

	var riderReq request.RideRequest

	riderID, _, err := h.service.GetRiderandTripID(c.Request.Context())
	riderReq.RiderID = riderID

	if err := ws.ReadJSON(&riderReq); err != nil {
		log.Println("ReadJSON error:", err)
		return
	}
	geohash := geohash.Encode(riderReq.Lat, riderReq.Lng)
	log.Println("hi!!!!")

	ctx, cancelCtx := context.WithCancel(c.Request.Context())
	defer cancelCtx()

	cancelChan := make(chan struct{})

	// ---- WS Reader Goroutine (listen for CANCEL / disconnect) ----
	go func() {
		for {
			var msg map[string]any
			if err := ws.ReadJSON(&msg); err != nil {
				// client disconnected
				cancelCtx()
				return
			}

			switch msg["type"] {
			case "CANCEL_RIDE":
				select {
				case cancelChan <- struct{}{}:
				default:
				}
				return
			}
		}
	}()

	log.Println("RIDER!!!")
    req := ride.Rider{
        ID: riderReq.RiderID,
        Latitude: riderReq.Lat,
        Longitude: riderReq.Lng,
        Luggage: riderReq.Luggage,
        Geohash: geohash,

    }
	if err := h.service.AddRiderPresence(ctx, req); err != nil {
		log.Println("AddRiderPresence error:", err)
		_ = ws.WriteJSON(gin.H{"type": "error", "message": "failed to add rider"})
		return
	}

	_ = ws.WriteJSON(gin.H{
		"type":   "status",
		"status": "PENDING",
		"msg":    "Searching for shared ride...",
	})

	select {
	case <-time.After(5 * time.Second):
	case <-cancelChan:
		h.rollback(ctx, riderReq,geohash)
		_ = ws.WriteJSON(gin.H{"type": "status", "status": "CANCELLED"})
		return
	case <-ctx.Done():
		h.rollback(ctx, riderReq, geohash)
		return
	}

	if _, err := h.service.RequestRide(ctx, req); err != nil {
		log.Println("RequestRide publish error:", err)
		h.rollback(ctx, riderReq, geohash)
		_ = ws.WriteJSON(gin.H{"type": "error", "message": "failed to start matching"})
		return
	}

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-cancelChan:
			h.rollback(ctx, riderReq, geohash)
			_ = ws.WriteJSON(gin.H{"type": "status", "status": "CANCELLED"})
			return

		case <-ctx.Done():
			h.rollback(ctx, riderReq, geohash)
			return

		case <-ticker.C:
			status, cabID, err := h.service.GetRiderStatus(ctx, riderReq.RiderID)
			if err != nil {
				log.Println("GetRiderStatus error:", err)
				continue
			}

			if status == "MATCHED" {
				_ = ws.WriteJSON(gin.H{
					"type":   "status",
					"status": "MATCHED",
					"cab_id": cabID,
					"msg":    "Driver found!",
				})
				return
			}

			_ = ws.WriteJSON(gin.H{
				"type":   "status",
				"status": "PENDING",
			})
		}
	}
}



func(h *RideHandler) CalculateFare(c *gin.Context){
    
    log.Printf("Handling CalculateFare request..")

	log.Printf("hi")
    r := request.GetReqBody[request.Location](c)
    // geohash := geohash.Encode(r.Pickup.Lat, r.Pickup.Lng)

    fare, err := h.service.CalculateFare(c.Request.Context(), r.Lat, r.Lng)
    if err != nil {
        log.Printf("Error in calculating fare to airport: %s", err.Error())
        c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
    }

    c.JSON(http.StatusCreated, gin.H{
        "fare_id": fare.ID,
        "fare": fare.Amount,
    })
}


func (h *RideHandler) rollback(ctx context.Context, req request.RideRequest, geoHash string) {
	// mare rider status as cancelled
	_ = h.service.MarkRiderCancelled(ctx, req.RiderID)

	// remove from geohash set waiting pool
	_ = h.service.RemoveFromWaitingPool(ctx, req.RiderID, geoHash)

	// check for assiged cab
	cabID, err := h.service.GetAssignedCabIfAny(ctx, req.RiderID)
	if err == nil && cabID != "" {
		// release can and recaculate the min tolerance
		_ = h.service.ReleaseCabSeat(ctx, cabID, req.RiderID)

		_ = h.service.NotifyDriverCancellation(ctx, cabID, req.RiderID)
	}

	_ = h.service.DeleteRiderRedisKeys(ctx, req.RiderID)
}



var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
