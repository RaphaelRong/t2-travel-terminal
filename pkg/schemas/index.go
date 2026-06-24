package schemas

import "time"

// RegionalIndex represents a regional market intelligence snapshot.
type RegionalIndex struct {
	Region      string    `json:"region"`
	Timestamp   time.Time `json:"timestamp"`
	DemandScore float64   `json:"demand_score"`
	SupplyScore float64   `json:"supply_score"`
	PriceIndex  float64   `json:"price_index"`
	ChannelHealth float64 `json:"channel_health"`
}
