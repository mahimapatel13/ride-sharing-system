package ride

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/mahimapatel13/ride-sharing-system/internal/infrastructure/queue"

	"github.com/mmcloughlin/geohash"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	// "github.com/mahimapatel13/ride-sharing-system/internal/infrastructure/worker"
)

type Service interface {

    AddRiderPresence(ctx context.Context, req Rider) error
    GetRiderStatus(ctx context.Context, riderID int) (status string, cabID string, err error)

	GetRiderandTripID(ctx context.Context) (int, int, error)
    RequestRide(ctx context.Context, ride Rider) (*Trip, error)
    CalculateFare(ctx context.Context, lat, lng float64 )(*Fare, error)

    // SaveTripDetails(ctx context.Context, trip Trip, rider Rider) error
    // MarkFareInGeohash(ctx cont, geohash string, fare float64) error
    MarkRiderCancelled(ctx context.Context, riderID int) error
    RemoveFromWaitingPool(ctx context.Context, riderID int, geohash string) error
    GetAssignedCabIfAny(ctx context.Context, riderID int) (string, error)
    ReleaseCabSeat(ctx context.Context, cabID string, riderID int) error
    NotifyDriverCancellation(ctx context.Context, cabID string, riderID int) error
    DeleteRiderRedisKeys(ctx context.Context, riderID int) error
}

type service struct {
    redisClient *redis.Client
    mqChannel *queue.MQChannel
	repo Repository
}

// NewRideService function initialises a new ride service 
func NewRideService(mqChannel *queue.MQChannel, redisClient *redis.Client, repo Repository) Service{
    return &service{
        mqChannel: mqChannel,
        redisClient: redisClient,
		repo: repo,
    }
}


func(s *service)GetRiderandTripID(ctx context.Context) (int, int,error){
	return s.repo.GetRiderandTripID(ctx)
}
func (s *service) AddRiderPresence(ctx context.Context, req Rider) error {
	riderKey := fmt.Sprintf("rider:%d", req.ID)

	pipe := s.redisClient.TxPipeline()

    // Add rider key value 
	pipe.HSet(ctx, riderKey, map[string]interface{}{
		"lat":            req.Latitude,
		"lng":            req.Longitude,
		"geohash":        req.Geohash,
		"status":         "PENDING",
		"last_update_ts": time.Now().Unix(),
	})

    // add rider to geohash set
	pipe.SAdd(ctx, fmt.Sprintf("pool:cell:%s:waiting", req.Geohash), req.ID)

	_, err := pipe.Exec(ctx)
	return err
}


func (s *service) GetRiderStatus(ctx context.Context, riderID int) (string, string, error) {
	riderKey := fmt.Sprintf("rider:%d", riderID)

    // get status from rider key value
	status, err := s.redisClient.HGet(ctx, riderKey, "status").Result()
	if err != nil {
		return "", "", err
	}

    // get cab details from rider key value
	cabID, err := s.redisClient.HGet(ctx, riderKey, "cab_id").Result()
	if err == redis.Nil {
		return status, "", nil
	}
	if err != nil {
		return "", "", err
	}

	return status, cabID, nil
}

func (s *service) MarkRiderCancelled(ctx context.Context, riderID int) error {
	riderKey := fmt.Sprintf("rider:%d", riderID)
    // set rider key value as cancelled
	return s.redisClient.HSet(ctx, riderKey, "status", "CANCELLED").Err()
}


func (s *service) RemoveFromWaitingPool(ctx context.Context, riderID int, geohash string) error {
    // remove from geohahs set
	return s.redisClient.SRem(ctx, fmt.Sprintf("pool:cell:%s:waiting", geohash), riderID).Err()
}


func (s *service) GetAssignedCabIfAny(ctx context.Context, riderID int) (string, error) {
	riderKey := fmt.Sprintf("rider:%d", riderID)
	cabID, err := s.redisClient.HGet(ctx, riderKey, "cab_id").Result()
	if err == redis.Nil {
		return "", nil
	}
	return cabID, err
}


