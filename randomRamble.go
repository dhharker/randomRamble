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

	ro := newOrchestrator()
	// Shut down on SIGINT or SIGKILL
	go shutdownOnSignal(ro)

	samplesChan := getSamples(config, ro)
	go doDisplay(samplesChan, ro)

	time.Sleep(1 * time.Second)
	ro.wg.Wait()
}

func getSamples(config *Config, ro *Orchestrator) chan *Sample {
	processedSamplesChan := make(chan *Sample)
	// defer close(processedSamplesChan)
	go func(config *Config, ro *Orchestrator, processedSamplesChan chan *Sample) {

		// Find USB serial ports with a TrueRNGpro V2 attached
		tpv2 := findTPV2Port(config.port)

		// Set TrueRNG mode to RAWBIN using port-knocking protocol
		if config.skipModeset {
			log.Printf("DANGER - skipping modeset.")
		} else if err := modeChange(USE_RNG_MODE, tpv2.Name); err != nil {
			log.Fatalf("Error setting mode: %v", err)
		}

		// // Connect to RNG
		rng := getConnected(tpv2.Name)

		// Get ready to read data
		signalReadSerialChan := make(chan time.Time)
		rawDataChan := make(chan *Sample)
		go readSerialOnDemand(rng, rawDataChan, signalReadSerialChan, ro)

		// Process data and send to output chan (passed from main)
		go doMath(rawDataChan, processedSamplesChan, ro)

		// Ticker to read data at intervals
		ticker := time.NewTicker(time.Duration(config.tickDelayMs) * time.Millisecond)
		go demandSerialReadOnTick(ticker, signalReadSerialChan, ro)

		// Shut down after run duration, if specified
		go shutdownAfterDelay(config.captureDuration, ro)

	}(config, ro, processedSamplesChan)
	return processedSamplesChan
}
