package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// Command is the data structure used for communicating with the client over websocket
type Command struct {
	// A Cms is one of: msg, info
	Cmd string // or another type?
	//Err error
	Msg  *msg  // nil for !msg
	Info *info // nil for !info
}

// msg represents a single message internally and in the Json
type msg struct {
	To   string // Set by the client
	Msg  string // Set by the client
	From string
	Time time.Time
}

// info represents communication betweenthe client and the server, for debuging and state, not for humans
type info struct {
	Success bool
	Reply   bool
	Msg     string
	Time    time.Time
}

// reprsents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	host     string
	templ    *template.Template
}

// One global lookup service
var ls lookupService

func main() {

	log.SetLevel(log.DebugLevel) // Should this be in init() ?

	port := os.Getenv("PORT")
	if port == "" {
		log.WithField("PORT", port).Fatal("$PORT must be set")
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}

	ls = startLookupService()

	r := mux.NewRouter()
	r.HandleFunc("/ws/{user}", handleWebsocket)
	r.HandleFunc("/api/hello", handleAPIHello)
	r.Handle("/{user}", &templateHandler{filename: "chat.html", host: fmt.Sprintf("%s:%s", host, port)}) // show some html and js - A simple chat frontend. Host is used in the JS

	fmt.Println("Listening on port", port, "..")

	err := http.ListenAndServe(":"+port, r)

	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	vars := mux.Vars(r)

	//headers := r.Header
	//log.Printf("%#v\n", headers)
	log.WithField("user", vars["user"]).WithField("host", r.Header.Get("Referer")).Info("User chatting")

	t.templ.Execute(w, struct {
		Host string
		User string
	}{t.host, vars["user"]})
}

func handleAPIHello(w http.ResponseWriter, r *http.Request) {
	//room1.toAll <- message{Msg: []byte("Hello From the API =)")}
	log.Fatal("API hello not defined")
}
