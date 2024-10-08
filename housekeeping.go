package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func shutdownAfterDelay(delay time.Duration, ro *Orchestrator) {
	if delay > 0 {
		time.Sleep(delay)
		ro.shutdown(fmt.Sprintf("shutdownAfterDelay %v", delay))
	}
}

func demandSerialReadOnTick(t *time.Ticker, signalReadSerialChan chan time.Time, ro *Orchestrator) {
	ro.wg.Add(1)
	// Make sure we do at least one read
	signalReadSerialChan <- time.Now()

	// On shutdown, stop the ticker and give all the other modules time to finish whatever they're doing
	// If we don't drain the chain then we get race conditions and app fails to quit
	// @TODO this could probably be made nicer by making all the shutdown stuff synchronous
	ro.onShutdown(func() {
		
		log.Printf("Orchestrator:Sync Stop ticker and wait")
		t.Stop()
		time.Sleep(1*time.Second)
		log.Printf("Orchestrator:Sync    (continue)")
	})
	for {
		select {
		case _, isFalse := <-ro.shutdownOnCloseChan:
			log.Printf("Orchestrator:Shutting down demandSerialReadOnTick")
			if isFalse {
				panic("demandSerialReadOnTick panic")
			}
			t.Stop()
			close(signalReadSerialChan)
			ro.wg.Done()
			return
		case tm := <-t.C:
			// log.Println("Tick: ", tm)
			if !ro.shutdownRequested {
				signalReadSerialChan <- tm
			}
		}
	}
}

func shutdownOnSignal(ro *Orchestrator) {
	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	ro.shutdown(fmt.Sprintf("shutdownOnSignal %v", sig))
}
