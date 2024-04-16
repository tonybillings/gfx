package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"os"
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

	help1 := gfx.NewLabel()
	help1.SetText("TAB/ARROW: switch views")
	help1.SetFontSize(.04)
	help1.SetPositionY(.1)

	help2 := gfx.NewLabel()
	help2.SetText("F11: toggle fullscreen mode")
	help2.SetFontSize(.04)
	help2.SetPositionY(-.1)

	container.AddChildren(help1, help2)
	return container
}

func tab1() gfx.WindowObject {
	triangle := gfx.NewShape2D()

	container := gfx.NewWindowObject(nil)
	container.AddChild(triangle)
	container.AddChild(label("Triangle / White", gfx.Blue))
	return container
}

func tab2() gfx.WindowObject {
	triangleLine := gfx.NewTriangle(.2)
	triangleLine.SetColor(gfx.Green)

	container := gfx.NewWindowObject(nil)
	container.AddChild(triangleLine)
	container.AddChild(label("Triangle Line / Green", gfx.Blue))
	return container
}

func tab3() gfx.WindowObject {
	quad := gfx.NewQuad()
	quad.SetTexture(gfx.NewTexture2D(testImage, testImage))

	container := gfx.NewWindowObject(nil)
	container.AddChild(quad)
	container.AddChild(label("Quad / Textured", gfx.Blue))
	return container
}

func tab4() gfx.WindowObject {
	square := gfx.NewSquare(.2)
	square.SetColor(gfx.Red)

	container := gfx.NewWindowObject(nil)
	container.AddChild(square)
	container.AddChild(label("Square Line / Red", gfx.Blue))
	return container
}

func tab5() gfx.WindowObject {
	dot := gfx.NewDot()
	dot.SetTexture(gfx.NewTexture2D(testImage, testImage))

	container := gfx.NewWindowObject(nil)
	container.AddChild(dot)
	container.AddChild(label("Dot / Textured", gfx.Blue))
	return container
}

func tab6() gfx.WindowObject {
	circle := gfx.NewCircle(.2)
	circle.SetColor(gfx.Orange)

	container := gfx.NewWindowObject(nil)
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

	container := gfx.NewWindowObject(nil)
	container.AddChild(pacman)
	container.AddChild(label("Pac-Man", gfx.Blue))
	return container
}

func tab8() gfx.WindowObject {
	arc := gfx.NewCircle(.1)
	arc.
		SetLength(.5).
		SetColor(gfx.Orange)

	container := gfx.NewWindowObject(nil)
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

	container := gfx.NewWindowObject(nil)
	container.AddChildren(carousel, topLeft, bottomCenter)
	return container
}

func New2DView() gfx.WindowObject {
	if imgBytes, err := os.ReadFile("test.png"); err == nil {
		gfx.Assets.Add(gfx.NewTexture2D(testImage, imgBytes))
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
	tabGroup.AddChild(tab9())
	return tabGroup
}
