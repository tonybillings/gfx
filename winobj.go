package gfx

import (
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"sync"
	"sync/atomic"
)

/******************************************************************************
 WindowObject
******************************************************************************/

// WindowObject objects are drawable objects that are "window-aware," contain
// additional rendering and positioning properties, and those that allow them
// to form a hierarchical structure of parent/child objects.
type WindowObject interface {
	DrawableObject
	Transform
	Resizer

	Window() *Window
	SetWindow(*Window) WindowObject

	Color() color.RGBA
	SetColor(color.RGBA) WindowObject

	Opacity() uint8
	SetOpacity(uint8) WindowObject

	BlurIntensity() float32
	SetBlurIntensity(float32) WindowObject
	BlurEnabled() bool
	SetBlurEnabled(bool) WindowObject

	Width() float32
	Height() float32
	HalfWidth() float32
	HalfHeight() float32

	MaintainAspectRatio() bool
	SetMaintainAspectRatio(bool) WindowObject

	Anchor() Anchor
	SetAnchor(Anchor) WindowObject

	Margin() *Margin
	SetMargin(Margin) WindowObject
	SetMarginTop(float32) WindowObject
	SetMarginRight(float32) WindowObject
	SetMarginBottom(float32) WindowObject
	SetMarginLeft(float32) WindowObject

	Bounds() *Bounds

	OnResize(func(int, int))
	RefreshLayout()

	Parent() WindowObject
	SetParent(WindowObject, ...bool) WindowObject
	AddChild(WindowObject) WindowObject
	AddChildren(...WindowObject) WindowObject
	RemoveChild(WindowObject)
	RemoveChildren()
	Child(string) WindowObject
	Children() []WindowObject
}

/******************************************************************************
 WindowObjectBase
******************************************************************************/

