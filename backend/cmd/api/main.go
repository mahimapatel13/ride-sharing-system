package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mahimapatel13/ride-sharing-system/internal/api/rest/router"
	"github.com/mahimapatel13/ride-sharing-system/internal/infrastructure/queue"
	"github.com/redis/go-redis/v9"

	"github.com/mahimapatel13/ride-sharing-system/internal/config"
	"github.com/mahimapatel13/ride-sharing-system/internal/infrastructure/worker"
)

func main() {
	log.Printf("Bootstrapping sytem..")

	log.Printf("Loading .env ..")
	cfg := config.LoadEnv()

    ctx := context.Background()

    // setting up redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisConfig.Address,
		Password: cfg.RedisConfig.Password,
		DB:       cfg.RedisConfig.DB,
		Protocol: cfg.RedisConfig.Protocol,
	})

	queueName := "ride-matching"

    // configuring RabbitMQ queue options
	queueOpts := queue.QueueOptions{
		Name:       queueName,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
	}

    // establishing connection to RabbitMQ
	mqConn, err := queue.ConnectRabbitMQ(cfg.RabbitMQConfig)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err.Error())
	}

    // creating a channel to communicate to RabbitMQ queue
	mqChan, err := queue.CreateChannel(mqConn, queueName)
	if err != nil {
		log.Fatalf("Failed to initialise RabbitMQ channel: %s", err.Error())
	}

    // declaring a persistent RabbitMQ queue 
	err = queue.DeclareQueue(mqChan.Channel,queueOpts)
    if err != nil {
        log.Fatalf("Failed to decalre RabbitMQ queue: %s", err.Error())
    }

    // intialising worker pool object
	workerPool := worker.NewPool(cfg.MaxWorkerCount, mqChan.Channel, queueName, redisClient)
    
    // starting the pool of workers
    workerPool.Run()

    // connecting to database
    db, err := pgxpool.New(ctx, fmt.Sprintf("user=%v password=%v host=%v port=%v dbname=%v", cfg.DatabaseConfig.User, cfg.DatabaseConfig.Password,cfg.DatabaseConfig.Host,cfg.DatabaseConfig.Port, cfg.DatabaseConfig.DatabaseName))

    if err != nil{
        log.Fatalf("Failed to connect to database: %s", err.Error())
    }

    // setting up gin router
    r := gin.New()
    r.Use(gin.Logger(), gin.Recovery())

	r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{
			"http://localhost:8080",

		},
        AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        // CRITICAL: This allows your Interceptor to read the token!
        ExposeHeaders:    []string{"Authorization"}, 
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))


    // health check point
    r.GET("/health", func(c *gin.Context) {
		if err := db.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "database unavailable", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

    // register routes
    router.RegisterRoutes(r, db, redisClient, mqChan)

    // configure server with timeouts
	srv := &http.Server{
		Addr:         ":8081",
		Handler:      r,
		ReadTimeout:  time.Duration(10) * time.Second,
		WriteTimeout: time.Duration(10) * time.Second,
		IdleTimeout:  time.Duration(10) * time.Second,
	}

    // Create a server context for graceful shutdown
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

    log.Println("server")

    quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	
	// Start server in a goroutine
	go func() {
		log.Println("Server starting", "port", 8080)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start", "error", err)
		}
		serverStopCtx()
	}()

    
	// Wait for shutdown signal
	select {
	case <-quit:
		log.Println("Shutdown signal received...")
	case <-serverCtx.Done():
		log.Println("Server stopped...")
	}

	// Create a deadline for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer shutdownCancel()

	// Shutdown the server
	log.Println("Shutting down server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown", "error", err)
	}

	log.Println("Server exited properly")

}
