package gfx

import (
	"image/color"
	"sync/atomic"
)

const (
	defaultButtonName = "Button"
)

type Button struct {
	View

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

	originalFillColor   color.RGBA
	originalBorderColor color.RGBA
	originalTextColor   color.RGBA
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (b *Button) Init(window *Window) (ok bool) {
	b.label.Init(window)
	b.bounds.Init(window)
	if !b.View.Init(window) {
		return false
	}
	b.originalFillColor = b.fill.Color()
	b.originalBorderColor = b.border.Color()
	b.originalTextColor = b.label.Color()
	return true
}

func (b *Button) Update(deltaTime int64) (ok bool) {
	if !b.enabled.Load() {
		return false
	}

	b.label.Update(deltaTime)
	b.bounds.Update(deltaTime)
	return b.View.Update(deltaTime)
}

func (b *Button) Draw(deltaTime int64) (ok bool) {
	if !b.visible.Load() {
		return false
	}

	if !b.View.Draw(deltaTime) {
		return false
	}

	return b.label.Draw(deltaTime)
}

func (b *Button) Close() {
	b.label.Close()
	b.bounds.Close()
	b.View.Close()
}

/******************************************************************************
 Button Functions
******************************************************************************/

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
		b.fill.SetColor(b.originalFillColor)
	}
	if b.mouseEnterBorderColorSet.Load() {
		b.border.SetColor(b.originalBorderColor)
	}
	if b.mouseEnterTextColorSet.Load() {
		b.label.SetColor(b.originalTextColor)
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
		b.fill.SetColor(b.originalFillColor)
	}
	if b.mouseDownBorderColorSet.Load() {
		b.border.SetColor(b.originalBorderColor)
	}
	if b.mouseDownTextColorSet.Load() {
		b.label.SetColor(b.originalTextColor)
	}
}

func (b *Button) OnDepressed(handler func(sender WindowObject, mouseState *MouseState)) *Button {
	b.bounds.OnPMouseDepressed(handler)
	return b
}

func (b *Button) OnClick(handler func(sender WindowObject, mouseState *MouseState)) *Button {
	b.bounds.OnPMouseClick(handler)
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
	b.View.MaintainAspectRatio(maintainAspectRatio)
	b.label.MaintainAspectRatio(maintainAspectRatio)
	b.bounds.MaintainAspectRatio(maintainAspectRatio)
	return b
}

func (b *Button) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	b.label.Resize(oldWidth, oldHeight, newWidth, newHeight)
	b.bounds.Resize(oldWidth, oldHeight, newWidth, newHeight)
	b.View.Resize(oldWidth, oldHeight, newWidth, newHeight)
}

/******************************************************************************
 New Button Function
******************************************************************************/

func NewButton(isCircular ...bool) *Button {
	b := &Button{
		View: *NewView(),
	}

	if len(isCircular) > 0 && isCircular[0] {
		b.fill = NewDot()
		b.border = NewCircle(thicknessEpsilon * 2)
		b.label = NewLabel()
		b.bounds = NewBoundingRadius()
	} else {
		b.fill = NewQuad()
		b.border = NewSquare(thicknessEpsilon * 2)
		b.label = NewLabel()
		b.bounds = NewBoundingBox()
	}

	b.SetName(defaultButtonName)
	b.fill.SetParent(b)
	b.border.SetParent(b)
	b.label.SetParent(b)
	b.bounds.SetParent(b)

	b.fill.SetColor(Black)
	b.border.SetColor(Black)

	b.bounds.OnMouseEnter(b.onMouseEnter)
	b.bounds.OnMouseLeave(b.onMouseLeave)
	b.bounds.OnPMouseDown(b.onMouseDown)
	b.bounds.OnPMouseUp(b.onMouseUp)

	return b
}
