package gfx

import (
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	_ "image/png"
	"sync"
	"sync/atomic"
)

const (
	sizeOfFloat32 = 4 // byte count
)

/******************************************************************************
 Interface
******************************************************************************/

type WindowObject interface {
	Name() string
	SetName(string) WindowObject

	Window() *Window
	SetWindow(*Window) WindowObject

	Initialized() bool
	SetInitialized(bool)

	Init(*Window) bool
	Update(int64) bool
	Draw(int64) bool
	Close()
	Closed() bool

	Resize(int32, int32, int32, int32)
	OnResize(func(int32, int32, int32, int32))
	RefreshLayout()

	Visible() bool
	SetVisibility(bool) WindowObject
	Enabled() bool
	SetEnabled(bool) WindowObject

	Color() color.RGBA
	SetColor(color.RGBA) WindowObject
	Opacity() uint8
	SetOpacity(uint8) WindowObject

	BlurIntensity() float32
	SetBlurIntensity(float32) WindowObject
	BlurEnabled() bool
	SetBlurEnabled(bool) WindowObject

	Origin() mgl32.Vec3
	SetOrigin(mgl32.Vec3) WindowObject

	Position() mgl32.Vec3
	WorldPosition() mgl32.Vec3
	SetPosition(mgl32.Vec3) WindowObject
	SetPositionX(float32) WindowObject
	SetPositionY(float32) WindowObject
	SetPositionZ(float32) WindowObject

	Rotation() mgl32.Vec3
	WorldRotation() mgl32.Vec3
	SetRotation(mgl32.Vec3) WindowObject
	SetRotationX(float32) WindowObject
	SetRotationY(float32) WindowObject
	SetRotationZ(float32) WindowObject

	Scale() mgl32.Vec3
	WorldScale() mgl32.Vec3
	SetScale(mgl32.Vec3) WindowObject
	SetScaleX(float32) WindowObject
	SetScaleY(float32) WindowObject
	SetScaleZ(float32) WindowObject
	Width() float32
	Height() float32
	HalfWidth() float32
	HalfHeight() float32

	MaintainAspectRatio(bool) WindowObject
	Anchor() Alignment
	SetAnchor(Alignment) WindowObject
	Margin() *Margin
	SetMargin(Margin) WindowObject

	Parent() WindowObject
	SetParent(WindowObject, ...bool) WindowObject
	AddChild(WindowObject) WindowObject
	AddChildren(...WindowObject) WindowObject
	RemoveChild(WindowObject)
	RemoveChildren()
	Child(string) WindowObject
	Children() []WindowObject

	Tag() any
	SetTag(any) WindowObject
}

/******************************************************************************
 WindowObjectBase
******************************************************************************/

