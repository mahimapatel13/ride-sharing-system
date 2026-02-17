package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mahimapatel13/ride-sharing-system/internal/api/rest/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/mahimapatel13/ride-sharing-system/internal/api/rest/request"
	"github.com/mahimapatel13/ride-sharing-system/internal/domain/ride"
	"github.com/mahimapatel13/ride-sharing-system/internal/api/rest/handlers"

)

func RegisterRideRoutes(
	r *gin.RouterGroup,
    redisClient *redis.Client,
    rideService ride.Service,

){
    h := handlers.NewRideHandler(rideService)

    ride := r.Group("/ride")
    {
        ride.POST("/fare", middleware.ReqValidate[request.Location](), h.CalculateFare)
        ride.GET("/request", h.RequestRide)
    }
}