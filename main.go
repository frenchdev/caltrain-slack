package main

/* Refactoring needed: */

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	_ "log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
	"sort"
	"log"
	"strings"
	"sync/atomic"
	"golang.org/x/net/websocket"

	"github.com/gocraft/web"
	_"github.com/polaris1119/go.net/websocket"

	"github.com/caltrain-slack/model"
	"os"
)

type Context struct {
	HelloCount int
}

type StopDir struct {
	StopName  string
	Direction string
}

type NextTrain struct {
	Direction 	string
	StopName 	string
	Next		string
}

var _MapStopByID *map[int]model.Stop
var _MapStopIDByName *map[string]int
var _MapTimesByIDWeekDay *map[int][]string
var _MapTimesByIDWeekEnd *map[int][]string

func cleanJson() {
	cmd := exec.Command("python $HOME/go/src/caltrain-slack/python/jsonCleaner.py")
	//cmd := exec.Command("cd $GOPATH")
	fmt.Println(cmd.Run())
}

// These two structures represent the response of the Slack API rtm.start.
// Only some fields are included. The rest are ignored by json.Unmarshal.

type responseRtmStart struct {
	Ok    bool         `json:"ok"`
	Error string       `json:"error"`
	Url   string       `json:"url"`
	Self  responseSelf `json:"self"`
}

type responseSelf struct {
	Id string `json:"id"`
}

// slackStart does a rtm.start, and returns a websocket URL and user ID. The
// websocket URL can be used to initiate an RTM session.
func slackStart(token string) (wsurl, id string, err error) {
	url := fmt.Sprintf("https://slack.com/api/rtm.start?token=%s", token)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}
	var respObj responseRtmStart
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return
	}

	if !respObj.Ok {
		err = fmt.Errorf("Slack error: %s", respObj.Error)
		return
	}

	wsurl = respObj.Url
	id = respObj.Self.Id
	return
}

// These are the messages read off and written into the websocket. Since this
// struct serves as both read and write, we include the "Id" field which is
// required only for writing.

