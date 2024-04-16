package gfx

import (
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"sync/atomic"
)

const (
	defaultCheckBoxName    = "CheckBox"
	defaultRadioButtonName = "RadioButton"
)

/******************************************************************************
 CheckButton
******************************************************************************/

type CheckButton struct {
	Button

	circular bool

	check   *Shape2D
	checked atomic.Bool

	onCheckedChangedHandlers []func(sender WindowObject, checked bool)
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (b *CheckButton) Init() (ok bool) {
	if b.Initialized() {
		return true
	}

	b.text.SetWindow(b.window)
	b.bounds.SetWindow(b.window)
	b.check.SetWindow(b.window)

	if !b.text.Init() || !b.bounds.Init() || !b.check.Init() {
		return false
	}

	b.initLayout()
	b.RefreshLayout()

	return b.View.Init()
}

func (b *CheckButton) Update(deltaTime int64) (ok bool) {
	if !b.Button.Update(deltaTime) {
		return false
	}

	return b.check.Update(deltaTime)
}

func (b *CheckButton) Close() {
	b.check.Close()
	b.Button.Close()
}

/******************************************************************************
 DrawableObject Implementation
******************************************************************************/

func (b *CheckButton) Draw(deltaTime int64) (ok bool) {
	if !b.Button.Draw(deltaTime) {
		return false
	}

	return b.check.Draw(deltaTime)
}

/******************************************************************************
 Resizer Implementation
******************************************************************************/

func (b *CheckButton) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	b.Button.Resize(oldWidth, oldHeight, newWidth, newHeight)
	b.check.Resize(oldWidth, oldHeight, newWidth, newHeight)
	b.text.Resize(oldWidth, oldHeight, newWidth, newHeight)
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (b *CheckButton) SetColor(rgba color.RGBA) WindowObject {
	b.check.SetColor(rgba)
	b.SetBorderColor(rgba)
	return b
}

func (b *CheckButton) SetMaintainAspectRatio(maintainAspectRatio bool) WindowObject {
	b.Button.SetMaintainAspectRatio(maintainAspectRatio)
	b.check.SetMaintainAspectRatio(maintainAspectRatio)
	return b
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
	b.check.SetParent(b)

	b.Button.defaultLayout()

	b.SetBorderThickness(0.15)

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

	b.text.SetAnchor(MiddleLeft)
	b.text.SetScaleX(1 / b.WorldScale().X())
	b.text.SetAlignment(Left)
	b.text.SetPaddingLeft(b.Scale().X() * 2.25)
	b.text.RefreshLayout()

	b.originalFillColor = b.fill.Color()
	b.originalBorderColor = b.border.Color()
	b.originalTextColor = b.text.Color()
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
		b.text.SetColor(b.mouseEnterTextColor)
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
		b.text.SetColor(b.originalTextColor)
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
		b.text.SetColor(b.mouseDownTextColor)
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
		b.text.SetColor(b.originalTextColor)
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

func (b *CheckButton) CheckShape() *Shape2D {
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

func newCheckButton(name string, circular ...bool) *CheckButton {
	b := &CheckButton{}
	b.View.WindowObjectBase = *NewWindowObject(nil)
	b.text = NewLabel()

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

	b.SetName(name)
	b.defaultLayout()

	return b
}

func NewCheckBox() *CheckButton {
	return newCheckButton(defaultCheckBoxName, false)
}

func NewRadioButton() *CheckButton {
	return newCheckButton(defaultRadioButtonName, true)
}
