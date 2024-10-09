package main

import "log"

var walkSum int64 = 0

func doDisplay(samplesChan chan *Sample, ro *Orchestrator) {
	ro.wg.Add(1)
	for {
		select {
		case _, isFalse := <-ro.shutdownOnCloseChan:
			log.Printf("Orchestrator:Shutting down doDisplay")
			if isFalse {
				panic("doDisplay panic")
			}
			ro.wg.Done()
			return
			// case <-numbersChan:
		case spl := <-samplesChan:
			walkSum += spl.walkSum
			log.Printf("#% 4.0f\t%2.3f\t% 7.0f", float32(spl.sampleCount), spl.entropy, float32(walkSum))
		}
	}
}
