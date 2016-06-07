package main

import (
	log "github.com/Sirupsen/logrus"
)

type lookupService struct {
	inc chan lookupRequest
}

type lookupRequest struct {
	query  string
	result chan inbox
}

func startLookupService() lookupService {

	ls := lookupService{}
	ls.inc = make(chan lookupRequest)

	go ls.run()
	return ls
}

func (ls lookupService) run() {
	log.Info("Running lookupService")

	inboxMap := make(map[string]inbox)

	for {
		select {
		case lreq := <-ls.inc:
			log.WithField("query", lreq.query).Debug("lookupService got request")

			// does the inbox exist? If so return it on the channel given
			if inbox, ok := inboxMap[lreq.query]; ok {
				lreq.result <- inbox
			} else {
				inbox := startInbox(lreq.query)
				inboxMap[lreq.query] = inbox
				lreq.result <- inbox
			}
		}
	}

}

// Helper fuction to make it slightly simpler to look up inboxes
// ask the lookup service for the right inbox. It will be created if it does not exist
// FIXME: Should take a pointer, right? Chan's will be copied?
func (ls lookupService) lookup(q string) inbox {

	lreq := lookupRequest{
		query:  q,
		result: make(chan inbox),
	}

	ls.inc <- lreq       // send reques to lookup service, blocking - Does it makes sense to ue a buffered chan here?
	box := <-lreq.result // wait (block) for response

	return box
}
