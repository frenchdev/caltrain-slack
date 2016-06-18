package model

//Stop is a stop for the caltrain
type Stop struct {
	LocationType       int64		`json:"location_type"`
	ParentStation      string		`json:"parent_station"`
	PlatformCode       string		`json:"platform_code"`
	StopCode           int64		`json:"stop_code"`
	StopID             int64		`json:"stop_id"`
	StopLat            float64		`json:"stop_lat"`
	StopLon            float64		`json:"stop_lon"`
	StopName           string		`json:"stop_name"`
	StopURL            string		`json:"stop_url"`
	WheelchairBoarding int64		`json:"wheelchair_boarding"`
	ZoneID             int64		`json:"zone_id"`
}

//NewStop create a new stop
// func NewStop(id int, name string, northbound string) *Stop {
// 	s := Stop{}
// 	s.id = id
// 	s.name = name
// 	if northbound == "NB" {
// 		s.northbound = true
// 	} else {
// 		s.northbound = false
// 	}
// 	return &s
// }
