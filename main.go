package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	_ "log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/frenchdev/caltrain-slack/model"
	"github.com/gocraft/web"
)

type Context struct {
	HelloCount int
}

type StopDir struct {
	StopName  string
	Direction string
}

var _MapStopByID *map[int]model.Stop
var _MapStopIDByName *map[string]int
var _MapTimesByID *map[int][]string

func cleanJson() {
	cmd := exec.Command("python $HOME/go/src/caltrain-slack/python/jsonCleaner.py")
	//cmd := exec.Command("cd $GOPATH")
	fmt.Println(cmd.Run())
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		//log.Fatal("$PORT must be set")
		// for testing
		port = "5000"
	}
	//cleanJson()
	_MapStopByID = getStops("./gtfs/stops.json")
	_MapStopIDByName = setMapStopIDByName(_MapStopByID)
	_MapTimesByID = setMapTimesByID(getStoptimes("./gtfs/stoptimes.json"))

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
	//fmt.Fprint(rw, "Northbound: ", req.PathParams["direction"])
	//fmt.Fprint(rw, "Stop Name: ", req.PathParams["stop_name"])

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

	hr, min, sec := time.Now().Clock()
	stringTime := strconv.Itoa(hr) + ":" + strconv.Itoa(min) + ":" + strconv.Itoa(sec)

	stopDir := direction + "_" + stopName
	stopID := (*_MapStopIDByName)[stopDir]

	nextTrains := (*_MapTimesByID)[stopID]

	idx := findTimeIdx(&stringTime, &nextTrains)
	if idx == -1 {
		idx = len(nextTrains) - 1
	}

	if len(nextTrains) > idx+3 {
		//fmt.Fprint(rw, json.Marshal(nextTrains[idx:idx+3]))
		fmt.Fprint(rw, nextTrains[idx:idx+3])
	} else {
		//fmt.Fprint(rw, json.Marshal(nextTrains[idx:]))
		fmt.Fprint(rw, nextTrains[idx:])
	}

	//fmt.Println(nextTrains[])
}

func getStops(stopsFilePath string) *map[int]model.Stop {
	stopsFilePath, _ = filepath.Abs(stopsFilePath)
	stopsFile, err := ioutil.ReadFile(stopsFilePath)
	if err != nil {
		fmt.Println("opening stops file: ", err)
	}

	var stops []model.Stop

	err = json.Unmarshal(stopsFile, &stops)
	if err != nil {
		fmt.Println("error:", err)
	}

	stopMap := make(map[int]model.Stop)

	for _, stop := range stops {
		stopMap[stop.StopID] = stop
	}

	return &stopMap
}

func getStoptimes(stoptimesFilePath string) *[]model.StopTime {
	stoptimesFilePath, _ = filepath.Abs(stoptimesFilePath)
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

// to translate request from Stop name to Stop ID
func setMapStopIDByName(stops *map[int]model.Stop) *map[string]int {
	stopIDByName := make(map[string]int)
	for k, v := range *stops {
		if v.PlatformCode == "NB" {
			_StopDir := "NB_" + v.StopName
			stopIDByName[_StopDir] = k
		} else {
			_StopDir := "SB_" + v.StopName
			stopIDByName[_StopDir] = k
		}
	}
	return &stopIDByName
}

func setMapTimesByID(stopTimes *[]model.StopTime) *map[int][]string {
	timesByID := make(map[int][]string)
	var _emptyList []string

	// init map
	for k, _ := range *_MapStopByID {
		timesByID[k] = _emptyList
	}

	for _, stopTime := range *stopTimes {
		if _, ok := timesByID[stopTime.StopID]; ok {
			timesByID[stopTime.StopID] = append(timesByID[stopTime.StopID], stopTime.DepartureTime)
		}
	}

	return &timesByID
}

func findTimeIdx(time *string, times *[]string) int {
	for k, v := range *times {
		if *time > v {
			return k - 1
		}
	}
	return -1
}
