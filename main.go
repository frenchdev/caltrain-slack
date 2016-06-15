package main

import (
	"caltrain-slack/model"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

func main() {
	content, err := ioutil.ReadFile("./gtfs/stops.txt")

	if err != nil {
		log.Fatal(err)
	}

	s := string(content)

	fmt.Println(getStops(s))
}

func getStops(s string) []*model.Stop {
	var stops []*model.Stop
	var stop *model.Stop
	lines := strings.Split(s, "\r\n")

	for _, line := range lines {
		stopLines := strings.Split(line, ",")
		for _, stopLine := range stopLines {
			id, err := strconv.Atoi(string(stopLine[0]))
			if err == nil {
				stop = model.NewStop(id, string(stopLine[2]), string(stopLine[9]))
				stops = append(stops, stop)
			}
		}
	}

	return stops
}
