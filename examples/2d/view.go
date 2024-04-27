package main

import (
	"fmt"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"os"
	"path"
	"time"
	"tonysoft.com/gfx"
)

const (
	testImage = "test.png"
)

func label(text string, rgba color.RGBA) *gfx.Label {
	lbl := gfx.NewLabel()
	lbl.
		SetText(text).
		SetFontSize(.1).
		SetColor(rgba).
		SetPositionY(-.7)
	return lbl
}

func tab0() gfx.WindowObject {
	container := gfx.NewView()
	container.SetName("home")

	help1 := gfx.NewLabel()
	help1.SetText("TAB/ARROW: switch views")
	help1.SetFontSize(.04)
	help1.SetPositionY(.4)

	help2 := gfx.NewLabel()
	help2.SetText("PGUP/PGDN: transition views")
	help2.SetFontSize(.04)
	help2.SetPositionY(.2)

	help3 := gfx.NewLabel()
	help3.SetText("HOME: transition to here")
	help3.SetFontSize(.04)
	help3.SetPositionY(0)

	help4 := gfx.NewLabel()
	help4.SetText("F11: toggle fullscreen mode")
	help4.SetFontSize(.04)
	help4.SetPositionY(-.2)

	help5 := gfx.NewLabel()
	help5.SetText("F12: export view to PNG")
	help5.SetFontSize(.04)
	help5.SetPositionY(-.4)

	container.AddChildren(help1, help2, help3, help4, help5)
	return container
}

func tab1() gfx.WindowObject {
	triangle := gfx.NewShape2D()

	container := gfx.NewWindowObject()
	container.SetName("tab1")
	container.AddChild(triangle)
	container.AddChild(label("Triangle / White", gfx.Blue))
	return container
}

func tab2() gfx.WindowObject {
	triangleLine := gfx.NewTriangle(.2)
	triangleLine.SetColor(gfx.Green)

	container := gfx.NewWindowObject()
	container.SetName("tab2")
	container.AddChild(triangleLine)
	container.AddChild(label("Triangle Line / Green", gfx.Blue))
	return container
}

func tab3(w *gfx.Window) gfx.WindowObject {
	quad := gfx.NewQuad()
	quad.SetTexture(gfx.NewTexture2D(testImage, testImage))

	container := gfx.NewWindowObject()
	container.SetName("tab3")
	container.AddChild(quad)
	container.AddChild(label("Quad / Textured", gfx.Blue))
	return container
}

func tab4() gfx.WindowObject {
	square := gfx.NewSquare(.2)
	square.SetColor(gfx.Red)

	container := gfx.NewWindowObject()
	container.SetName("tab4")
	container.AddChild(square)
	container.AddChild(label("Square Line / Red", gfx.Blue))
	return container
}

func tab5() gfx.WindowObject {
	dot := gfx.NewDot()
	dot.SetTexture(gfx.NewTexture2D(testImage, testImage))

	container := gfx.NewWindowObject()
	container.SetName("tab5")
	container.AddChild(dot)
	container.AddChild(label("Dot / Textured", gfx.Blue))
	return container
}

func tab6() gfx.WindowObject {
	circle := gfx.NewCircle(.2)
	circle.SetColor(gfx.Orange)

	container := gfx.NewWindowObject()
	container.SetName("tab6")
	container.AddChild(circle)
	container.AddChild(label("Circle Line / Orange", gfx.Blue))
	return container
}

func tab7() gfx.WindowObject {
	pacman := gfx.NewDot()
	pacman.
		SetLength(.8).
		SetColor(gfx.Yellow).
		SetRotationZ(mgl32.DegToRad(-30))

	container := gfx.NewWindowObject()
	container.SetName("tab7")
	container.AddChild(pacman)
	container.AddChild(label("Pac-Man", gfx.Blue))
	return container
}

func tab8() gfx.WindowObject {
	arc := gfx.NewCircle(.1)
	arc.
		SetLength(.5).
		SetColor(gfx.Orange)

	container := gfx.NewWindowObject()
	container.SetName("tab8")
	container.AddChild(arc)
	container.AddChild(label("Arc / Orange", gfx.Blue))
	return container
}

func tab9() gfx.WindowObject {
	carousel := NewCarousel()

	topLeft := gfx.NewLabel()
	topLeft.SetText("TopLeft anchored, Right aligned")
	topLeft.SetMaintainAspectRatio(false)
	topLeft.SetFillColor(gfx.Purple)
	topLeft.SetColor(gfx.Magenta)
	topLeft.SetScale(mgl32.Vec3{.5, .04})
	topLeft.SetAnchor(gfx.TopLeft)
	topLeft.SetAlignment(gfx.Right)

	bottomCenter := gfx.NewLabel()
	bottomCenter.SetText("BottomCenter anchored, Center aligned")
	bottomCenter.SetMaintainAspectRatio(false)
	bottomCenter.SetFillColor(gfx.Purple)
	bottomCenter.SetColor(gfx.Magenta)
	bottomCenter.SetScale(mgl32.Vec3{1, .04})
	bottomCenter.SetAnchor(gfx.BottomCenter)
	bottomCenter.SetAlignment(gfx.Centered)

	container := gfx.NewWindowObject()
	container.SetName("tab9")
	container.AddChildren(carousel, topLeft, bottomCenter)
	return container
}

func New2DView(window *gfx.Window) gfx.WindowObject {
	if imgBytes, err := os.ReadFile("test.png"); err == nil {
		gfx.Assets.Add(gfx.NewTexture2D(testImage, imgBytes))
	} else {
		panic(err)
	}

	tabGroup := gfx.NewTabGroup()
	tabGroup.AddChild(tab0())
	tabGroup.AddChild(tab1())
	tabGroup.AddChild(tab2())
	tabGroup.AddChild(tab3(window))
	tabGroup.AddChild(tab4())
	tabGroup.AddChild(tab5())
	tabGroup.AddChild(tab6())
	tabGroup.AddChild(tab7())
	tabGroup.AddChild(tab8())
	tabGroup.AddChild(tab9())

	window.AddKeyEventHandler(glfw.KeyPageUp, glfw.Press, func(window *gfx.Window, key glfw.Key, action glfw.Action) {
		tabGroup.TransitionPrevious()
	})

	window.AddKeyEventHandler(glfw.KeyPageDown, glfw.Press, func(window *gfx.Window, key glfw.Key, action glfw.Action) {
		tabGroup.TransitionNext()
	})

	window.AddKeyEventHandler(glfw.KeyHome, glfw.Press, func(window *gfx.Window, key glfw.Key, action glfw.Action) {
		tabGroup.Transition(tabGroup.IndexOf("home"))
	})

	window.AddKeyEventHandler(glfw.KeyF12, glfw.Press, func(window *gfx.Window, key glfw.Key, action glfw.Action) {
		go func() {
			pngBytes := window.ToPNG()
			if homeDir, err := os.UserHomeDir(); err != nil {
				panic(fmt.Errorf("could not determine your home directory: %w", err))
			} else {
				if err = os.WriteFile(path.Join(homeDir, fmt.Sprintf("gfx_window_%d.png", time.Now().UnixMilli())), pngBytes, 0664); err != nil {
					panic(fmt.Errorf("failed to save PNG to storage: %w", err))
				}
			}
		}()
	})

	return tabGroup
}
