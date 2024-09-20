package main

import (
	"log"
	"os"
	"runtime"
)

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
	log.Printf("Setting RNG to raw binary mode...")
	modeChange("MODE_RAW_BIN", tpv2.Name)

	// Main Loop
	// Parse RAWBIN format and return random 10-bit numbers A and B
	// Read data from TrueRNG (discarding first 64 bytes if >threshold time since last read)
	// (1) Integrate over a fixed window of a second or two
	// (2) Integrate over a longer window with more recent values given a higher weighting
	// (3) Integrate over n windows something like Wolf's app for the Wyrdoscope works
	// Realtime display of results
	// Probably use fyne https://fyne.io/blog/2019/03/19/building-cross-platform-gui.html for GUI
	// 1+ histograms for (1),(2) above?
	// A fun shortest-window or weighted-window feedback e.g. tone, colour, size
	// Gamification???
	// Do this via a webserver so people can use the RNG remotely

}
