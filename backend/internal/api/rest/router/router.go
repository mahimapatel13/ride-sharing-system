package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mahimapatel13/ride-sharing-system/internal/domain/ride"
	"github.com/mahimapatel13/ride-sharing-system/internal/infrastructure/database/postgres/repositories"
	"github.com/mahimapatel13/ride-sharing-system/internal/infrastructure/queue"
	"github.com/redis/go-redis/v9"
)

func RegisterRoutes(
	r *gin.Engine,
    pool *pgxpool.Pool,
    redisClient *redis.Client,
    mqChannel *queue.MQChannel,
){

    // api versioning
    v1 := r.Group("/api/v1")

    rideRepo := repositories.NewRideRepository(pool)

    rideService := ride.NewRideService(mqChannel, redisClient, rideRepo)
    RegisterRideRoutes(v1, redisClient,rideService)
}