package main

import (
	"flag"
	"log"
	"time"
)

type Config struct {
	fileName        string
	captureDuration time.Duration
	tickDelayMs     int
	port            string
	skipModeset     bool
	showGui         bool
}

func getConfig() *Config {

	c := Config{}

	flag.StringVar(&c.fileName, "filename", "", "Base filename to record data. Leave blank to not save data.")
	flag.DurationVar(&c.captureDuration, "duration", 0, "How long to capture for e.g. 10s. Leave blank to run until closed.")
	flag.IntVar(&c.tickDelayMs, "rate", 1000, "Delay between samples in ms e.g. 100 to sample at 10Hz.")
	flag.StringVar(&c.port, "port", "", "USBTTY Port with TrueRNGpro V2. Leave blank to auto-detect.")
	flag.BoolVar(&c.skipModeset, "skipmodeset", false, "DANGER skip modeset on init. Modeset is slow. Only use if mode already set.")
	flag.BoolVar(&c.showGui, "gui", false, "Show gui")

	flag.Parse()

	if c.fileName == "" {
		log.Println("Not saving output to file.")
	} else {
		log.Printf("Writing output to %s.csv", c.fileName)
	}

	if int(c.captureDuration) == 0 {
		log.Println("Will capture indefinitely until program is terminated.")
	} else if c.captureDuration < 0 {
		log.Fatal("Cannot capture for negative duration.")
	} else if c.captureDuration <= time.Duration(c.tickDelayMs*int(time.Millisecond)) {
		log.Fatal("Capture duration must exceed capture rate.")
	} else {
		log.Printf("Will capture for %v or until program is terminated.", c.captureDuration)
	}

	if c.tickDelayMs < 1 {
		log.Fatal("Min sample rate 10ms")
	}
	log.Printf("Will sample every %vms (~%2.1fHz)", c.tickDelayMs, float64(1000/c.tickDelayMs))

	if c.port == "" {
		log.Println("Will attempt to auto detect which port the TrueRNGproV2 is connected to.")
	} else {
		log.Printf("Expecting to find a TrueRNGpro V2 on port: %s", c.port)
	}

	return &c
}
