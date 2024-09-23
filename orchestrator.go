package main

import (
	"log"
	"sync"
)

type Orchestrator struct {
	wg                  sync.WaitGroup
	shutdownOnCloseChan chan struct{}
	shutdownRequested   bool
}

func (ro *Orchestrator) shutdown(reason string) {
	if !ro.shutdownRequested {
		log.Printf("Orchestrator:Shutdown %v", reason)
		ro.shutdownRequested = true
		close(ro.shutdownOnCloseChan)
	} else {
		log.Printf("Orchestrator:Shutdown AGAIN %v", reason)
	}
	ro.wg.Wait()
}

func newOrchestrator() *Orchestrator {
	ro := &Orchestrator{}
	ro.shutdownOnCloseChan = make(chan struct{})
	return ro
}
