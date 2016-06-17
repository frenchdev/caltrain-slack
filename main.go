package main

import (
	"caltrain-slack/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+port, nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	// var (
	// 	response string
	// 	err      error
	// )
	switch r.URL.Path[1:] {
	case "next":

		//response := url.ParseQuery(r.URL.RawQuery)
		//jsonString := json.Marshal(response)
		fmt.Println("query: ", r.URL.RawQuery)
		fmt.Fprintf(w, "coucou")
	default:
		response := "Not implemented"
		fmt.Fprintf(w, response)
	}

}

func nextCaltrain() string {
	nothing := "The next caltrain is"
	return nothing
}

func getStops(stopsFilePath string) *map[int]model.Stop {
	stopsFile, err := ioutil.ReadFile(stopsFilePath)
	if err != nil {
		fmt.Println("opening stops file: ", err)
	}

	var stops []model.Stop

	err = json.Unmarshal(stopsFile, &stops)
	if err != nil {
		fmt.Println("error:", err)
	}

	var stopMap map[int]model.Stop

	for _, stop := range stops {
		stopMap[stop.StopID] = stop
	}

	return &stopMap
}

func getStoptimes(stoptimesFilePath string) *[]model.StopTime {
	stoptimesFile, err := ioutil.ReadFile(stoptimesFilePath)
	if err != nil {
		fmt.Println("opening stops file: ", err)
	}

	var stoptimes []model.StopTime

	err = json.Unmarshal(stoptimesFile, &stoptimes)
	if err != nil {
		fmt.Println("error:", err)
	}

	return &stoptimes
}

func mapStop(stops *[]model.Stop) *map[int]model.Stop {
	var stopMap map[int]model.Stop

	for _, stop := range *stops {
		stopMap[stop.StopID] = stop
	}

	return &stopMap
}
