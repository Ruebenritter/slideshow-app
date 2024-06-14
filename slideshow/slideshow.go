package slideshow

import (
	"time"
)

type Slideshow struct {
	images        []string
	slideDuration time.Duration
	timer         *time.Timer
	ticker        *time.Ticker
	remaningTime  time.Duration
	paused        bool
	currentIndex  int
	stopChan      chan bool
	progressChan  chan float64
}

// golang constructor
func NewSlideshow(images []string, duration time.Duration) *Slideshow {
	return &Slideshow{
		images:        images,
		slideDuration: duration,
		timer:         time.NewTimer(duration),
		ticker:        time.NewTicker(time.Second),
		remaningTime:  duration,
		paused:        false,
		currentIndex:  0,
		stopChan:      make(chan bool), // or chan struct{}?
		progressChan:  make(chan float64),
	}
}

// set images func with images parameter
func (s *Slideshow) SetImages(images []string) {
	s.images = images
}

// start func
func (s *Slideshow) Start() {
	s.resetTimerAndTicker()
	go func() {
		for {
			select {
			case <-s.timer.C:
				s.NextSlide((s.currentIndex + 1) % len(s.images))

			case <-s.ticker.C:
				s.updateProgress()
			case <-s.stopChan:
				return
			}

		}
	}()
}

// next slide func with index parameter
func (s *Slideshow) NextSlide(index int) string {
	s.currentIndex = index
	s.resetTimerAndTicker()
	return s.images[s.currentIndex]
}

func (s *Slideshow) resetTimerAndTicker() {
	s.timer.Stop()
	s.ticker.Stop()
	s.timer = time.NewTimer(s.slideDuration)
	s.ticker = time.NewTicker(time.Second)
}

func (s *Slideshow) updateProgress() {
	elapsed := s.slideDuration.Seconds() - s.remaningTime.Seconds()
	progress := float64(elapsed) / s.slideDuration.Seconds()
	s.progressChan <- progress
}

func (s *Slideshow) ProgressChan() chan float64 {
	return s.progressChan
}

func (s *Slideshow) Stop() {
	close(s.stopChan)
	s.timer.Stop()
	s.ticker.Stop()
}
