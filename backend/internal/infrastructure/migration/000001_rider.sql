CREATE SCHEMA IF NOT EXISTS rider_schema;

CREATE TABLE rider_schema.trip (
    id SERIAL PRIMARY KEY,
    cab_id TEXT NOT NULL,
    status TEXT NOT NULL, 
    pickup_lat DOUBLE PRECISION NOT NULL,
    pickup_lng DOUBLE PRECISION NOT NULL,
    drop_lat DOUBLE PRECISION NOT NULL,
    drop_lng DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX idx_trip_status
ON rider_schema.trip (status);

CREATE INDEX idx_trip_created_at
ON rider_schema.trip (created_at);

CREATE INDEX idx_trip_cab_id
ON rider_schema.trip (cab_id);

CREATE TABLE rider_schema.rider_trip (
    id SERIAL PRIMARY KEY,
    trip_id INT NOT NULL,
    joined_at TIMESTAMP NOT NULL DEFAULT now(),

    CONSTRAINT fk_trip
        FOREIGN KEY (trip_id)
        REFERENCES rider_schema.trip(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_rider_trip UNIQUE (id, trip_id)
);

CREATE INDEX idx_rider_trip_trip_id
ON rider_schema.rider_trip (trip_id);

CREATE INDEX idx_rider_trip_joined_at
ON rider_schema.rider_trip (joined_at);
