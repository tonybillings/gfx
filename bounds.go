package gfx

import (
	"github.com/go-gl/mathgl/mgl32"
	"sync"
	"sync/atomic"
)

const (
	defaultBoundingRadiusName = "BoundingRadius"
	defaultBoundingBoxName    = "BoundingBox"
)

/******************************************************************************
 BoundingObject
******************************************************************************/

type BoundingObject interface {
	WindowObject

	OnMouseEnter(func(WindowObject, *MouseState))
	OnMouseLeave(func(WindowObject, *MouseState))
	OnPMouseDown(func(WindowObject, *MouseState))
	OnSMouseDown(func(WindowObject, *MouseState))
	OnPMouseUp(func(WindowObject, *MouseState))
	OnSMouseUp(func(WindowObject, *MouseState))
	OnPMouseDepressed(func(WindowObject, *MouseState))
	OnSMouseDepressed(func(WindowObject, *MouseState))
	OnPMouseClick(func(WindowObject, *MouseState))
	OnSMouseClick(func(WindowObject, *MouseState))

	MouseSurface() MouseSurface
	SetMouseSurface(MouseSurface)

	LocalMouse() *MouseState
	MouseOver() bool
}

/******************************************************************************
 BoundingObjectBase
******************************************************************************/

type BoundingObjectBase struct {
	WindowObjectBase

	onMouseEnterHandlers      []func(WindowObject, *MouseState)
	onMouseLeaveHandlers      []func(WindowObject, *MouseState)
	onPMouseDownHandlers      []func(WindowObject, *MouseState)
	onSMouseDownHandlers      []func(WindowObject, *MouseState)
	onPMouseUpHandlers        []func(WindowObject, *MouseState)
	onSMouseUpHandlers        []func(WindowObject, *MouseState)
	onPMouseDepressedHandlers []func(WindowObject, *MouseState)
	onSMouseDepressedHandlers []func(WindowObject, *MouseState)
	onPMouseClickHandlers     []func(WindowObject, *MouseState)
	onSMouseClickHandlers     []func(WindowObject, *MouseState)

	mouseSurface    MouseSurface
	mouseOver       atomic.Bool
	mouseState      MouseState
	winMouseState   *MouseState
	mouseStateMutex sync.Mutex
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (o *BoundingObjectBase) Init() bool {
	if o.Initialized() {
		return true
	}

	if o.window != nil {
		o.window.EnableMouseTracking()
		if o.mouseSurface == nil {
			o.mouseSurface = o.window
		}
	}

	if o.mouseSurface == nil {
		panic("mouseSurface cannot be nil")
	}

	return o.WindowObjectBase.Init()
}

/******************************************************************************
 BoundingObject Implementation
******************************************************************************/

func (o *BoundingObjectBase) OnMouseEnter(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onMouseEnterHandlers = append(o.onMouseEnterHandlers, handler)
}

func (o *BoundingObjectBase) OnMouseLeave(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onMouseLeaveHandlers = append(o.onMouseLeaveHandlers, handler)
}

func (o *BoundingObjectBase) OnPMouseDown(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onPMouseDownHandlers = append(o.onPMouseDownHandlers, handler)
}

func (o *BoundingObjectBase) OnSMouseDown(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onSMouseDownHandlers = append(o.onSMouseDownHandlers, handler)
}

func (o *BoundingObjectBase) OnPMouseUp(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onPMouseUpHandlers = append(o.onPMouseUpHandlers, handler)
}

func (o *BoundingObjectBase) OnSMouseUp(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onSMouseUpHandlers = append(o.onSMouseUpHandlers, handler)
}

func (o *BoundingObjectBase) OnPMouseDepressed(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onPMouseDepressedHandlers = append(o.onPMouseDepressedHandlers, handler)
}

func (o *BoundingObjectBase) OnSMouseDepressed(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onSMouseDepressedHandlers = append(o.onSMouseDepressedHandlers, handler)
}

func (o *BoundingObjectBase) OnPMouseClick(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onPMouseClickHandlers = append(o.onPMouseClickHandlers, handler)
}

func (o *BoundingObjectBase) OnSMouseClick(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onSMouseClickHandlers = append(o.onSMouseClickHandlers, handler)
}

func (o *BoundingObjectBase) MouseSurface() MouseSurface {
	o.mouseStateMutex.Lock()
	surface := o.mouseSurface
	o.mouseStateMutex.Unlock()
	return surface
}

func (o *BoundingObjectBase) SetMouseSurface(surface MouseSurface) {
	o.mouseStateMutex.Lock()
	o.mouseSurface = surface
	o.mouseStateMutex.Unlock()
}

func (o *BoundingObjectBase) LocalMouse() *MouseState {
	o.mouseStateMutex.Lock()
	ms := o.mouseState
	o.mouseStateMutex.Unlock()
	return &ms
}

func (o *BoundingObjectBase) MouseOver() bool {
	return o.mouseOver.Load()
}

/******************************************************************************
 BoundingObjectBase Functions
******************************************************************************/

func (o *BoundingObjectBase) beginUpdate() (*MouseState, mgl32.Vec3, [2]float32) {
	winMouse := o.mouseSurface.Mouse()
	position := o.WorldPosition()
	worldScale := o.WorldScale()
	scale := [2]float32{}
	if o.maintainAspectRatio {
		width := float32(o.mouseSurface.Width())
		height := float32(o.mouseSurface.Height())
		switch {
		case width > height:
			scale[0] = worldScale[0] * (height / width)
			scale[1] = worldScale[1]
		case height > width:
			scale[0] = worldScale[0]
			scale[1] = worldScale[1] * (width / height)
		default:
			scale[0] = worldScale[0]
			scale[1] = worldScale[1]
		}
	} else {
		scale[0] = worldScale[0]
		scale[1] = worldScale[1]
	}

	return winMouse, position, scale
}

func (o *BoundingObjectBase) endUpdate(winMouse *MouseState, mouseOver bool, xLocal, yLocal float32) {
	raiseMouseEnter := false
	raiseMouseLeave := false

	raisePMouseDown := false
	raisePMouseDepressed := false
	raisePMouseUp := false
	raisePMouseClick := false

	raiseSMouseDown := false
	raiseSMouseDepressed := false
	raiseSMouseUp := false
	raiseSMouseClick := false

	lastMouseOver := o.mouseOver.Load()

	if !lastMouseOver && mouseOver {
		o.mouseOver.Store(true)
		raiseMouseEnter = true
	}

	if lastMouseOver && !mouseOver {
		o.mouseOver.Store(false)
		raiseMouseLeave = true
	}

	raisePMouseDown, raisePMouseUp, raisePMouseClick, raisePMouseDepressed = o.endUpdatePrimaryButton(mouseOver, winMouse.PrimaryDown)
	raiseSMouseDown, raiseSMouseUp, raiseSMouseClick, raiseSMouseDepressed = o.endUpdateSecondaryButton(mouseOver, winMouse.SecondaryDown)

	o.mouseStateMutex.Lock()
	o.mouseState.X = xLocal
	o.mouseState.Y = yLocal
	o.mouseStateMutex.Unlock()
	localMouse := o.mouseState
	o.winMouseState = winMouse

	if raiseMouseEnter {
		o.raiseMouseEnter(&localMouse)
	}

	if raiseMouseLeave {
		o.raiseMouseLeave(&localMouse)
	}

	if raisePMouseDown {
		o.raisePMouseDown(&localMouse)
	}

	if raisePMouseDepressed {
		o.raisePMouseDepressed(&localMouse)
	}

	if raisePMouseUp {
		o.raisePMouseUp(&localMouse)
	}

	if raisePMouseClick {
		o.raisePMouseClick(&localMouse)
	}

	if raiseSMouseDown {
		o.raiseSMouseDown(&localMouse)
	}

	if raiseSMouseDepressed {
		o.raiseSMouseDepressed(&localMouse)
	}

	if raiseSMouseUp {
		o.raiseSMouseUp(&localMouse)
	}

	if raiseSMouseClick {
		o.raiseSMouseClick(&localMouse)
	}
}

func (o *BoundingObjectBase) endUpdatePrimaryButton(mouseOver, winMousePrimaryDown bool) (
	raisePMouseDown, raisePMouseUp, raisePMouseClick, raisePMouseDepressed bool) {
	if !o.winMouseState.PrimaryDown && mouseOver && winMousePrimaryDown {
		o.mouseStateMutex.Lock()
		o.mouseState.PrimaryDown = true
		o.mouseStateMutex.Unlock()
		raisePMouseDown = true
	} else if o.mouseState.PrimaryDown && mouseOver && !winMousePrimaryDown {
		o.mouseStateMutex.Lock()
		o.mouseState.PrimaryDown = false
		o.mouseStateMutex.Unlock()
		raisePMouseUp = true
		raisePMouseClick = true
	} else if o.mouseState.PrimaryDown && !mouseOver && !winMousePrimaryDown {
		o.mouseStateMutex.Lock()
		o.mouseState.PrimaryDown = false
		o.mouseStateMutex.Unlock()
		raisePMouseUp = true
	} else if o.mouseState.PrimaryDown && winMousePrimaryDown {
		raisePMouseDepressed = true
	}

	return
}

func (o *BoundingObjectBase) endUpdateSecondaryButton(mouseOver, winMouseSecondaryDown bool) (
	raiseSMouseDown, raiseSMouseUp, raiseSMouseClick, raiseSMouseDepressed bool) {
	if !o.winMouseState.SecondaryDown && mouseOver && winMouseSecondaryDown {
		o.mouseStateMutex.Lock()
		o.mouseState.SecondaryDown = true
		o.mouseStateMutex.Unlock()
		raiseSMouseDown = true
	} else if o.mouseState.SecondaryDown && mouseOver && !winMouseSecondaryDown {
		o.mouseStateMutex.Lock()
		o.mouseState.SecondaryDown = false
		o.mouseStateMutex.Unlock()
		raiseSMouseUp = true
		raiseSMouseClick = true
	} else if o.mouseState.SecondaryDown && !mouseOver && !winMouseSecondaryDown {
		o.mouseStateMutex.Lock()
		o.mouseState.SecondaryDown = false
		o.mouseStateMutex.Unlock()
		raiseSMouseUp = true
	} else if o.mouseState.SecondaryDown && winMouseSecondaryDown {
		raiseSMouseDepressed = true
	}

	return
}

func (o *BoundingObjectBase) raiseMouseEnter(localMouse *MouseState) {
	for _, f := range o.onMouseEnterHandlers {
		if parent := o.parent.Load(); parent != nil {
			f(*parent, localMouse)
		} else {
			f(o, localMouse)
		}
	}
}

func (o *BoundingObjectBase) raiseMouseLeave(localMouse *MouseState) {
	for _, f := range o.onMouseLeaveHandlers {
		if parent := o.parent.Load(); parent != nil {
			f(*parent, localMouse)
		} else {
			f(o, localMouse)
		}
	}
}

func (o *BoundingObjectBase) raisePMouseDown(localMouse *MouseState) {
	for _, f := range o.onPMouseDownHandlers {
		if parent := o.parent.Load(); parent != nil {
			f(*parent, localMouse)
		} else {
			f(o, localMouse)
		}
	}
}

func (o *BoundingObjectBase) raisePMouseDepressed(localMouse *MouseState) {
	for _, f := range o.onPMouseDepressedHandlers {
		if parent := o.parent.Load(); parent != nil {
			f(*parent, localMouse)
		} else {
			f(o, localMouse)
		}
	}
}

func (o *BoundingObjectBase) raisePMouseUp(localMouse *MouseState) {
	for _, f := range o.onPMouseUpHandlers {
		if parent := o.parent.Load(); parent != nil {
			f(*parent, localMouse)
		} else {
			f(o, localMouse)
		}
	}
}

func (o *BoundingObjectBase) raisePMouseClick(localMouse *MouseState) {
	for _, f := range o.onPMouseClickHandlers {
		if parent := o.parent.Load(); parent != nil {
			f(*parent, localMouse)
		} else {
			f(o, localMouse)
		}
	}
}

func (o *BoundingObjectBase) raiseSMouseDown(localMouse *MouseState) {
	for _, f := range o.onSMouseDownHandlers {
		if parent := o.parent.Load(); parent != nil {
			f(*parent, localMouse)
		} else {
			f(o, localMouse)
		}
	}
}

func (o *BoundingObjectBase) raiseSMouseDepressed(localMouse *MouseState) {
	for _, f := range o.onSMouseDepressedHandlers {
		if parent := o.parent.Load(); parent != nil {
			f(*parent, localMouse)
		} else {
			f(o, localMouse)
		}
	}
}

func (o *BoundingObjectBase) raiseSMouseUp(localMouse *MouseState) {
	for _, f := range o.onSMouseUpHandlers {
		if parent := o.parent.Load(); parent != nil {
			f(*parent, localMouse)
		} else {
			f(o, localMouse)
		}
	}
}

func (o *BoundingObjectBase) raiseSMouseClick(localMouse *MouseState) {
	for _, f := range o.onSMouseClickHandlers {
		if parent := o.parent.Load(); parent != nil {
			f(*parent, localMouse)
		} else {
			f(o, localMouse)
		}
	}
}

/******************************************************************************
 New BoundingObjectBase Function
******************************************************************************/

func NewBoundingObject() *BoundingObjectBase {
	bo := &BoundingObjectBase{
		WindowObjectBase: *NewWindowObject(),
	}

	bo.winMouseState = &bo.mouseState
	return bo
}

/******************************************************************************
 BoundingBox
******************************************************************************/

type BoundingBox struct {
	BoundingObjectBase
}

func (b *BoundingBox) Update(_ int64) bool {
	if !b.enabled.Load() {
		return false
	}

	winMouse, position, scale := b.beginUpdate()
	if scale[0] == 0 || scale[1] == 0 {
		return true
	}

	xScaleHalf := scale[0] * 0.5
	yScaleHalf := scale[1] * 0.5
	left := position[0] - xScaleHalf*2.0
	top := position[1] + yScaleHalf*2.0
	right := position[0] + xScaleHalf*2.0
	bottom := position[1] - yScaleHalf*2.0
	width := right - left
	height := top - bottom
	mouseOver := winMouse.X >= left && winMouse.X <= right && winMouse.Y <= top && winMouse.Y >= bottom
	xLocal := (((winMouse.X - left) / width) * 2.0) - 1.0
	yLocal := (((winMouse.Y - bottom) / height) * 2.0) - 1.0

	b.endUpdate(winMouse, mouseOver, xLocal, yLocal)

	return true
}

func NewBoundingBox() *BoundingBox {
	bb := &BoundingBox{
		BoundingObjectBase: *NewBoundingObject(),
	}

	bb.visible.Store(false)
	bb.SetName(defaultBoundingBoxName)
	return bb
}

/******************************************************************************
 BoundingRadius
******************************************************************************/

type BoundingRadius struct {
	BoundingObjectBase
}

func (r *BoundingRadius) Update(_ int64) bool {
	if !r.enabled.Load() {
		return false
	}

	winMouse, position, scale := r.beginUpdate()
	if scale[0] == 0 || scale[1] == 0 {
		return true
	}

	xScaleHalf := scale[0] * 0.5
	yScaleHalf := scale[1] * 0.5
	left := position[0] - xScaleHalf*2.0
	top := position[1] - yScaleHalf*2.0
	right := position[0] + xScaleHalf*2.0
	bottom := position[1] + yScaleHalf*2.0
	width := right - left
	height := top - bottom
	dx := (winMouse.X - position[0]) / scale[0]
	dy := (winMouse.Y - position[1]) / scale[1]
	mouseOver := dx*dx+dy*dy <= 1.0
	xLocal := (((winMouse.X - left) / width) * 2.0) - 1.0
	yLocal := (((winMouse.Y - bottom) / height) * -2.0) + 1.0

	r.endUpdate(winMouse, mouseOver, xLocal, yLocal)

	return true
}

func NewBoundingRadius() *BoundingRadius {
	br := &BoundingRadius{
		BoundingObjectBase: *NewBoundingObject(),
	}

	br.visible.Store(false)
	br.SetName(defaultBoundingRadiusName)
	return br
}
