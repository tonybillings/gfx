package view

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tonybillings/gfx"
)

func newBrushControls(brush *gfx.BasicBrush) *gfx.View {
	brushColorPreview := gfx.NewView()
	brushColorPreview.
		SetBorderColor(gfx.White).
		SetBorderThickness(.1).
		SetAnchor(gfx.TopRight).
		SetMargin(gfx.Margin{Top: .06, Right: .05}).
		SetScale(mgl32.Vec3{.3, .1})

	brushColorRSlider := gfx.NewSlider(gfx.Vertical, true)
	brushColorRSlider.
		SetFillColor(gfx.Transparent).
		SetAnchor(gfx.MiddleLeft).
		SetScaleX(.3333).
		SetScaleY(1)
	brushColorRSlider.Button().
		SetText("R").
		SetFontSize(.6).
		SetTextColor(gfx.Red).
		SetMouseEnterFillColor(gfx.Darken(gfx.LightGray, .2)).
		SetMouseDownFillColor(gfx.Darken(gfx.LightGray, .35)).
		SetColor(gfx.LightGray)
	brushColorRSlider.Rail().SetColor(gfx.White)
	brushColorRSlider.OnValueChanging(func(_ gfx.WindowObject, value float32) {
		color := brush.Color()
		color.R = uint8(value * 255)
		brush.SetColor(color)
		brushColorPreview.SetFillColor(color)
	})
	brushColorRSlider.OnValueChanged(func(_ gfx.WindowObject, value float32) {
		color := brush.Color()
		color.R = uint8(value * 255)
		brush.SetColor(color)
		brushColorPreview.SetFillColor(color)
	})

	brushColorGSlider := gfx.NewSlider(gfx.Vertical, true)
	brushColorGSlider.
		SetFillColor(gfx.Transparent).
		SetAnchor(gfx.MiddleLeft).
		SetMarginLeft(.1).
		SetScaleX(.3333).
		SetScaleY(1)
	brushColorGSlider.Button().
		SetText("G").
		SetFontSize(.6).
		SetTextColor(gfx.Green).
		SetMouseEnterFillColor(gfx.Darken(gfx.LightGray, .2)).
		SetMouseDownFillColor(gfx.Darken(gfx.LightGray, .35)).
		SetColor(gfx.LightGray)
	brushColorGSlider.Rail().SetColor(gfx.White)
	brushColorGSlider.OnValueChanging(func(_ gfx.WindowObject, value float32) {
		color := brush.Color()
		color.G = uint8(value * 255)
		brush.SetColor(color)
		brushColorPreview.SetFillColor(color)
	})
	brushColorGSlider.OnValueChanged(func(_ gfx.WindowObject, value float32) {
		color := brush.Color()
		color.G = uint8(value * 255)
		brush.SetColor(color)
		brushColorPreview.SetFillColor(color)
	})

	brushColorBSlider := gfx.NewSlider(gfx.Vertical, true)
	brushColorBSlider.
		SetFillColor(gfx.Transparent).
		SetAnchor(gfx.MiddleLeft).
		SetMarginLeft(.2).
		SetScaleX(.3333).
		SetScaleY(1)
	brushColorBSlider.Button().
		SetText("B").
		SetFontSize(.6).
		SetTextColor(gfx.Blue).
		SetMouseEnterFillColor(gfx.Darken(gfx.LightGray, .2)).
		SetMouseDownFillColor(gfx.Darken(gfx.LightGray, .35)).
		SetColor(gfx.LightGray)
	brushColorBSlider.Rail().SetColor(gfx.White)
	brushColorBSlider.OnValueChanging(func(_ gfx.WindowObject, value float32) {
		color := brush.Color()
		color.B = uint8(value * 255)
		brush.SetColor(color)
		brushColorPreview.SetFillColor(color)
	})
	brushColorBSlider.OnValueChanged(func(_ gfx.WindowObject, value float32) {
		color := brush.Color()
		color.B = uint8(value * 255)
		brush.SetColor(color)
		brushColorPreview.SetFillColor(color)
	})

	brushHeadRound := gfx.NewButton()
	brushHeadSquare := gfx.NewButton()
	brushHeadRoundIcon := gfx.NewDot()
	brushHeadSquareIcon := gfx.NewQuad()

	brushHeadRound.
		SetMouseEnterBorderColor(gfx.White).
		SetBorderThickness(.2).
		SetBorderColor(gfx.Purple).
		SetFillColor(gfx.Transparent).
		SetAnchor(gfx.MiddleRight).
		SetMarginRight(.15).
		SetScale(mgl32.Vec3{.13, .2})
	brushHeadRound.OnClick(func(_ gfx.WindowObject, _ *gfx.MouseState) {
		brushHeadSquare.SetBorderColor(gfx.Purple)
		brushHeadSquareIcon.SetColor(gfx.Purple)

		brushHeadRound.SetBorderColor(gfx.White)
		brushHeadRoundIcon.SetColor(gfx.White)

		brush.SetBrushHead(gfx.RoundBrushHead)
	})
	brushHeadRoundIcon.
		SetColor(gfx.White).
		SetScale(mgl32.Vec3{.5, .5})
	brushHeadRound.AddChild(brushHeadRoundIcon)

	brushHeadSquare.
		SetMouseEnterBorderColor(gfx.White).
		SetBorderThickness(.2).
		SetBorderColor(gfx.Purple).
		SetFillColor(gfx.Transparent).
		SetAnchor(gfx.MiddleRight).
		SetMarginRight(.054).
		SetScale(mgl32.Vec3{.13, .2})
	brushHeadSquare.OnClick(func(_ gfx.WindowObject, _ *gfx.MouseState) {
		brushHeadRound.SetBorderColor(gfx.Purple)
		brushHeadRoundIcon.SetColor(gfx.Purple)

		brushHeadSquare.SetBorderColor(gfx.White)
		brushHeadSquareIcon.SetColor(gfx.White)

		brush.SetBrushHead(gfx.SquareBrushHead)
	})
	brushHeadSquareIcon.
		SetColor(gfx.Purple).
		SetScale(mgl32.Vec3{.5, .5})
	brushHeadSquare.AddChild(brushHeadSquareIcon)

	brushSizeSlider := gfx.NewSlider(gfx.Horizontal, false)
	brushSizeSlider.
		SetFillColor(gfx.Transparent).
		SetAnchor(gfx.BottomRight).
		SetMargin(gfx.Margin{Bottom: .010, Right: .035}).
		SetScale(mgl32.Vec3{.35, .33})
	brushSizeSlider.Button().
		SetMouseEnterFillColor(gfx.Darken(gfx.White, .2)).
		SetMouseDownFillColor(gfx.Darken(gfx.LightGray, .35)).
		SetColor(gfx.White)
	brushSizeSlider.OnValueChanging(func(_ gfx.WindowObject, value float32) {
		brush.SetSize(value*.1 + 0.005)
	})

	brushControlsLabel := gfx.NewLabel()
	brushControlsLabel.
		SetText("Brush Controls").
		SetFontSize(.1).
		SetColor(gfx.Purple).
		SetAnchor(gfx.BottomCenter).
		SetMarginBottom(-.045)

	brushControls := gfx.NewView()
	brushControls.
		SetBorderThickness(.01).
		SetBorderColor(gfx.Purple).
		SetFillColor(gfx.Opacity(gfx.Purple, .3)).
		SetScale(mgl32.Vec3{.3, .2}).
		SetPosition(mgl32.Vec3{-.5, .4})

	brushControls.AddChildren(
		brushColorPreview,
		brushColorRSlider, brushColorGSlider, brushColorBSlider,
		brushHeadRound, brushHeadSquare,
		brushSizeSlider,
		brushControlsLabel)

	return brushControls
}

