package gfx

import (
	"image/color"
	"sync/atomic"
)

const (
	defaultButtonName = "Button"
)

/******************************************************************************
 Button
******************************************************************************/

type Button struct {
	View

	text   *Label
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
 Object Implementation
******************************************************************************/

func (b *Button) Init() (ok bool) {
	if b.Initialized() {
		return true
	}

	b.text.SetWindow(b.window)
	b.bounds.SetWindow(b.window)

	if !b.text.Init() || !b.bounds.Init() {
		return false
	}

	b.initLayout()
	b.RefreshLayout()

	return b.View.Init()
}

func (b *Button) Update(deltaTime int64) (ok bool) {
	if !b.View.Update(deltaTime) {
		return false
	}

	b.text.Update(deltaTime)
	b.bounds.Update(deltaTime)

	return true
}

func (b *Button) Close() {
	b.text.Close()
	b.bounds.Close()
	b.View.Close()
}

/******************************************************************************
 DrawableObject Implementation
******************************************************************************/

func (b *Button) Draw(deltaTime int64) (ok bool) {
	if !b.View.Draw(deltaTime) {
		return false
	}

	return b.text.Draw(deltaTime)
}

/******************************************************************************
 Resizer Implementation
******************************************************************************/

func (b *Button) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	b.View.Resize(oldWidth, oldHeight, newWidth, newHeight)
	b.text.Resize(oldWidth, oldHeight, newWidth, newHeight)
	b.bounds.Resize(oldWidth, oldHeight, newWidth, newHeight)
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (b *Button) SetWindow(window *Window) WindowObject {
	b.View.SetWindow(window)
	b.text.SetWindow(window)
	b.bounds.SetWindow(window)
	return b
}

func (b *Button) SetMaintainAspectRatio(maintainAspectRatio bool) WindowObject {
	b.View.SetMaintainAspectRatio(maintainAspectRatio)
	b.text.SetMaintainAspectRatio(maintainAspectRatio)
	b.bounds.SetMaintainAspectRatio(maintainAspectRatio)
	return b
}

/******************************************************************************
 Button Functions
******************************************************************************/

func (b *Button) defaultLayout() {
	b.fill.SetParent(b)
	b.border.SetParent(b)
	b.text.SetParent(b)
	b.bounds.SetParent(b)

	b.fill.SetColor(Black)
	b.border.SetColor(Black)
}

func (b *Button) initLayout() {
	b.bounds.OnMouseEnter(b.onMouseEnter)
	b.bounds.OnMouseLeave(b.onMouseLeave)
	b.bounds.OnPMouseDown(b.onMouseDown)
	b.bounds.OnPMouseUp(b.onMouseUp)

	b.originalFillColor = b.fill.Color()
	b.originalBorderColor = b.border.Color()
	b.originalTextColor = b.text.Color()
}

func (b *Button) onMouseEnter(_ WindowObject, _ *MouseState) {
	if b.mouseEnterFillColorSet.Load() {
		b.fill.SetColor(b.mouseEnterFillColor)
	}
	if b.mouseEnterBorderColorSet.Load() {
		b.border.SetColor(b.mouseEnterBorderColor)
	}
	if b.mouseEnterTextColorSet.Load() {
		b.text.SetColor(b.mouseEnterTextColor)
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
		b.text.SetColor(b.originalTextColor)
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
		b.text.SetColor(b.mouseDownTextColor)
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
		b.text.SetColor(b.originalTextColor)
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
	return b.text.Text()
}

func (b *Button) SetText(text string) *Button {
	b.text.SetText(text)
	return b
}

func (b *Button) SetTextColor(rgba color.RGBA) *Button {
	b.text.SetColor(rgba)
	return b
}

func (b *Button) SetMouseEnterTextColor(rgba color.RGBA) *Button {
	b.mouseEnterTextColorSet.Store(true)
	b.stateMutex.Lock()
	b.mouseEnterTextColor = rgba
	b.stateMutex.Unlock()
	return b
}

func (b *Button) SetFontSize(size float32) *Button {
	b.text.SetFontSize(size)
	return b
}

func (b *Button) SetMouseDownTextColor(rgba color.RGBA) *Button {
	b.mouseDownTextColorSet.Store(true)
	b.stateMutex.Lock()
	b.mouseDownTextColor = rgba
	b.stateMutex.Unlock()
	return b
}

func (b *Button) Label() *Label {
	return b.text
}

/******************************************************************************
 New Button Function
******************************************************************************/

func NewButton(circular ...bool) *Button {
	b := &Button{}
	b.View.WindowObjectBase = *NewWindowObject(nil)
	b.text = NewLabel()

	if len(circular) > 0 && circular[0] {
		b.fill = NewDot()
		b.border = NewCircle(thicknessEpsilon * 2)
		b.bounds = NewBoundingRadius()
	} else {
		b.fill = NewQuad()
		b.border = NewSquare(thicknessEpsilon * 2)
		b.bounds = NewBoundingBox()
	}

	b.SetName(defaultButtonName)

	b.defaultLayout()

	return b
}
