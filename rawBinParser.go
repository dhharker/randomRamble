package main

import "log"

func doParseBinaryFormat(binSampleChan chan *Sample, parsedSampleChan chan *Sample, ro *Orchestrator) {

	ro.wg.Add(1)
	for {
		select {
		case _, isFalse := <-ro.shutdownOnCloseChan:
			log.Printf("Orchestrator:Shutting down doParseBinaryFormat")
			if isFalse {
				panic("doParseBinaryFormat panic")
			}
			ro.wg.Done()
			return
		case spl := <-binSampleChan:
			// parse
			if !ro.shutdownRequested {
				parsedSampleChan <- parseRawValues(spl)
			}
		}
	}
}

func parseRawValues(spl *Sample) *Sample {
	locked := false
	seq := 0
	masks := []byte{0x00, 0x40, 0x80, 0xC0}
	for _, cb := range spl.sample {
		header := 0xC0 & cb
		if locked || header == 0x00 && !locked {
			locked = true
			switch header {
			case masks[0]:
				spl.rawValues[seq][0] = ((uint16(cb & 0x0F)) << 6)
			case masks[1]:
				spl.rawValues[seq][0] |= (uint16(cb & 0x3F))
			case masks[2]:
				spl.rawValues[seq][1] = ((uint16(cb & 0x0F)) << 6)
			case masks[3]:
				spl.rawValues[seq][1] |= (uint16(cb & 0x3F))
				seq++
			default:
			}
		}
	}
	return spl
}
