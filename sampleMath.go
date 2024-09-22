package main

import (
	"log"
	"math"
)

// goroutine to take a sample and do some maths on it
func doMath(sampleChan chan *Sample, outputChan chan *Sample, stopMathChan chan bool) {

	for {
		select {
		case <-stopMathChan:
			log.Printf("Debug bits: %v", debugBits)
			log.Printf("      Ones: %v", debugBitsOnes)
			log.Printf("    Zeroes: %v", debugBitsZeroes)
			log.Printf("     Histo: %v", histo)
			var dbbr = [8]float32{}
			for i := range debugBits {
				dbbr[i] = float32(debugBitsOnes[i]) / float32(debugBitsZeroes[i])
			}
			log.Printf("    Ratios: %v", dbbr)
			return
		case spl := <-sampleChan:
			// log.Printf("S %v", spl.values)
			calculateWalkDeltas(spl)
			spl.entropy = entropy(spl.values)

			// Send the sample off to display
			outputChan <- spl
		}
	}
}

var EACH_BIT = [...]byte{0b10000000, 0b01000000, 0b00100000, 0b00010000, 0b00001000, 0b00000100, 0b00000010, 0b00000001}
var debugBits = [8]int64{}
var debugBitsOnes = [8]int64{}
var debugBitsZeroes = [8]int64{}
var histo = [256]int{}

// rollingMean := ringbuffer.New(1024)

func getWalkDelta(x byte) int8 {
	var acc int8 = 0
	// log.Printf("%8b", x)
	for a, b := range EACH_BIT {
		if b&x != 0 {
			acc++
			debugBits[a]++
			debugBitsOnes[a]++
		} else {
			acc--
			debugBits[a]--
			debugBitsZeroes[a]++
		}
	}
	return acc
}

func calculateWalkDeltas(spl *Sample) {
	var s int64 = 0
	for i, p := range spl.values {
		histo[p]++
		spl.walkDeltas[i] = getWalkDelta(p)
		s += int64(spl.walkDeltas[i])
	}
	spl.walkSum = s
}

func entropy(data []byte) float64 {

	l := float64(0)
	m := map[byte]float64{}
	for _, r := range data {
		m[r]++
		l++
	}
	var hm float64
	for _, c := range m {
		hm += c * math.Log2(c)
	}
	return math.Log2(l) - hm/l
}
