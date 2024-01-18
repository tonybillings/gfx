package gfx

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"sync"
)

/******************************************************************************
 Keyboard Input
******************************************************************************/

type KeyEventHandler struct {
	Key      glfw.Key
	Action   glfw.Action
	Callback func(*Window, glfw.Key, glfw.Action)
}

/******************************************************************************
 Mouse Input
******************************************************************************/

type MouseState struct {
	x, y                float64
	leftDown, rightDown bool
	stateMutex          sync.Mutex
}

func (s *MouseState) Get() (x, y float64, leftIsDown, rightIsDown bool) {
	s.stateMutex.Lock()
	x = s.x
	y = s.y
	leftIsDown = s.leftDown
	rightIsDown = s.rightDown
	s.stateMutex.Unlock()
	return
}

func (s *MouseState) UpdatePosition(x, y float64) {
	s.stateMutex.Lock()
	s.x = x
	s.y = y
	s.stateMutex.Unlock()
}

func (s *MouseState) UpdateButton(buttonIndex int, isDown bool) {
	s.stateMutex.Lock()
	switch buttonIndex {
	case 0:
		s.leftDown = isDown
	case 1:
		s.rightDown = isDown
	default:
		panic("invalid mouse button index (expecting 0 or 1)")
	}
	s.stateMutex.Unlock()
}

/******************************************************************************
 BoundingBox
******************************************************************************/

type BoundingBox struct {
	WindowObjectBase

	mouseOver         bool
	clickBegun        bool
	lastMouseLeftDown bool

	onMouseEnterHandlers []func(WindowObject, *MouseState)
	onMouseLeaveHandlers []func(WindowObject, *MouseState)
	onMouseDownHandlers  []func(WindowObject, *MouseState)
	onMouseUpHandlers    []func(WindowObject, *MouseState)
	onClickHandlers      []func(WindowObject, *MouseState)
}

func (b *BoundingBox) Init(window *Window) bool {
	if !b.WindowObjectBase.Init(window) {
		return false
	}

	b.window.EnableMouseTracking()
	b.initialized.Store(true)
	return true
}

func (b *BoundingBox) Update(_ int64) bool {
	x, y, leftDown, rightDown := b.window.Mouse().Get()
	xNorm := float32(((x / float64(b.window.Width())) * 2.0) - 1.0)
	yNorm := float32(((y / float64(b.window.Height())) * -2.0) + 1.0)
	position := b.WorldPosition()
	worldScale := b.WorldScale()
	scale := [2]float32{}
	if b.maintainAspectRatio {
		width := float32(b.window.width.Load())
		height := float32(b.window.height.Load())
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
	xScaleHalf := scale[0] * 0.5
	yScaleHalf := scale[1] * 0.5
	left := position[0] - xScaleHalf*2.0
	top := position[1] - yScaleHalf*2.0
	right := position[0] + xScaleHalf*2.0
	bottom := position[1] + yScaleHalf*2.0

	mouseState := &MouseState{x, y, leftDown, rightDown, sync.Mutex{}}

	if !b.mouseOver && xNorm >= left && xNorm <= right && yNorm >= top && yNorm <= bottom {
		b.mouseOver = true
		if !b.clickBegun {
			for _, f := range b.onMouseEnterHandlers {
				if parent := b.parent.Load(); parent != nil {
					f(*parent, mouseState)
				} else {
					f(b, mouseState)
				}
			}
		}
	}

	if b.mouseOver && (xNorm < left || xNorm > right || yNorm < top || yNorm > bottom) {
		b.mouseOver = false
		if !b.clickBegun {
			for _, f := range b.onMouseLeaveHandlers {
				if parent := b.parent.Load(); parent != nil {
					f(*parent, mouseState)
				} else {
					f(b, mouseState)
				}
			}
		}
	}

	if !b.clickBegun && b.mouseOver && leftDown && !b.lastMouseLeftDown {
		b.clickBegun = true
		for _, f := range b.onMouseDownHandlers {
			if parent := b.parent.Load(); parent != nil {
				f(*parent, mouseState)
			} else {
				f(b, mouseState)
			}
		}
	} else if b.clickBegun && b.mouseOver && !leftDown {
		b.clickBegun = false
		for _, f := range b.onClickHandlers {
			if parent := b.parent.Load(); parent != nil {
				f(*parent, mouseState)
			} else {
				f(b, mouseState)
			}
		}
		for _, f := range b.onMouseUpHandlers {
			if parent := b.parent.Load(); parent != nil {
				f(*parent, mouseState)
			} else {
				f(b, mouseState)
			}
		}
	} else if b.clickBegun && !b.mouseOver && !leftDown {
		b.clickBegun = false
		for _, f := range b.onMouseUpHandlers {
			if parent := b.parent.Load(); parent != nil {
				f(*parent, mouseState)
			} else {
				f(b, mouseState)
			}
		}
	}

	b.lastMouseLeftDown = leftDown
	return true
}

func (b *BoundingBox) OnMouseEnter(handler func(sender WindowObject, mouseState *MouseState)) {
	b.onMouseEnterHandlers = append(b.onMouseEnterHandlers, handler)
}

func (b *BoundingBox) OnMouseLeave(handler func(sender WindowObject, mouseState *MouseState)) {
	b.onMouseLeaveHandlers = append(b.onMouseLeaveHandlers, handler)
}

func (b *BoundingBox) OnMouseDown(handler func(sender WindowObject, mouseState *MouseState)) {
	b.onMouseDownHandlers = append(b.onMouseDownHandlers, handler)
}

func (b *BoundingBox) OnMouseUp(handler func(sender WindowObject, mouseState *MouseState)) {
	b.onMouseUpHandlers = append(b.onMouseUpHandlers, handler)
}

func (b *BoundingBox) OnClick(handler func(sender WindowObject, mouseState *MouseState)) {
	b.onClickHandlers = append(b.onClickHandlers, handler)
}

func NewBoundingBox() *BoundingBox {
	return &BoundingBox{
		WindowObjectBase: *NewObject(nil),
	}
}
