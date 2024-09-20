package main

import (
	"fmt"
	"log"
	"time"

	"go.bug.st/serial"
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

// Function to change mode by setting different baud rates
func modeChange(MODE string, PORT string) error {
	log.Printf("Setting RNG to %s...", MODE)

	// Function to open and close the port at specific baud rate
	knockSequence := func(baud int) error {
		mode := &serial.Mode{
			BaudRate: baud,
		}
		port, err := serial.Open(PORT, mode)
		if err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
		return port.Close()
	}

	// Knock sequence to activate mode change
	if err := knockSequence(110); err != nil {
		return err
	}
	if err := knockSequence(300); err != nil {
		return err
	}
	if err := knockSequence(110); err != nil {
		return err
	}

	// Setting the baud rate based on the MODE input
	var baud int
	switch MODE {
	case "MODE_NORMAL":
		baud = 300
	case "MODE_PSDEBUG":
		baud = 1200
	case "MODE_RNGDEBUG":
		baud = 2400
	case "MODE_RNG1WHITE":
		baud = 4800
	case "MODE_RNG2WHITE":
		baud = 9600
	case "MODE_RAW_BIN":
		baud = 19200
	case "MODE_RAW_ASC":
		baud = 38400
	case "MODE_UNWHITENED":
		baud = 57600
	case "MODE_NORMAL_ASC":
		baud = 115200
	case "MODE_NORMAL_ASC_SLOW":
		baud = 230400
	default:
		return fmt.Errorf("invalid mode: %v", MODE)
	}

	// Final mode change
	mode := &serial.Mode{
		BaudRate: baud,
	}
	port, err := serial.Open(PORT, mode)
	if err != nil {
		return err
	}
	defer port.Close()

	return nil
}
