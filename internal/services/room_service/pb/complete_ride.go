package roomservice

// CompleteRideRequest — запрос на завершение поездки
type CompleteRideRequest struct {
	RoomId     string  `json:"room_id"`
	DriverId   string  `json:"driver_id"`
	TotalPrice float32 `json:"total_price"`
	DistanceKm float32 `json:"distance_km"`
}

// CompleteRideResponse — ответ после завершения поездки
type CompleteRideResponse struct {
	Success       bool    `json:"success"`
	TotalPrice    float32 `json:"total_price"`
	CostPerMember float32 `json:"cost_per_member"`
	PaymentsCount int32   `json:"payments_count"`
}