type WindowObjectBase struct {
	name atomic.Pointer[string]

	window *Window

	initialized atomic.Bool

	vertices    []float32
	vertexCount int32

	color mgl32.Vec4

	blurIntensity float32
	blurEnabled   bool

	origin   mgl32.Vec3
	position mgl32.Vec3
	rotation mgl32.Vec3
	scale    mgl32.Vec3

	onResizeHandlers []func(int32, int32, int32, int32)

	maintainAspectRatio bool
	anchor              Alignment
	margin              Margin

	stateMutex   sync.Mutex
	stateChanged atomic.Bool

	visible atomic.Bool
	enabled atomic.Bool

	closing atomic.Bool
	closed  atomic.Bool

	parent   atomic.Pointer[WindowObject]
	children []WindowObject

	tag atomic.Value
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (o *WindowObjectBase) initChildren(window *Window) (ok bool) {
	ok = true
	for _, c := range o.children {
		initOk := c.Init(window)
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

func (o *WindowObjectBase) Name() string {
	if n := o.name.Load(); n != nil {
		return *n
	}
	return ""
}

func (o *WindowObjectBase) SetName(name string) WindowObject {
	o.name.Store(&name)
	return o
}

func (o *WindowObjectBase) Window() *Window {
	return o.window
}

func (o *WindowObjectBase) SetWindow(window *Window) WindowObject {
	o.stateMutex.Lock()
	o.window = window
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) Initialized() bool {
	return o.initialized.Load()
}

func (o *WindowObjectBase) SetInitialized(initialized bool) {
	o.initialized.Store(initialized)
}

func (o *WindowObjectBase) Init(window *Window) (ok bool) {
	if o.Initialized() {
		return true
	}

	o.window = window

	if !o.closed.Load() && !o.closing.Load() {
		return o.initChildren(window)
	} else {
		return false
	}
}

func (o *WindowObjectBase) Update(deltaTime int64) (ok bool) {
	if o.enabled.Load() {
		return o.updateChildren(deltaTime)
	} else {
		return false
	}
}

func (o *WindowObjectBase) Draw(deltaTime int64) (ok bool) {
	if o.closing.Load() {
		o.closed.Store(true)
		o.closing.Store(false)
		return false
	}

	if o.visible.Load() {
		return o.drawChildren(deltaTime)
	} else {
		return false
	}
}

func (o *WindowObjectBase) Close() {
	if o.closed.Load() || o.closing.Load() {
		return
	}
	o.closing.Store(true)
	o.closeChildren()
}

func (o *WindowObjectBase) Closed() bool {
	return o.closed.Load()
}

func (o *WindowObjectBase) RefreshLayout() {
	window := o.Window()
	if window == nil {
		return
	}

	left := float32(-1)
	right := float32(1)
	top := float32(1)
	bottom := float32(-1)
	if p := o.Parent(); p != nil {
		if view, ok := p.(*View); ok {
			if view.maintainAspectRatio {
				left *= window.ScaleX(view.Scale().X())
				top *= window.ScaleY(view.Scale().Y())
			} else {
				left *= view.Scale().X()
				top *= view.Scale().Y()
			}

			right = left * -1
			bottom = top * -1
		}
	}

	margin := o.Margin()
	margin.Left = window.ScaleX(margin.Left)
	margin.Right = window.ScaleX(margin.Right)
	margin.Top = window.ScaleY(margin.Top)
	margin.Bottom = window.ScaleY(margin.Bottom)

	leftOffset := o.HalfWidth() + margin.Left
	rightOffset := o.HalfWidth() + margin.Right
	topOffset := o.HalfHeight() + margin.Top
	bottomOffset := o.HalfHeight() + margin.Bottom

	switch o.Anchor() {
	case TopLeft:
		o.SetPosition(mgl32.Vec3{left + leftOffset, top - topOffset})
	case MiddleLeft:
		o.SetPosition(mgl32.Vec3{left + leftOffset, -margin.Top + margin.Bottom})
	case BottomLeft:
		o.SetPosition(mgl32.Vec3{left + leftOffset, bottom + bottomOffset})
	case TopCenter:
		o.SetPosition(mgl32.Vec3{-margin.Right + margin.Left, 1 - topOffset})
	case Center:
		o.SetPosition(mgl32.Vec3{-margin.Right + margin.Left, -margin.Top + margin.Bottom})
	case BottomCenter:
		o.SetPosition(mgl32.Vec3{-margin.Right + margin.Left, -1 + bottomOffset})
	case TopRight:
		o.SetPosition(mgl32.Vec3{right - rightOffset, top - topOffset})
	case MiddleRight:
		o.SetPosition(mgl32.Vec3{right - rightOffset, -margin.Top + margin.Bottom})
	case BottomRight:
		o.SetPosition(mgl32.Vec3{right - rightOffset, bottom + bottomOffset})
	}
}

func (o *WindowObjectBase) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	if o.Closed() {
		return
	}

	o.RefreshLayout()

	for _, c := range o.children {
		if !c.Closed() {
			c.Resize(oldWidth, oldHeight, newWidth, newHeight)
		}
	}

	for _, f := range o.onResizeHandlers {
		f(oldWidth, oldHeight, newWidth, newHeight)
	}
}

func (o *WindowObjectBase) OnResize(handler func(oldWidth, oldHeight, newWidth, newHeight int32)) {
	o.onResizeHandlers = append(o.onResizeHandlers, handler)
}

func (o *WindowObjectBase) Visible() bool {
	return o.visible.Load()
}

func (o *WindowObjectBase) SetVisibility(visible bool) WindowObject {
	o.visible.Store(visible)
	return o
}

func (o *WindowObjectBase) Enabled() bool {
	return o.enabled.Load()
}

func (o *WindowObjectBase) SetEnabled(enabled bool) WindowObject {
	o.enabled.Store(enabled)
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
	o.stateChanged.Store(true)
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) Origin() mgl32.Vec3 {
	o.stateMutex.Lock()
	origin := mgl32.Vec3{o.origin[0], o.origin[1], o.origin[2]}
	o.stateMutex.Unlock()
	return origin
}

func (o *WindowObjectBase) SetOrigin(origin mgl32.Vec3) WindowObject {
	o.stateMutex.Lock()
	o.origin[0] = origin.X()
	o.origin[1] = origin.Y()
	o.origin[2] = origin.Z()
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) Position() mgl32.Vec3 {
	o.stateMutex.Lock()
	position := mgl32.Vec3{o.position[0], o.position[1], o.position[2]}
	o.stateMutex.Unlock()
	return position
}

func (o *WindowObjectBase) WorldPosition() mgl32.Vec3 {
	o.stateMutex.Lock()
	position := mgl32.Vec3{o.position[0], o.position[1], o.position[2]}
	o.stateMutex.Unlock()

	if p := o.Parent(); p != nil {
		parentPos := p.WorldPosition()
		position = position.Add(parentPos)
	}

	return position
}

func (o *WindowObjectBase) SetPosition(position mgl32.Vec3) WindowObject {
	o.stateMutex.Lock()
	o.position[0] = position.X()
	o.position[1] = position.Y()
	o.position[2] = position.Z()
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) SetPositionX(x float32) WindowObject {
	o.stateMutex.Lock()
	o.position[0] = x
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) SetPositionY(y float32) WindowObject {
	o.stateMutex.Lock()
	o.position[1] = y
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) SetPositionZ(z float32) WindowObject {
	o.stateMutex.Lock()
	o.position[2] = z
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) Rotation() mgl32.Vec3 {
	o.stateMutex.Lock()
	rotation := mgl32.Vec3{o.rotation[0], o.rotation[1], o.rotation[2]}
	o.stateMutex.Unlock()
	return rotation
}

func (o *WindowObjectBase) WorldRotation() mgl32.Vec3 {
	o.stateMutex.Lock()
	rotation := mgl32.Vec3{o.rotation[0], o.rotation[1], o.rotation[2]}
	o.stateMutex.Unlock()

	if p := o.Parent(); p != nil {
		parentRot := p.WorldRotation()
		rotation = rotation.Add(parentRot)
	}

	return rotation
}

func (o *WindowObjectBase) SetRotation(rotation mgl32.Vec3) WindowObject {
	o.stateMutex.Lock()
	o.rotation[0] = rotation.X()
	o.rotation[1] = rotation.Y()
	o.rotation[2] = rotation.Z()
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) SetRotationX(x float32) WindowObject {
	o.stateMutex.Lock()
	o.rotation[0] = x
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) SetRotationY(y float32) WindowObject {
	o.stateMutex.Lock()
	o.rotation[1] = y
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) SetRotationZ(z float32) WindowObject {
	o.stateMutex.Lock()
	o.rotation[2] = z
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) Scale() mgl32.Vec3 {
	o.stateMutex.Lock()
	scale := mgl32.Vec3{o.scale[0], o.scale[1], o.scale[2]}
	o.stateMutex.Unlock()
	return scale
}

func (o *WindowObjectBase) WorldScale() mgl32.Vec3 {
	o.stateMutex.Lock()
	scale := mgl32.Vec3{o.scale[0], o.scale[1], o.scale[2]}
	o.stateMutex.Unlock()

	if p := o.Parent(); p != nil {
		parentScale := p.WorldScale()
		scale = mgl32.Vec3{o.scale[0] * parentScale.X(), o.scale[1] * parentScale.Y(), o.scale[2] * parentScale.Z()}
	}

	return scale
}

func (o *WindowObjectBase) SetScale(scale mgl32.Vec3) WindowObject {
	o.stateMutex.Lock()
	o.scale[0] = scale.X()
	o.scale[1] = scale.Y()
	o.scale[2] = scale.Z()
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) SetScaleX(x float32) WindowObject {
	o.stateMutex.Lock()
	o.scale[0] = x
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) SetScaleY(y float32) WindowObject {
	o.stateMutex.Lock()
	o.scale[1] = y
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) SetScaleZ(z float32) WindowObject {
	o.stateMutex.Lock()
	o.scale[2] = z
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

func (o *WindowObjectBase) MaintainAspectRatio(maintainAspectRatio bool) WindowObject {
	o.stateMutex.Lock()
	o.maintainAspectRatio = maintainAspectRatio
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) Anchor() Alignment {
	o.stateMutex.Lock()
	a := o.anchor
	o.stateMutex.Unlock()
	return a
}

func (o *WindowObjectBase) SetAnchor(anchor Alignment) WindowObject {
	o.stateMutex.Lock()
	o.anchor = anchor
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
	o.stateMutex.Unlock()
	return o
}

func (o *WindowObjectBase) Parent() WindowObject {
	if p := o.parent.Load(); p != nil {
		return *p
	}
	return nil
}

func (o *WindowObjectBase) SetParent(parent WindowObject, recursive ...bool) WindowObject {
	o.parent.Store(&parent)
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
	o.stateChanged.Store(true)
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
	}
	o.stateChanged.Store(true)
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

	o.stateChanged.Store(true)
	o.stateMutex.Unlock()
}

func (o *WindowObjectBase) RemoveChildren() {
	o.stateMutex.Lock()
	o.children = make([]WindowObject, 0)
	o.stateChanged.Store(true)
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

func (o *WindowObjectBase) Tag() any {
	return o.tag.Load()
}

func (o *WindowObjectBase) SetTag(value any) WindowObject {
	o.tag.Store(value)
	return o
}

/******************************************************************************
 New WindowObjectBase Function
******************************************************************************/

func NewObject(parent WindowObject) *WindowObjectBase {
	w := &WindowObjectBase{
		blurIntensity:       0.0001,
		color:               RgbaToFloatArray(White),
		position:            [3]float32{0, 0, 0},
		rotation:            [3]float32{0, 0, 0},
		scale:               [3]float32{1, 1, 1},
		maintainAspectRatio: true,
		anchor:              NoAlignment,
		children:            make([]WindowObject, 0),
	}

	w.SetParent(parent)
	w.stateChanged.Store(true)
	w.enabled.Store(true)
	w.visible.Store(true)

	return w
}
