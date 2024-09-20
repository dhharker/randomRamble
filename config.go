package main

import (
	"flag"
	"log"
	"os"
	"time"
)

type Config struct {
	fileName        string
	captureDuration time.Duration
	port            string
}

func getConfig() *Config {

	c := Config{}

	flag.StringVar(&c.fileName, "filename", "", "Base filename to record data. Leave blank to not save data.")
	flag.DurationVar(&c.captureDuration, "duration", 0, "How long to capture for e.g. 10s. Leave blank to run until closed.")
	flag.StringVar(&c.port, "port", "", "USBTTY Port with TrueRNGpro V2. Leave blank to auto-detect.")

	flag.Parse()

	if c.fileName == "" {
		log.Println("Not saving output to file.")
	} else {
		log.Printf("Writing output to %s.csv", c.fileName)
	}

	if int(c.captureDuration) == 0 {
		log.Println("Will capture indefinitely until program is terminated.")
	} else if c.captureDuration < 0 {
		log.Println("Cannot capture for negative duration.")
		os.Exit(2)
	} else {
		log.Printf("Will capture for %v or until program is terminated.", c.captureDuration)
	}

	if c.port == "" {
		log.Println("Will attempt to auto detect which port the TrueRNGproV2 is connected to.")
	} else {
		log.Printf("Expecting to find a TrueRNGpro V2 on port: %s", c.port)
	}

	return &c
}
