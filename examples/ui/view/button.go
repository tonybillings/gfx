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
	blur *= 1.01
	button.(*gfx.Button).SetBlurIntensity(blur)
}

func anchorTopRight(button gfx.WindowObject) {
	button.SetPositionX(1 - (button.Window().ScaleY(button.Scale().X()) * button.Window().AspectRatioInv()))
	button.SetPositionY(1 - (button.Window().ScaleX(button.Scale().Y()) * button.Window().AspectRatio()))
}

func anchorBottomRight(button gfx.WindowObject) {
	button.SetPositionX(1 - (button.Window().ScaleY(button.Scale().X()) * button.Window().AspectRatioInv()))
	button.SetPositionY(-1 + (button.Window().ScaleX(button.Scale().Y()) * button.Window().AspectRatio()))
}

// NewButtonView In this example, the buttons are aligned/anchored to the
// corners, regardless of their size (scale) or the size/ratio of the window.
func NewButtonView(window *gfx.Window) gfx.WindowObject {
	gfx.AddEmbeddedAsset("button.png", textures.Assets)

	btnWidth := float32(.25) // try changing the width/height!
	btnHeight := float32(.25)
	textSize := float32(.2)

	button1 := gfx.NewButton()
	button1.
		SetTexture("button.png").
		SetBorderThickness(.1).
		SetBorderColor(gfx.Magenta).
		SetMouseEnterBorderColor(gfx.Purple).
		SetMouseDownBorderColor(gfx.Blue).
		SetFillColor(gfx.Blue).
		SetMouseEnterFillColor(gfx.Red).
		SetMouseDownFillColor(gfx.Orange).
		SetText("Click me!").
		SetTextSize(textSize).
		SetTextColor(gfx.Red).
		SetMouseEnterTextColor(gfx.Green).
		SetMouseDownTextColor(gfx.Yellow).
		OnClick(onClick).
		MaintainAspectRatio(false).
		SetScaleX(btnWidth).SetScaleY(btnHeight).
		SetPositionX(-1 + btnWidth).SetPositionY(1 - btnHeight)

	button2 := gfx.NewButton()
	button2.
		SetBorderThickness(.05).
		SetBorderColor(gfx.White).
		SetFillColor(color.RGBA{R: 160, G: 160, B: 255, A: 255}).
		SetMouseDownFillColor(gfx.Darken(button2.FillColor(), .2)).
		SetMouseEnterFillColor(gfx.Lighten(button2.FillColor(), .2)).
		SetText("Click me!").
		SetTextSize(textSize).
		SetTextColor(gfx.White).
		OnClick(onClick).
		MaintainAspectRatio(true).
		SetScaleX(btnWidth).SetScaleY(btnHeight).
		SetPositionX(1 - (window.ScaleY(btnWidth) * window.AspectRatioInv())).
		SetPositionY(-1 + (window.ScaleX(btnHeight) * window.AspectRatio())).
		OnResize(func(_, _, _, _ int32) {
			anchorBottomRight(button2)
		})

	button3 := gfx.NewButton()
	button3.
		SetBorderThickness(0).
		SetBorderColor(gfx.Blue).
		SetFillColor(gfx.Orange).
		SetMouseDownFillColor(gfx.Darken(button3.FillColor(), .2)).
		SetMouseEnterFillColor(gfx.Lighten(button3.FillColor(), .2)).
		SetText("Click me!").
		SetTextSize(textSize).
		SetTextColor(gfx.White).
		OnClick(onClick).
		MaintainAspectRatio(false).
		SetScaleX(btnWidth).SetScaleY(btnHeight).
		SetPositionX(-1 + btnWidth).SetPositionY(-1 + btnHeight)

	button4 := gfx.NewButton()
	button4.
		SetTexture("button.png").
		SetBorderThickness(.1).
		SetBorderColor(gfx.Magenta).
		SetFillColor(gfx.Blue).
		SetMouseDownFillColor(gfx.Darken(button4.FillColor(), .4)).
		SetMouseEnterFillColor(color.RGBA{R: 40, G: 40, B: 255, A: 255}).
		SetText("Click me!").
		SetTextSize(textSize).
		SetTextColor(gfx.Red).
		OnClick(onClick).
		MaintainAspectRatio(true).
		SetScaleX(btnWidth).SetScaleY(btnHeight).
		SetPositionX(1 - (window.ScaleY(btnWidth) * window.AspectRatioInv())).
		SetPositionY(1 - (window.ScaleX(btnHeight) * window.AspectRatio())).
		OnResize(func(_, _, _, _ int32) {
			anchorTopRight(button4)
		})

	btnWidth = .1
	btnHeight = .1

	button5 := gfx.NewButton(true) // true = will be a circular button
	button5.
		SetBorderThickness(.05).
		SetBorderColor(gfx.Opacity(gfx.White, .1)).
		SetFillColor(gfx.Red).
		SetMouseEnterFillColor(color.RGBA{R: 255, G: 50, B: 50, A: 255}).
		OnDepressed(onDepressed). // will trigger once per game tick when left mouse button is depressed
		MaintainAspectRatio(false).
		SetBlurEnabled(true).
		SetScaleX(btnWidth).SetScaleY(btnHeight).
		SetPositionX(-.8 + (window.ScaleY(btnWidth) * window.AspectRatioInv()))

	button6 := gfx.NewButton(true) // true = will be a circular button
	button6.
		SetBorderThickness(.05).
		SetBorderColor(gfx.Opacity(gfx.White, .1)).
		SetFillColor(gfx.Green).
		SetMouseEnterFillColor(color.RGBA{R: 150, G: 255, B: 150, A: 255}).
		OnDepressed(onDepressed). // will trigger once per game tick when left mouse button is depressed
		MaintainAspectRatio(true).
		SetBlurEnabled(true).
		SetScaleX(btnWidth).SetScaleY(btnHeight).
		SetPositionX(.8 - (window.ScaleY(btnWidth) * window.AspectRatioInv()))

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
