package main

import (
	"log"
	// "fyne.io/fyne/v2/app"
	// "fyne.io/fyne/v2/widget"
)

func doDisplay(numbersChan chan float64, stopDisplayChan chan bool) {
	var min, max float64

	// a := app.New()
	// w := a.NewWindow("Hello World")

	// w.SetContent(widget.NewLabel("Hello World!"))
	// w.ShowAndRun()

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
