package gfx

import (
	"image/color"
	"sync"
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
	disabledFillColorSet     atomic.Bool
	disabledBorderColorSet   atomic.Bool
	disabledTextColorSet     atomic.Bool

	mouseEnterFillColor   color.RGBA
	mouseEnterBorderColor color.RGBA
	mouseEnterTextColor   color.RGBA
	mouseDownFillColor    color.RGBA
	mouseDownBorderColor  color.RGBA
	mouseDownTextColor    color.RGBA
	disabledFillColor     color.RGBA
	disabledBorderColor   color.RGBA
	disabledTextColor     color.RGBA

	originalFillColor   color.RGBA
	originalBorderColor color.RGBA
	originalTextColor   color.RGBA

	clickDispatcher          chan *MouseState
	depressedDispatcher      chan *MouseState
	onMouseClickHandlers     []func(WindowObject, *MouseState)
	onMouseDepressedHandlers []func(WindowObject, *MouseState)
	eventHandlersMutex       sync.Mutex
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
	b.initDispatchers()

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
	if !b.Initialized() {
		return
	}

	if b.clickDispatcher != nil {
		close(b.clickDispatcher)
	}

	if b.depressedDispatcher != nil {
		close(b.depressedDispatcher)
	}

	b.text.Close()
	b.bounds.Close()
	b.View.Close()
}

func (b *Button) SetEnabled(enabled bool) Object {
	b.View.SetEnabled(enabled)
	b.bounds.SetEnabled(enabled)

	b.stateMutex.Lock()
	if enabled {
		b.fill.SetColor(b.originalFillColor)
		b.border.SetColor(b.originalBorderColor)
		b.text.SetColor(b.originalTextColor)
	} else {
		if b.disabledFillColorSet.Load() {
			b.fill.SetColor(b.disabledFillColor)
		}
		if b.disabledBorderColorSet.Load() {
			b.border.SetColor(b.disabledBorderColor)
		}
		if b.disabledTextColorSet.Load() {
			b.text.SetColor(b.disabledTextColor)
		}
	}
	b.stateMutex.Unlock()

	return b
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

func (b *Button) Resize(newWidth, newHeight int) {
	b.View.Resize(newWidth, newHeight)
	b.text.Resize(newWidth, newHeight)
	b.bounds.Resize(newWidth, newHeight)
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

func (b *Button) SetParent(parent WindowObject, recursive ...bool) WindowObject {
	b.View.SetParent(parent, recursive...)
	b.fill.SetParent(b)
	b.border.SetParent(b)
	b.text.SetParent(b)
	b.bounds.SetParent(b)
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

func (b *Button) initDispatchers() {
	b.clickDispatcher = make(chan *MouseState, 64)
	b.depressedDispatcher = make(chan *MouseState, 1024)
	b.bounds.OnPMouseClick(func(_ WindowObject, ms *MouseState) {
		select {
		case b.clickDispatcher <- ms:
		default:
		}
	})
	b.bounds.OnPMouseDepressed(func(_ WindowObject, ms *MouseState) {
		select {
		case b.depressedDispatcher <- ms:
		default:
		}
	})

	go b.handleClick()
	go b.handleDepressed()
}

func (b *Button) handleClick() {
	for {
		select {
		case ms, ok := <-b.clickDispatcher:
			if !ok {
				return
			}

			b.eventHandlersMutex.Lock()
			for _, handler := range b.onMouseClickHandlers {
				handler(b, ms)
			}
			b.eventHandlersMutex.Unlock()
		}
	}
}

func (b *Button) handleDepressed() {
	for {
		select {
		case ms, ok := <-b.depressedDispatcher:
			if !ok {
				return
			}

			b.eventHandlersMutex.Lock()
			for _, handler := range b.onMouseDepressedHandlers {
				handler(b, ms)
			}
			b.eventHandlersMutex.Unlock()
		}
	}
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
	b.fill.SetColor(b.originalFillColor)
	b.border.SetColor(b.originalBorderColor)
	b.text.SetColor(b.originalTextColor)
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
	b.fill.SetColor(b.originalFillColor)
	b.border.SetColor(b.originalBorderColor)
	b.text.SetColor(b.originalTextColor)
}

func (b *Button) OnDepressed(handler func(sender WindowObject, mouseState *MouseState)) *Button {
	b.eventHandlersMutex.Lock()
	b.onMouseDepressedHandlers = append(b.onMouseDepressedHandlers, handler)
	b.eventHandlersMutex.Unlock()
	return b
}

func (b *Button) OnClick(handler func(sender WindowObject, mouseState *MouseState)) *Button {
	b.eventHandlersMutex.Lock()
	b.onMouseClickHandlers = append(b.onMouseClickHandlers, handler)
	b.eventHandlersMutex.Unlock()
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

func (b *Button) SetDisabledFillColor(rgba color.RGBA) *Button {
	b.disabledFillColorSet.Store(true)
	b.stateMutex.Lock()
	b.disabledFillColor = rgba
	b.stateMutex.Unlock()
	return b
}

func (b *Button) SetDisabledBorderColor(rgba color.RGBA) *Button {
	b.disabledBorderColorSet.Store(true)
	b.stateMutex.Lock()
	b.disabledBorderColor = rgba
	b.stateMutex.Unlock()
	return b
}

func (b *Button) SetDisabledTextColor(rgba color.RGBA) *Button {
	b.disabledTextColorSet.Store(true)
	b.stateMutex.Lock()
	b.disabledTextColor = rgba
	b.stateMutex.Unlock()
	return b
}

func (b *Button) SetMouseSurface(surface MouseSurface) {
	b.bounds.SetMouseSurface(surface)
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

func (b *Button) SetFontSize(size float32) *Button {
	b.text.SetFontSize(size)
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
	b.View.WindowObjectBase = *NewWindowObject()
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
