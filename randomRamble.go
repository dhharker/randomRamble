package main

import (
	"log"
	"os"
	"runtime"
	"time"
)

// Note: don't put this too high; 256 resulted in a bunch of null reads which screws up the stats
const READ_BUFFER_SIZE int = 128
const BIN_VALS_SIZE = READ_BUFFER_SIZE / 4 // 2 bytes * 2 RNGs

type Sample struct {
	// Incremented every time READ_BUFFER_SIZE bytes are read from the RNG into a Sample
	sampleCount uint64
	// Time that sampling was initiated
	sampleTime time.Time
	// is it a WhiteSample or a RawSample?
	sampleType RngMode

	// Data read from RNG over serial port
	sample []byte

	// +/- sum of zeroes (-1) and 1s (+1) in each value
	walkDeltas [READ_BUFFER_SIZE]int8
	// Sum of walkDeltas
	walkSum int64
	// Entropy of values
	entropy float64

	// Used in raw mode where we have a stream of pairs of normally distributed 10-bit ADC readings, one from each RNG
	// Array of tuples containing the raw ADC reading from generators A and B
	rawValues [BIN_VALS_SIZE][2]uint16
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

	// Get Samples from RNG
	samplesChan := getSamples(config, ro)

	if config.showGui {
		ro.wg.Add(1)
		_, w := gui(samplesChan, ro)
		// Blocks until the GUI quits
		w.ShowAndRun()
		ro.wg.Done()
	} else {
		go doDisplay(samplesChan, ro)
		time.Sleep(1 * time.Second)
	}

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
		} else if err := modeChange(config.mode, tpv2.Name); err != nil {
			log.Fatalf("Error setting mode: %v", err)
		}

		// // Connect to RNG
		rng := getConnected(tpv2.Name)

		// Get ready to read data
		signalReadSerialChan := make(chan time.Time)
		rngSerialOutputChan := make(chan *Sample)
		go readSerialOnDemand(rng, config.mode, rngSerialOutputChan, signalReadSerialChan, ro)

		if config.mode == RngRawMode {
			// In RAWBIN mode we must parse the binary format to get the raw values for ADCs A and B
			parsedBinaryFormatChan := make(chan *Sample)
			go doParseBinaryFormat(rngSerialOutputChan, parsedBinaryFormatChan, ro)
			go doRawStats(config, parsedBinaryFormatChan, processedSamplesChan, ro)
		} else {
			// Process data and send to output chan (passed from main)
			go doWhiteStats(rngSerialOutputChan, processedSamplesChan, ro)
		}

		// Ticker to trigger read data at intervals
		ticker := time.NewTicker(time.Duration(config.tickDelayMs) * time.Millisecond)
		go demandSerialReadOnTick(ticker, signalReadSerialChan, ro)

		// Shut down after run duration, if specified
		go shutdownAfterDelay(config.captureDuration, ro)

	}(config, ro, processedSamplesChan)
	return processedSamplesChan
}
