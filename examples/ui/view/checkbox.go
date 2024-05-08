package view

import (
	"context"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/examples/ui/signal"
)

func newCheckBoxView(window *gfx.Window, signalSampleCount int, signalSource *signal.Source) gfx.WindowObject {
	defaultRate := 250.0
	defaultFreq1 := 1.0
	defaultFreq2 := 10.0
	defaultLoPassCutoff := 2.0
	defaultHiPassCutoff := 9.0
	defaultCoeffCount := 100

	signalSource.SetSampleRate(defaultRate)
	signalSource.SetFrequencyComponent1(defaultFreq1)
	signalSource.SetFrequencyComponent2(defaultFreq2)

	loPassFilter := gfx.NewLowPassFilter()
	loPassFilter.GenerateCoefficients(defaultCoeffCount, defaultRate, defaultLoPassCutoff)
	loPassFilter.SetEnabled(false)

	hiPassFilter := gfx.NewHighPassFilter()
	hiPassFilter.GenerateCoefficients(defaultCoeffCount, defaultRate, defaultHiPassCutoff)
	hiPassFilter.SetEnabled(false)

	signalLine := gfx.NewSignalLine("", signalSampleCount)
	signalLine.Label().SetFontSize(.2).SetPaddingLeft(.01)
	signalLine.SetAnchor(gfx.Center)
	signalLine.SetColor(gfx.Purple)
	signalLine.SetThickness(1)
	signalLine.SetScale(mgl32.Vec3{.9, .4})
	signalLine.SetBorderColor(gfx.White)
	signalLine.SetBorderThickness(.02)
	signalLine.SetFillColor(gfx.Darken(gfx.DarkGray, .7))
	signalLine.EnableInspector()
	signalLine.SetInspectorAnchor(gfx.TopCenter)
	signalLine.SetInspectorMargin(gfx.Margin{Top: .02})
	signalLine.AddFilter(loPassFilter)
	signalLine.AddFilter(hiPassFilter)

	updateSignalLineColor := func() {
		if loPassFilter.Enabled() && hiPassFilter.Enabled() {
			signalLine.SetColor(gfx.Magenta)
		} else if loPassFilter.Enabled() {
			signalLine.SetColor(gfx.Red)
		} else if hiPassFilter.Enabled() {
			signalLine.SetColor(gfx.Blue)
		} else {
			signalLine.SetColor(gfx.Purple)
		}
	}

	loPassCheckBox := gfx.NewCheckBox()
	hiPassCheckBox := gfx.NewCheckBox()

	loPassCheckBox.SetScale(mgl32.Vec3{.033, .033})
	loPassCheckBox.SetText("Low-Pass Filter")
	loPassCheckBox.SetColor(gfx.Red)
	loPassCheckBox.SetTextColor(gfx.Red)
	loPassCheckBox.SetMouseEnterColor(gfx.Lighten(gfx.Red, 100))
	loPassCheckBox.SetPositionY(signalLine.Position().Y() - (signalLine.Scale().Y() + loPassCheckBox.Scale().Y() + 0.05))
	loPassCheckBox.SetPositionX(window.ScaleX(-.7))
	loPassCheckBox.OnCheckedChanged(func(_ gfx.WindowObject, checked bool) {
		loPassFilter.SetEnabled(checked)
		updateSignalLineColor()
	})

	hiPassCheckBox.SetScale(mgl32.Vec3{.033, .033})
	hiPassCheckBox.SetText("High-Pass Filter")
	hiPassCheckBox.SetTextColor(gfx.Blue)
	hiPassCheckBox.SetColor(gfx.Blue)
	hiPassCheckBox.SetMouseEnterColor(gfx.Lighten(gfx.Blue, 80))
	hiPassCheckBox.SetPositionY(signalLine.Position().Y() - (signalLine.Scale().Y() + loPassCheckBox.Scale().Y() + 0.05))
	hiPassCheckBox.SetPositionX(window.ScaleX(.2))
	hiPassCheckBox.OnCheckedChanged(func(_ gfx.WindowObject, checked bool) {
		hiPassFilter.SetEnabled(checked)
		updateSignalLineColor()
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
	view.AddChildren(signalLine, loPassCheckBox, hiPassCheckBox)

	return view
}

func NewCheckBoxView(ctx context.Context, window *gfx.Window, signalSampleCount int) gfx.WindowObject {
	source := signal.NewSource()
	source.Run(ctx)

	container := gfx.NewWindowObject()
	container.AddChild(newCheckBoxView(window, signalSampleCount, source))

	return container
}
