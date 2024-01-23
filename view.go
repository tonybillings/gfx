package gfx

import (
	"image/color"
)

const (
	defaultViewName = "View"
)

type View struct {
	WindowObjectBase

	fill   *Shape
	border *Shape
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (v *View) Init(window *Window) (ok bool) {
	v.fill.Init(window)
	v.border.Init(window)

	return v.WindowObjectBase.Init(window)
}

func (v *View) Update(deltaTime int64) (ok bool) {
	if !v.enabled.Load() {
		return false
	}

	v.fill.Update(deltaTime)
	v.border.Update(deltaTime)

	return v.WindowObjectBase.Update(deltaTime)
}

func (v *View) Draw(deltaTime int64) (ok bool) {
	if !v.visible.Load() {
		return false
	}

	v.fill.Draw(deltaTime)
	v.border.Draw(deltaTime)

	return v.WindowObjectBase.Draw(deltaTime)
}

func (v *View) Close() {
	v.fill.Close()
	v.border.Close()
	v.WindowObjectBase.Close()
}

/******************************************************************************
 View Functions
******************************************************************************/

func (v *View) Texture() string {
	return v.fill.Texture()
}

func (v *View) SetTexture(pathToPng string) *View {
	v.fill.SetTexture(pathToPng)
	return v
}

func (v *View) FillColor() color.RGBA {
	return v.fill.Color()
}

func (v *View) SetFillColor(rgba color.RGBA) *View {
	v.fill.SetColor(rgba)
	return v
}

func (v *View) SetColor(rgba color.RGBA) WindowObject {
	v.SetFillColor(rgba)
	return v.SetBorderColor(rgba)
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

func (v *View) MaintainAspectRatio(maintainAspectRatio bool) WindowObject {
	v.fill.MaintainAspectRatio(maintainAspectRatio)
	v.border.MaintainAspectRatio(maintainAspectRatio)
	return v
}

func (v *View) BlurEnabled() bool {
	return v.fill.BlurEnabled()
}

func (v *View) SetBlurEnabled(isEnabled bool) WindowObject {
	v.fill.SetBlurEnabled(isEnabled)
	return v
}

func (v *View) BlurIntensity() float32 {
	return v.fill.BlurIntensity()
}

func (v *View) SetBlurIntensity(intensity float32) WindowObject {
	v.fill.SetBlurIntensity(intensity)
	return v
}

func (v *View) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	v.fill.Resize(oldWidth, oldHeight, newWidth, newHeight)
	v.border.Resize(oldWidth, oldHeight, newWidth, newHeight)
	v.WindowObjectBase.Resize(oldWidth, oldHeight, newWidth, newHeight)
}

/******************************************************************************
 New View Function
******************************************************************************/

func NewView() *View {
	v := &View{
		WindowObjectBase: *NewObject(nil),
		fill:             NewQuad(),
		border:           NewSquare(thicknessEpsilon * 2),
	}

	v.SetName(defaultViewName)
	v.fill.SetParent(v)
	v.border.SetParent(v)
	v.fill.SetColor(Black)
	v.border.SetColor(Black)
	return v
}
