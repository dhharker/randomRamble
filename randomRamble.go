package main

import (
	"log"
	"os"
	"runtime"
	"time"
)

const READ_BUFFER_SIZE int = 128
const MAX_PAIRS_RB = READ_BUFFER_SIZE / 4

type Sample struct {
	sampleCount uint64
	sampleTime  time.Time
	rawData     []byte
	values      [MAX_PAIRS_RB][2]int16
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
	} else if err := modeChange("MODE_RAW_BIN", tpv2.Name); err != nil {
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

	// Parse raw data into a Sample
	stopParserChan := make(chan bool)
	defer close(stopParserChan)
	sampleChan := make(chan *Sample)
	defer close(sampleChan)
	go doParsing(rawDataChan, sampleChan, stopParserChan)

	// Get ready to process data
	stopMathChan := make(chan bool)
	defer close(stopMathChan)
	numbersChan := make(chan float64)
	defer close(numbersChan)
	go doMath(sampleChan, numbersChan, stopMathChan)

	// Display the data to the user somehow
	stopDisplayChan := make(chan bool)
	defer close(stopDisplayChan)
	go doDisplay(numbersChan, stopDisplayChan)

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
	stopParserChan <- true
	stopMathChan <- true
	stopDisplayChan <- true
	rng.Close()
	log.Printf("Bye!")

}
