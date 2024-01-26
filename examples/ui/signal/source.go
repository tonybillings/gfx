package signal

import (
	"context"
	"math"
	"sync"
	"time"
)

type Source struct {
	freqComp1      float64
	freqComp2      float64
	sampleRate     float64
	outputChannels []chan []float64
	stateMutex     sync.Mutex
}

func (s *Source) Run(ctx context.Context) {
	s.outputChannels = make([]chan []float64, 0)

	go func(ctx context.Context) {
		x := 0
		for {
			select {
			case <-ctx.Done():
				s.stateMutex.Lock()
				for _, c := range s.outputChannels {
					close(c)
				}
				s.stateMutex.Unlock()
				return
			default:
			}

			s.stateMutex.Lock()

			loFreq := (2.0 * math.Pi * s.freqComp1) / s.sampleRate
			hiFreq := (2.0 * math.Pi * s.freqComp2) / s.sampleRate
			sleepTime := 1000.0 / s.sampleRate

			sineLo := math.Sin(float64(x) * loFreq)
			sineHi := math.Sin(float64(x) * hiFreq)
			value := []float64{sineLo + sineHi}

			for _, c := range s.outputChannels {
				select {
				case c <- value:
					break
				default:
				}
			}

			s.stateMutex.Unlock()

			x++
			time.Sleep(time.Duration(sleepTime) * time.Millisecond)
		}
	}(ctx)
}

func (s *Source) FrequencyComponent1() float64 {
	s.stateMutex.Lock()
	value := s.freqComp1
	s.stateMutex.Unlock()
	return value
}

func (s *Source) FrequencyComponent2() float64 {
	s.stateMutex.Lock()
	value := s.freqComp2
	s.stateMutex.Unlock()
	return value
}

func (s *Source) SampleRate() float64 {
	s.stateMutex.Lock()
	value := s.sampleRate
	s.stateMutex.Unlock()
	return value
}

func (s *Source) SetFrequencyComponent1(value float64) {
	s.stateMutex.Lock()
	s.freqComp1 = value
	s.stateMutex.Unlock()
}

func (s *Source) SetFrequencyComponent2(value float64) {
	s.stateMutex.Lock()
	s.freqComp2 = value
	s.stateMutex.Unlock()
}

func (s *Source) SetSampleRate(value float64) {
	s.stateMutex.Lock()
	s.sampleRate = value
	s.stateMutex.Unlock()
}

func (s *Source) GetOutputChan() <-chan []float64 {
	c := make(chan []float64, 1000)
	s.stateMutex.Lock()
	s.outputChannels = append(s.outputChannels, c)
	s.stateMutex.Unlock()
	return c
}

func NewSource() *Source {
	return &Source{}
}
