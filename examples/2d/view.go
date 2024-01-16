package main

import (
	"context"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"math"
	"math/rand"
	"os"
	"time"
	"tonysoft.com/gfx"
)

func label(text string, rgba color.RGBA) *gfx.Label {
	lbl := gfx.NewLabel()
	lbl.
		SetText(text).
		SetColor(rgba).
		SetScale(mgl32.Vec3{.1, .1, .1}).
		SetPositionY(-.7)
	return lbl
}

func tab0() gfx.WindowObject {
	lbl := gfx.NewLabel()
	lbl.
		SetText("Use TAB/ARROW keys to navigate").
		SetFontSize(.03)
	return lbl
}

func tab1() gfx.WindowObject {
	triangle := gfx.NewShape()

	container := gfx.NewObject(nil)
	container.AddChild(triangle)
	container.AddChild(label("Triangle / White", gfx.Blue))
	return container
}

func tab2() gfx.WindowObject {
	triangleLine := gfx.NewTriangle(.2)
	triangleLine.SetColor(gfx.Green)

	container := gfx.NewObject(nil)
	container.AddChild(triangleLine)
	container.AddChild(label("Triangle Line / Green", gfx.Blue))
	return container
}

func tab3() gfx.WindowObject {
	quad := gfx.NewQuad()
	quad.SetTexture("test.png")

	container := gfx.NewObject(nil)
	container.AddChild(quad)
	container.AddChild(label("Quad / Textured", gfx.Blue))
	return container
}

func tab4() gfx.WindowObject {
	square := gfx.NewSquare(.2)
	square.SetColor(gfx.Red)

	container := gfx.NewObject(nil)
	container.AddChild(square)
	container.AddChild(label("Square Line / Red", gfx.Blue))
	return container
}

func tab5() gfx.WindowObject {
	dot := gfx.NewDot()
	dot.SetTexture("test.png")

	container := gfx.NewObject(nil)
	container.AddChild(dot)
	container.AddChild(label("Dot / Textured", gfx.Blue))
	return container
}

func tab6() gfx.WindowObject {
	circle := gfx.NewCircle(.2)
	circle.SetColor(gfx.Orange)

	container := gfx.NewObject(nil)
	container.AddChild(circle)
	container.AddChild(label("Circle Line / Orange", gfx.Blue))
	return container
}

func tab7() gfx.WindowObject {
	pacman := gfx.NewDot()
	pacman.
		SetLength(.8).
		SetRotationZ(mgl32.DegToRad(-30)).
		SetColor(gfx.Yellow)

	container := gfx.NewObject(nil)
	container.AddChild(pacman)
	container.AddChild(label("Pac-Man", gfx.Blue))
	return container
}

func tab8() gfx.WindowObject {
	arc := gfx.NewCircle(.1)
	arc.
		SetLength(.5).
		SetColor(gfx.Orange)

	container := gfx.NewObject(nil)
	container.AddChild(arc)
	container.AddChild(label("Arc / Orange", gfx.Blue))
	return container
}

func tab9(ctx context.Context, window *gfx.Window) gfx.WindowObject {
	sg := gfx.NewSignalGroup(signalSampleCount, 3, []color.RGBA{gfx.Green, gfx.Blue, gfx.Orange, gfx.Purple}...)
	signalCount := 20

	thicknessHelp := gfx.NewLabel()
	thicknessHelp.
		SetText("UP/DOWN: change line thickness").
		SetFontSize(.03)

	container := gfx.NewObject(nil)
	container.AddChild(sg)
	container.AddChild(thicknessHelp)

	go func(ctx context.Context, sg *gfx.SignalGroup) {
		labels := []string{"Sine / Green", "Square / Blue", "Saw / Orange", "Random / Purple"}
		labelsIdx := 0

		for i := 0; i < signalCount; i++ {
			sg.New(labels[labelsIdx])
			labelsIdx = (labelsIdx + 1) % len(labels)
		}

		window.AddKeyEventHandler(glfw.KeyUp, glfw.Press, func(_ *gfx.Window, _ glfw.Key, _ glfw.Action) {
			for _, s := range sg.Signals() {
				s.SetThickness(s.Thickness() + 1)
			}
		})

		window.AddKeyEventHandler(glfw.KeyDown, glfw.Press, func(_ *gfx.Window, _ glfw.Key, _ glfw.Action) {
			for _, s := range sg.Signals() {
				s.SetThickness(s.Thickness() - 1)
			}
		})

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

func NewTestView(ctx context.Context, window *gfx.Window) gfx.WindowObject {
	if texBytes, err := os.ReadFile("test.png"); err == nil {
		gfx.AddAsset("test.png", texBytes)
	} else {
		panic(err)
	}

	tabGroup := gfx.NewTabGroup()
	tabGroup.AddChild(tab0())
	tabGroup.AddChild(tab1())
	tabGroup.AddChild(tab2())
	tabGroup.AddChild(tab3())
	tabGroup.AddChild(tab4())
	tabGroup.AddChild(tab5())
	tabGroup.AddChild(tab6())
	tabGroup.AddChild(tab7())
	tabGroup.AddChild(tab8())
	tabGroup.AddChild(tab9(ctx, window))
	return tabGroup
}
