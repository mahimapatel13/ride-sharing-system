package worker

import (
	// "log"
	// "sync"

	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/mahimapatel13/ride-sharing-system/internal/domain/ride"
	"github.com/mmcloughlin/geohash"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

var queueName string
var airportlat float64
var airportlng float64

type Job struct {
	ID       int32
	Delivery amqp.Delivery
}

// Pool represents the worker pool structure
type Pool struct {
	WorkerCount   int
	WorkerChannel chan chan Job
	JobQueue      *amqp.Channel
    RedisClient  *redis.Client
	Stopped       chan bool
}

// Worker represents the actual worker doing the job
type Worker struct {
	ID            int
	JobChannel    chan Job
	WorkerChannel chan chan Job // used to communicate between dispatcher and worker
    RedisClient   *redis.Client
	Quit          chan bool
}

// NewPool returns contructs and returns new Pool object
func NewPool(workerCount int, jobQueueChannel *amqp.Channel, queue string, rdb *redis.Client) Pool {
	queueName = queue 
	airportlat = 23.3179
	airportlng = 77.349225
	return Pool{
		WorkerCount:   workerCount,
		WorkerChannel: make(chan chan Job),
		JobQueue:      jobQueueChannel,
        RedisClient: rdb,
		Stopped:       make(chan bool),
	}
}

func (p *Pool) Run() {
	log.Println("Spawning the workers")

	for i := range p.WorkerCount {
		worker := Worker{
			ID:            i + 1,
			JobChannel:    make(chan Job),
			WorkerChannel: p.WorkerChannel,
            RedisClient: p.RedisClient,
			Quit:          make(chan bool),
		}
		worker.start()
	}

	p.allocate()
}

// RabbitMQ message conumer and job dispatcher
func (p *Pool) allocate() {
	q := p.JobQueue
	
	msgs, err := q.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	go func() {
		for {
			select {
			case d, ok := <-msgs:
				if !ok {
					log.Println("RabbitMQ channel closed, stopping allocator")
					return
				}

				job := Job{Delivery: d}

				workerCh := <-p.WorkerChannel
				workerCh <- job

			case <-p.Stopped:
				log.Println("Allocator received stop signal, shutting down")
				return
			}
		}
	}()
}

func (w *Worker) start() {

	go func() {
		for {
			w.WorkerChannel <- w.JobChannel // when the worker is available place channel in queue
			select {
			case job := <-w.JobChannel: // worker has recived job
				w.matchRide(job)
			case <-w.Quit:
				return
			}
		}
	}()
}

// work is the task performer
// This is the main function that matches riders
// It looks for riders in the same geo cell and in neighbouring 8 geo cells
// the function checks compatibility against each cab avaible in the geo area
// the cab data maintains a special "min tolerance distance" attribute which 
// is updated any time a new passengers boards the cab
// the matchride function checks if detour <= min tolerance distance AND 
// the capcity of the cab is not exceede
// if not exceeded while matching, then we lock the cab and insert passenger into the cab
// if no compatible match found, we simply book a cab only for a single passenger
func (w *Worker) matchRide(job Job) {

	log.Printf("------")

    // fetch from redis all the active users
    j := job.Delivery.Body
    var rider ride.Rider
    err := json.Unmarshal(j, &rider);
	
    if(err != nil){
		log.Println(fmt.Sprintf("Error while reading unmarshalling job for JOB ID %d : %s",job.ID,err.Error()))
        return
    }

	job.ID = int32(rider.ID)

    cells := geohash.Neighbors(rider.Geohash)
    
    cells = append([]string{rider.Geohash}, rider.Geohash)

    cabIDSet := make(map[string]struct{}) 

    for _, cell := range cells{
        geohash := cell
        ids, err := w.RedisClient.SMembers(context.Background(),fmt.Sprintf("cell:%s:cabs",geohash)).Result()
        if err != nil {
            log.Println(fmt.Printf("Error while reading from Redis set for JOB ID %d : %s", job.ID,err.Error()))
            return
        }

        for _, id := range ids {
			cabIDSet[id] = struct{}{}
		}
        
    }
    
    bestCabID := ""
    bestScore := math.MaxFloat64

    now := time.Now().Unix()

	var totaldistance float64

    for cabID := range cabIDSet {
        cab, err := w.RedisClient.HGetAll(context.Background(),fmt.Sprintf("cab:%s", cabID)).Result()

        if err != nil || len(cab) == 0{
            continue
        }

        if cab["status"] != "AVAILABLE"{
            continue
        }

        lastUpdate, err := strconv.ParseInt(cab["last_update_ts"], 10, 64)
		if err != nil {
			continue
		}
		if now-lastUpdate > 30 { 
			continue
		} 

        // parse fields
		cabLat, _ := strconv.ParseFloat(cab["lat"], 64)
		cabLng, _ := strconv.ParseFloat(cab["lng"], 64)
		passengerCount, _ := strconv.Atoi(cab["passenger_count"])
        luggageCount, _ := strconv.Atoi(cab["luggage_count"])
		minTolerance, _ := strconv.ParseFloat(cab["min_tolerance_km"], 64)
        capacity, _ :=strconv.Atoi(cab["capacity"])

        // compute distance (km)
		totaldistance = haversineKm(rider.Latitude, rider.Longitude, cabLat, cabLng)

        // tolerance rule
		if passengerCount + luggageCount + 1 + rider.Luggage <= capacity {
			if totaldistance > minTolerance {
				continue
			}
		}
        score := totaldistance

        if score < bestScore{
            bestScore = score
            bestCabID = cabID
        }
    }

	d := haversineKm(rider.Latitude, rider.Longitude, airportlat, airportlng) 
	tolerance := computeRiderToleranceKm(rider.Tolerance, d)
	if bestCabID == "" {
		log.Printf("No cab found for rider %d, creating a new cab", rider.ID)

		ctx := context.Background()

		cabID := randomCabID()
		cabKey := fmt.Sprintf("cab:%s", cabID)
		cabRidersKey := fmt.Sprintf("cab:%s:riders", cabID)
		riderKey := fmt.Sprintf("rider:%d", rider.ID)
		cellCabsKey := fmt.Sprintf("cell:%s:cabs", rider.Geohash)

		now := time.Now().Unix()

		pipe := w.RedisClient.TxPipeline()

		pipe.HSet(ctx, cabKey, map[string]interface{}{
			"lat":               rider.Latitude,
			"lng":               rider.Longitude,
			"passenger_count":   1,
			"luggage_count":     rider.Luggage,
			"capacity":          4, // or from config
			"min_tolerance_km":  tolerance,
			"status":            "AVAILABLE", // still available for more riders
			"last_update_ts":    now,
		})

		pipe.SAdd(ctx, cabRidersKey, rider.ID)

		pipe.SAdd(ctx, cellCabsKey, cabID)

		pipe.HSet(ctx, riderKey, map[string]interface{}{
			"status":         "MATCHED",
			"cab_id":         cabID,
			"last_update_ts": now,
		})

		waitKey := fmt.Sprintf("pool:cell:%s:waiting", rider.Geohash)
		pipe.SRem(ctx, waitKey, rider.ID)

		if _, err := pipe.Exec(ctx); err != nil {
			log.Println("Failed to create new cab and assign rider:", err)
			_ = job.Delivery.Nack(false, true) // retry
			return
		}

		_ = job.Delivery.Ack(false)
		return
	}

	
    success, err := w.tryAssignCab(context.Background(), bestCabID, rider.ID, tolerance)
	if err != nil {
		log.Printf("Assignment error: %v", err)
		_ = job.Delivery.Nack(false, true)
		return
	}

	if success {
		log.Printf(fmt.Sprintf("Assigned rider %d to cab %s", rider.ID, bestCabID))
	
		_ = job.Delivery.Ack(false)
		return
	}

	log.Printf("Race lost assigning cab %s, retrying job %d", bestCabID, job.ID)
	_ = job.Delivery.Nack(false, true)

	_ = job.Delivery.Ack(false)
	log.Printf("Processed by Worker [%d]", w.ID)
	log.Printf("Processed Job With ID [%d] & content: [%v]", job.ID, job.Delivery)
	log.Printf("-------")
}

func (w *Worker) tryAssignCab(ctx context.Context, cabID string, riderID int, riderToleranceKm float64) (bool, error) {
	lua := `
        -- KEYS[1] = cab key
        -- KEYS[2] = rider key
        -- KEYS[3] = cab riders set key
        -- ARGV[1] = riderID
        -- ARGV[2] = rider_tolerance_km

        local status = redis.call("HGET", KEYS[1], "status")
        if status ~= "AVAILABLE" then
            return 0
        end

        local passenger_count = tonumber(redis.call("HGET", KEYS[1], "passenger_count") or "0")
        local luggage_count = tonumber(redis.call("HGET", KEYS[1], "luggage_count") or "0")
        local capacity = tonumber(redis.call("HGET", KEYS[1], "capacity") or "4")

        if (passenger_count + luggage_count) >= capacity then
            return 0
        end

        redis.call("HINCRBY", KEYS[1], "passenger_count", 1)

        local rider_tol = tonumber(ARGV[2])
        local old_min_tol = redis.call("HGET", KEYS[1], "min_tolerance_km")

        if old_min_tol == false then
            -- If not set yet, set to rider's tolerance
            redis.call("HSET", KEYS[1], "min_tolerance_km", rider_tol)
        else
            local old_tol = tonumber(old_min_tol)
            if rider_tol < old_tol then
                redis.call("HSET", KEYS[1], "min_tolerance_km", rider_tol)
            end
        end

        redis.call("HSET", KEYS[2], "cab_id", string.sub(KEYS[1], 5))
        redis.call("HSET", KEYS[2], "status", "MATCHED")

        redis.call("SADD", KEYS[3], ARGV[1])

        if (passenger_count + luggage_count + 1) >= capacity then
            redis.call("HSET", KEYS[1], "status", "FULL")
        end

        return 1
    `

	cabKey := fmt.Sprintf("cab:%s", cabID)
	riderKey := fmt.Sprintf("rider:%d", riderID)
	cabRidersKey := fmt.Sprintf("cab:%s:riders", cabID)

	res, err := w.RedisClient.Eval(
		ctx,
		lua,
		[]string{cabKey, riderKey, cabRidersKey},
		strconv.Itoa(riderID),
		fmt.Sprintf("%f", riderToleranceKm),
	).Result()

	if err != nil {
		return false, err
	}

	ok, okType := res.(int64)
	if !okType {
		return false, fmt.Errorf("unexpected Lua return type: %T", res)
	}

	if ok != 1 {
		return false, fmt.Errorf("failed to assign rider:%d to cab:%s", riderID, cabID)
	}

	return true, nil
}

func computeRiderToleranceKm(detourFactor, directKm float64) float64 {
	
	tolerance := directKm * detourFactor 

	return tolerance
}



func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in km

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func randomCabID() string {
	return uuid.NewString() 
}


