package main

import (
	"log"
	"math"

	"github.com/montanaflynn/stats"
)

const RING_BUFFER_DURATION_SECONDS = 5

// This ALWAYS refers to calculations done with BIN_VALS_SIZE values
type AggregateStats struct {
	mean     float64
	variance float64
	sum      float64
	stdS     float64
}

// Calculated for the whole of the window buffer by adding/subtracting the new/old data values only
type RunningStats struct {
	sum  float64
	mean float64
}

func doRawStats(config *Config, sampleChan chan *Sample, outputChan chan *Sample, ro *Orchestrator) {

	// reads per second * seconds in ring * values per read
	windowBufferSize := nextPowerOf2(uint32(math.Ceil(1000 / float64(config.tickDelayMs) * RING_BUFFER_DURATION_SECONDS * float64(BIN_VALS_SIZE))))

	// Store recent ADC values
	bufAs := NewWindowBuffer[float64](int(windowBufferSize))
	bufBs := NewWindowBuffer[float64](int(windowBufferSize))

	// Track stats of contents of WindowBuffers
	var bufAStats, bufBStats RunningStats

	// If the window buffers were filled before we WriteShift()ed the latest data in then we can do calcs.
	alreadyFilled := false

	ro.wg.Add(1)
	for {
		select {
		case _, isFalse := <-ro.shutdownOnCloseChan:
			log.Printf("Orchestrator:Shutting down doRawStats")
			log.Printf("Window buffer size: %v samples (%v bytes)", windowBufferSize, windowBufferSize*64)
			if isFalse {
				panic("doRawStats panic")
			}
			close(outputChan)
			ro.wg.Done()
			return
		case spl := <-sampleChan:

			// Array of ADC values for both ADCs
			rngA, rngB := unzip2dAry(spl.rawValues)

			// If the window buffer was already filled before we put more data in then we can do some stats (yay!)
			alreadyFilled = bufAs.filled

			// Put the latest data into the buffer and save what was discarded
			subtractAs := bufAs.WriteShift(rngA[:])
			subtractBs := bufBs.WriteShift(rngB[:])

			agA, err := aggregate(rngA[:])
			if err != nil {
				log.Printf("Error aggregating A %v", err)
			}
			agB, err := aggregate(rngB[:])
			if err != nil {
				log.Printf("Error aggregating B %v", err)
			}

			// Add the sum of the new data to the running sum for the wb
			bufAStats.sum += agA.sum
			bufBStats.sum += agB.sum

			// If the buffers were already full when we added data, then subtract the sums of the removed data from the running sums
			if alreadyFilled {
				agAR, err := aggregate(subtractAs[:])
				if err != nil {
					log.Printf("Error aggregating AR %v", err)
				}
				agBR, err := aggregate(subtractBs[:])
				if err != nil {
					log.Printf("Error aggregating BR %v", err)
				}
				bufAStats.sum -= agAR.sum
				bufBStats.sum -= agBR.sum

				// Update running means
				bufAStats.mean = bufAStats.sum / float64(windowBufferSize)
				bufBStats.mean = bufBStats.sum / float64(windowBufferSize)
			} else {
				// Update running means
				bufAStats.mean = bufAStats.sum / float64(bufAs.writePtr)
				bufBStats.mean = bufBStats.sum / float64(bufBs.writePtr)
			}

			log.Printf("%v %3.1f %3.1f", alreadyFilled, bufAStats.mean, bufBStats.mean)

			// 	cov, _ := stats.Correlation(rngA[:], rngB[:])
			// log.Printf("Correlation: %2.2f", cov)

			// agB, _ := aggregate(rngb)
			// log.Printf("AG AM: %2.1f\tAV: %2.1f\tBM: %2.1f\tBV: %2.1f", agA.mean, agA.variance, agB.mean, agB.variance)
			if !ro.shutdownRequested {
				outputChan <- spl
			}
		}
	}
}

func aggregate(vals []float64) (*AggregateStats, error) {

	mean, err := stats.Mean(vals[:])
	if err != nil {
		return nil, err
	}

	variance, err := stats.Variance(vals[:])
	if err != nil {
		return nil, err
	}

	sum, err := stats.Sum(vals[:])
	if err != nil {
		return nil, err
	}

	stdS, err := stats.StandardDeviationSample(vals[:])
	if err != nil {
		return nil, err
	}

	as := &AggregateStats{
		mean:     mean,
		variance: variance,
		sum:      sum,
		stdS:     stdS,
	}

	return as, nil
}

func unzip2dAry(input [BIN_VALS_SIZE][2]uint16) ([BIN_VALS_SIZE]float64, [BIN_VALS_SIZE]float64) {
	var outa, outb [BIN_VALS_SIZE]float64
	for i, pair := range input {
		outa[i] = float64(pair[0])
		outb[i] = float64(pair[1])
	}
	return outa, outb
}

func nextPowerOf2(v uint32) uint32 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}

type T interface{}

// Write forever, keep last size values
type WindowBuffer[T any] struct {
	data []T
	// how many datas
	size int
	// have at least size datas been written yet?
	filled   bool
	writePtr int
}

func NewWindowBuffer[T any](size int) *WindowBuffer[T] {
	// var w WindowBuffer[T]
	w := &WindowBuffer[T]{
		data: make([]T, size),
		size: size,
	}

	return w
}

// Write values to the WB and return an array containing what was overwritten
func (w *WindowBuffer[T]) WriteShift(d []T) []T {
	var overwritten []T
	for _, v := range d {
		overwritten = append(overwritten, w.data[w.writePtr])
		w.data[w.writePtr] = v
		w.writePtr++
		if w.writePtr == w.size {
			w.writePtr = 0
			if !w.filled {
				w.filled = true
				log.Println("Window buffer filled!")
			}
		}
	}
	return overwritten
}
