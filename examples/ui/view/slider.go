package view

import (
	"context"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/examples/ui/signal"
)

func newRawSignalView(window *gfx.Window, signalSampleCount int, signalSource *signal.Source) gfx.WindowObject {
	const rateMax = 500.0
	const loFreqMax = 5.0
	const hiFreqMax = 30.0

	defaultRate := 250.0
	defaultLoFreq := 1.0
	defaultHiFreq := 10.0

	signalSource.SetSampleRate(defaultRate)
	signalSource.SetFrequencyComponent1(defaultLoFreq)
	signalSource.SetFrequencyComponent2(defaultHiFreq)

	signalLine := gfx.NewSignalLine("Raw", signalSampleCount)
	signalLine.Label().SetFontSize(.2).SetMarginLeft(.01)
	signalLine.SetAnchor(gfx.TopCenter)
	signalLine.SetMarginLeft(.3)
	signalLine.SetColor(gfx.Purple)
	signalLine.SetThickness(2)
	signalLine.SetScale(mgl32.Vec3{.6, .2})
	signalLine.SetBorderColor(gfx.White)
	signalLine.SetBorderThickness(.02)
	signalLine.SetFillColor(gfx.Darken(gfx.DarkGray, .7))
	signalLine.EnableInspector()
	signalLine.SetInspectorAnchor(gfx.TopRight)

	controls := gfx.NewView()
	controls.SetScaleX(.3)
	controls.SetScaleY(.2)
	controls.SetPositionX(-signalLine.WorldScale().X())
	controls.SetPositionY(1 - controls.WorldScale().Y())
	controls.OnResize(func(oldWidth, oldHeight, newWidth, newHeight int32) {
		controls.SetPositionX(-signalLine.WorldScale().X())
	})

	controlsLabel := gfx.NewLabel()
	controlsLabel.SetFontSize(.1)
	controlsLabel.SetAnchor(gfx.MiddleLeft)
	controlsLabel.SetPosition(mgl32.Vec3{window.ScaleX(-controls.WorldScale().X()), window.ScaleY(controls.WorldScale().Y() + .03)})
	controls.OnResize(func(_, _, _, _ int32) {
		controlsLabel.SetPosition(mgl32.Vec3{window.ScaleX(-controls.WorldScale().X()), window.ScaleY(controls.WorldScale().Y() + .03)})
	})

	updateControlsLabel := func() {
		controlsLabel.SetText(fmt.Sprintf("%.0fHz + %.0fHz signals @ %.0fHz sample rate",
			signalSource.FrequencyComponent1(),
			signalSource.FrequencyComponent2(),
			signalSource.SampleRate()))
	}
	updateControlsLabel()

	rateSlider := gfx.NewSlider(gfx.Vertical, true)
	loFreqSlider := gfx.NewSlider(gfx.Horizontal, true)
	hiFreqSlider := gfx.NewSlider(gfx.Horizontal, true)
	controls.AddChildren(controlsLabel, rateSlider, loFreqSlider, hiFreqSlider)

	rateSlider.SetScaleY(1)
	rateSlider.SetScaleX(.3333)
	rateSlider.SetAnchor(gfx.MiddleLeft)
	rateSlider.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		rate := (value * rateMax) + 20.0
		if rate > rateMax {
			rate = rateMax
		}
		signalSource.SetSampleRate(float64(rate))
		updateControlsLabel()
	})
	rateSlider.SetValue(float32(defaultRate / rateMax))

	loFreqSlider.SetScaleX(.6666)
	loFreqSlider.SetScaleY(.5)
	loFreqSlider.SetAnchor(gfx.MiddleLeft)
	loFreqSlider.SetMargin(gfx.Margin{Left: rateSlider.WorldScale().X() * 2, Bottom: loFreqSlider.WorldScale().Y()})
	loFreqSlider.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		freq := (value * loFreqMax) + 1.0
		if freq > loFreqMax {
			freq = loFreqMax
		}
		signalSource.SetFrequencyComponent1(float64(freq))
		updateControlsLabel()
	})
	loFreqSlider.SetValue(float32(defaultLoFreq / loFreqMax))

	hiFreqSlider.SetScaleX(.6666)
	hiFreqSlider.SetScaleY(.5)
	hiFreqSlider.SetAnchor(gfx.MiddleLeft)
	hiFreqSlider.SetMargin(gfx.Margin{Left: rateSlider.WorldScale().X() * 2, Top: hiFreqSlider.WorldScale().Y()})
	hiFreqSlider.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		freq := (value * hiFreqMax) + 0.0001
		if freq > hiFreqMax {
			freq = hiFreqMax
		}
		signalSource.SetFrequencyComponent2(float64(freq))
		updateControlsLabel()
	})
	hiFreqSlider.SetValue(float32(defaultHiFreq / hiFreqMax))

	go func() {
		outputChan := signalSource.GetOutputChan()
		for data := range outputChan {
			signalLine.AddSamples(data)
		}
	}()

	view := gfx.NewView()
	view.SetPositionY(-.08)
	view.AddChildren(controls, signalLine)

	return view
}

