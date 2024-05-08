package view

import (
	"context"
	"fmt"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/examples/ui/signal"
)

func newExtendedRawSignalView(window *gfx.Window, signalSampleCount int, signalSource *signal.Source) gfx.WindowObject {
	const rateMax = 500.0
	const freqMax = 100.0

	defaultRate := 250.0
	defaultFreq1 := 20.0
	defaultFreq2 := 40.0
	defaultFreq3 := 80.0

	freq1 := defaultFreq1
	freq2 := defaultFreq2
	freq3 := defaultFreq3

	signalSource.SetSampleRate(defaultRate)
	signalSource.SetFrequencyComponent1(defaultFreq1)
	signalSource.SetFrequencyComponent2(defaultFreq2)
	signalSource.SetFrequencyComponent3(defaultFreq3)

	signalLine := gfx.NewSignalLine("", signalSampleCount)
	signalLine.Label().SetPaddingLeft(.01)
	signalLine.SetAnchor(gfx.Center)
	signalLine.SetMarginLeft(.3)
	signalLine.SetColor(gfx.Orange)
	signalLine.SetThickness(1)
	signalLine.SetScale(mgl32.Vec3{.6, .4})
	signalLine.SetBorderColor(gfx.White)
	signalLine.SetBorderThickness(.02)
	signalLine.SetFillColor(gfx.Darken(gfx.DarkGray, .7))
	signalLine.EnableInspector()
	signalLine.SetInspectorAnchor(gfx.TopCenter)
	signalLine.SetInspectorMargin(gfx.Margin{Top: .02})

	controls := gfx.NewView()
	controls.SetScaleX(.3)
	controls.SetScaleY(.4)
	controls.SetPositionX(-signalLine.WorldScale().X())
	controls.OnResize(func(_, _ int) {
		controls.SetPositionX(-signalLine.WorldScale().X())
	})

	controlsLabel := gfx.NewLabel()
	controlsLabel.SetAnchor(gfx.TopLeft)
	controlsLabel.SetAlignment(gfx.Left)
	controlsLabel.SetScaleX(1 / window.ScaleX(controls.WorldScale().X()))
	controlsLabel.SetFontSize(.05)
	controlsLabel.SetPaddingBottom(.05)

	updateControlsLabel := func() {
		controlsLabel.SetText(fmt.Sprintf("%.0fHz + %.0fHz + %.0fHz signals @ %.0fHz sample rate",
			signalSource.FrequencyComponent1(),
			signalSource.FrequencyComponent2(),
			signalSource.FrequencyComponent3(),
			signalSource.SampleRate()))
	}
	updateControlsLabel()

	rateSlider := gfx.NewSlider(gfx.Vertical, false)
	freqSlider1 := gfx.NewSlider(gfx.Horizontal, true)
	freqSlider2 := gfx.NewSlider(gfx.Horizontal, true)
	freqSlider3 := gfx.NewSlider(gfx.Horizontal, true)
	controls.AddChildren(controlsLabel, rateSlider, freqSlider1, freqSlider2, freqSlider3)

	rateSlider.SetScaleY(1)
	rateSlider.SetScaleX(.3333)
	rateSlider.SetAnchor(gfx.MiddleLeft)
	rateSlider.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		rate := float64((value * rateMax) + 20.0)
		if rate > rateMax {
			rate = rateMax
		}

		signalSource.SetSampleRate(rate)
		signalLine.SetFFTSampleRate(rate)
		updateControlsLabel()
	})
	rateSlider.SetValue(float32(defaultRate / rateMax))

	freqSlider1.SetScaleX(.6666)
	freqSlider1.SetScaleY(.3333)
	freqSlider1.SetAnchor(gfx.MiddleLeft)
	freqSlider1.SetMargin(gfx.Margin{Left: rateSlider.WorldScale().X() * 2, Bottom: freqSlider1.WorldScale().Y() * 2})
	freqSlider1.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		freq1 = float64((value * freqMax) + 1.0)
		if freq1 > freqMax {
			freq1 = freqMax
		}
		signalSource.SetFrequencyComponent1(freq1)
		updateControlsLabel()
	})
	freqSlider1.SetValue(float32(defaultFreq1 / freqMax))

	freqSlider2.SetScaleX(.6666)
	freqSlider2.SetScaleY(.3333)
	freqSlider2.SetAnchor(gfx.MiddleLeft)
	freqSlider2.SetMargin(gfx.Margin{Left: rateSlider.WorldScale().X() * 2})
	freqSlider2.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		freq2 = float64((value * freqMax) + 0.0001)
		if freq2 > freqMax {
			freq2 = freqMax
		}
		signalSource.SetFrequencyComponent2(freq2)
		updateControlsLabel()
	})
	freqSlider2.SetValue(float32(defaultFreq2 / freqMax))

	freqSlider3.SetScaleX(.6666)
	freqSlider3.SetScaleY(.3333)
	freqSlider3.SetAnchor(gfx.MiddleLeft)
	freqSlider3.SetMargin(gfx.Margin{Left: rateSlider.WorldScale().X() * 2, Top: freqSlider3.WorldScale().Y() * 2})
	freqSlider3.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		freq3 = float64((value * freqMax) + 0.0001)
		if freq3 > freqMax {
			freq3 = freqMax
		}
		signalSource.SetFrequencyComponent3(freq3)
		updateControlsLabel()
	})
	freqSlider3.SetValue(float32(defaultFreq3 / freqMax))

	timeDomainButton := gfx.NewRadioButton()
	freqDomainButton := gfx.NewRadioButton()

	timeDomainButton.SetScale(mgl32.Vec3{.033, .033})
	timeDomainButton.SetText("Time Domain")
	timeDomainButton.SetColor(gfx.Orange)
	timeDomainButton.SetTextColor(gfx.Orange)
	timeDomainButton.SetMouseEnterColor(gfx.Darken(gfx.Orange, .4))
	timeDomainButton.SetPositionY(signalLine.Position().Y() - (signalLine.Scale().Y() + timeDomainButton.Scale().Y() + 0.05))
	timeDomainButton.SetPositionX(signalLine.Scale().X() * -.3)
	timeDomainButton.SetChecked(true)
	timeDomainButton.OnCheckedChanged(func(_ gfx.WindowObject, checked bool) {
		if checked {
			signalLine.DisableFFT()
			signalLine.SetColor(gfx.Orange)
			freqDomainButton.SetChecked(false)
		}
	})

	freqDomainButton.SetScale(mgl32.Vec3{.033, .033})
	freqDomainButton.SetText("Frequency Domain")
	freqDomainButton.SetTextColor(gfx.Green)
	freqDomainButton.SetColor(gfx.Green)
	freqDomainButton.SetMouseEnterColor(gfx.Darken(gfx.Green, .4))
	freqDomainButton.SetPositionY(signalLine.Position().Y() - (signalLine.Scale().Y() + timeDomainButton.Scale().Y() + 0.05))
	freqDomainButton.SetPositionX(signalLine.Scale().X() * .3)
	freqDomainButton.OnCheckedChanged(func(_ gfx.WindowObject, checked bool) {
		if checked {
			signalLine.EnableFFT(defaultRate)
			signalLine.SetColor(gfx.Green)
			timeDomainButton.SetChecked(false)
		}
	})

	window.AddKeyEventHandler(signalLine, glfw.KeyUp, glfw.Press, func(_ *gfx.Window, _ glfw.Key, _ glfw.Action) {
		if !signalLine.Enabled() {
			return
		}

		signalLine.SetThickness(signalLine.Thickness() + 1)
	})

	window.AddKeyEventHandler(signalLine, glfw.KeyDown, glfw.Press, func(_ *gfx.Window, _ glfw.Key, _ glfw.Action) {
		if !signalLine.Enabled() {
			return
		}

		signalLine.SetThickness(signalLine.Thickness() - 1)
	})

	go func() {
		outputChan := signalSource.GetOutputChan()
		for data := range outputChan {
			signalLine.AddSamples(data)
		}
	}()

	view := gfx.NewView()
	view.AddChildren(controls, signalLine, timeDomainButton, freqDomainButton)

	return view
}

func NewRadioButtonView(ctx context.Context, window *gfx.Window, signalSampleCount int) gfx.WindowObject {
	source := signal.NewSource()
	source.Run(ctx)

	container := gfx.NewWindowObject()
	container.AddChild(newExtendedRawSignalView(window, signalSampleCount, source))

	return container
}
