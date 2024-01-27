package view

import (
	"image/color"
	"strconv"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/examples/ui/textures"
)

func onClick(button gfx.WindowObject, _ *gfx.MouseState) {
	clickCount := 0
	if count, ok := button.Tag().(int); ok { // use a map[string]any to store multiple values
		clickCount = count
	}
	clickCount++
	button.SetTag(clickCount)
	button.(*gfx.Button).SetText(strconv.Itoa(clickCount))
}

func onDepressed(button gfx.WindowObject, _ *gfx.MouseState) {
	blur := button.(*gfx.Button).BlurIntensity()
	if blur < 1 {
		blur = 1
	}
	blur *= 1.02
	button.(*gfx.Button).SetBlurIntensity(blur)
}

// NewButtonView In this example, the buttons are aligned/anchored to the
// corners, regardless of their size (scale) or the size/ratio of the window.
func NewButtonView(window *gfx.Window) gfx.WindowObject {
	gfx.AddEmbeddedAsset("button.png", textures.Assets)

	btnWidth := float32(.25) // try changing the width/height!
	btnHeight := float32(.25)
	textSize := float32(.2)

	// window.SwapMouseButtons(true) // uncomment for lefties!

	button1 := gfx.NewButton()
	button1.
		SetMouseEnterBorderColor(gfx.Purple).
		SetMouseDownBorderColor(gfx.Blue).
		SetMouseEnterFillColor(gfx.Red).
		SetMouseDownFillColor(gfx.Orange).
		SetMouseEnterTextColor(gfx.Green).
		SetMouseDownTextColor(gfx.Yellow).
		SetText("Click me!").
		SetTextSize(textSize).
		SetTextColor(gfx.Red).
		OnClick(onClick).
		MaintainAspectRatio(false).(*gfx.Button). // *see note...
		SetTexture("button.png").
		SetBorderColor(gfx.Magenta).
		SetBorderThickness(.1).
		SetFillColor(gfx.Blue).
		SetScaleX(btnWidth).SetScaleY(btnHeight).
		SetAnchor(gfx.TopLeft)

	// *Be careful when method-chaining, particularly when calling shadowed methods.
	// Here, we want to call Button's version of MaintainAspectRatio, so where this
	// call exists in the chain is important.  Also, since this function returns a
	// WindowObject interface, we use type assertion to keep the chain going.

	button2 := gfx.NewButton()
	button2.
		SetText("Click me!").
		SetTextSize(textSize).
		SetTextColor(gfx.White).
		OnClick(onClick).
		SetBorderThickness(.05).
		SetBorderColor(gfx.White).
		SetFillColor(color.RGBA{R: 160, G: 160, B: 255, A: 255}).
		SetScaleX(btnWidth).SetScaleY(btnHeight).
		SetAnchor(gfx.BottomRight)

	// Can't put these calls in the chain above, even with type assertion, if
	// we want to call button2.FillColor() after it has been set.
	button2.
		SetMouseDownFillColor(gfx.Darken(button2.FillColor(), .2)).
		SetMouseEnterFillColor(gfx.Lighten(button2.FillColor(), .2))

	button3 := gfx.NewButton()
	button3.
		SetMouseDownFillColor(gfx.Darken(gfx.Orange, .2)).
		SetMouseEnterFillColor(gfx.Lighten(gfx.Orange, .2)).
		SetText("Click me!").
		SetTextSize(textSize).
		SetTextColor(gfx.White).
		OnClick(onClick).
		MaintainAspectRatio(false).(*gfx.Button).
		SetBorderThickness(0).
		SetBorderColor(gfx.Blue).
		SetFillColor(gfx.Orange).
		SetScaleX(btnWidth).SetScaleY(btnHeight).
		SetAnchor(gfx.BottomLeft)

	button4 := gfx.NewButton()
	button4.
		SetMouseDownFillColor(gfx.Darken(button4.FillColor(), .4)).
		SetMouseEnterFillColor(color.RGBA{R: 40, G: 40, B: 255, A: 255}).
		SetText("Click me!").
		SetTextSize(textSize).
		SetTextColor(gfx.Red).
		OnClick(onClick).
		SetTexture("button.png").
		SetBorderThickness(.1).
		SetBorderColor(gfx.Magenta).
		SetFillColor(gfx.Blue).
		SetScaleX(btnWidth).SetScaleY(btnHeight).
		SetAnchor(gfx.TopRight)

	btnWidth = .1
	btnHeight = .1

	button5 := gfx.NewButton(true) // true = will be a circular button
	button5.
		SetText("Depress").SetTextSize(.24).SetTextColor(gfx.Green).
		SetMouseEnterFillColor(color.RGBA{R: 255, G: 50, B: 50, A: 255}).
		OnDepressed(onDepressed). // will trigger once per game tick when left mouse button is depressed
		MaintainAspectRatio(false).(*gfx.Button).
		SetBorderThickness(.05).
		SetBorderColor(gfx.Opacity(gfx.White, .1)).
		SetFillColor(gfx.Red).
		SetBlurEnabled(true).
		SetScaleX(btnWidth).SetScaleY(btnHeight).
		SetPositionX(-.8 + window.ScaleX(btnWidth))

	button6 := gfx.NewButton(true) // true = will be a circular button
	button6.
		SetText("Depress").SetTextSize(.24).SetTextColor(gfx.Red).
		SetMouseEnterFillColor(color.RGBA{R: 150, G: 255, B: 150, A: 255}).
		OnDepressed(onDepressed). // will trigger once per game tick when left mouse button is depressed
		SetBorderThickness(.05).
		SetBorderColor(gfx.Opacity(gfx.White, .1)).
		SetFillColor(gfx.Green).
		SetBlurEnabled(true).
		SetScaleX(btnWidth).SetScaleY(btnHeight).
		SetPositionX(.8 - window.ScaleX(btnWidth))

	lbl1 := gfx.NewLabel()
	lbl1.
		SetText("<- false").
		SetFontSize(.05).
		SetPositionY(.1)

	lbl2 := gfx.NewLabel()
	lbl2.
		SetText("MaintainAspectRatio").
		SetFontSize(.05)

	lbl3 := gfx.NewLabel()
	lbl3.
		SetText("true ->").
		SetFontSize(.05).
		SetPositionY(-.1)

	container := gfx.NewObject(nil)
	container.AddChild(button1)
	container.AddChild(button2)
	container.AddChild(button3)
	container.AddChild(button4)
	container.AddChild(button5)
	container.AddChild(button6)
	container.AddChild(lbl1)
	container.AddChild(lbl2)
	container.AddChild(lbl3)

	return container
}
