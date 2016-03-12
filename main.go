package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
)

// messae represents a single message
type message struct {
	From string
	To   string
	Msg  string
	Time time.Time
}

// One global room, for now
//var room1 *room

var ls lookupService

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		log.WithField("PORT", port).Fatal("$PORT must be set")
	}

	//room1 = startRoom()
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
