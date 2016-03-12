package main

import "log"

type inbox struct {
	clientName string
	messages   chan message
}

func startInbox(clientName string) inbox {
	log.Println("Starting new inbox for:", clientName)

	inbox := inbox{
		clientName: clientName,
		messages:   make(chan message, 20),
	}

	go inbox.run()
	return inbox
}

func (i inbox) run() {

	//for {
	//	select {
	//	case inc := <-i.incoming:
	//		log.Println("inbox", i.clientName, "got message:", inc.Msg)
	//	}
	//}
}
