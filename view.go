package gfx

import (
	"image/color"
)

const (
	defaultViewName = "View"
)

/******************************************************************************
 View
******************************************************************************/

type View struct {
	WindowObjectBase

	fill   *Shape2D
	border *Shape2D
}

/******************************************************************************
Object Implementation
******************************************************************************/

func (v *View) Init() (ok bool) {
	if v.Initialized() {
		return true
	}

	if !v.fill.Init() || !v.border.Init() {
		return false
	}

	v.RefreshLayout()

	return v.WindowObjectBase.Init()
}

func (v *View) Update(deltaTime int64) (ok bool) {
	if !v.enabled.Load() || !v.initialized.Load() {
		return false
	}

	v.fill.Update(deltaTime)
	v.border.Update(deltaTime)

	return v.updateChildren(deltaTime)
}

func (v *View) Close() {
	v.fill.Close()
	v.border.Close()
	v.WindowObjectBase.Close()
}

/******************************************************************************
 DrawableObject Implementation
******************************************************************************/

func (v *View) Draw(deltaTime int64) (ok bool) {
	if !v.visible.Load() || !v.initialized.Load() {
		return false
	}

	v.fill.Draw(deltaTime)
	v.border.Draw(deltaTime)

	return v.drawChildren(deltaTime)
}

/******************************************************************************
 Resizer Implementation
******************************************************************************/

func (v *View) Resize(newWidth, newHeight int) {
	v.WindowObjectBase.Resize(newWidth, newHeight)
	v.fill.Resize(newWidth, newHeight)
	v.border.Resize(newWidth, newHeight)
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (v *View) SetWindow(window *Window) WindowObject {
	v.WindowObjectBase.SetWindow(window)
	v.fill.SetWindow(window)
	v.border.SetWindow(window)
	return v
}

func (v *View) SetColor(rgba color.RGBA) WindowObject {
	v.SetFillColor(rgba)
	return v.SetBorderColor(rgba)
}

func (v *View) SetMaintainAspectRatio(maintainAspectRatio bool) WindowObject {
	v.maintainAspectRatio = maintainAspectRatio
	v.fill.SetMaintainAspectRatio(maintainAspectRatio)
	v.border.SetMaintainAspectRatio(maintainAspectRatio)
	return v
}

func (v *View) BlurEnabled() bool {
	return v.fill.BlurEnabled()
}

func (v *View) SetBlurEnabled(enabled bool) WindowObject {
	v.fill.SetBlurEnabled(enabled)
	return v
}

func (v *View) BlurIntensity() float32 {
	return v.fill.BlurIntensity()
}

func (v *View) SetBlurIntensity(intensity float32) WindowObject {
	v.fill.SetBlurIntensity(intensity)
	return v
}

func (v *View) AddChild(child WindowObject) WindowObject {
	if child == nil {
		return v
	}
	v.WindowObjectBase.AddChild(child)
	child.SetParent(v)
	return v
}

func (v *View) AddChildren(children ...WindowObject) WindowObject {
	if children == nil || len(children) == 0 {
		return v
	}

	v.WindowObjectBase.AddChildren(children...)
	v.stateMutex.Lock()
	for _, c := range v.children {
		c.SetParent(v)
	}
	v.stateMutex.Unlock()
	return v
}

/******************************************************************************
 View Functions
******************************************************************************/

func (v *View) defaultLayout() {
	v.fill.SetParent(v)
	v.border.SetParent(v)
	v.fill.SetColor(Black)
	v.border.SetColor(Black)
}

func (v *View) SetTexture(texture *Texture2D) *View {
	v.fill.SetTexture(texture)
	return v
}

func (v *View) FillColor() color.RGBA {
	return v.fill.Color()
}

func (v *View) SetFillColor(rgba color.RGBA) *View {
	v.fill.SetColor(rgba)
	return v
}

func (v *View) BorderThickness() float32 {
	return v.border.Thickness()
}

func (v *View) SetBorderThickness(thickness float32) *View {
	if thickness <= thicknessEpsilon {
		thickness = thicknessEpsilon * 2
	}
	v.border.SetThickness(thickness)
	return v
}

func (v *View) BorderColor() color.RGBA {
	return v.border.Color()
}

func (v *View) SetBorderColor(rgba color.RGBA) *View {
	v.border.SetColor(rgba)
	return v
}

/******************************************************************************
 New View Function
******************************************************************************/

func NewView() *View {
	v := &View{
		WindowObjectBase: *NewWindowObject(),
		fill:             NewQuad(),
		border:           NewSquare(thicknessEpsilon * 2),
	}

	v.SetName(defaultViewName)
	v.defaultLayout()

	return v
}
