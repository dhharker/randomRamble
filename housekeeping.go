package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func shutdownAfterDelay(delay time.Duration, shutdownCtlChan chan bool) {
	if delay > 0 {
		time.Sleep(delay)
		log.Printf("Run duration of %v elapsed. Shutting down.", delay)
		shutdownCtlChan <- true
	}
}

func demandSerialReadOnTick(t *time.Ticker, signalReadSerialChan chan time.Time, stopTickerChan chan bool) {
	for {
		select {
		case <-stopTickerChan:
			// log.Println("Stopping ticker...")
			t.Stop()
			return
		// interval task
		case tm := <-t.C:
			// log.Println("Tick: ", tm)
			signalReadSerialChan <- tm
		}
	}
}

func shutdownOnSignal(shutdownCtlChan chan bool) {
	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	log.Printf("Caught signal %v shutting down", sig)
	shutdownCtlChan <- true
}
