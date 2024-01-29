package gfx

import (
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"sync/atomic"
)

const (
	defaultCheckButtonName = "CheckButton"
	defaultCheckBoxName    = "CheckBox"
	defaultRadioButtonName = "RadioButton"
)

/******************************************************************************
 CheckButton
******************************************************************************/

type CheckButton struct {
	Button

	circular bool

	check   *Shape
	checked atomic.Bool

	onCheckedChangedHandlers []func(sender WindowObject, checked bool)
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (b *CheckButton) Init(window *Window) (ok bool) {
	if !b.label.Init(window) || !b.bounds.Init(window) || !b.check.Init(window) {
		return false
	}

	if !b.View.Init(window) {
		return false
	}

	b.initLayout()
	b.RefreshLayout()

	return true
}

func (b *CheckButton) Update(deltaTime int64) (ok bool) {
	if !b.Button.Update(deltaTime) {
		return false
	}

	return b.check.Update(deltaTime)
}

func (b *CheckButton) Draw(deltaTime int64) (ok bool) {
	if !b.Button.Draw(deltaTime) {
		return false
	}

	return b.check.Draw(deltaTime)
}

func (b *CheckButton) Close() {
	b.check.Close()
	b.Button.Close()
}

func (b *CheckButton) SetColor(rgba color.RGBA) WindowObject {
	b.check.SetColor(rgba)
	return b.SetBorderColor(rgba)
}

func (b *CheckButton) MaintainAspectRatio(maintainAspectRatio bool) WindowObject {
	b.Button.MaintainAspectRatio(maintainAspectRatio)
	b.check.MaintainAspectRatio(maintainAspectRatio)
	return b
}

func (b *CheckButton) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	b.Button.Resize(oldWidth, oldHeight, newWidth, newHeight)
	b.check.Resize(oldWidth, oldHeight, newWidth, newHeight)
	b.label.SetMarginLeft(b.window.ScaleX(b.Scale().X() * 1.25))
}

func (b *CheckButton) SetWindow(window *Window) WindowObject {
	b.Button.SetWindow(window)
	b.check.SetWindow(window)
	return b
}

/******************************************************************************
 CheckButton Functions
******************************************************************************/

func (b *CheckButton) defaultLayout() {
	b.Button.defaultLayout()

	b.SetBorderThickness(0.15)

	b.label.SetAnchor(MiddleLeft)

	b.check.SetColor(Black)
	b.check.SetScale(mgl32.Vec3{.5, .5})
	b.check.SetVisibility(false)
}

func (b *CheckButton) initLayout() {
	b.bounds.OnMouseEnter(b.onMouseEnter)
	b.bounds.OnMouseLeave(b.onMouseLeave)
	b.bounds.OnPMouseDown(b.onMouseDown)
	b.bounds.OnPMouseUp(b.onMouseUp)
	b.bounds.OnPMouseClick(b.onClick)

	b.label.SetMarginLeft(b.window.ScaleX(b.Scale().X() * 1.25))

	b.originalFillColor = b.fill.Color()
	b.originalBorderColor = b.border.Color()
	b.originalTextColor = b.label.Color()
}

func (b *CheckButton) onClick(sender WindowObject, _ *MouseState) {
	prevState := b.checked.Load()

	if b.circular {
		b.checked.Store(true)
	} else {
		b.checked.Store(!prevState)
	}

	newState := b.checked.Load()
	b.check.SetVisibility(newState)

	if newState != prevState {
		for _, f := range b.onCheckedChangedHandlers {
			f(sender, newState)
		}
	}
}

func (b *CheckButton) onMouseEnter(_ WindowObject, _ *MouseState) {
	if b.mouseEnterFillColorSet.Load() {
		b.fill.SetColor(b.mouseEnterFillColor)
	}
	if b.mouseEnterBorderColorSet.Load() {
		b.border.SetColor(b.mouseEnterBorderColor)
		b.check.SetColor(b.mouseEnterBorderColor)
	}
	if b.mouseEnterTextColorSet.Load() {
		b.label.SetColor(b.mouseEnterTextColor)
	}
}

func (b *CheckButton) onMouseLeave(_ WindowObject, _ *MouseState) {
	if b.mouseEnterFillColorSet.Load() {
		b.fill.SetColor(b.originalFillColor)
	}
	if b.mouseEnterBorderColorSet.Load() {
		b.border.SetColor(b.originalBorderColor)
		b.check.SetColor(b.originalBorderColor)
	}
	if b.mouseEnterTextColorSet.Load() {
		b.label.SetColor(b.originalTextColor)
	}
}

func (b *CheckButton) onMouseDown(_ WindowObject, _ *MouseState) {
	if b.mouseDownFillColorSet.Load() {
		b.fill.SetColor(b.mouseDownFillColor)
	}
	if b.mouseDownBorderColorSet.Load() {
		b.border.SetColor(b.mouseDownBorderColor)
		b.check.SetColor(b.mouseDownBorderColor)
	}
	if b.mouseDownTextColorSet.Load() {
		b.label.SetColor(b.mouseDownTextColor)
	}
}

func (b *CheckButton) onMouseUp(_ WindowObject, _ *MouseState) {
	if b.mouseDownFillColorSet.Load() {
		b.fill.SetColor(b.originalFillColor)
	}
	if b.mouseDownBorderColorSet.Load() {
		b.border.SetColor(b.originalBorderColor)
		b.check.SetColor(b.originalBorderColor)
	}
	if b.mouseDownTextColorSet.Load() {
		b.label.SetColor(b.originalTextColor)
	}
}

func (b *CheckButton) OnClick(handler func(sender WindowObject, mouseState *MouseState)) *CheckButton {
	b.bounds.OnPMouseClick(handler)
	return b
}

func (b *CheckButton) OnCheckedChanged(handler func(sender WindowObject, checked bool)) *CheckButton {
	b.onCheckedChangedHandlers = append(b.onCheckedChangedHandlers, handler)
	return b
}

func (b *CheckButton) Checked() bool {
	return b.checked.Load()
}

func (b *CheckButton) SetChecked(checked bool) {
	b.checked.Store(checked)
	b.check.SetVisibility(checked)
}

func (b *CheckButton) CheckShape() *Shape {
	return b.check
}

func (b *CheckButton) SetMouseEnterColor(rgba color.RGBA) *CheckButton {
	b.Button.SetMouseEnterBorderColor(rgba)
	return b
}

func (b *CheckButton) SetMouseDownColor(rgba color.RGBA) *CheckButton {
	b.Button.SetMouseDownBorderColor(rgba)
	return b
}

/******************************************************************************
 New CheckButton Functions
******************************************************************************/

func newCheckButton(circular ...bool) *CheckButton {
	b := &CheckButton{}
	b.View.WindowObjectBase = *NewObject(nil)
	b.label = NewLabel()

	if len(circular) > 0 && circular[0] {
		b.circular = true
		b.fill = NewDot()
		b.border = NewCircle(thicknessEpsilon * 2)
		b.bounds = NewBoundingRadius()
		b.check = NewDot()
	} else {
		b.fill = NewQuad()
		b.border = NewSquare(thicknessEpsilon * 2)
		b.bounds = NewBoundingBox()
		b.check = NewQuad()
	}

	b.SetName(defaultCheckButtonName)
	b.check.SetParent(b)

	b.defaultLayout()

	return b
}

func NewCheckBox() *CheckButton {
	return newCheckButton(false)
}

func NewRadioButton() *CheckButton {
	return newCheckButton(true)
}
