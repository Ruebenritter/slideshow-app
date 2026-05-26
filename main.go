package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func main() {
	BuildMainWindow()
}

func showSlideshow(a fyne.App, slideshowObj *Slideshow) {
	w := a.NewWindow("Slideshow")

	imageCanvas := canvas.NewImageFromFile(slideshowObj.Images[slideshowObj.CurrentIndex])
	imageCanvas.FillMode = canvas.ImageFillContain
	imageCanvas.SetMinSize(fyne.NewSize(1920/2, 800))

	currentIndexLabel := widget.NewLabel(fmt.Sprint(slideshowObj.CurrentIndex+1) + " of " + fmt.Sprint(len(slideshowObj.Images)))
	centeredLabel := container.New(layout.NewCenterLayout(), currentIndexLabel)

	progressBar := widget.NewProgressBar()
	progressBar.Max = float64(slideshowObj.SlideDuration.Seconds())

	nextButton := widget.NewButton("Next", func() {
		slideshowObj.NextSlide((slideshowObj.CurrentIndex + 1) % len(slideshowObj.Images))
	})

	prevButton := widget.NewButton("Previous", func() {
		slideshowObj.NextSlide((slideshowObj.CurrentIndex - 1 + len(slideshowObj.Images)) % len(slideshowObj.Images))
	})

	var pauseButton *widget.Button
	pauseButton = widget.NewButton("Pause", func() {
		if slideshowObj.IsPaused() {
			slideshowObj.Pause()
			pauseButton.SetText("Pause")
		} else {
			slideshowObj.Pause()
			pauseButton.SetText("Resume")
		}
	})

	stopButton := widget.NewButton("Stop", func() {
		slideshowObj.Stop()
		w.Close()
	})

	buttons := container.NewHBox(prevButton, pauseButton, nextButton)
	slideGroup := container.NewVBox(centeredLabel, buttons, progressBar, stopButton)
	centeredButtonGrop := container.New(layout.NewCenterLayout(), slideGroup)

	split := container.NewHSplit(centeredButtonGrop, imageCanvas)
	split.Offset = 0.33

	w.SetContent(split)
	w.Resize(fyne.NewSize(1280, 720))

	go func() {
		slideshowObj.Start()
		for {
			select {
			case progress := <-slideshowObj.ProgressChan():
				progressBar.SetValue(progress)
			case img := <-slideshowObj.ImageChan():
				imageCanvas.File = img
				imageCanvas.Refresh()
				currentIndexLabel.SetText(fmt.Sprint(slideshowObj.CurrentIndex+1) + " of " + fmt.Sprint(len(slideshowObj.Images)))
				progressBar.SetValue(0)
			case <-slideshowObj.StopChan:
				return
			}
		}
	}()
	w.Show()
}
