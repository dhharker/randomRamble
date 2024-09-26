package main

import "log"

func parseRawBinaryFormat(binSampleChan chan *Sample, parsedSampleChan chan *Sample, ro *Orchestrator) {

	ro.wg.Add(1)
	for {
		select {
		case _, isFalse := <-ro.shutdownOnCloseChan:
			log.Printf("Orchestrator:Shutting down parseRawBinaryFormat")
			if isFalse {
				panic("parseRawBinaryFormat panic")
			}
			ro.wg.Done()
			return
		case spl := <-binSampleChan:
			// parse
			if !ro.shutdownRequested {
				parsedSampleChan <- spl
			}
		}
	}
}
