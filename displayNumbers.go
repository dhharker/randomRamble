package main

import "log"

func doDisplay(numbersChan chan *Sample, stopDisplayChan chan bool) {

	for {
		select {
		case <-stopDisplayChan:
			return
			// case <-numbersChan:
		case spl := <-numbersChan:
			log.Printf("#%v %2.3f", spl.sampleCount, spl.entropy)
		}
	}
}