type Message struct {
	Id      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

func getMessage(ws *websocket.Conn) (m Message, err error) {
	err = websocket.JSON.Receive(ws, &m)
	return
}

var counter uint64

func postMessage(ws *websocket.Conn, m Message) error {
	m.Id = atomic.AddUint64(&counter, 1)
	return websocket.JSON.Send(ws, m)
}

// Starts a websocket-based Real Time API session and return the websocket
// and the ID of the (bot-)user whom the token belongs to.
func slackConnect(token string) (*websocket.Conn, string) {
	wsurl, id, err := slackStart(token)
	if err != nil {
		log.Fatal(err)
	}

	ws, err := websocket.Dial(wsurl, "", "https://api.slack.com/")
	if err != nil {
		log.Fatal(err)
	}

	return ws, id
}

func main() {
	//port := os.Getenv("PORT")
	//
	//if port == "" {
	//	//log.Fatal("$PORT must be set")
	//	// for testing
	//	port = "5001"
	//}

	token := ""
	if len(os.Args) > 1 {
		token = os.Args[1]
	} else {
		fmt.Println("Error no token")
		os.Exit(4)
	}

	//cleanJson()
	_MapStopByID = getStops("./gtfs/stops.json")
	_MapStopIDByName = setMapStopIDByName(_MapStopByID)
	_MapTimesByIDWeekDay, _MapTimesByIDWeekEnd = setMapTimesByID(getStoptimes("./gtfs/stoptimes.json"))

	// start a websocket-based Real Time API session
	ws, id := slackConnect(token)
	fmt.Println("caltrainbot ready, ^C exits")

	for {
		// read each incoming message
		m, err := getMessage(ws)
		if err != nil {
			log.Fatal(err)
		}

		// see if we're mentioned
		if m.Type == "message" && strings.HasPrefix(m.Text, "<@"+id+">") {
			// if so try to parse if
			parts := strings.Fields(m.Text)
			fmt.Println(parts)
			if len(parts) >= 4 && parts[1] == "next" {
				// looks good, get the quote and reply with the result
				go func(m Message) {
					m.Text = SearchNext(parts[2], strings.Join(parts[3:], " "))
					postMessage(ws, m)
				}(m)
				// NOTE: the Message object is copied, this is intentional
			} else {
				// huh?
				m.Text = fmt.Sprintf("sorry, that does not compute\n")
				postMessage(ws, m)
			}
		}
	}

	//router := web.New(Context{}).
	//	Middleware(web.LoggerMiddleware).
	//	Middleware(web.ShowErrorsMiddleware).
	//	NotFound((*Context).NotFound).
	//	Get("/next/:direction/:stop_name", (*Context).FindStop).
	//	Get("/stop/:id", (*Context).GetStopDetails)
	//http.ListenAndServe(":"+port, router)
}

func (c *Context) NotFound(rw web.ResponseWriter, r *web.Request) {
	rw.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(rw, "Not Found")
}

func (c *Context) FindStop(rw web.ResponseWriter, req *web.Request) {
	direction := req.PathParams["direction"]
	if direction != "NB" && direction != "SB" {
		rw.WriteHeader(http.StatusBadRequest)
	}
	stopName := req.PathParams["stop_name"]
	if stopName == "" {
		rw.WriteHeader(http.StatusBadRequest)
	}

	fmt.Fprint(rw, SearchNext(direction, stopName))
}

func SearchNext(dir string, stop string) string {
	hr, min, sec := time.Now().Clock()
	stringTime := strconv.Itoa(hr) + ":" + strconv.Itoa(min) + ":" + strconv.Itoa(sec)

	stopDir := dir + "_" + stop
	stopID := (*_MapStopIDByName)[stopDir]
	var nextTrains []string

	if time.Now().Weekday() == time.Saturday || time.Now().Weekday() == time.Sunday {
		nextTrains = (*_MapTimesByIDWeekEnd)[stopID]
	} else {
		nextTrains = (*_MapTimesByIDWeekDay)[stopID]
	}

	if nextTrains != nil && len(nextTrains) > 0 {
		idx := findTimeIdx(&stringTime, &nextTrains)
		if idx == -1 {
			idx = len(nextTrains) - 1
		}
		if len(nextTrains) > idx+3 {
			nextTrains = nextTrains[idx:idx+3]
		} else if len(nextTrains) > idx { // should be a else only
			nextTrains = nextTrains[idx:]
		}
		var _nextTrains []NextTrain
		for _, h := range nextTrains {
			next := NextTrain{Direction:dir, StopName:stop, Next:h}
			_nextTrains = append(_nextTrains, next)
		}
		b, _ := json.Marshal(_nextTrains)
		return string(b)
	} else {
		return "{}"
	}
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

func setMapTimesByID(stopTimes *[]model.StopTime) (*map[int][]string, *map[int][]string) {
	timesByIDWeekDay := make(map[int][]string)
	timesByIDWeekEnd := make(map[int][]string)
	var _emptyList []string

	// init map
	for k, _ := range *_MapStopByID {
		timesByIDWeekDay[k] = _emptyList
		timesByIDWeekEnd[k] = _emptyList
	}
	for _, stopTime := range *stopTimes {
		if strings.HasPrefix(stopTime.TripID, "8") || strings.HasPrefix(stopTime.TripID, "4") {
			if _, ok := timesByIDWeekEnd[stopTime.StopID]; ok {
				timesByIDWeekEnd[stopTime.StopID] = append(timesByIDWeekEnd[stopTime.StopID], stopTime.DepartureTime)
			}
		} else {
			if _, ok := timesByIDWeekDay[stopTime.StopID]; ok {
				timesByIDWeekDay[stopTime.StopID] = append(timesByIDWeekDay[stopTime.StopID], stopTime.DepartureTime)
			}
		}
	}
	for _, v := range timesByIDWeekDay {
		sort.Strings(v)
	}
	for _, v := range timesByIDWeekEnd {
		sort.Strings(v)
	}
	return &timesByIDWeekDay, &timesByIDWeekEnd
}

func findTimeIdx(time *string, times *[]string) int {
	return sort.Search(len(*times), func(i int) bool { return (*times)[i] >= *time })
}

func (c *Context) GetStopDetails(rw web.ResponseWriter, req *web.Request) {
	_id := req.PathParams["id"]
	if _id == "" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(_id)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Println("error:", err)
	}
	if v, ok := (*_MapStopByID)[id]; ok {
		b, err := json.Marshal(v)
		if err != nil {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Println("error:", err)
			return
		}
		fmt.Fprint(rw, string(b))
	} else {
		rw.WriteHeader(http.StatusNotFound)
	}
}
