package ride

// Rider represents the Rider entity stored in Redis
type Rider struct {
	ID        int
	Geohash   string
	Longitude float64
	Latitude  float64
	Luggage   int
	Tolerance float64
}

type Trip struct {
	TripID     int
	Passengers []Rider
	Luggage    int
}

type Fare struct {
	ID     int
	Amount float64
}
