package model

//StopTime is a stop time of the caltrain
type StopTime struct {
	ArrivalTime   string `json:"arrival_time"`
	DepartureTime string `json:"departure_time"`
	DropOffType   int    `json:"drop_off_type"`
	PickupType    int    `json:"pickup_type"`
	StopID        int    `json:"stop_id"`
	StopSequence  int    `json:"stop_sequence"`
	TripID        string `json:"trip_id"`
}
