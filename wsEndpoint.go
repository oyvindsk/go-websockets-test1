package main

import (
	"io"
	"net/http"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true }, // FIXME : Remove
	}
)

func handleWebsocket(w http.ResponseWriter, r *http.Request) {

	// Deny all but HTTP GET
	if r.Method != "GET" {
		log.WithField("method", r.Method).Error("Disallowed http method")
		http.Error(w, "Method not allowed", 405)
		return
	}

	// Upgrade connection to Websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Upgrading to websockets failed:", err)
		http.Error(w, "Error Upgrading to websockets", 400)
		return
	}

	// Get clientname from the url
	myClientName := path.Base(r.URL.Path)
	log.WithFields(log.Fields{"client": myClientName, "subprotocol": ws.Subprotocol()}).Info("New connection from client")

	// This is an ok websokset connection. Get or create our inbox: the place our messages are stored and a go routine to write those messages to the websocket
	// First, ask the lookup service for the right inbox. It will be created if it does not exist
	lreq := lookupRequest{
		query:  myClientName,
		result: make(chan inbox),
	}

	ls.inc <- lreq           // send reques to lookup service, blocking - Does it makes sense to ue a buffered chan here?
	myInbox := <-lreq.result // wait (block) for response

	// Hook this ws conection up to the inbox so we get future messages
	go myInbox.deliverTo(ws) // start a new go routine to run and deliver queued and future messages to this websocket.

	log.WithField("inbox", myInbox.clientName).Info("Got my inbox")

	// Take messages from the client and send them to X
	//      lookup to get channel
	//      put message in channel
	for {
		mt, data, err := ws.ReadMessage()
		if err != nil {
			if err == io.EOF {
				log.Info("Websocket closed!")
			} else {
				log.Error("Error reading websocket message:", err)
			}
			break
		}

		if mt != websocket.TextMessage {
			log.WithField("mt", mt).Warn("Unknown Message! (Probably binary)")
			continue
		}

		log.WithField("msg", data).Debug("Read from ws")

		// Decode message
		// FIXME
		msg := message{
			From: myClientName,
			To:   "olga",
			Msg:  "hello hello hello!",
			Time: time.Now(),
		}
		log.WithFields(log.Fields{"from": msg.From, "to": msg.To, "msg": msg.Msg, "time": msg.Time}).Debug("Message for delivery")

		// Lookup who it's for
		lreq = lookupRequest{
			query:  msg.To,
			result: make(chan inbox),
		}

		ls.inc <- lreq      // send reques to lookup service, blocking
		to := <-lreq.result // wait (block) for response
		log.WithField("inbox", to.clientName).Debug("Got inbox for delivery")

		// Deliver to channel
		log.WithFields(log.Fields{"from": msg.From, "to": msg.To}).Debug("sending message from:", msg.From, "==>", msg.To)
		to.messages <- msg

		// FIXME : don't block when the abosve channel is full

		// Cant write to the ws in this go routine - ws.WriteMessage(mt, []byte("ok"))
		myInbox.messages <- message{From: myClientName, To: myClientName, Msg: "OK", Time: time.Now()}
	}

}