func (s *service) ReleaseCabSeat(ctx context.Context, cabID string, riderID int) error {
	lua := `
    
    -- KEYS[1] = cab:{id}
    -- KEYS[2] = cab:{id}:riders
    -- ARGV[1] = riderID

    redis.call("SREM", KEYS[2], ARGV[1])
    redis.call("HINCRBY", KEYS[1], "passenger_count", -1)

    local capacity = tonumber(redis.call("HGET", KEYS[1], "capacity"))
    local pc = tonumber(redis.call("HGET", KEYS[1], "passenger_count"))

    if pc < capacity then
    redis.call("HSET", KEYS[1], "status", "AVAILABLE")
    end

    return 1

    `

	cabKey := fmt.Sprintf("cab:%s", cabID)
	cabRidersKey := fmt.Sprintf("cab:%s:riders", cabID)

	_, err := s.redisClient.Eval(ctx, lua, []string{cabKey, cabRidersKey}, riderID).Result()
	return err
}


func (s *service) NotifyDriverCancellation(ctx context.Context, cabID string, riderID int) error {
    message := map[string]interface{}{
        "cabID": cabID,
        "riderID": riderID,
        "event": "RIDER_CANCELLED",
    }
    body, _ := json.Marshal(message)
    return s.publishMessage(body)
}

func (s *service) DeleteRiderRedisKeys(ctx context.Context, riderID int) error {
    // delete rider key
	riderKey := fmt.Sprintf("rider:%d", riderID)
	return s.redisClient.Del(ctx, riderKey).Err()
}




func (s *service) RequestRide(ctx context.Context, req Rider) (*Trip, error) {
	if req.ID == 0 {
		log.Println("Invalid rider ID")
		return nil, fmt.Errorf("invalid rider ID")
	}

	geohash := geohash.Encode(req.Latitude, req.Longitude)

	riderKey := fmt.Sprintf("rider:%d", req.ID)

	err := s.redisClient.HSet(ctx, riderKey, map[string]interface{}{
		"lat":     req.Latitude,
		"lng":     req.Longitude,
		"geohash": req.Geohash,
		"status":  "PENDING",
	}).Err()

	if err != nil {
		log.Printf("failed to store ride request in redis: %v", err)
		return nil, err
	}

    rider := Rider{
        ID: req.ID,
        Geohash: geohash,
        Longitude: req.Longitude,
        Latitude: req.Latitude,
    }
	
	body, err := json.Marshal(rider)
	if err != nil {
		log.Printf("failed to marshal ride request: %v", err)
		return nil, err
	}

	if err := s.publishMessage(body); err != nil {
		log.Printf("failed to publish ride request: %v", err)
		return nil, err
	}

	trip := &Trip{
		
	}

	return trip, nil
}

func (s *service) CalculateFare(ctx context.Context, lat, lng float64) (*Fare, error) {
	gh := geohash.Encode(lat, lng)

	distanceKm := haversineKmToAirport(lat, lng)

	const basePerKm = 12.0 // tweak as needed
	baseFare := distanceKm * basePerKm

	waitingKey := fmt.Sprintf("pool:cell:%s:waiting", gh)

	log.Println("hii rediis")
	demand, err := s.redisClient.SCard(ctx, waitingKey).Result()
	if err != nil {
		return nil, err
	}

	multiplier := 1.0 + float64(demand)*0.05
	if multiplier > 2.0 {
		multiplier = 2.0
	}

	finalFare := baseFare * multiplier

	const minFare = 50.0
	if finalFare < minFare {
		finalFare = minFare
	}

	f := &Fare{
		ID:     int(time.Now().Unix()), // or some generator
		Amount: math.Round(finalFare*100) / 100, // round to 2 decimals
	}

	cacheKey := fmt.Sprintf("fare:%s:%f:%f", gh, lat, lng)
	_ = s.redisClient.Set(ctx, cacheKey, f.Amount, 30*time.Second).Err()

	return f, nil
}


func(s *service) publishMessage(rider []byte) error{
    err := s.mqChannel.Channel.Publish(
        "",
        s.mqChannel.QueueName,
        true,
        false,
        amqp.Publishing{
            ContentType: "text/plain",
            Body: rider,
        },
    )

    return err
}

func haversineKmToAirport(lat1, lon1 float64) float64 {
	const R = 6371.0 // Earth radius in km

    lat2 := 23.263567
    lon2 := 77.363717
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