func newLowPassFilterView(window *gfx.Window, signalSampleCount int, signalSource *signal.Source) gfx.WindowObject {
	const rateMax = 500.0
	const cutoffMax = 100.0
	const coeffCountMax = 500

	defaultRate := 250.0
	defaultCutoff := 2.0
	defaultCoeffCount := 60

	filter := gfx.NewLowPassFilter()
	filter.GenerateCoefficients(defaultCoeffCount, defaultRate, defaultCutoff)

	signalLine := gfx.NewSignalLine("Low-Pass", signalSampleCount)
	signalLine.Label().SetFontSize(.2).SetMarginLeft(.01)
	signalLine.SetAnchor(gfx.TopCenter)
	signalLine.SetMarginLeft(.3)
	signalLine.SetColor(gfx.Red)
	signalLine.SetThickness(2)
	signalLine.SetScale(mgl32.Vec3{.6, .2})
	signalLine.SetBorderColor(gfx.White)
	signalLine.SetBorderThickness(.02)
	signalLine.SetFillColor(gfx.Darken(gfx.DarkGray, .7))
	signalLine.EnableInspector()
	signalLine.SetInspectorAnchor(gfx.TopRight)
	signalLine.AddFilter(filter)

	controls := gfx.NewView()
	controls.SetScaleX(.3)
	controls.SetScaleY(.2)
	controls.SetPositionX(-signalLine.WorldScale().X())
	controls.SetPositionY(1 - controls.WorldScale().Y())
	controls.OnResize(func(oldWidth, oldHeight, newWidth, newHeight int32) {
		controls.SetPositionX(-signalLine.WorldScale().X())
	})

	controlsLabel := gfx.NewLabel()
	controlsLabel.SetFontSize(.1)
	controlsLabel.SetAnchor(gfx.MiddleLeft)
	controlsLabel.SetPosition(mgl32.Vec3{window.ScaleX(-controls.WorldScale().X()), window.ScaleY(controls.WorldScale().Y() + .03)})
	controls.OnResize(func(_, _, _, _ int32) {
		controlsLabel.SetPosition(mgl32.Vec3{window.ScaleX(-controls.WorldScale().X()), window.ScaleY(controls.WorldScale().Y() + .03)})
	})

	updateControlsLabel := func() {
		controlsLabel.SetText(fmt.Sprintf("Rate: %.0fHz, Cutoff: %.1fHz, Coeff: %d",
			filter.SampleRate(),
			filter.CutoffFrequencies()[0],
			filter.CoefficientCount()))
	}
	updateControlsLabel()

	rateSlider := gfx.NewSlider(gfx.Vertical)
	cutoffSlider := gfx.NewSlider(gfx.Vertical)
	coeffSlider := gfx.NewSlider(gfx.Vertical)
	controls.AddChildren(controlsLabel, rateSlider, cutoffSlider, coeffSlider)

	rateSlider.SetScaleX(.3333)
	rateSlider.SetScaleY(1)
	rateSlider.SetAnchor(gfx.MiddleLeft)
	rateSlider.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		rate := (value * rateMax) + 20.0
		if rate > rateMax {
			rate = rateMax
		}
		filter.GenerateCoefficients(filter.CoefficientCount(), float64(rate), filter.CutoffFrequencies()...)
		updateControlsLabel()
	})
	rateSlider.SetValue(float32(defaultRate / rateMax))

	cutoffSlider.SetScaleX(.3333)
	cutoffSlider.SetScaleY(1)
	cutoffSlider.SetAnchor(gfx.MiddleLeft)
	cutoffSlider.SetMarginLeft(rateSlider.WorldScale().X() * 2)
	cutoffSlider.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		cutoff := (value * cutoffMax) + 0.0001
		if cutoff > cutoffMax {
			cutoff = cutoffMax
		}
		filter.GenerateCoefficients(filter.CoefficientCount(), filter.SampleRate(), float64(cutoff))
		updateControlsLabel()
	})
	cutoffSlider.SetValue(float32(defaultCutoff / cutoffMax))

	coeffSlider.SetScaleX(.3333)
	coeffSlider.SetScaleY(1)
	coeffSlider.SetAnchor(gfx.MiddleLeft)
	coeffSlider.SetMarginLeft(rateSlider.WorldScale().X() * 4)
	coeffSlider.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		coeff := (value * coeffCountMax) + 1
		if coeff > coeffCountMax {
			coeff = coeffCountMax
		}
		filter.GenerateCoefficients(int(coeff), filter.SampleRate(), filter.CutoffFrequencies()...)
		updateControlsLabel()
	})
	coeffSlider.SetValue(float32(defaultCoeffCount) / float32(coeffCountMax))

	go func() {
		outputChan := signalSource.GetOutputChan()
		for data := range outputChan {
			signalLine.AddSamples(data)
		}
	}()

	view := gfx.NewView()
	view.SetPositionY(-.58)
	view.AddChildren(controls, signalLine)

	return view
}

