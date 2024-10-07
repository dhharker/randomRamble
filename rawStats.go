package main

import (
	"log"
	"math"

	"github.com/montanaflynn/stats"
)

const RING_BUFFER_DURATION_SECONDS = 1

// This ALWAYS refers to calculations done with BIN_VALS_SIZE values
type ChunkStats struct {
	mean               float64
	variance           float64
	sum                float64
	sStdDev            float64
	deviations         []float64
	deviationsInStDevs []float64
}

// Calculated for the whole of the window buffer by adding/subtracting the new/old data values only
type WindowStats struct {
	sum  float64
	mean float64
}

// Stats for entire run
type RunningStats struct {
	correlationSum float64
}

func doRawStats(config *Config, sampleChan chan *Sample, outputChan chan *Sample, ro *Orchestrator) {

	// reads per second * seconds in ring * values per read
	adcValueWindowBufferSize := nextPowerOf2(uint32(math.Ceil(1000 / float64(config.tickDelayMs) * RING_BUFFER_DURATION_SECONDS * float64(BIN_VALS_SIZE))))

	// Store recent ADC values
	bufAs := NewWindowBuffer[float64](int(adcValueWindowBufferSize))
	bufBs := NewWindowBuffer[float64](int(adcValueWindowBufferSize))

	// Track stats of contents of WindowBuffers
	var bufAStats, bufBStats WindowStats

	// Track stats for run
	var runningStats RunningStats

	// If the window buffers were filled before we WriteShift()ed the latest data in then we can do calcs.
	alreadyFilled := false

	ro.wg.Add(1)
	for {
		select {
		case _, isFalse := <-ro.shutdownOnCloseChan:
			log.Printf("Orchestrator:Shutting down doRawStats")
			log.Printf("Window buffer size: %v samples (%v bytes)", adcValueWindowBufferSize, adcValueWindowBufferSize*64)
			log.Printf("Window buffer duration: ~%1.1fs", (float64(adcValueWindowBufferSize)/float64(BIN_VALS_SIZE))/(1000/float64(config.tickDelayMs)))
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

			sumAs, err := stats.Sum(rngA[:])
			if err != nil {
				log.Printf("Error sum A %v", err)
			}
			sumBs, err := stats.Sum(rngB[:])
			if err != nil {
				log.Printf("Error sum B %v", err)
			}

			// Add the sum of the new data to the running sum for the wb
			bufAStats.sum += sumAs
			bufBStats.sum += sumBs

			// If the buffers were already full when we added data, then subtract the sums of the removed data from the running sums
			if alreadyFilled {
				sumARs, err := stats.Sum(subtractAs[:])
				if err != nil {
					log.Printf("Error aggregating AR %v", err)
				}
				sumBRs, err := stats.Sum(subtractBs[:])
				if err != nil {
					log.Printf("Error aggregating BR %v", err)
				}
				bufAStats.sum -= sumARs
				bufBStats.sum -= sumBRs
			}

			// Update running means
			if bufAs.filled {
				bufAStats.mean = bufAStats.sum / float64(adcValueWindowBufferSize)
				bufBStats.mean = bufBStats.sum / float64(adcValueWindowBufferSize)
			} else {
				bufAStats.mean = bufAStats.sum / float64(bufAs.writePtr)
				bufBStats.mean = bufBStats.sum / float64(bufBs.writePtr)
			}

			chunkAStats := chunkStats(rngA[:], bufAStats.mean)
			chunkBStats := chunkStats(rngB[:], bufBStats.mean)

			cor, _ := stats.Correlation(chunkAStats.deviationsInStDevs, chunkBStats.deviationsInStDevs)
			var corEmo = ""
			switch {
			case cor > 0:
				corEmo = "\u2795"
			case cor < 0:
				corEmo = "\u2796"
			default:
				corEmo = "\u27A1"
			}
			runningStats.correlationSum += cor
			log.Printf("Correlation(tally): %v\t% 2.2f\t(% 2.2f)", corEmo, cor, runningStats.correlationSum)

			// agB, _ := aggregate(rngb)
			// log.Printf("AG AM: %2.1f\tAV: %2.1f\tBM: %2.1f\tBV: %2.1f", agA.mean, agA.variance, agB.mean, agB.variance)
			if !ro.shutdownRequested {
				outputChan <- spl
			}
		}
	}
}

func chunkStats(vals []float64, windowMean float64) *ChunkStats {

	deviations := getDeviations(vals[:], windowMean)
	sqDeviations := squares(deviations)
	sumSqDevs, _ := stats.Sum(sqDeviations)
	sStdDev := math.Sqrt(sumSqDevs / float64(len(vals)-1))
	deviationsInStDevs := scale(deviations, sStdDev)

	as := &ChunkStats{
		sStdDev:            sStdDev,
		deviations:         deviations,
		deviationsInStDevs: deviationsInStDevs,
	}

	return as
}

func getDeviations(vals []float64, mean float64) []float64 {
	o := make([]float64, len(vals))
	for i, v := range vals {
		o[i] = v - mean
	}
	return o
}

func squares(vals []float64) []float64 {
	o := make([]float64, len(vals))
	for i, v := range vals {
		o[i] = math.Pow(v, 2)
	}
	return o
}

func scale(vals []float64, scaleBy float64) []float64 {
	o := make([]float64, len(vals))
	for i, v := range vals {
		o[i] = v / scaleBy
	}
	return o
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
