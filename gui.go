package main

import (
	"fmt"
	"log"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
)

func gui(samplesChan chan *Sample, ro *Orchestrator) (fyne.App, fyne.Window) {
	log.Println("Initialising GUI")
	a := app.New()
	w := a.NewWindow("Hello World")

	w.SetContent(widget.NewLabel("I'm a potato!"))
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
			w.SetContent(widget.NewLabel(fmt.Sprintf("%v", spl.walkSum)))

		}
	}()

	log.Println("return gui to main()")
	// app.Run has to be called in main func because ??? fyne reasons
	return a, w
}
