package main

import (
	"log"

	"go.bug.st/serial/enumerator"
)

// Find USB serial ports with a TrueRNGpro V2 attached
// portName optional ignore ports other than this one
func findTPV2Port(portName string) *enumerator.PortDetails {

	ports, err := enumerator.GetDetailedPortsList()
	var retPorts []*enumerator.PortDetails

	if err != nil {
		log.Fatal(err)
	}

	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}

	if portName != "" {
		log.Printf("Searching for TrueRNGpro V2 device only on port %s", portName)
	}

	for _, port := range ports {
		if port.IsUSB && port.VID == "04d8" && port.PID == "ebb5" {
			if portName == "" || (portName != "" && port.Name == portName) {
				log.Printf("Found TrueRNGpro V2 on port: %s\n", port.Name)
				retPorts = append(retPorts, port)
			}
		}
	}

	if len(retPorts) == 0 {
		log.Fatal("No device.")
	}

	if len(retPorts) > 1 {
		log.Fatal("More than one TrueRNGpro V2 device detected and no port specified.\nSpecify port name with -port.")
	}

	return retPorts[0]
}
