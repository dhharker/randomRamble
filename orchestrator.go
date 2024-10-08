package main

import (
	"log"
	"sync"
)

type Orchestrator struct {
	wg                  sync.WaitGroup
	shutdownOnCloseChan chan struct{}
	shutdownRequested   bool
	syncSDC             []func()
}

func (ro *Orchestrator) shutdown(reason string) {
	if !ro.shutdownRequested {
		log.Printf("Orchestrator:Shutdown %v", reason)
		// Iterate syncSDC functions (block until they've all run)
		for _, f := range ro.syncSDC {
			f()
		}
		ro.shutdownRequested = true
		close(ro.shutdownOnCloseChan)
	} else {
		log.Printf("Orchestrator:Shutdown AGAIN %v", reason)
	}
	ro.wg.Wait()
}

func (ro *Orchestrator) onShutdown(fn func()) {
	ro.syncSDC = append(ro.syncSDC, fn)
}

func newOrchestrator() *Orchestrator {
	ro := &Orchestrator{}
	ro.shutdownOnCloseChan = make(chan struct{})
	return ro
}
