package main

func doDisplay(numbersChan chan *Sample, stopDisplayChan chan bool) {

	for {
		select {
		case <-stopDisplayChan:
			return
		case <-numbersChan:
			// case spl := <-numbersChan:
			// log.Printf("#%v %v", spl.sampleCount, spl.walkSums.sumEq)
		}
	}
}