func newCanvasControls(canvas *gfx.Canvas, brush *gfx.BasicBrush, exportDirectory string) *gfx.View {
	bgColorPreview := gfx.NewView()
	bgColorPreview.
		SetFillColor(gfx.White).
		SetBorderColor(gfx.White).
		SetBorderThickness(.1).
		SetAnchor(gfx.TopRight).
		SetMargin(gfx.Margin{Top: .06, Right: .05}).
		SetScale(mgl32.Vec3{.3, .1})

	bgColorRSlider := gfx.NewSlider(gfx.Vertical, true)
	bgColorRSlider.
		SetFillColor(gfx.Transparent).
		SetAnchor(gfx.MiddleLeft).
		SetScaleX(.3333).
		SetScaleY(1)
	bgColorRSlider.Button().
		SetText("R").
		SetFontSize(.6).
		SetTextColor(gfx.Red).
		SetMouseEnterFillColor(gfx.Darken(gfx.LightGray, .2)).
		SetMouseDownFillColor(gfx.Darken(gfx.LightGray, .35)).
		SetColor(gfx.LightGray)
	bgColorRSlider.Rail().SetColor(gfx.White)
	bgColorRSlider.OnValueChanging(func(_ gfx.WindowObject, value float32) {
		color := canvas.FillColor()
		color.R = uint8(value * 255)
		canvas.SetFillColor(color)
		bgColorPreview.SetFillColor(color)
	})
	bgColorRSlider.OnValueChanged(func(_ gfx.WindowObject, value float32) {
		color := canvas.FillColor()
		color.R = uint8(value * 255)
		canvas.SetFillColor(color)
		bgColorPreview.SetFillColor(color)
	})
	bgColorRSlider.SetValue(1)

	bgColorGSlider := gfx.NewSlider(gfx.Vertical, true)
	bgColorGSlider.
		SetFillColor(gfx.Transparent).
		SetAnchor(gfx.MiddleLeft).
		SetMarginLeft(.1).
		SetScaleX(.3333).
		SetScaleY(1)
	bgColorGSlider.Button().
		SetText("G").
		SetFontSize(.6).
		SetTextColor(gfx.Green).
		SetMouseEnterFillColor(gfx.Darken(gfx.LightGray, .2)).
		SetMouseDownFillColor(gfx.Darken(gfx.LightGray, .35)).
		SetColor(gfx.LightGray)
	bgColorGSlider.Rail().SetColor(gfx.White)
	bgColorGSlider.OnValueChanging(func(_ gfx.WindowObject, value float32) {
		color := canvas.FillColor()
		color.G = uint8(value * 255)
		canvas.SetFillColor(color)
		bgColorPreview.SetFillColor(color)
	})
	bgColorGSlider.OnValueChanged(func(_ gfx.WindowObject, value float32) {
		color := canvas.FillColor()
		color.G = uint8(value * 255)
		canvas.SetFillColor(color)
		bgColorPreview.SetFillColor(color)
	})
	bgColorGSlider.SetValue(1)

	bgColorBSlider := gfx.NewSlider(gfx.Vertical, true)
	bgColorBSlider.
		SetFillColor(gfx.Transparent).
		SetAnchor(gfx.MiddleLeft).
		SetMarginLeft(.2).
		SetScaleX(.3333).
		SetScaleY(1)
	bgColorBSlider.Button().
		SetText("B").
		SetFontSize(.6).
		SetTextColor(gfx.Blue).
		SetMouseEnterFillColor(gfx.Darken(gfx.LightGray, .2)).
		SetMouseDownFillColor(gfx.Darken(gfx.LightGray, .35)).
		SetColor(gfx.LightGray)
	bgColorBSlider.Rail().SetColor(gfx.White)
	bgColorBSlider.OnValueChanging(func(_ gfx.WindowObject, value float32) {
		color := canvas.FillColor()
		color.B = uint8(value * 255)
		canvas.SetFillColor(color)
		bgColorPreview.SetFillColor(color)
	})
	bgColorBSlider.OnValueChanged(func(_ gfx.WindowObject, value float32) {
		color := canvas.FillColor()
		color.B = uint8(value * 255)
		canvas.SetFillColor(color)
		bgColorPreview.SetFillColor(color)
	})
	bgColorBSlider.SetValue(1)

	undoButton := gfx.NewButton()
	undoButton.
		SetText("Undo").
		SetFontSize(.5).
		SetMouseDownFillColor(gfx.Darken(gfx.White, .5)).
		SetMouseDownBorderColor(gfx.White).
		SetMouseEnterBorderColor(gfx.White).
		SetBorderThickness(.2).
		SetBorderColor(gfx.Purple).
		SetFillColor(gfx.Transparent).
		SetAnchor(gfx.MiddleRight).
		SetMargin(gfx.Margin{Bottom: .025, Right: .05}).
		SetScale(mgl32.Vec3{.3, .15})
	undoButton.OnClick(func(_ gfx.WindowObject, _ *gfx.MouseState) {
		brush.Undo()
	})

	resetButton := gfx.NewButton()
	resetButton.
		SetText("Reset").
		SetFontSize(.5).
		SetMouseDownFillColor(gfx.Darken(gfx.White, .5)).
		SetMouseDownBorderColor(gfx.White).
		SetMouseEnterBorderColor(gfx.White).
		SetBorderThickness(.2).
		SetBorderColor(gfx.Purple).
		SetFillColor(gfx.Transparent).
		SetAnchor(gfx.MiddleRight).
		SetMargin(gfx.Margin{Top: .05, Right: .05}).
		SetScale(mgl32.Vec3{.3, .15})
	resetButton.OnClick(func(_ gfx.WindowObject, _ *gfx.MouseState) {
		canvas.Clear()
	})

	exportButton := gfx.NewButton()
	exportButton.
		SetText("Export").
		SetFontSize(.5).
		SetMouseDownFillColor(gfx.Darken(gfx.White, .5)).
		SetMouseDownBorderColor(gfx.White).
		SetMouseEnterBorderColor(gfx.White).
		SetBorderThickness(.2).
		SetBorderColor(gfx.Purple).
		SetFillColor(gfx.Transparent).
		SetAnchor(gfx.MiddleRight).
		SetMargin(gfx.Margin{Top: .125, Right: .05}).
		SetScale(mgl32.Vec3{.3, .15})
	exportButton.OnClick(func(_ gfx.WindowObject, _ *gfx.MouseState) {
		canvas.Export(exportDirectory)
	})

	canvasControlsLabel := gfx.NewLabel()
	canvasControlsLabel.
		SetText("Canvas Controls").
		SetFontSize(.1).
		SetColor(gfx.Purple).
		SetAnchor(gfx.BottomCenter).
		SetMarginBottom(-.045)

	canvasControls := gfx.NewView()
	canvasControls.
		SetBorderThickness(.01).
		SetBorderColor(gfx.Purple).
		SetFillColor(gfx.Opacity(gfx.Purple, .3)).
		SetScale(mgl32.Vec3{.3, .2}).
		SetPosition(mgl32.Vec3{-.5, -.4})

	canvasControls.AddChildren(
		bgColorPreview,
		bgColorRSlider, bgColorGSlider, bgColorBSlider,
		undoButton, resetButton, exportButton,
		canvasControlsLabel)

	return canvasControls
}

func NewCanvasView(exportDirectory ...string) gfx.WindowObject {
	exportDir := ""
	if len(exportDirectory) > 0 {
		exportDir = exportDirectory[0]
	}

	canvas := gfx.NewCanvas()
	canvas.
		SetFillColor(gfx.White).
		SetBorderColor(gfx.Purple).
		SetBorderThickness(.02).
		SetScale(mgl32.Vec3{.75, .75}).
		SetPositionX(.2)

	brush := gfx.NewBasicBrush()
	brush.
		SetBrushHead(gfx.RoundBrushHead).
		SetSize(0.005).
		SetColor(gfx.Black)
	brush.SetCanvas(canvas) // this is what links the two together...

	canvas.AddChild(brush) // ...not this; this is just convenient for life-cycle management

	brushControls := newBrushControls(brush)
	canvasControls := newCanvasControls(canvas, brush, exportDir)

	container := gfx.NewWindowObject()
	container.AddChildren(brushControls, canvasControls, canvas)

	return container
}
