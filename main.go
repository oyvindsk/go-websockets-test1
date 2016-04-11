package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Command is the data structure used for communicating with the client over websocket
type Command struct {
	Cmd string // or another type?
	//Err error
	Msg *message
}

// message represents a single message internally and in the Json
type message struct {
	From string
	To   string
	Msg  string
	Time time.Time
}

// One global lookup service
var ls lookupService

func main() {

	log.SetLevel(log.DebugLevel) // Should this be in init() ?

	port := os.Getenv("PORT")
	if port == "" {
		log.WithField("PORT", port).Fatal("$PORT must be set")
	}

	ls = startLookupService()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws/", handleWebsocket)
	mux.HandleFunc("/api/hello", handleAPIHello)

	fmt.Println("Listening on port", port, "..")

	err := http.ListenAndServe(":"+port, mux)

	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}

func handleAPIHello(w http.ResponseWriter, r *http.Request) {
	//room1.toAll <- message{Msg: []byte("Hello From the API =)")}
	log.Fatal("API hello not defined")
}
