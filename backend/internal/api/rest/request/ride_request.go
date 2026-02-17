package request

import "github.com/gin-gonic/gin"

type Location struct {
	Lat float64 `json:"lat" validate:"latitude"`
	Lng float64 `json:"lng" validate:"longitude"`
}

type CalculateFareRequest struct {
	Pickup Location `json:"pickup" validate:"required"`
}

type RideRequest struct {
	RiderID   int     `json:"rider_id"`
	Luggage   int     `json:"luggage" validate:"gte=0"`
	Lat       float64 `json:"lat" validate:"required,latitude"`
	Lng       float64 `json:"lng" validate:"required,longitude"`
	Tolerance float64 `json:"tolerance" validate:"required"`
}

func GetReqBody[T any](c *gin.Context) T {
	val, _ := c.Get("reqBody")
	return val.(T)
}
