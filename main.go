package main

import (
	"caltrain-slack/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	_"log"
	"net/http"
	"os"
	"github.com/gocraft/web"
	"os/exec"
)

type Context struct {
	HelloCount int64
}

type StopDir struct {
	StopName string
	Direction string
}
//type MapStopByID map[int]model.Stop
//type MapStopIDByName map[StopDir]int
var _MapStopByID map[int64]model.Stop
var _MapStopIDByName *map[string]int64

func cleanJson() {
	cmd := exec.Command("python $GOPATH/src/caltrain-slack/python/jsonCleaner.py")
	//cmd := exec.Command("cd %GOPATH%")
	fmt.Println(cmd.Run())
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		//log.Fatal("$PORT must be set")
		// for testing
		port = "5000"
	}
	cleanJson()
	_MapStopByID = getStops("src/caltrain-slack/gtfs/stops.json")
	_MapStopIDByName = setMapStopIDByName(&_MapStopByID)

	//http.HandleFunc("/", handler)
	//http.ListenAndServe(":"+port, nil)

	router := web.New(Context{}).
		Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).
		NotFound((*Context).NotFound).
		Get("/next/:direction/:stop_name", (*Context).FindStop)
	http.ListenAndServe("localhost:"+port, router)
}

func (c *Context) NotFound(rw web.ResponseWriter, r *web.Request) {
	rw.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(rw, "Not Found")
}

func (c *Context) FindStop(rw web.ResponseWriter, req *web.Request) {
	fmt.Fprint(rw, "Northbound: ", req.PathParams["direction"])
	fmt.Fprint(rw, "Stop Name: ", req.PathParams["stop_name"])

	//if req.PathParams["stop_name"] == nil || req.PathParams["direction"] == nil {
	//	rw.Header().Set("Location", "/")
	//	rw.WriteHeader(http.StatusMovedPermanently)
	//}

	direction := req.PathParams["direction"]
	if direction != "NB" && direction != "SB" {
		rw.Header().Set("Location", "/")
		rw.WriteHeader(http.StatusMovedPermanently)
	}
	stopName := req.PathParams["direction"]
	if stopName == "" {
		rw.Header().Set("Location", "/")
		rw.WriteHeader(http.StatusMovedPermanently)
	}

	// returns the stop ID for that stop & direction combo
	//m := make(map[string]int)

	stopDir := direction + "_" + stopName
	stopID := (*_MapStopIDByName)[stopDir]

	// while times are not set
	// let's write the stop details
	nextTrains := _MapStopByID[stopID]
	fmt.Println(nextTrains)
	//fmt.Fprintf(rw, json.Marshal(string(nextTrains)))
}

//func handler(w http.ResponseWriter, r *http.Request) {
//	// var (
//	// 	response string
//	// 	err      error
//	// )
//	switch r.URL.Path[1:] {
//	case "next":
//
//		//response := url.ParseQuery(r.URL.RawQuery)
//		//jsonString := json.Marshal(response)
//		fmt.Println("query: ", r.URL.RawQuery)
//		fmt.Fprintf(w, "coucou")
//	default:
//		response := "Not implemented"
//		fmt.Fprintf(w, response)
//	}
//}
//
//func nextCaltrain() string {
//	nothing := "The next caltrain is"
//	return nothing
//}

func getStops(stopsFilePath string) map[int64]model.Stop {
	stopsFile, err := ioutil.ReadFile(stopsFilePath)
	if err != nil {
		fmt.Println("opening stops file: ", err)
	}

	var stops []model.Stop

	err = json.Unmarshal(stopsFile, &stops)
	if err != nil {
		fmt.Println("error:", err)
	}

	var stopMap map[int64]model.Stop

	for _, stop := range stops {
		//if _, err := strconv.Atoi(stop.StopID); err == nil {
		stopMap[stop.StopID] = stop
		//}
	}

	return stopMap
}

//func getStoptimes(stoptimesFilePath string) *[]model.StopTime {
//	stoptimesFile, err := ioutil.ReadFile(stoptimesFilePath)
//	if err != nil {
//		fmt.Println("opening stops file: ", err)
//	}
//
//	var stoptimes []model.StopTime
//
//	err = json.Unmarshal(stoptimesFile, &stoptimes)
//	if err != nil {
//		fmt.Println("error:", err)
//	}
//
//	return &stoptimes
//}

//func mapStop(stops *[]model.Stop) *map[int]model.Stop {
//	var stopMap map[int]model.Stop
//
//	for _, stop := range *stops {
//		stopMap[stop.StopID] = stop
//	}
//
//	return &stopMap
//}

// to translate request from Stop name to Stop ID
func setMapStopIDByName(stops *map[int64]model.Stop) *map[string]int64 {
	var stopIDByName map[string]int64
	for k, v := range *stops {
		if v.PlatformCode == "NB" {
			_StopDir :=  "NB_" + v.StopName
			stopIDByName[_StopDir] = k
		} else {
			_StopDir :=  "SB_" + v.StopName
			stopIDByName[_StopDir] = k
		}
	}
	return &stopIDByName
}
