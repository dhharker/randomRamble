package main

import (
	"fmt"
	"log"
	"time"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

var RngModeStrings = map[RngMode]string{
	RngRawMode:   "MODE_RAW_BIN",
	RngWhiteMode: "MODE_RNG1WHITE",
}

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
func modeChange(rngMode RngMode, PORT string) error {
	var MODE = RngModeStrings[rngMode]

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
	// log.Printf("   Knock once...")
	if err := knockSequence(110); err != nil {
		return err
	}
	// log.Printf("   Knock twice...")
	if err := knockSequence(300); err != nil {
		return err
	}
	// log.Printf("   Knock thrice...")
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
	// log.Printf("   Fourth and final knock...")
	port, err := serial.Open(PORT, mode)
	if err != nil {
		return err
	}
	port.Close()

	// log.Printf("   Done knocking!")
	return nil
}

// Connect to RNG ready to read data
func getConnected(portName string) serial.Port {

	// Connect to RNG
	mode := &serial.Mode{
		BaudRate: 9600,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		log.Fatalf("Failed to open serial port '%v': %v", portName, err)
	}

	// RNG sends data once DTR true
	port.SetDTR(true)
	// If it goes south, die
	port.SetReadTimeout(3 * time.Second)

	return port
}

// goroutine to read samples from RNG
func readSerialOnDemand(port serial.Port, sampleType RngMode, readChan chan *Sample, signalChan chan time.Time, ro *Orchestrator) {
	ro.wg.Add(1)
	var sampleCounter uint64 = 0

	for {
		var readTime time.Time

		select {
		case _, isFalse := <-ro.shutdownOnCloseChan:
			log.Printf("Orchestrator:Shutting down readSerialOnDemand")
			if isFalse {
				panic("readSerialOnDemand panic")
			}
			log.Printf("Stopping serial reader after %v samples * %v bytes = %v bytes total", sampleCounter, CAPTURE_SAMPLE_BYTES, int(sampleCounter)*CAPTURE_SAMPLE_BYTES)
			pcErr := port.Close()
			log.Printf("Closed port %v", pcErr)
			close(readChan)
			log.Printf("Closed readchan")
			ro.wg.Done()
			return
		case readTime = <-signalChan:

			if ro.shutdownRequested {
				return
			}

			// Create a readBuffer to hold incoming data
			readBuffer := make([]byte, 128)
			sampleBuffer := make([]byte, CAPTURE_SAMPLE_BYTES)
			sampleSlice := sampleBuffer[:0]

			// Flush the buffer first so we're getting freshest data
			port.ResetInputBuffer()

			// Keep track of total bytes read until we have enough
			bytesRead := 0

			// Attempt to read from the serial port
			for bytesRead < CAPTURE_SAMPLE_BYTES {
				n, err := port.Read(readBuffer)
				if err != nil {
					log.Println("Error reading from serial:", err)
					continue
				}
				sampleSlice = append(sampleSlice, readBuffer[0:n]...)
				bytesRead += n

			}

			sampleCounter++
			spl := &Sample{
				sampleCount: sampleCounter,
				sampleTime:  readTime,
				sampleType:  sampleType,
				sample:      sampleSlice[:CAPTURE_SAMPLE_BYTES],
				walkSum:     0,
			}

			// Send the data to the main program via readChan
			if !ro.shutdownRequested {
				readChan <- spl
			}
		}
	}
}
