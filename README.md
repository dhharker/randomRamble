# Random Ramble

`randomRamble` connects to TrueRNGpro V2 over USB serial and runs real-time analyses to quantify and display deviations in the randomness of the output(s) from the TrueRNG.

## Meta

This is the first thing I've ever written in Go, and haven't written a full piece of software for over about 5 years. So this is going to be a learning journey and perhaps not The Best Way of doing things.


### pkg-config error finding libusb-1.0 on linux mint 22

Needed to run this to fix error:
```
$ pkg-config --cflags --libs libusb-1.0
```

### Dev

Install fyne helper and deps https://docs.fyne.io/started/
```
sudo apt-get install golang gcc libgl1-mesa-dev xorg-dev
go install fyne.io/fyne/v2/cmd/fyne@latest
go mod tidy
```

### @TODO

- [ ] Close channels in orchestrator shutdown blocks in data pipeline goroutines
- [ ] Write like, literally any tests at all


## Maybes
- Analysis
    - [ ] (1) Integrate over a fixed window of a second or two
    - [ ] (2) Integrate over a longer window with more recent values given a higher weighting
    - [ ] (3) Integrate over n windows something like Wolf's app for the Wyrdoscope works
- [ ] Realtime display of results
    - [ ] Probably use fyne https://fyne.io/blog/2019/03/19/building-cross-platform-gui.html for GUI
    - [ ] 1+ histograms for (1),(2) above?
    - [ ] A fun shortest-window or weighted-window feedback e.g. tone, colour, size
    - [ ] Gamification???
    - [ ] Do this via a webserver so people can use the RNG remotely

