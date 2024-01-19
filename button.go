package gfx

import (
	"image/color"
	"sync/atomic"
)

const (
	defaultButtonName = "Button"
)

type Button struct {
	WindowObjectBase

	fill   *Shape
	border *Shape
	label  *Label
	bounds BoundingObject

	mouseEnterFillColorSet   atomic.Bool
	mouseEnterBorderColorSet atomic.Bool
	mouseEnterTextColorSet   atomic.Bool
	mouseDownFillColorSet    atomic.Bool
	mouseDownBorderColorSet  atomic.Bool
	mouseDownTextColorSet    atomic.Bool

	mouseEnterFillColor   color.RGBA
	mouseEnterBorderColor color.RGBA
	mouseEnterTextColor   color.RGBA
	mouseDownFillColor    color.RGBA
	mouseDownBorderColor  color.RGBA
	mouseDownTextColor    color.RGBA

	mouseLeaveFillColor   color.RGBA
	mouseLeaveBorderColor color.RGBA
	mouseLeaveTextColor   color.RGBA
	mouseUpFillColor      color.RGBA
	mouseUpBorderColor    color.RGBA
	mouseUpTextColor      color.RGBA
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (b *Button) Init(window *Window) (ok bool) {
	if !b.WindowObjectBase.Init(window) {
		return false
	}

	b.initialized.Store(true)
	return true
}

/******************************************************************************
 Button Functions
******************************************************************************/

func (b *Button) Texture() string {
	return b.fill.Texture()
}

func (b *Button) SetTexture(pathToPng string) *Button {
	b.fill.SetTexture(pathToPng)
	return b
}

func (b *Button) FillColor() color.RGBA {
	return b.fill.Color()
}

func (b *Button) SetFillColor(rgba color.RGBA) *Button {
	b.fill.SetColor(rgba)
	b.stateMutex.Lock()
	b.mouseLeaveFillColor = rgba
	b.mouseUpFillColor = rgba
	b.stateMutex.Unlock()
	return b
}

func (b *Button) SetMouseEnterFillColor(rgba color.RGBA) *Button {
	b.mouseEnterFillColorSet.Store(true)
	b.stateMutex.Lock()
	b.mouseEnterFillColor = rgba
	b.stateMutex.Unlock()
	return b
}

func (b *Button) SetMouseDownFillColor(rgba color.RGBA) *Button {
	b.mouseDownFillColorSet.Store(true)
	b.stateMutex.Lock()
	b.mouseDownFillColor = rgba
	b.stateMutex.Unlock()
	return b
}

func (b *Button) SetColor(rgba color.RGBA) WindowObject {
	b.SetFillColor(rgba)
	return b.SetBorderColor(rgba)
}

func (b *Button) BorderThickness() float32 {
	return b.border.Thickness()
}

func (b *Button) SetBorderThickness(thickness float32) *Button {
	if thickness <= thicknessEpsilon {
		thickness = thicknessEpsilon * 2
	}
	b.border.SetThickness(thickness)
	return b
}

func (b *Button) BorderColor() color.RGBA {
	return b.border.Color()
}

func (b *Button) SetBorderColor(rgba color.RGBA) *Button {
	b.border.SetColor(rgba)
	b.stateMutex.Lock()
	b.mouseLeaveBorderColor = rgba
	b.mouseUpBorderColor = rgba
	b.stateMutex.Unlock()
	return b
}

func (b *Button) SetMouseEnterBorderColor(rgba color.RGBA) *Button {
	b.mouseEnterBorderColorSet.Store(true)
	b.stateMutex.Lock()
	b.mouseEnterBorderColor = rgba
	b.stateMutex.Unlock()
	return b
}

func (b *Button) SetMouseDownBorderColor(rgba color.RGBA) *Button {
	b.mouseDownBorderColorSet.Store(true)
	b.stateMutex.Lock()
	b.mouseDownBorderColor = rgba
	b.stateMutex.Unlock()
	return b
}

func (b *Button) Text() string {
	return b.label.Text()
}

func (b *Button) SetText(text string) *Button {
	b.label.SetText(text)
	return b
}

func (b *Button) SetTextColor(rgba color.RGBA) *Button {
	b.label.SetColor(rgba)
	b.stateMutex.Lock()
	b.mouseLeaveTextColor = rgba
	b.mouseUpTextColor = rgba
	b.stateMutex.Unlock()
	return b
}

func (b *Button) SetMouseEnterTextColor(rgba color.RGBA) *Button {
	b.mouseEnterTextColorSet.Store(true)
	b.stateMutex.Lock()
	b.mouseEnterTextColor = rgba
	b.stateMutex.Unlock()
	return b
}

func (b *Button) SetMouseDownTextColor(rgba color.RGBA) *Button {
	b.mouseDownTextColorSet.Store(true)
	b.stateMutex.Lock()
	b.mouseDownTextColor = rgba
	b.stateMutex.Unlock()
	return b
}

func (b *Button) SetTextSize(size float32) *Button {
	b.label.SetFontSize(size)
	return b
}

func (b *Button) Label() *Label {
	return b.label
}

func (b *Button) MaintainAspectRatio(maintainAspectRatio bool) WindowObject {
	b.fill.MaintainAspectRatio(maintainAspectRatio)
	b.border.MaintainAspectRatio(maintainAspectRatio)
	b.label.MaintainAspectRatio(maintainAspectRatio)
	b.bounds.MaintainAspectRatio(maintainAspectRatio)
	return b
}

func (b *Button) BlurEnabled() bool {
	return b.fill.BlurEnabled()
}

func (b *Button) SetBlurEnabled(isEnabled bool) WindowObject {
	b.fill.SetBlurEnabled(isEnabled)
	return b
}

func (b *Button) BlurIntensity() float32 {
	return b.fill.BlurIntensity()
}

func (b *Button) SetBlurIntensity(intensity float32) WindowObject {
	b.fill.SetBlurIntensity(intensity)
	return b
}

func (b *Button) onMouseEnter(_ WindowObject, _ *MouseState) {
	if b.mouseEnterFillColorSet.Load() {
		b.fill.SetColor(b.mouseEnterFillColor)
	}
	if b.mouseEnterBorderColorSet.Load() {
		b.border.SetColor(b.mouseEnterBorderColor)
	}
	if b.mouseEnterTextColorSet.Load() {
		b.label.SetColor(b.mouseEnterTextColor)
	}
}

func (b *Button) onMouseLeave(_ WindowObject, _ *MouseState) {
	if b.mouseEnterFillColorSet.Load() {
		b.fill.SetColor(b.mouseLeaveFillColor)
	}
	if b.mouseEnterBorderColorSet.Load() {
		b.border.SetColor(b.mouseLeaveBorderColor)
	}
	if b.mouseEnterTextColorSet.Load() {
		b.label.SetColor(b.mouseLeaveTextColor)
	}
}

func (b *Button) onMouseDown(_ WindowObject, _ *MouseState) {
	if b.mouseDownFillColorSet.Load() {
		b.fill.SetColor(b.mouseDownFillColor)
	}
	if b.mouseDownBorderColorSet.Load() {
		b.border.SetColor(b.mouseDownBorderColor)
	}
	if b.mouseDownTextColorSet.Load() {
		b.label.SetColor(b.mouseDownTextColor)
	}
}

func (b *Button) onMouseUp(_ WindowObject, _ *MouseState) {
	if b.mouseDownFillColorSet.Load() {
		b.fill.SetColor(b.mouseUpFillColor)
	}
	if b.mouseDownBorderColorSet.Load() {
		b.border.SetColor(b.mouseUpBorderColor)
	}
	if b.mouseDownTextColorSet.Load() {
		b.label.SetColor(b.mouseUpTextColor)
	}
}

func (b *Button) OnDepressed(handler func(sender WindowObject, mouseState *MouseState)) *Button {
	b.bounds.OnDepressed(handler)
	return b
}

func (b *Button) OnClick(handler func(sender WindowObject, mouseState *MouseState)) *Button {
	b.bounds.OnClick(handler)
	return b
}

/******************************************************************************
 New Button Function
******************************************************************************/

func NewButton(isCircular ...bool) *Button {
	var b *Button
	if len(isCircular) > 0 && isCircular[0] {
		b = &Button{
			WindowObjectBase: *NewObject(nil),
			fill:             NewDot(),
			border:           NewCircle(thicknessEpsilon * 2),
			label:            NewLabel(),
			bounds:           NewBoundingRadius(),
		}
	} else {
		b = &Button{
			WindowObjectBase: *NewObject(nil),
			fill:             NewQuad(),
			border:           NewSquare(thicknessEpsilon * 2),
			label:            NewLabel(),
			bounds:           NewBoundingBox(),
		}
	}

	b.SetName(defaultButtonName)
	b.AddChild(b.fill)
	b.AddChild(b.border)
	b.AddChild(b.label)
	b.AddChild(b.bounds)
	b.bounds.SetParent(b)

	b.bounds.OnMouseEnter(b.onMouseEnter)
	b.bounds.OnMouseLeave(b.onMouseLeave)
	b.bounds.OnMouseDown(b.onMouseDown)
	b.bounds.OnMouseUp(b.onMouseUp)

	return b
}
