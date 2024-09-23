package main

import "log"

var walkSum int64 = 0

func doDisplay(numbersChan chan *Sample, ro *Orchestrator) {
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
		case spl := <-numbersChan:
			walkSum += spl.walkSum
			log.Printf("#%v %2.3f  %v", spl.sampleCount, spl.entropy, walkSum)
		}
	}
}
