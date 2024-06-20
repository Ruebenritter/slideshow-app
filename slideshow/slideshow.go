package slideshow

import (
	"time"
)

type Slideshow struct {
	Images        []string
	SlideDuration time.Duration
	timer         *time.Timer
	ticker        *time.Ticker
	elapsedTime   time.Duration
	paused        bool
	CurrentIndex  int
	StopChan      chan bool
	progressChan  chan float64
	imageChan     chan string
}

// golang constructor
func NewSlideshow(images []string, duration time.Duration) *Slideshow {
	return &Slideshow{
		Images:        images,
		SlideDuration: duration,
		timer:         time.NewTimer(duration),
		ticker:        time.NewTicker(time.Second),
		elapsedTime:   0,
		paused:        false,
		CurrentIndex:  0,
		StopChan:      make(chan bool), // or chan struct{}?
		progressChan:  make(chan float64),
		imageChan:     make(chan string),
	}
}

// set images func with images parameter
func (s *Slideshow) SetImages(images []string) {
	s.Images = images
}

func (s *Slideshow) IsPaused() bool {
	return s.paused
}

func (s *Slideshow) ImageChan() chan string {
	return s.imageChan
}

// start func
func (s *Slideshow) Start() {
	s.resetTimerAndTicker()
	go func() {
		for {
			select {
			case <-s.timer.C:
				s.NextSlide((s.CurrentIndex + 1) % len(s.Images))
			case <-s.ticker.C:
				s.updateProgress()
			case <-s.StopChan:
				return
			}

		}
	}()
}

func (s *Slideshow) Pause() {
	if s.paused {
		s.timer.Reset(s.SlideDuration - s.elapsedTime)
		s.ticker.Reset(time.Second)
	} else {
		s.timer.Stop()
		s.ticker.Stop()
	}
	s.paused = !s.paused
}

// next slide func with index parameter
func (s *Slideshow) NextSlide(index int) string {
	s.CurrentIndex = index
	s.Start()
	s.imageChan <- s.Images[s.CurrentIndex]
	return s.Images[s.CurrentIndex]
}

func (s *Slideshow) resetTimerAndTicker() {
	if s.timer != nil && s.ticker != nil {
		s.timer.Stop()
		s.ticker.Stop()
	}
	s.timer = time.NewTimer(s.SlideDuration)
	s.ticker = time.NewTicker(time.Second)
	s.elapsedTime = 0
}

func (s *Slideshow) updateProgress() {
	s.elapsedTime += time.Second
	progress := float64(s.elapsedTime.Seconds())
	s.progressChan <- progress
}

func (s *Slideshow) ProgressChan() chan float64 {
	return s.progressChan
}

func (s *Slideshow) Stop() {
	s.StopChan <- true
	s.timer.Stop()
	s.ticker.Stop()
}
