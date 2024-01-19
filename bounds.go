package gfx

import "sync"

/******************************************************************************
 Interface
******************************************************************************/

type BoundingObject interface {
	WindowObject

	OnMouseEnter(func(WindowObject, *MouseState))
	OnMouseLeave(func(WindowObject, *MouseState))
	OnMouseDown(func(WindowObject, *MouseState))
	OnMouseUp(func(WindowObject, *MouseState))
	OnDepressed(func(WindowObject, *MouseState))
	OnClick(func(WindowObject, *MouseState))
}

/******************************************************************************
 BoundingObjectBase
******************************************************************************/

type BoundingObjectBase struct {
	WindowObjectBase

	mouseOver         bool
	clickBegun        bool
	lastMouseLeftDown bool

	onMouseEnterHandlers []func(WindowObject, *MouseState)
	onMouseLeaveHandlers []func(WindowObject, *MouseState)
	onMouseDownHandlers  []func(WindowObject, *MouseState)
	onMouseUpHandlers    []func(WindowObject, *MouseState)
	onDepressedHandlers  []func(WindowObject, *MouseState)
	onClickHandlers      []func(WindowObject, *MouseState)
}

func (o *BoundingObjectBase) Init(window *Window) bool {
	if !o.WindowObjectBase.Init(window) {
		return false
	}

	o.window.EnableMouseTracking()
	o.initialized.Store(true)
	return true
}

func (o *BoundingObjectBase) OnMouseEnter(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onMouseEnterHandlers = append(o.onMouseEnterHandlers, handler)
}

func (o *BoundingObjectBase) OnMouseLeave(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onMouseLeaveHandlers = append(o.onMouseLeaveHandlers, handler)
}

func (o *BoundingObjectBase) OnMouseDown(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onMouseDownHandlers = append(o.onMouseDownHandlers, handler)
}

func (o *BoundingObjectBase) OnMouseUp(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onMouseUpHandlers = append(o.onMouseUpHandlers, handler)
}

func (o *BoundingObjectBase) OnDepressed(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onDepressedHandlers = append(o.onDepressedHandlers, handler)
}

func (o *BoundingObjectBase) OnClick(handler func(sender WindowObject, mouseState *MouseState)) {
	o.onClickHandlers = append(o.onClickHandlers, handler)
}

func NewBoundingObject() *BoundingObjectBase {
	return &BoundingObjectBase{
		WindowObjectBase: *NewObject(nil),
	}
}

/******************************************************************************
 BoundingBox
******************************************************************************/

type BoundingBox struct {
	BoundingObjectBase
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
	} else if b.clickBegun && leftDown {
		for _, f := range b.onDepressedHandlers {
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

func NewBoundingBox() *BoundingBox {
	return &BoundingBox{
		BoundingObjectBase: *NewBoundingObject(),
	}
}

/******************************************************************************
 BoundingRadius
******************************************************************************/

type BoundingRadius struct {
	BoundingObjectBase
}

func (r *BoundingRadius) Update(_ int64) bool {
	x, y, leftDown, rightDown := r.window.Mouse().Get()
	xNorm := float32(((x / float64(r.window.Width())) * 2.0) - 1.0)
	yNorm := float32(((y / float64(r.window.Height())) * -2.0) + 1.0)
	position := r.WorldPosition()
	worldScale := r.WorldScale()
	scale := [2]float32{}
	if r.maintainAspectRatio {
		width := float32(r.window.width.Load())
		height := float32(r.window.height.Load())
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

	if scale[0] == 0 || scale[1] == 0 {
		return true
	}

	dx := (xNorm - position[0]) / scale[0]
	dy := (yNorm - position[1]) / scale[1]
	mouseOver := dx*dx+dy*dy <= 1.0

	mouseState := &MouseState{x, y, leftDown, rightDown, sync.Mutex{}}

	if !r.mouseOver && mouseOver {
		r.mouseOver = true
		if !r.clickBegun {
			for _, f := range r.onMouseEnterHandlers {
				if parent := r.parent.Load(); parent != nil {
					f(*parent, mouseState)
				} else {
					f(r, mouseState)
				}
			}
		}
	}

	if r.mouseOver && !mouseOver {
		r.mouseOver = false
		if !r.clickBegun {
			for _, f := range r.onMouseLeaveHandlers {
				if parent := r.parent.Load(); parent != nil {
					f(*parent, mouseState)
				} else {
					f(r, mouseState)
				}
			}
		}
	}

	if !r.clickBegun && r.mouseOver && leftDown && !r.lastMouseLeftDown {
		r.clickBegun = true
		for _, f := range r.onMouseDownHandlers {
			if parent := r.parent.Load(); parent != nil {
				f(*parent, mouseState)
			} else {
				f(r, mouseState)
			}
		}
	} else if r.clickBegun && r.mouseOver && !leftDown {
		r.clickBegun = false
		for _, f := range r.onClickHandlers {
			if parent := r.parent.Load(); parent != nil {
				f(*parent, mouseState)
			} else {
				f(r, mouseState)
			}
		}
		for _, f := range r.onMouseUpHandlers {
			if parent := r.parent.Load(); parent != nil {
				f(*parent, mouseState)
			} else {
				f(r, mouseState)
			}
		}
	} else if r.clickBegun && !r.mouseOver && !leftDown {
		r.clickBegun = false
		for _, f := range r.onMouseUpHandlers {
			if parent := r.parent.Load(); parent != nil {
				f(*parent, mouseState)
			} else {
				f(r, mouseState)
			}
		}
	} else if r.clickBegun && leftDown {
		for _, f := range r.onDepressedHandlers {
			if parent := r.parent.Load(); parent != nil {
				f(*parent, mouseState)
			} else {
				f(r, mouseState)
			}
		}
	}

	r.lastMouseLeftDown = leftDown
	return true
}

func NewBoundingRadius() *BoundingRadius {
	return &BoundingRadius{
		BoundingObjectBase: *NewBoundingObject(),
	}
}
