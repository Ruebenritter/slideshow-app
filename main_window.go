package main

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func BuildMainWindow() {
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

		// images come in lexical order, so we shuffle before we slice the first n images for better randomness
		// ToDo: Decide whether user should have the option to disable shuffling.
		// ToDo: Pick the correct amount of images instead of checking all.
		// → look at random file in directory, validate extension and add to list until n = amount
		shuffleImages(&images)

		slideshow := NewSlideshow(images[:amount], timePerImage)

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
