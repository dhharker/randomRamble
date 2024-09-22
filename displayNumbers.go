package main

import "log"

var walkSum int64 = 0

func doDisplay(numbersChan chan *Sample, stopDisplayChan chan bool) {

	for {
		select {
		case <-stopDisplayChan:
			return
			// case <-numbersChan:
		case spl := <-numbersChan:
			walkSum += spl.walkSum
			log.Printf("#%v %2.3f  %v", spl.sampleCount, spl.entropy, walkSum)
		}
	}
}
