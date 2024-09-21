package main

import (
	"log"
	"os"
	"runtime"
	"time"
)

const READ_BUFFER_SIZE int = 64
const USE_RNG_MODE = "MODE_RNG1WHITE"

type Sample struct {
	sampleCount uint64
	sampleTime  time.Time
	values      []byte
	walkDeltas  [READ_BUFFER_SIZE]int8
	walkSum     int64
	entropy     float64
}

func main() {
	log.Println("Random Ramble")

	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		log.Printf("RandomRamble developed on linux, might work on mac, windows not supported.")
		os.Exit(1)
	}

	// Parse arguments/env return config object
	config := getConfig()

	// Find USB serial ports with a TrueRNGpro V2 attached
	tpv2 := findTPV2Port(config.port)

	// Set TrueRNG mode to RAWBIN using port-knocking protocol
	if config.skipModeset {
		log.Printf("DANGER - skipping modeset.")
	} else if err := modeChange(USE_RNG_MODE, tpv2.Name); err != nil {
		log.Fatalf("Error setting mode: %v", err)
	}

	// // Connect to rng
	rng := getConnected(tpv2.Name)
	defer rng.Close()

	// Get ready to read data
	signalReadSerialChan := make(chan time.Time)
	defer close(signalReadSerialChan)
	stopReaderChan := make(chan bool)
	defer close(stopReaderChan)
	rawDataChan := make(chan *Sample)
	defer close(rawDataChan)
	go readSerialOnDemand(rng, rawDataChan, signalReadSerialChan, stopReaderChan)

	// Process data
	stopMathChan := make(chan bool)
	defer close(stopMathChan)
	processedSamplesChan := make(chan *Sample)
	defer close(processedSamplesChan)
	go doMath(rawDataChan, processedSamplesChan, stopMathChan)

	// Display the data to the user somehow
	stopDisplayChan := make(chan bool)
	defer close(stopDisplayChan)
	go doDisplay(processedSamplesChan, stopDisplayChan)

	// Ticker to read data at intervals
	stopTickerChan := make(chan bool)
	defer close(stopTickerChan)
	ticker := time.NewTicker(time.Duration(config.tickDelayMs) * time.Millisecond)
	go demandSerialReadOnTick(ticker, signalReadSerialChan, stopTickerChan)

	// As soon as anything writes to this, we shut down.
	shutdownCtlChan := make(chan bool)
	defer close(shutdownCtlChan)

	// Shut down after run duration, if specified
	go shutdownAfterDelay(config.captureDuration, shutdownCtlChan)

	// Shut down on SIGINT or SIGKILL
	go shutdownOnSignal(shutdownCtlChan)

	// block until someone writes to the shutdown chan
	<-shutdownCtlChan

	// CLEAN UP AND SHUT DOWN
	log.Println("Shutting down...")
	stopTickerChan <- true
	stopReaderChan <- true
	stopMathChan <- true
	stopDisplayChan <- true
	rng.Close()
	log.Printf("Bye!")

}