func newHighPassFilterView(window *gfx.Window, signalSampleCount int, signalSource *signal.Source) gfx.WindowObject {
	const rateMax = 500.0
	const cutoffMax = 100.0
	const coeffCountMax = 500

	defaultRate := 250.0
	defaultCutoff := 9.0
	defaultCoeffCount := 60

	filter := gfx.NewHighPassFilter()
	filter.GenerateCoefficients(defaultCoeffCount, defaultRate, defaultCutoff)

	signalLine := gfx.NewSignalLine("High-Pass", signalSampleCount)
	signalLine.Label().SetFontSize(.2).SetMarginLeft(.01)
	signalLine.SetAnchor(gfx.TopCenter)
	signalLine.SetMarginLeft(.3)
	signalLine.SetColor(gfx.Blue)
	signalLine.SetThickness(2)
	signalLine.SetScale(mgl32.Vec3{.6, .2})
	signalLine.SetBorderColor(gfx.White)
	signalLine.SetBorderThickness(.02)
	signalLine.SetFillColor(gfx.Darken(gfx.DarkGray, .7))
	signalLine.EnableInspector()
	signalLine.SetInspectorAnchor(gfx.TopRight)
	signalLine.AddFilter(filter)

	controls := gfx.NewView()
	controls.SetScaleX(.3)
	controls.SetScaleY(.2)
	controls.SetPositionX(-signalLine.WorldScale().X())
	controls.SetPositionY(1 - controls.WorldScale().Y())
	controls.OnResize(func(oldWidth, oldHeight, newWidth, newHeight int32) {
		controls.SetPositionX(-signalLine.WorldScale().X())
	})

	controlsLabel := gfx.NewLabel()
	controlsLabel.SetFontSize(.1)
	controlsLabel.SetAnchor(gfx.MiddleLeft)
	controlsLabel.SetPosition(mgl32.Vec3{window.ScaleX(-controls.WorldScale().X()), window.ScaleY(controls.WorldScale().Y() + .03)})
	controls.OnResize(func(_, _, _, _ int32) {
		controlsLabel.SetPosition(mgl32.Vec3{window.ScaleX(-controls.WorldScale().X()), window.ScaleY(controls.WorldScale().Y() + .03)})
	})

	updateControlsLabel := func() {
		controlsLabel.SetText(fmt.Sprintf("Rate: %.0fHz, Cutoff: %.1fHz, Coeff: %d",
			filter.SampleRate(),
			filter.CutoffFrequencies()[0],
			filter.CoefficientCount()))
	}
	updateControlsLabel()

	rateSlider := gfx.NewSlider(gfx.Vertical)
	cutoffSlider := gfx.NewSlider(gfx.Vertical)
	coeffSlider := gfx.NewSlider(gfx.Vertical)
	controls.AddChildren(controlsLabel, rateSlider, cutoffSlider, coeffSlider)

	rateSlider.SetScaleX(.3333)
	rateSlider.SetScaleY(1)
	rateSlider.SetAnchor(gfx.MiddleLeft)
	rateSlider.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		rate := (value * rateMax) + 20.0
		if rate > rateMax {
			rate = rateMax
		}
		filter.GenerateCoefficients(filter.CoefficientCount(), float64(rate), filter.CutoffFrequencies()...)
		updateControlsLabel()
	})
	rateSlider.SetValue(float32(defaultRate / rateMax))

	cutoffSlider.SetScaleX(.3333)
	cutoffSlider.SetScaleY(1)
	cutoffSlider.SetAnchor(gfx.MiddleLeft)
	cutoffSlider.SetMarginLeft(rateSlider.WorldScale().X() * 2)
	cutoffSlider.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		cutoff := (value * cutoffMax) + 0.0001
		if cutoff > cutoffMax {
			cutoff = cutoffMax
		}
		filter.GenerateCoefficients(filter.CoefficientCount(), filter.SampleRate(), float64(cutoff))
		updateControlsLabel()
	})
	cutoffSlider.SetValue(float32(defaultCutoff / cutoffMax))

	coeffSlider.SetScaleX(.3333)
	coeffSlider.SetScaleY(1)
	coeffSlider.SetAnchor(gfx.MiddleLeft)
	coeffSlider.SetMarginLeft(rateSlider.WorldScale().X() * 4)
	coeffSlider.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		coeff := (value * coeffCountMax) + 1
		if coeff > coeffCountMax {
			coeff = coeffCountMax
		}
		filter.GenerateCoefficients(int(coeff), filter.SampleRate(), filter.CutoffFrequencies()...)
		updateControlsLabel()
	})
	coeffSlider.SetValue(float32(defaultCoeffCount) / float32(coeffCountMax))

	go func() {
		outputChan := signalSource.GetOutputChan()
		for data := range outputChan {
			signalLine.AddSamples(data)
		}
	}()

	view := gfx.NewView()
	view.SetPositionY(-1.08)
	view.AddChildren(controls, signalLine)

	return view
}

