package main

import (
	"log"
)

func doDisplay(numbersChan chan float64, stopDisplayChan chan bool) {
	var min, max float64

	for {
		select {
		case <-stopDisplayChan:
			log.Printf("Max/min: %4.0f / %4.0f", max, min)
			return
		case n := <-numbersChan:
			if n > max {
				max = n
			}
			if n < min {
				min = n
			}
			// log.Printf("%4.0f", n)
		}
	}
}
