package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mahimapatel13/ride-sharing-system/internal/domain/ride"
)

type repository struct {
	pool *pgxpool.Pool
}

func NewRideRepository(pool *pgxpool.Pool) ride.Repository{
    return &repository{
        pool: pool,
    }
}

func(r *repository) GetRiderandTripID(ctx context.Context) (int, int, error){
    tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var tripID int
	var riderTripID int

	err = tx.QueryRow(ctx, `
		INSERT INTO rider_schema.trip (cab_id, status, pickup_lat, pickup_lng, drop_lat, drop_lng)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`,
		"",           
		"CREATED",     
		0.0, 0.0,      
		0.0, 0.0,      
	).Scan(&tripID)

	if err != nil {
		return 0, 0, err
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO rider_schema.rider_trip (trip_id)
		VALUES ($1)
		RETURNING id
	`, tripID).Scan(&riderTripID)

	if err != nil {
		return 0, 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, 0, err
	}

	return tripID, riderTripID, nil
}