package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/Ruebenritter/slideshow-app/slideshow"

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
	timeEntry.SetPlaceHolder("Enter time per image in seconds")

	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("Enter amount of images")

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
		timeEntry,
		amountEntry,
		startButton,
	))

	w.Resize(fyne.NewSize(400, 400))
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

func showSlideshow(a fyne.App, slideshow *Slideshow) {
	w := a.NewWindow("Slideshow")

	imageCanvas := canvas.NewImageFromFile(slideshow.images[slideshow.index])
	imageCanvas.FillMode = canvas.ImageFillContain
	imageCanvas.SetMinSize(fyne.NewSize(1920/2, 800))

	currentIndexLabel := widget.NewLabel(fmt.Sprint(slideshow.index) + " of " + fmt.Sprint(len(slideshow.images)))

	progressBar := widget.NewProgressBar()
	progressBar.Max = float64(slideshow.slideDuration.Seconds())

	updateImage := func() {
		imageCanvas.File = slideshow.images[slideshow.index]
		imageCanvas.Refresh()
		currentIndexLabel.SetText(fmt.Sprint(slideshow.index) + " of " + fmt.Sprint(len(slideshow.images)))

		slideshow.remaningTime = slideshow.slideDuration
		slideshow.paused = false
		startTimer(slideshow, progressBar)
	}

	nextButton := widget.NewButton("Next", func() {
		slideshow.index = (slideshow.index + 1) % len(slideshow.images)
		// stop timer and ticker
		slideshow.stopChan <- true
		updateImage()
	})

	prevButton := widget.NewButton("Previous", func() {
		slideshow.index = (slideshow.index - 1 + len(slideshow.images)) % len(slideshow.images)
		// stop timer and ticker
		slideshow.stopChan <- true
		updateImage()
	})

	var pauseButton *widget.Button
	pauseButton = widget.NewButton("Pause", func() {
		if slideshow.paused {
			slideshow.paused = false
			startTimer(slideshow, progressBar)
			pauseButton.SetText("Pause")
		} else {
			slideshow.paused = true
			if slideshow.timer != nil {
				slideshow.timer.Stop()
			}
			if slideshow.ticker != nil {
				slideshow.ticker.Stop()
			}
			pauseButton.SetText("Resume")
		}
	})

	stopButton := widget.NewButton("Stop", func() {
		slideshow.stopChan <- true
		w.Close()
	})

	buttons := container.NewHBox(prevButton, pauseButton, nextButton)
	slideGroup := container.NewVBox(currentIndexLabel, buttons, progressBar, stopButton)

	split := container.NewHSplit(slideGroup, imageCanvas)
	split.Offset = 0.33

	w.SetContent(split)
	w.Resize(fyne.NewSize(1280, 720))

	updateImage()
	w.Show()
}

func (s *Slideshow) nextImage() {
	s.index = (s.index + 1) % len(s.images)
}

func startTimer(slideshow *Slideshow, progressBar *widget.ProgressBar) {
	if slideshow.timer != nil {
		slideshow.timer.Stop()
	}

	slideshow.timer = time.NewTimer(slideshow.remaningTime)
	slideshow.ticker = time.NewTicker(time.Second)
	progressBar.SetValue(slideshow.slideDuration.Seconds() - slideshow.remaningTime.Seconds())

	go func() {
		for {
			select {
			case <-slideshow.timer.C:
				slideshow.index = (slideshow.index + 1) % len(slideshow.images)
				updateImage := func() {
					progressBar.SetValue(0)
					slideshow.remaningTime = slideshow.slideDuration
				}
				updateImage()
				slideshow.nextImage()
				return
			case <-slideshow.ticker.C:
				if !slideshow.paused {
					slideshow.remaningTime -= time.Second
					progressBar.SetValue(progressBar.Value + 1)
				}
			case <-slideshow.stopChan:
				slideshow.timer.Stop()
				slideshow.ticker.Stop()
				return
			}
		}
	}()

}
