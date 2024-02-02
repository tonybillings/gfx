package view

import (
	"context"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"math"
	"math/rand"
	"time"
	"tonysoft.com/gfx"
)

func NewSignalsView(ctx context.Context, window *gfx.Window, signalSampleCount int) gfx.WindowObject {
	sg := gfx.NewSignalGroup(signalSampleCount, 3, []color.RGBA{gfx.Green, gfx.Blue, gfx.Orange, gfx.Purple}...)
	sg.EnableInspector()
	sg.SetInspectorAnchor(gfx.TopCenter)
	sg.SetInspectorMargin(gfx.Margin{Top: .05})

	const signalCount = 20

	labels := []string{"Sine / Green", "Square / Blue", "Saw / Orange", "Random / Purple"}
	labelsIdx := 0

	for i := 0; i < signalCount; i++ {
		sg.New(labels[labelsIdx])
		labelsIdx = (labelsIdx + 1) % len(labels)
	}

	window.AddKeyEventHandler(glfw.KeyUp, glfw.Press, func(_ *gfx.Window, _ glfw.Key, _ glfw.Action) {
		if !sg.Enabled() {
			return
		}
		for _, s := range sg.Signals() {
			s.SetThickness(s.Thickness() + 1)
		}
	})

	window.AddKeyEventHandler(glfw.KeyDown, glfw.Press, func(_ *gfx.Window, _ glfw.Key, _ glfw.Action) {
		if !sg.Enabled() {
			return
		}
		for _, s := range sg.Signals() {
			s.SetThickness(s.Thickness() - 1)
		}
	})

	thicknessHelp := gfx.NewLabel()
	thicknessHelp.
		SetText("UP/DOWN: change line thickness").
		SetFontSize(.03)

	container := gfx.NewObject(nil)
	container.AddChild(sg)
	container.AddChild(thicknessHelp)

	go func(ctx context.Context, sg *gfx.SignalGroup) {
		<-window.GetReadyChan()

		x := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			sineValue := math.Sin(float64(x) * .025)
			for i := 0; i < signalCount; i += len(labels) {
				sg.Signals()[i].AddSamples([]float64{sineValue})
			}

			squareValue := 0.0
			if sineValue > 0 {
				squareValue = 1.0
			} else {
				squareValue = -1.0
			}
			for i := 1; i < signalCount; i += len(labels) {
				sg.Signals()[i].AddSamples([]float64{squareValue})
			}

			sawValue := 2.0*float64(x%100)/100.0 - 1.0
			for i := 2; i < signalCount; i += len(labels) {
				sg.Signals()[i].AddSamples([]float64{sawValue})
			}

			for i := 3; i < signalCount; i += len(labels) {
				sg.Signals()[i].AddSamples([]float64{rand.Float64()})
			}

			x++
			time.Sleep(4 * time.Millisecond)
		}
	}(ctx, sg)

	return container
}

func NewSignalInspectorView(ctx context.Context, window *gfx.Window, signalSampleCount int) gfx.WindowObject {
	inspectorHelp1 := gfx.NewLabel()
	inspectorHelp1.
		SetText("CTRL/CMD: show signal inspector").
		SetFontSize(.03).
		SetPositionY(.35)

	inspectorHelp2 := gfx.NewLabel()
	inspectorHelp2.
		SetText("F5 and F6: export signal data to CSV").
		SetFontSize(.03).
		SetPositionY(.25)

	sig := gfx.NewSignalLine("Inspect me!", signalSampleCount)
	sig.Label().SetFontSize(.1).SetPaddingLeft(0.01).SetColor(gfx.Green)
	sig.SetPositionY(.7)
	sig.SetScale(mgl32.Vec3{.9, .25, 1})
	sig.SetBorderColor(gfx.White)
	sig.SetBorderThickness(.02)
	sig.SetFillColor(gfx.Darken(gfx.DarkGray, .7))
	sig.EnableInspector()
	sig.SetInspectorAnchor(gfx.TopRight)
	sig.SetInspectorMargin(gfx.Margin{Top: .05, Right: .05})
	sig.EnableDataExportKey(glfw.KeyF5)

	//sig.SetInspectorFontSize(.1)  // probably don't ever need to call this method
	//sig.SetInspectorPanelSize(.5) // use this method to change the inspector scale
	//sig.SetInspectorMargin(gfx.Margin{
	//	Top:    0.05,
	//	Right:  0.05,
	//	Bottom: 0.05,
	//	Left:   0.05}) // use this to adjust the spacing around the inspector

	// You can customize the inspector using these methods:
	//sig.InspectorPanel().SetFillColor(gfx.Red)
	//sig.InspectorPanel().SetBorderColor(gfx.Green)
	//sig.InspectorPanel().SetBorderThickness(.15)

	// You can also get a reference to the labels like so:
	//sig.InspectorPanel().Child("min").SetColor(gfx.Green)
	// ...other label names: max, avg, std, delta_avg, delta_std, sample

	sg := gfx.NewSignalGroup(signalSampleCount, 3, []color.RGBA{gfx.Blue, gfx.Orange, gfx.Purple}...)
	sg.SetPositionY(-.4)
	sg.SetScale(mgl32.Vec3{.9, .5, 1})
	sg.EnableInspector()
	sg.SetInspectorAnchor(gfx.BottomRight)
	sg.SetInspectorMargin(gfx.Margin{Bottom: .05, Right: .05})
	sg.EnableDataExportKey(glfw.KeyF6)

	signalCount := 3

	labels := []string{"Square", "Saw", "Random"}
	labelsIdx := 0

	for i := 0; i < signalCount; i++ {
		sg.New(labels[labelsIdx])
		labelsIdx = (labelsIdx + 1) % len(labels)
	}

	window.AddKeyEventHandler(glfw.KeyUp, glfw.Press, func(_ *gfx.Window, _ glfw.Key, _ glfw.Action) {
		if !sg.Enabled() {
			return
		}
		for _, s := range sg.Signals() {
			s.SetThickness(s.Thickness() + 1)
		}
	})

	window.AddKeyEventHandler(glfw.KeyDown, glfw.Press, func(_ *gfx.Window, _ glfw.Key, _ glfw.Action) {
		if !sg.Enabled() {
			return
		}
		for _, s := range sg.Signals() {
			s.SetThickness(s.Thickness() - 1)
		}
	})

	go func(ctx context.Context, sg *gfx.SignalGroup) {
		<-window.GetReadyChan()

		x := 0
		randomSample := rand.Float64()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			sineValue := math.Sin(float64(x) * .025)
			sig.AddSamples([]float64{sineValue})

			squareValue := 0.0
			if sineValue > 0 {
				squareValue = 1.0
			} else {
				squareValue = -1.0
			}
			for i := 0; i < signalCount; i += len(labels) {
				sg.Signals()[i].AddSamples([]float64{squareValue})
			}

			sawValue := 2.0*float64(x%100)/100.0 - 1.0
			for i := 1; i < signalCount; i += len(labels) {
				sg.Signals()[i].AddSamples([]float64{sawValue})
			}

			if x%10 == 0 { // we'll add the same sample several times to "stretch out" the signal
				randomSample = rand.Float64()
			}
			for i := 2; i < signalCount; i += len(labels) {
				sg.Signals()[i].AddSamples([]float64{randomSample})
			}

			x++
			time.Sleep(4 * time.Millisecond)
		}
	}(ctx, sg)

	container := gfx.NewObject(nil)
	container.AddChild(inspectorHelp1)
	container.AddChild(inspectorHelp2)
	container.AddChild(sig)
	container.AddChild(sg)

	return container
}
