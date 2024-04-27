package _test

import (
	"math"
	"testing"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/_test"
)

func setupSignal() *gfx.Signal {
	return gfx.NewSignal("TestSignal", 10)
}

func TestSignalBasic(t *testing.T) {
	s := setupSignal()
	if s.Label() != "TestSignal" {
		t.Errorf("expected label 'TestSignal', got %s", s.Label())
	}
	if len(s.Data()) != 10 {
		t.Errorf("expected data length 10, got %d", len(s.Data()))
	}
	if s.MinValue() != math.Inf(1) {
		t.Errorf("expected min value +Inf, got %f", s.MinValue())
	}
	if s.MaxValue() != math.Inf(-1) {
		t.Errorf("expected max value -Inf, got %f", s.MaxValue())
	}
}

func TestAddSamples(t *testing.T) {
	s := setupSignal()
	s.AddSamples([]float64{1.0, 2.0, 3.0, 4.0, 5.0})
	if s.Data()[4] != 5.0 {
		t.Errorf("expedcted data point 5.0, got %f", s.Data()[4])
	}
	if s.MinValue() != 1.0 {
		t.Errorf("expected min data 1.0, got %f", s.MinValue())
	}
	if s.MaxValue() != 5.0 {
		t.Errorf("expected max data 5.0, got %f", s.MaxValue())
	}
}

func TestAverage(t *testing.T) {
	s := setupSignal()
	s.AddSamples([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	expectedAverage := 5.5

	avg := s.Average()
	if avg != expectedAverage {
		t.Errorf("expected average %v, got %v", expectedAverage, avg)
	}
}

func TestStdDev(t *testing.T) {
	s := setupSignal()
	s.AddSamples([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	expectedStdDev := math.Sqrt(8.25)

	stdDev := s.StdDev()
	if math.Abs(stdDev-expectedStdDev) > 0.001 {
		t.Errorf("expected standard deviation ~%v, got %v", expectedStdDev, stdDev)
	}
}

func TestDeltaAverage(t *testing.T) {
	s := setupSignal()
	s.AddSamples([]float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100})
	expectedDeltaAvg := 10.0

	deltaAvg := s.DeltaAverage()
	if deltaAvg != expectedDeltaAvg {
		t.Errorf("expected delta average %v, got %v", expectedDeltaAvg, deltaAvg)
	}
}

func TestDeltaStdDev(t *testing.T) {
	s := setupSignal()
	s.AddSamples([]float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100})
	expectedDeltaStdDev := 0.0

	deltaStdDev := s.DeltaStdDev(true)
	if deltaStdDev != expectedDeltaStdDev {
		t.Errorf("expected delta standard deviation %v, got %v", expectedDeltaStdDev, deltaStdDev)
	}
}

func TestSignalWithFilter(t *testing.T) {
	s := setupSignal()
	s.AddFilter(_test.NewOddValueFilter())

	s.AddSamples([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	if len(s.FilteredData()) != 10 {
		t.Errorf("filtered data length mismatch: expected 10, got %d", len(s.FilteredData()))
	}

	for i, v := range s.Data() {
		if int(v) != i+1 {
			t.Errorf("expected data value %d, got %d", i+1, int(v))
		}
	}

	for _, v := range s.FilteredData() {
		if int(v)%2 == 1 {
			t.Errorf("expected only even values, got %f", v)
		}
	}
}

func TestSignalWithTransformer(t *testing.T) {
	s := setupSignal()
	s.AddTransformer(_test.NewPlusNTransformer())

	s.AddSamples([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	if len(s.TransformedData()) != 10 {
		t.Errorf("transformed data length mismatch: expected 10, got %d", len(s.FilteredData()))
	}

	for i, v := range s.Data() {
		if int(v) != i+1 {
			t.Errorf("expected data value %d, got %d", i+1, int(v))
		}
	}

	for i, v := range s.TransformedData() {
		tV := int(s.Data()[i]) + int(s.TransformedDataLabels()[i])
		if int(v) != tV {
			t.Errorf("expected transformed data value %d, got %d", tV, int(v))
		}
	}
}
