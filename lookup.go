package main

import "log"

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
	log.Println("Running lookupService")

	inboxMap := make(map[string]inbox)

	for {
		select {
		case lreq := <-ls.inc:
			log.Println("lookupService got request for:", lreq.query)

			// does the inbox exist? If so return it on the channel given
			if inbox,ok := inboxMap[lreq.query]; ok {
				lreq.result <- inbox
			} else {
				inbox := startInbox(lreq.query)
                inboxMap[ lreq.query ] = inbox
				lreq.result <- inbox
			}
		}
	}

}
