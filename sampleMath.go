package main

import (
	"log"
)

// goroutine to take a sample and do some maths on it
func doMath(sampleChan chan *Sample, numbersChan chan *Sample, stopMathChan chan bool) {
	var walkSums = &WalkSums{
		sum:   0,
		sumA:  0,
		sumB:  0,
		sumEq: 0,
	}
	for {
		select {
		case <-stopMathChan:
			log.Printf("Debug bits: %v", debugBits)
			log.Printf("      Ones: %v", debugBitsOnes)
			log.Printf("    Zeroes: %v", debugBitsZeroes)
			log.Printf("    Blanks: %v", numBlanks)
			log.Printf("     Histo: %v", histo)
			var dbbr = [8]float32{}
			for i, _ := range debugBits {
				dbbr[i] = float32(debugBitsOnes[i]) / float32(debugBitsZeroes[i])
			}
			log.Printf("    Ratios: %v", dbbr)
			return
		case spl := <-sampleChan:
			log.Printf("S %v", spl.rawValues)
			pruneSampleAndGetDeltas(spl)
			sd := sumDeltas(spl)
			walkSums.sum += sd.sum
			walkSums.sumA += sd.sumA
			walkSums.sumB += sd.sumB
			walkSums.sumEq += sd.sumEq
			wsSnapshot := walkSums
			spl.walkSums = *wsSnapshot

			// Send the sample off to display
			numbersChan <- spl
		}
	}
}

func prune(tenBit uint16) byte {
	// They implement arithmetic shifts if the left operand is a signed integer and logical shifts if it is an unsigned integer.

	// return byte((0b0000000111111110 & tenBit) >> 1)
	// return byte((0b0000001111111100 & uint16(tenBit)) >> 2)
	// return byte(0b0000000111111110 & (uint16(tenBit)) >> 1)
	// return byte(tenBit) ^ byte((tenBit>>2)&0b11000000)
	return byte(tenBit)
}

var EACH_BIT = [...]byte{0b10000000, 0b01000000, 0b00100000, 0b00010000, 0b00001000, 0b00000100, 0b00000010, 0b00000001}
var debugBits = [8]int64{}
var debugBitsOnes = [8]int64{}
var debugBitsZeroes = [8]int64{}
var numBlanks = [3]int{0, 0, 0}
var histo = [1024]int{}

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

func pruneSampleAndGetDeltas(spl *Sample) {
	for i, p := range spl.rawValues {
		spl.pruned[i][0] = prune(p[0])
		spl.pruned[i][1] = prune(p[1])
		if p[0] == 0x0000 {
			numBlanks[0]++
			if p[1] == 0x0000 {
				numBlanks[2]++
			}
		} else if p[1] == 0x0000 {
			numBlanks[1]++
		}
		histo[p[0]]++
		histo[p[1]]++
		spl.walkDeltas[i][0] = getWalkDelta(spl.pruned[i][0])
		spl.walkDeltas[i][1] = getWalkDelta(spl.pruned[i][1])
	}
}

func sumDeltas(spl *Sample) *WalkSums {
	s := &WalkSums{
		sum:   0,
		sumA:  0,
		sumB:  0,
		sumEq: 0,
	}
	for _, d := range spl.walkDeltas {
		s.sum += int64(d[0] + d[1])
		s.sumA += int64(d[0])
		s.sumB += int64(d[1])
		if d[0] == d[1] {
			s.sumEq += int64(d[0])
		}
	}
	return s
}

/*
type Deviances [MAX_PAIRS_RB][2]int16
type Agreeances [MAX_PAIRS_RB]int16

func deviances(spl *Sample) Deviances {
	var devs Deviances
	for i, p := range spl.rawValues {
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

case spl := <-sampleChan:
			// log.Printf("Sample #%v %v", spl.sampleCount, spl.values)
			// log.Printf("Deviances %v", deviances(spl))
			// log.Printf("Agreeances %v", agreeances(deviances(spl)))
			// agSum = 0
			// ags := agreeances(deviances(spl))
			// // log.Printf("%v", ags)
			// for _, a := range ags {
			// 	agSum += a
			// }
			// numbersChan <- float64(agSum) / float64(len(ags))

		}
*/
