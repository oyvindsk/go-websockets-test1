package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

type inbox struct {
	clientName string
	//messages   chan message
	commands chan Command
	wsConn   *websocket.Conn
}

func startInbox(clientName string) inbox {
	log.WithField("client", clientName).Info("Starting new inbox")

	inbox := inbox{
		clientName: clientName,
		//messages:   make(chan message, 200),
		commands: make(chan Command, 200),
	}

	//go inbox.run()
	return inbox
}

// Take messages from whoever and send them to the client
func (i inbox) deliverTo(ws *websocket.Conn) {

	log.WithField("client", i.clientName).Info("Writing messages to ws connection")
	i.wsConn = ws

	for inc := range i.commands {
		log.WithFields(log.Fields{"client": i.clientName, "msg": inc.Msg}).Debug("Inbox got message")
		var err error
		switch inc.Cmd {
		case "msg":
			err = ws.WriteJSON(inc)
		case "info":
			err = ws.WriteJSON(inc)
		default:
			log.WithFields(log.Fields{"client": i.clientName, "Cmd": inc.Cmd, "msg": inc.Msg}).Error("Unknown CMD")
		}
		if err != nil {
			log.WithField("client", i.clientName).Error("error writing message to ws:", err)
			break
			// todo: disconnect client?
		}
	}

	//for {
	//	select {
	//	case inc := <-i.incoming:
	//		log.Println("inbox", i.clientName, "got message:", inc.Msg)
	//	}
	//}
}
