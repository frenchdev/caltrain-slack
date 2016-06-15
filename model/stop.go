package model

//Stop is a stop for the caltrain
type Stop struct {
	id         int
	name       string
	northbound bool
}

//NewStop create a new stop
func NewStop(id int, name string, northbound string) *Stop {
	s := Stop{}
	s.id = id
	s.name = name
	if northbound == "NB" {
		s.northbound = true
	} else {
		s.northbound = false
	}
	return &s
}
