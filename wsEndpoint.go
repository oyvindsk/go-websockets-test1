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

		// Parse the JSON commands from the client
		cmd := Command{}
		err := ws.ReadJSON(&cmd) // Blocks!
		if err != nil {
			if err == io.EOF {
				log.Info("Websocket closed!")
				// FIXME - WHat to do? Does this even work with ReadJSON?
			} else {
				log.WithField("client", myClientName).Error("Could not parse JSON:", err)
				myInbox.commands <- Command{Cmd: "Error =(", Msg: &message{Msg: "Err"}}
			}
			continue
		}

		log.WithField("cmd", cmd).Debug("Read from ws")

		switch cmd.Cmd {
		case "msg":
			if cmd.Msg == nil {
				log.WithFields(log.Fields{"cmd": cmd.Cmd, "client": myClientName}).Warning("Msg er NIL")
				myInbox.commands <- Command{Cmd: "Error =(", Msg: &message{Msg: "Command msg needs Msg"}}
			} else {
				log.Printf("Got MSG %+v", cmd.Msg)
				cmd.Msg.From = myClientName
				cmd.Msg.Time = time.Now()
				//FIXME sjekk To
			}
		default:
			log.WithField("cmd", cmd.Cmd).Warning("Unknow CMD from client")
		}

		log.WithFields(log.Fields{"from": cmd.Msg.From, "to": cmd.Msg.To, "msg": cmd.Msg.Msg, "time": cmd.Msg.Time}).Debug("Message for delivery")

		// Lookup who it's for
		lreq = lookupRequest{
			query:  cmd.Msg.To,
			result: make(chan inbox),
		}

		ls.inc <- lreq      // send reques to lookup service, blocking
		to := <-lreq.result // wait (block) for response
		log.WithField("inbox", to.clientName).Debug("Got inbox for delivery")

		// Deliver to channel
		log.WithFields(log.Fields{"from": cmd.Msg.From, "to": cmd.Msg.To}).Debug("sending message from:", cmd.Msg.From, "==>", cmd.Msg.To)
		//to.messages <- msg
		to.commands <- cmd

		// FIXME : don't block when the abosve channel is full

		// Cant write to the ws in this go routine - ws.WriteMessage(mt, []byte("ok"))
		//myInbox.messages <- message{From: myClientName, To: myClientName, Msg: "OK", Time: time.Now()}
		myInbox.commands <- Command{Msg: &message{From: myClientName, To: myClientName, Msg: "OK", Time: time.Now()}}
	}

}
