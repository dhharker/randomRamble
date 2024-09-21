package main

func doParsing(rawDataChan chan *Sample, sampleChan chan *Sample, stopParserChan chan bool) {
	for {
		select {
		case <-stopParserChan:
			return
		case spl := <-rawDataChan:
			parseRngBinaryFormat(spl)
			sampleChan <- spl
		}
	}
}

func parseRngBinaryFormat(spl *Sample) *Sample {
	locked := false
	seq := 0
	masks := []byte{0x00, 0x40, 0x80, 0xC0}
	for _, cb := range spl.rawData {
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
