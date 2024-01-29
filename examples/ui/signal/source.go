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
	freqComp3      float64
	freqComp4      float64
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

			freq1 := (2.0 * math.Pi * s.freqComp1) / s.sampleRate
			freq2 := (2.0 * math.Pi * s.freqComp2) / s.sampleRate
			freq3 := (2.0 * math.Pi * s.freqComp3) / s.sampleRate
			freq4 := (2.0 * math.Pi * s.freqComp4) / s.sampleRate

			sine1 := math.Sin(float64(x) * freq1)
			sine2 := math.Sin(float64(x) * freq2)
			sine3 := math.Sin(float64(x) * freq3)
			sine4 := math.Sin(float64(x) * freq4)
			value := []float64{sine1 + sine2 + sine3 + sine4}

			for _, c := range s.outputChannels {
				select {
				case c <- value:
					break
				default:
				}
			}

			sleepTime := 1000.0 / s.sampleRate
			s.stateMutex.Unlock()
			time.Sleep(time.Duration(sleepTime) * time.Millisecond)
			x++
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

func (s *Source) FrequencyComponent3() float64 {
	s.stateMutex.Lock()
	value := s.freqComp3
	s.stateMutex.Unlock()
	return value
}

func (s *Source) FrequencyComponent4() float64 {
	s.stateMutex.Lock()
	value := s.freqComp4
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

func (s *Source) SetFrequencyComponent3(value float64) {
	s.stateMutex.Lock()
	s.freqComp3 = value
	s.stateMutex.Unlock()
}

func (s *Source) SetFrequencyComponent4(value float64) {
	s.stateMutex.Lock()
	s.freqComp4 = value
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