type WindowObjectBase struct {
	DrawableObjectBase
	ObjectTransform

	window *Window

	maintainAspectRatio bool

	color mgl32.Vec4

	blurIntensity float32
	blurEnabled   bool

	anchor Anchor
	margin Margin
	bounds Bounds

	anchorSet   bool
	positionBak mgl32.Vec3

	onResizeHandlers []func(int, int)

	parent   atomic.Pointer[WindowObject]
	children []WindowObject

	stateMutex sync.Mutex
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (o *WindowObjectBase) Init() (ok bool) {
	if ok = o.initChildren(); !ok {
		return false
	}

	return o.ObjectBase.Init()
}

func (o *WindowObjectBase) Update(deltaTime int64) (ok bool) {
	if ok = o.ObjectBase.Update(deltaTime); !ok {
		return
	}

	if o.boundsChanged {
		o.ObjectTransform.stateMutex.Lock()
		o.boundsChanged = false
		o.ObjectTransform.stateMutex.Unlock()
		o.RefreshLayout()
	}

	return o.updateChildren(deltaTime)
}

func (o *WindowObjectBase) Close() {
	if !o.Initialized() {
		return
	}

	o.closeChildren()
	o.ObjectBase.Close()
}

/******************************************************************************
 DrawableObject Implementation
******************************************************************************/

func (o *WindowObjectBase) Draw(deltaTime int64) (ok bool) {
	if ok = o.DrawableObjectBase.Draw(deltaTime); !ok {
		return
	}

	return o.drawChildren(deltaTime)
}

/******************************************************************************
 Resizer Implementation
******************************************************************************/

func (o *WindowObjectBase) Resize(newWidth, newHeight int) {
	if !o.initialized.Load() {
		return
	}

	o.RefreshLayout()

	for _, c := range o.children {
		c.Resize(newWidth, newHeight)
	}

	for _, f := range o.onResizeHandlers {
		f(newWidth, newHeight)
	}
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (o *WindowObjectBase) Window() *Window {
	return o.window
}

func (o *WindowObjectBase) SetWindow(window *Window) WindowObject {
	o.stateMutex.Lock()
	o.window = window
	for _, c := range o.children {
		c.SetWindow(window)
	}
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) Color() color.RGBA {
	o.stateMutex.Lock()
	rgba := FloatArrayToRgba(o.color)
	o.stateMutex.Unlock()
	return rgba
}

func (o *WindowObjectBase) SetColor(rgba color.RGBA) WindowObject {
	o.stateMutex.Lock()
	o.color[0] = float32(rgba.R) / 255.0
	o.color[1] = float32(rgba.G) / 255.0
	o.color[2] = float32(rgba.B) / 255.0
	o.color[3] = float32(rgba.A) / 255.0
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) Opacity() uint8 {
	o.stateMutex.Lock()
	a := uint8(o.color[3] * 255.0)
	o.stateMutex.Unlock()
	return a
}

func (o *WindowObjectBase) SetOpacity(alpha uint8) WindowObject {
	o.stateMutex.Lock()
	o.color[3] = float32(alpha) / 255.0
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) BlurIntensity() float32 {
	o.stateMutex.Lock()
	intensity := o.blurIntensity
	o.stateMutex.Unlock()
	return intensity
}

func (o *WindowObjectBase) SetBlurIntensity(intensity float32) WindowObject {
	o.stateMutex.Lock()
	o.blurIntensity = intensity
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) BlurEnabled() bool {
	o.stateMutex.Lock()
	enabled := o.blurEnabled
	o.stateMutex.Unlock()
	return enabled
}

func (o *WindowObjectBase) SetBlurEnabled(enabled bool) WindowObject {
	o.stateMutex.Lock()
	o.blurEnabled = enabled
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) Width() float32 {
	width := o.WorldScale().X() * 2.0
	if o.maintainAspectRatio {
		if w := o.Window(); w != nil {
			width = w.ScaleX(width)
		} else {
			panic("width can only be calculated once Window has been set on the object")
		}
	}
	return width
}

func (o *WindowObjectBase) Height() float32 {
	height := o.WorldScale().Y() * 2.0
	if o.maintainAspectRatio {
		if w := o.Window(); w != nil {
			height = w.ScaleY(height)
		} else {
			panic("height can only be calculated once Window has been set on the object")
		}
	}
	return height
}

func (o *WindowObjectBase) HalfWidth() float32 {
	return o.Width() * 0.5
}

func (o *WindowObjectBase) HalfHeight() float32 {
	return o.Height() * 0.5
}

func (o *WindowObjectBase) MaintainAspectRatio() bool {
	return o.maintainAspectRatio
}

func (o *WindowObjectBase) SetMaintainAspectRatio(maintainAspectRatio bool) WindowObject {
	o.maintainAspectRatio = maintainAspectRatio
	o.boundsChanged = true
	return o
}

func (o *WindowObjectBase) Anchor() Anchor {
	o.stateMutex.Lock()
	a := o.anchor
	o.stateMutex.Unlock()
	return a
}

func (o *WindowObjectBase) SetAnchor(anchor Anchor) WindowObject {
	o.stateMutex.Lock()
	o.anchor = anchor
	o.boundsChanged = true
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) Margin() *Margin {
	o.stateMutex.Lock()
	m := o.margin
	o.stateMutex.Unlock()
	return &m
}

func (o *WindowObjectBase) SetMargin(margin Margin) WindowObject {
	o.stateMutex.Lock()
	o.margin = margin
	o.boundsChanged = true
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) SetMarginTop(margin float32) WindowObject {
	o.stateMutex.Lock()
	o.margin.Top = margin
	o.boundsChanged = true
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) SetMarginRight(margin float32) WindowObject {
	o.stateMutex.Lock()
	o.margin.Right = margin
	o.boundsChanged = true
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) SetMarginBottom(margin float32) WindowObject {
	o.stateMutex.Lock()
	o.margin.Bottom = margin
	o.boundsChanged = true
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) SetMarginLeft(margin float32) WindowObject {
	o.stateMutex.Lock()
	o.margin.Left = margin
	o.boundsChanged = true
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) Bounds() *Bounds {
	o.stateMutex.Lock()
	bounds := o.bounds
	o.stateMutex.Unlock()
	return &bounds
}

func (o *WindowObjectBase) OnResize(handler func(newWidth, newHeight int)) {
	o.onResizeHandlers = append(o.onResizeHandlers, handler)
}

func (o *WindowObjectBase) RefreshLayout() {
	window := o.Window()
	if window == nil {
		return
	}

	top, right, bottom, left := o.getParentBounds()

	margin := o.Margin()
	margin.Left = window.ScaleX(margin.Left)
	margin.Right = window.ScaleX(margin.Right)
	margin.Top = window.ScaleY(margin.Top)
	margin.Bottom = window.ScaleY(margin.Bottom)

	halfWidth := o.HalfWidth()
	halfHeight := o.HalfHeight()

	leftOffset := halfWidth + margin.Left
	rightOffset := halfWidth + margin.Right
	topOffset := halfHeight + margin.Top
	bottomOffset := halfHeight + margin.Bottom

	if anchor := o.Anchor(); anchor == NoAnchor {
		if o.anchorSet {
			o.anchorSet = false
			o.SetPosition(o.positionBak)
		}
	} else {
		if !o.anchorSet {
			o.anchorSet = true
			o.positionBak = o.Position()
		}

		parentPos := mgl32.Vec3{}
		if p := o.Parent(); p != nil {
			parentPos = p.WorldPosition()
		}

		o.stateMutex.Lock()
		switch anchor {
		case TopLeft:
			o.position = mgl32.Vec3{(left + leftOffset) - parentPos.X(), (top - topOffset) - parentPos.Y()}
		case MiddleLeft:
			o.position = mgl32.Vec3{(left + leftOffset) - parentPos.X(), -margin.Top + margin.Bottom}
		case BottomLeft:
			o.position = mgl32.Vec3{(left + leftOffset) - parentPos.X(), (bottom + bottomOffset) - parentPos.Y()}
		case TopCenter:
			o.position = mgl32.Vec3{-margin.Right + margin.Left, (top - topOffset) - parentPos.Y()}
		case Center:
			o.position = mgl32.Vec3{-margin.Right + margin.Left, -margin.Top + margin.Bottom}
		case BottomCenter:
			o.position = mgl32.Vec3{-margin.Right + margin.Left, (bottom + bottomOffset) - parentPos.Y()}
		case TopRight:
			o.position = mgl32.Vec3{(right - rightOffset) - parentPos.X(), (top - topOffset) - parentPos.Y()}
		case MiddleRight:
			o.position = mgl32.Vec3{(right - rightOffset) - parentPos.X(), -margin.Top + margin.Bottom}
		case BottomRight:
			o.position = mgl32.Vec3{(right - rightOffset) - parentPos.X(), (bottom + bottomOffset) - parentPos.Y()}
		}
		o.stateMutex.Unlock()
	}

	pos := o.WorldPosition()
	o.stateMutex.Lock()
	o.bounds.Top = pos.Y() + halfHeight
	o.bounds.Right = pos.X() + halfWidth
	o.bounds.Bottom = pos.Y() - halfHeight
	o.bounds.Left = pos.X() - halfWidth
	o.stateMutex.Unlock()
}

func (o *WindowObjectBase) Parent() WindowObject {
	if p := o.parent.Load(); p != nil {
		return *p
	}
	return nil
}

func (o *WindowObjectBase) SetParent(parent WindowObject, recursive ...bool) WindowObject {
	o.parent.Store(&parent)
	o.SetParentTransform(parent)

	if len(recursive) > 0 {
		if recursive[0] {
			for _, c := range o.children {
				c.SetParent(parent)
			}
		}
	}
	return o
}

func (o *WindowObjectBase) AddChild(child WindowObject) WindowObject {
	if child == nil {
		return o
	}

	o.stateMutex.Lock()
	o.children = append(o.children, child)
	child.SetParent(o)
	child.SetWindow(o.window)
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) AddChildren(children ...WindowObject) WindowObject {
	if children == nil || len(children) == 0 {
		return o
	}

	o.stateMutex.Lock()
	o.children = append(o.children, children...)
	for _, c := range o.children {
		c.SetParent(o)
		c.SetWindow(o.Window())
	}
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) RemoveChild(child WindowObject) {
	if child == nil {
		return
	}

	o.stateMutex.Lock()
	removeAt := -1
	for i, c := range o.children {
		if c == child {
			removeAt = i
			break
		}
	}

	if removeAt != -1 {
		o.children = append(o.children[:removeAt], o.children[removeAt+1:]...)
	}

	o.stateMutex.Unlock()
}

func (o *WindowObjectBase) RemoveChildren() {
	o.stateMutex.Lock()
	o.children = make([]WindowObject, 0)
	o.stateMutex.Unlock()
}

func (o *WindowObjectBase) Child(name string) WindowObject {
	if name == "" {
		return nil
	}

	for _, c := range o.children {
		if c.Name() == name {
			return c
		}
		gc := c.Child(name)
		if gc != nil {
			return gc
		}
	}

	return nil
}

func (o *WindowObjectBase) Children() []WindowObject {
	return o.children
}

/******************************************************************************
 WindowObjectBase Functions
******************************************************************************/

func (o *WindowObjectBase) initChildren() (ok bool) {
	ok = true
	for _, c := range o.children {
		c.SetWindow(o.window)
		initOk := c.Init()
		ok = ok && initOk
	}
	return ok
}

func (o *WindowObjectBase) updateChildren(deltaTime int64) (ok bool) {
	ok = true
	for _, c := range o.children {
		updateOk := c.Update(deltaTime)
		ok = ok && updateOk
	}
	return ok
}

func (o *WindowObjectBase) drawChildren(deltaTime int64) (ok bool) {
	ok = true
	for _, c := range o.children {
		drawOk := c.Draw(deltaTime)
		ok = ok && drawOk
	}
	return ok
}

func (o *WindowObjectBase) closeChildren() {
	for _, c := range o.children {
		c.Close()
	}
}

func (o *WindowObjectBase) getParentBounds() (top, right, bottom, left float32) {
	if p := o.Parent(); p == nil {
		top = 1
		right = 1
		bottom = -1
		left = -1
	} else {
		pBounds := p.Bounds()
		top = pBounds.Top
		right = pBounds.Right
		bottom = pBounds.Bottom
		left = pBounds.Left
	}
	return
}

/******************************************************************************
 New WindowObjectBase Function
******************************************************************************/

func NewWindowObject() *WindowObjectBase {
	w := &WindowObjectBase{
		ObjectTransform:     *NewObjectTransform(),
		blurIntensity:       0.0001,
		color:               RgbaToFloatArray(White),
		maintainAspectRatio: true,
		anchor:              NoAnchor,
	}

	w.enabled.Store(true)
	w.visible.Store(true)

	return w
}