func newBandPassFilterView(window *gfx.Window, signalSampleCount int, signalSource *signal.Source) gfx.WindowObject {
	const rateMax = 500.0
	const cutoffMax = 100.0
	const coeffCountMax = 500

	defaultRate := 250.0
	defaultCutoffLo := 0.2
	defaultCutoffHi := 1.8
	defaultCoeffCount := 400

	filter := gfx.NewBandPassFilter()
	filter.GenerateCoefficients(defaultCoeffCount, defaultRate, defaultCutoffLo, defaultCutoffHi)

	signalLine := gfx.NewSignalLine("Band-Pass", signalSampleCount)
	signalLine.Label().SetFontSize(.2).SetMarginLeft(.01)
	signalLine.SetAnchor(gfx.TopCenter)
	signalLine.SetMarginLeft(.3)
	signalLine.SetColor(gfx.Red)
	signalLine.SetThickness(2)
	signalLine.SetScale(mgl32.Vec3{.6, .2})
	signalLine.SetBorderColor(gfx.White)
	signalLine.SetBorderThickness(.02)
	signalLine.SetFillColor(gfx.Darken(gfx.DarkGray, .7))
	signalLine.EnableInspector()
	signalLine.SetInspectorAnchor(gfx.TopRight)
	signalLine.AddFilter(filter)

	controls := gfx.NewView()
	controls.SetScaleX(.4)
	controls.SetScaleY(.2)
	controls.SetPositionX(-signalLine.WorldScale().X() - window.ScaleX(.1))
	controls.SetPositionY(1 - controls.WorldScale().Y())
	controls.OnResize(func(oldWidth, oldHeight, newWidth, newHeight int32) {
		controls.SetPositionX(-signalLine.WorldScale().X() - window.ScaleX(.1))
	})

	controlsLabel := gfx.NewLabel()
	controlsLabel.SetFontSize(.1)
	controlsLabel.SetAnchor(gfx.MiddleLeft)
	controlsLabel.SetPosition(mgl32.Vec3{window.ScaleX(-controls.WorldScale().X()), window.ScaleY(controls.WorldScale().Y() + .03)})
	controls.OnResize(func(_, _, _, _ int32) {
		controlsLabel.SetPosition(mgl32.Vec3{window.ScaleX(-controls.WorldScale().X()), window.ScaleY(controls.WorldScale().Y() + .03)})
	})

	updateControlsLabel := func() {
		controlsLabel.SetText(fmt.Sprintf("Rate: %.0fHz, LoCut: %.1fHz, HiCut: %.1fHz, Coeff: %d",
			filter.SampleRate(),
			filter.CutoffFrequencies()[0],
			filter.CutoffFrequencies()[1],
			filter.CoefficientCount()))
	}
	updateControlsLabel()

	rateSlider := gfx.NewSlider(gfx.Vertical)
	cutoffSliderLo := gfx.NewSlider(gfx.Vertical)
	cutoffSliderHi := gfx.NewSlider(gfx.Vertical)
	coeffSlider := gfx.NewSlider(gfx.Vertical)
	controls.AddChildren(controlsLabel, rateSlider, cutoffSliderLo, cutoffSliderHi, coeffSlider)

	rateSlider.SetScaleX(.25)
	rateSlider.SetScaleY(1)
	rateSlider.SetAnchor(gfx.MiddleLeft)
	rateSlider.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		rate := (value * rateMax) + 20.0
		if rate > rateMax {
			rate = rateMax
		}
		filter.GenerateCoefficients(filter.CoefficientCount(), float64(rate), filter.CutoffFrequencies()...)
		updateControlsLabel()
	})
	rateSlider.SetValue(float32(defaultRate / rateMax))

	cutoffSliderLo.SetScaleX(.25)
	cutoffSliderLo.SetScaleY(1)
	cutoffSliderLo.SetAnchor(gfx.MiddleLeft)
	cutoffSliderLo.SetMarginLeft(rateSlider.WorldScale().X() * 2)
	cutoffSliderLo.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		cutoff := (value * cutoffMax) + 0.0001
		if cutoff > cutoffMax {
			cutoff = cutoffMax
		}
		filter.GenerateCoefficients(filter.CoefficientCount(), filter.SampleRate(), float64(cutoff), filter.CutoffFrequencies()[1])
		updateControlsLabel()
	})
	cutoffSliderLo.SetValue(float32(defaultCutoffLo / cutoffMax))

	cutoffSliderHi.SetScaleX(.25)
	cutoffSliderHi.SetScaleY(1)
	cutoffSliderHi.SetAnchor(gfx.MiddleLeft)
	cutoffSliderHi.SetMarginLeft(rateSlider.WorldScale().X() * 4)
	cutoffSliderHi.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		cutoff := (value * cutoffMax) + 0.0001
		if cutoff > cutoffMax {
			cutoff = cutoffMax
		}
		filter.GenerateCoefficients(filter.CoefficientCount(), filter.SampleRate(), filter.CutoffFrequencies()[0], float64(cutoff))
		updateControlsLabel()
	})
	cutoffSliderHi.SetValue(float32(defaultCutoffHi / cutoffMax))

	coeffSlider.SetScaleX(.25)
	coeffSlider.SetScaleY(1)
	coeffSlider.SetAnchor(gfx.MiddleLeft)
	coeffSlider.SetMarginLeft(rateSlider.WorldScale().X() * 6)
	coeffSlider.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		coeff := (value * coeffCountMax) + 1
		if coeff > coeffCountMax {
			coeff = coeffCountMax
		}
		filter.GenerateCoefficients(int(coeff), filter.SampleRate(), filter.CutoffFrequencies()...)
		updateControlsLabel()
	})
	coeffSlider.SetValue(float32(defaultCoeffCount) / float32(coeffCountMax))

	go func() {
		outputChan := signalSource.GetOutputChan()
		for data := range outputChan {
			signalLine.AddSamples(data)
		}
	}()

	view := gfx.NewView()
	view.SetPositionY(-1.58)
	view.AddChildren(controls, signalLine)

	return view
}

func NewSliderView(ctx context.Context, window *gfx.Window, signalSampleCount int) gfx.WindowObject {
	source := signal.NewSource()
	source.Run(ctx)

	container := gfx.NewObject(nil)
	container.AddChild(newRawSignalView(window, signalSampleCount, source))
	container.AddChild(newLowPassFilterView(window, signalSampleCount, source))
	container.AddChild(newHighPassFilterView(window, signalSampleCount, source))
	container.AddChild(newBandPassFilterView(window, signalSampleCount, source))

	return container
}
