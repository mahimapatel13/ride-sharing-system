# Ride Sharing System

powered by geohash & geospatial indexing provided by redis & golang worker pool

# Architecture Overview
<img width="7479" height="3657" alt="image" src="https://github.com/user-attachments/assets/4103e9a9-e047-48a8-a7d8-22f33c83f13a" />

link: https://excalidraw.com/#json=_qJcyO3x9C_Oe9gp2KnE_,rMweYx-wltCiSE8PMT8exg

# Ride Matching Algorithm Overview
<img width="1857" height="1039" alt="1" src="https://github.com/user-attachments/assets/c7319373-bc32-4e40-92e6-0b4944bcf3cc" />
<img width="1870" height="1049" alt="2" src="https://github.com/user-attachments/assets/2092b6a6-21a2-4d0c-b13c-76e16d6bcd27" />
<img width="1867" height="1044" alt="3" src="https://github.com/user-attachments/assets/016dfb53-4648-462f-82bd-502cbd2aa76d" />
<img width="1868" height="1049" alt="4" src="https://github.com/user-attachments/assets/51ae5f43-96b9-4ca6-b3eb-2d9eae68dbb7" />


## Techstack
- Golang (backend)
- RabbitMQ (buffer queue for worker pool)
- Redis (for faster lookup without network overhead)
- Postgres Database (for persistent storage)
- Typescript (frontend)

## Usage
Configure database setting in .env using .env.example

Ensure that redis and rabbit mq are running on the following port. Pull docker image and run their instances.
```
redis at localhost:6379
rabbit mq at localhost:5672
```
then run
```
cd backend
go build cmd/api/main.go
go run ./main

```

# Directory Structure
This project follows modular architecture to ensure sepearation of concerns.

```text
backend/
├── cmd/                               # Application entry points
│   ├── api/                           # REST + WebSocket API server
│   │   └── main.go                    # Bootstraps HTTP server, router, dependencies
│   └── worker/                        # Background worker process
│       └── main.go                    # Bootstraps worker, connects to Redis & RabbitMQ
│
├── internal/                          # Private application code
│   ├── api/
│   │   └── rest/
│   │       ├── handlers/              # HTTP / WebSocket handlers
│   │       │   └── ride_handler.go    # Ride request, WS handling, polling logic
│   │       ├── middleware/            # HTTP middlewares (auth, logging, etc.)
│   │       ├── request/               # Request DTOs
│   │       │   └── ride_request.go
│   │       ├── response/              # Response DTOs
│   │       └── router/                # Route registration
│   │           ├── router.go
│   │           └── ride_routes.go
│   │
│   ├── config/                        # Configuration loading (env, configs)
│   │
│   ├── domain/
│   │   └── ride/                      # Core ride domain (business logic)
│   │       ├── model.go               # Domain models (Rider, Cab, Fare, etc.)
│   │       ├── repository.go          # Repository interfaces (ports)
│   │       └── service.go             # Domain services (RequestRide, CalculateFare, etc.)
│   │
│   ├── infrastructure/
│   │   ├── database/
│   │   │   └── postgres/
│   │   │       └── repositories/      # PostgreSQL implementations of repositories
│   │   │           └── ride_repository.go
│   │   └── migration/                 # Database migrations
│   │       └── 000001_rider.sql
│   │
│   ├── queue/                         # Message queue abstraction (RabbitMQ)
│   │   ├── queue.go                   # Queue connection & setup
│   │   └── service.go                 # Publisher helpers
│   │
│   └── worker/                        # Background workers
│       └── ride_matching_worker.go    # Ride matching logic (Redis + Lua + MQ)
│
├── go.mod
└── go.sum
```
## A minimal dashboard :
<img width="1913" height="723" alt="Screenshot 2026-02-18 002856" src="https://github.com/user-attachments/assets/2f376a88-48d5-48cd-b21e-daaa40abf86e" />

