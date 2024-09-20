package main

// goroutine to take a sample and do some maths on it
func doMath(sampleChan chan *Sample, numbersChan chan float64, stopMathChan chan bool) {
	var agSum int16
	for {
		select {
		case <-stopMathChan:
			return
		case spl := <-sampleChan:
			// log.Printf("Sample #%v %v", spl.sampleCount, spl.values)
			// log.Printf("Deviances %v", deviances(spl))
			// log.Printf("Agreeances %v", agreeances(deviances(spl)))
			agSum = 0
			ags := agreeances(deviances(spl))
			// log.Printf("%v", ags)
			for _, a := range ags {
				agSum += a
			}
			numbersChan <- float64(agSum) / float64(len(ags))
		}
	}
}

type Deviances [MAX_PAIRS_RB][2]int16
type Agreeances [MAX_PAIRS_RB]int16

func deviances(spl *Sample) Deviances {
	var devs Deviances
	for i, p := range spl.values {
		devs[i][0] = p[0] - 512
		devs[i][1] = p[1] - 512
	}
	return devs
}

func agreeances(d Deviances) Agreeances {
	var ags Agreeances
	for i, p := range d {
		ags[i] = p[0] + p[1]
	}
	return ags
}
