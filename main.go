package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/Ruebenritter/slideshow-app/slideshow"

	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Slideshow Setup")

	var dirPath string
	dirEntryDisplay := widget.NewLabel("No directory selected")

	selectionDirButton := widget.NewButton("Select Directory", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if uri == nil {
				return
			}
			dirPath = uri.Path()
			dirEntryDisplay.SetText(dirPath)
		}, w)
	})

	timeEntry := widget.NewEntry()
	// set min width of time entry to 100px
	timeEntry.Resize(fyne.NewSize(300, timeEntry.MinSize().Height))
	timeEntry.SetPlaceHolder("Enter time per image in seconds")

	// add row of buttons with preset times between 60s and 300s
	presetTimes := []int{60, 120, 180, 240, 300}
	presetTimeButtons := make([]fyne.CanvasObject, len(presetTimes))
	for i, t := range presetTimes {
		time := t
		button := widget.NewButton(strconv.Itoa(time), func() {
			timeEntry.SetText(strconv.Itoa(time))
		})
		presetTimeButtons[i] = button
	}

	presetTimeRow := container.NewGridWithColumns(2, timeEntry, container.NewHBox(presetTimeButtons...))

	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("Enter amount of images")

	presetAmounts := []int{10, 25, 50, 75, 100}
	presetAmountButtons := make([]fyne.CanvasObject, len(presetAmounts))
	for i, a := range presetAmounts {
		amount := a
		button := widget.NewButton(strconv.Itoa(amount), func() {
			amountEntry.SetText(strconv.Itoa(amount))
		})
		presetAmountButtons[i] = button
	}

	presetAmountRow := container.NewGridWithColumns(2, amountEntry, container.NewHBox(presetAmountButtons...))

	startButton := widget.NewButton("Start Slideshow", func() {
		timePerImage, err1 := time.ParseDuration(timeEntry.Text + "s")
		var amount int
		_, err2 := fmt.Sscanf(amountEntry.Text, "%d", &amount)

		if dirPath == "" {
			dialog.ShowError(fmt.Errorf("no directory selected"), w)
			return
		}

		if err1 != nil {
			dialog.ShowError(fmt.Errorf("invalid time format"), w)
			return
		}

		if err2 != nil {
			dialog.ShowError(fmt.Errorf("invalid amount format"), w)
			return
		}

		images := getImagesFromDir(dirPath)
		if len(images) == 0 {
			dialog.ShowError(fmt.Errorf("no images found in directory"), w)
			return
		}

		if amount > len(images) {
			dialog.ShowError(fmt.Errorf("amount of images is greater than the amount of images in the directory! Amount will be set to max"), w)
			amount = len(images)
		}

		rand.Shuffle(len(images), func(i, j int) {
			images[i], images[j] = images[j], images[i]
		})

		slideshow := slideshow.NewSlideshow(images[:amount], timePerImage)

		showSlideshow(a, slideshow)
	})

	grid := container.New(layout.NewGridLayout(2), selectionDirButton, dirEntryDisplay)
	w.SetContent(container.NewVBox(
		grid,
		presetTimeRow,
		presetAmountRow,
		startButton,
	))

	w.Resize(fyne.NewSize(1200, 800))
	w.ShowAndRun()
}

func getImagesFromDir(dir string) []string {
	var images []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := filepath.Ext(info.Name())
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
				images = append(images, path)
			}
		}
		return nil
	})
	return images
}

func showSlideshow(a fyne.App, slideshowObj *slideshow.Slideshow) {
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
