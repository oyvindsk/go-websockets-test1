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
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithField("err", err).Println("Upgrading to websockets")
		http.Error(w, "Error Upgrading to websockets", 400)
		return
	}

	myClientName := path.Base(r.URL.Path)
	log.Println("New connection from client:", myClientName)

	// OK websokset connection. Get or create our inbox (a running go routine)
	lreq := lookupRequest{
		query:  myClientName,
		result: make(chan inbox),
	}

	ls.inc <- lreq           // send reques to lookup service, blocking
	myInbox := <-lreq.result // wait (block) for response
	log.Println("Got my inbox:", myInbox.clientName)

	// Hook this ws conection up to the inbox so we get future messages

	// Get pending messages and send them to the client
GetMessages:
	for {
		select {
		case msg := <-myInbox.messages:
			log.Println("stored msg for", myInbox.clientName, msg.Msg)
		default:
			log.Println("No stored msg for", myInbox.clientName)
			break GetMessages
		}
	}

	// Take messages from the client and send them to X
	//      lookup to get channle
	//      put message in channel

	// Take messages from whoever and send them to the client
	//      connect out inbox channle to the websocket

	//room1.join <- ws

	for {
		mt, data, err := ws.ReadMessage()
		ctx := log.Fields{"mt": mt, "data": string(data), "err": err}
		if err != nil {
			if err == io.EOF {
				log.WithFields(ctx).Info("Websocket closed!")
			} else {
				log.WithFields(ctx).Error("Error reading websocket message")
			}
			break
		}
		switch mt {
		case websocket.TextMessage:
			log.WithFields(ctx).Println("Read from ws")

			// Decode message
			// FIXME
			msg := message{
				From: "ole",
				To:   "olga",
				Msg:  "hello hello hello!",
				Time: time.Now(),
			}
			log.Println("msg:", msg)

			// Lookup who it's for
			lreq = lookupRequest{
				query:  msg.To,
				result: make(chan inbox),
			}

			ls.inc <- lreq      // send reques to lookup service, blocking
			to := <-lreq.result // wait (block) for response
			log.Println("got inbox:", to.clientName)

			// Deliver to channel
			log.Println("sending message from:", msg.From, "==>", msg.To)
			to.messages <- msg

			// FIXME : don't block when the abosve channel is full

			ws.WriteMessage(mt, []byte("ok"))
		default:
			log.WithFields(ctx).Warning("Unknown Message!")
		}
	}

}
