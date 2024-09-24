package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

func gui(samplesChan chan *Sample, ro *Orchestrator) (fyne.App, fyne.Window) {
	log.Println("Initialising GUI")
	a := app.New()
	w := a.NewWindow("Hello World")

	str := binding.NewString()
	str.Set("I am a string.")
	label := widget.NewLabelWithData(str)

	w.SetContent(label)
	// Close the GUI if orchestrator requests
	log.Println("Initialising hook orchestrator GUI quit")
	go func() {
		for {
			select {
			case _, isFalse := <-ro.shutdownOnCloseChan:
				log.Printf("Orchestrator:Shutting down GUI")
				if isFalse {
					panic("GUI panic")
				}
				a.Quit()
				log.Printf("Orchestrator:Shut down GUI done")
				return
			}
		}
	}()

	log.Println("Initialising hook update window on sample")
	go func() {
		for spl := range samplesChan {

			str.Set(fmt.Sprintf("%v", spl.walkSum))
		}
	}()

	log.Println("return gui to main()")
	// app.Run has to be called in main func because ??? fyne reasons
	return a, w
}
