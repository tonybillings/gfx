package gfx

import "math"

const (
	defaultLowPassFilterName  = "LowPass"
	defaultHighPassFilterName = "HighPass"
	defaultBandPassFilterName = "BandPass"
)

/******************************************************************************
 Filter
******************************************************************************/

type Filter interface {
	GenerateCoefficients(int, float64, ...float64) []float64
	Apply(int, []float64) float64
	Enabled() bool
	SetEnabled(bool)
	Name() string
	SetName(string)
	CoefficientCount() int
	SampleRate() float64
	CutoffFrequencies() []float64
}

/******************************************************************************
 FilterBase
******************************************************************************/

type FilterBase struct {
	coefficients     []float64
	coefficientCount int
	enabled          bool
	name             string
	rate             float64
	cutoffs          []float64
}

func (f *FilterBase) Apply(index int, input []float64) (output float64) {
	if !f.enabled {
		return input[index]
	}

	var sum float64
	n := len(f.coefficients)
	inputLength := len(input)
	for j := 0; j < n; j++ {
		idx := (index - j + inputLength) % inputLength
		sum += input[idx] * f.coefficients[j]
	}
	return sum
}

func (f *FilterBase) Enabled() bool {
	return f.enabled
}

func (f *FilterBase) SetEnabled(enabled bool) {
	f.enabled = enabled
}

func (f *FilterBase) Name() string {
	return f.name
}

func (f *FilterBase) SetName(name string) {
	f.name = name
}

func NewFilterBase(name string) *FilterBase {
	return &FilterBase{
		enabled: true,
		name:    name,
	}
}

func (f *FilterBase) CoefficientCount() int {
	return f.coefficientCount
}

func (f *FilterBase) SampleRate() float64 {
	return f.rate
}

func (f *FilterBase) CutoffFrequencies() []float64 {
	return f.cutoffs
}

/******************************************************************************
 LowPassFilter
******************************************************************************/

type LowPassFilter struct {
	FilterBase
}

func (f *LowPassFilter) GenerateCoefficients(coefficientCount int, sampleRate float64, cutoffFrequencies ...float64) []float64 {
	f.coefficientCount = coefficientCount
	f.rate = sampleRate
	f.cutoffs = cutoffFrequencies
	cutoffFrequency := cutoffFrequencies[0]

	result := make([]float64, coefficientCount)
	M := float64(coefficientCount - 1)
	fc := cutoffFrequency / sampleRate

	var sum float64
	for i := 0; i < coefficientCount; i++ {
		if i == int(M/2) {
			result[i] = 2 * fc
		} else {
			result[i] = math.Sin(2*math.Pi*fc*(float64(i)-M/2)) / (math.Pi * (float64(i) - M/2))
		}

		result[i] *= 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/M)
		sum += result[i]
	}

	for i := range result {
		result[i] /= sum
	}

	f.coefficients = result
	return result
}

func NewLowPassFilter() Filter {
	return &LowPassFilter{
		FilterBase: *NewFilterBase(defaultLowPassFilterName),
	}
}

/******************************************************************************
 HighPassFilter
******************************************************************************/

type HighPassFilter struct {
	FilterBase
}

func (f *HighPassFilter) GenerateCoefficients(coefficientCount int, sampleRate float64, cutoffFrequencies ...float64) []float64 {
	f.coefficientCount = coefficientCount
	f.rate = sampleRate
	f.cutoffs = cutoffFrequencies
	cutoffFrequency := cutoffFrequencies[0]

	result := make([]float64, coefficientCount)
	M := float64(coefficientCount - 1)
	fc := cutoffFrequency / sampleRate

	for i := 0; i < coefficientCount; i++ {
		if i == int(M/2) {
			result[i] = 2 * fc
		} else {
			result[i] = math.Sin(2*math.Pi*fc*(float64(i)-M/2)) / (math.Pi * (float64(i) - M/2))
		}

		result[i] *= 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/M)
	}

	for i := range result {
		if i == int(M/2) {
			result[i] = 1 - result[i]
		} else {
			result[i] = -result[i]
		}
	}

	f.coefficients = result
	return result
}

func NewHighPassFilter() Filter {
	return &HighPassFilter{
		FilterBase: *NewFilterBase(defaultHighPassFilterName),
	}
}

/******************************************************************************
 BandPassFilter
******************************************************************************/

type BandPassFilter struct {
	FilterBase
}

func (f *BandPassFilter) GenerateCoefficients(coefficientCount int, sampleRate float64, cutoffFrequencies ...float64) []float64 {
	f.coefficientCount = coefficientCount
	f.rate = sampleRate
	f.cutoffs = cutoffFrequencies
	lowCutoffFreq := cutoffFrequencies[0]
	highCutoffFreq := cutoffFrequencies[1]

	lowPassCoefficients := NewLowPassFilter().GenerateCoefficients(coefficientCount, sampleRate, lowCutoffFreq)
	highPassCoefficients := NewHighPassFilter().GenerateCoefficients(coefficientCount, sampleRate, highCutoffFreq)
	result := make([]float64, coefficientCount)
	M := float64(coefficientCount - 1)

	for i := range result {
		if i == int(M/2) {
			result[i] = lowPassCoefficients[i] + highPassCoefficients[i] - 1
		} else {
			result[i] = lowPassCoefficients[i] + highPassCoefficients[i]
		}
	}

	f.coefficients = result
	return result
}

func NewBandPassFilter() Filter {
	return &BandPassFilter{
		FilterBase: *NewFilterBase(defaultBandPassFilterName),
	}
}
