package ride

import "context"

type Repository interface {
	GetRiderandTripID(ctx context.Context) (int, int, error)
}