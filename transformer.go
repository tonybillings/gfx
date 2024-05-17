package gfx

import (
	"gonum.org/v1/gonum/dsp/fourier"
	"math/cmplx"
)

const (
	defaultFftTransformerName = "FFT"
)

/******************************************************************************
 Transformer
******************************************************************************/

// Transformer instances are used to transform the data being added to a
// signal, such as to translate from the time domain to the frequency domain.
type Transformer interface {
	// Name shall return the unique name given to the transformer.
	Name() string
	SetName(string)

	// Enabled shall return true if this transformer should actually have an
	// effect on the data being passed through it.
	Enabled() bool
	SetEnabled(bool)

	// Transform shall perform the actual transformations to the data,
	// with the input provided to the src argument and the output in the
	// dst argument, optionally returning additional data as a float array.
	Transform(dst, src []float64) []float64
}

/******************************************************************************
 TransformerBase
******************************************************************************/

type TransformerBase struct {
	enabled bool
	name    string
	rate    float64
}

func (t *TransformerBase) Name() string {
	return t.name
}

func (t *TransformerBase) SetName(name string) {
	t.name = name
}

func (t *TransformerBase) Enabled() bool {
	return t.enabled
}

func (t *TransformerBase) SetEnabled(enabled bool) {
	t.enabled = enabled
}

func (t *TransformerBase) Transform(dst, src []float64) []float64 {
	copy(dst, src)
	return dst
}

func NewTransformerBase(name string) *TransformerBase {
	return &TransformerBase{
		enabled: true,
		name:    name,
	}
}

/******************************************************************************
 FastFourierTransformer
******************************************************************************/

type FastFourierTransformer struct {
	TransformerBase

	dataSize int

	sampleRate   float64
	fft          *fourier.FFT
	resolution   float64
	coefficients []complex128
	freqDomain   []float64
}

func (t *TransformerBase) SampleRate() float64 {
	return t.rate
}

func (t *FastFourierTransformer) SetSampleRate(rate float64) {
	t.sampleRate = rate
	t.resolution = rate / float64(t.dataSize)
}

func (t *FastFourierTransformer) Transform(dst, src []float64) []float64 {
	if !t.enabled {
		return nil
	}

	t.fft.Coefficients(t.coefficients, src)

	for i := 0; i < len(t.coefficients)-1; i++ {
		coeff := t.coefficients[i]
		mag := cmplx.Abs(coeff)
		freq := float64(i) * t.resolution
		idx1 := i * 2
		idx2 := i*2 + 1
		dst[idx1] = mag
		dst[idx2] = mag
		t.freqDomain[idx1] = freq
		t.freqDomain[idx2] = freq
	}

	return t.freqDomain
}

func NewFastFourierTransformer(dataSize int, sampleRate float64) *FastFourierTransformer {
	return &FastFourierTransformer{
		TransformerBase: *NewTransformerBase(defaultFftTransformerName),
		sampleRate:      sampleRate,
		dataSize:        dataSize,
		fft:             fourier.NewFFT(dataSize),
		resolution:      sampleRate / float64(dataSize),
		coefficients:    make([]complex128, (dataSize/2)+1),
		freqDomain:      make([]float64, dataSize),
	}
}
