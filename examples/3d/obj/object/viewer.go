package object

import (
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"math"
	"tonysoft.com/gfx"
)

func newRotationControl(window *gfx.Window, obj *gfx.Shape3D) *gfx.View {
	panel := gfx.NewView()
	panel.SetAnchor(gfx.MiddleLeft)
	panel.SetMarginLeft(.1)
	panel.SetScale(mgl32.Vec3{0.3, .2})
	panel.SetFillColor(gfx.Opacity(gfx.Purple, 0.3))
	panel.SetBorderColor(gfx.Purple)
	panel.SetBorderThickness(0.02)

	rotSliderX := gfx.NewSlider(gfx.Vertical)
	rotSliderY := gfx.NewSlider(gfx.Vertical)
	rotSliderZ := gfx.NewSlider(gfx.Vertical)

	panel.AddChild(rotSliderX)
	panel.AddChild(rotSliderY)
	panel.AddChild(rotSliderZ)

	rotSliderX.SetWindow(window)
	rotSliderX.SetScaleX(.3333)
	rotSliderX.SetScaleY(.9)
	rotSliderX.SetAnchor(gfx.MiddleLeft)
	rotSliderX.SetFillColor(gfx.Transparent)
	rotSliderX.Rail().SetColor(gfx.Darken(gfx.Purple, .6))
	rotSliderX.Button().SetFillColor(gfx.Purple)
	rotSliderX.Button().SetMouseEnterFillColor(gfx.Darken(gfx.Purple, .3))
	rotSliderX.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		obj.SetRotationX(value * math.Pi)
	})

	rotSliderY.SetWindow(window)
	rotSliderY.SetScaleX(.3333)
	rotSliderY.SetScaleY(.9)
	rotSliderY.SetAnchor(gfx.MiddleLeft)
	rotSliderY.SetMarginLeft(rotSliderX.WorldScale().X() * 2)
	rotSliderY.SetFillColor(gfx.Transparent)
	rotSliderY.Rail().SetColor(gfx.Darken(gfx.Purple, .6))
	rotSliderY.Button().SetFillColor(gfx.Purple)
	rotSliderY.Button().SetMouseEnterFillColor(gfx.Darken(gfx.Purple, .3))
	rotSliderY.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		obj.SetRotationY(value * math.Pi)
	})

	rotSliderZ.SetWindow(window)
	rotSliderZ.SetScaleX(.3333)
	rotSliderZ.SetScaleY(.9)
	rotSliderZ.SetAnchor(gfx.MiddleLeft)
	rotSliderZ.SetMarginLeft(rotSliderX.WorldScale().X() * 4)
	rotSliderZ.SetFillColor(gfx.Transparent)
	rotSliderZ.Rail().SetColor(gfx.Darken(gfx.Purple, .6))
	rotSliderZ.Button().SetFillColor(gfx.Purple)
	rotSliderZ.Button().SetMouseEnterFillColor(gfx.Darken(gfx.Purple, .3))
	rotSliderZ.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		obj.SetRotationZ(value * math.Pi)
	})

	rotControlLabel := gfx.NewLabel()
	rotControlLabel.SetText("Object Rotation")
	rotControlLabel.SetColor(gfx.Purple)
	rotControlLabel.SetFontSize(.1)
	rotControlLabel.SetAnchor(gfx.TopCenter)
	rotControlLabel.SetMarginTop(.02)
	panel.AddChild(rotControlLabel)

	rotSliderXLabel := gfx.NewLabel()
	rotSliderXLabel.SetColor(gfx.Purple)
	rotSliderXLabel.SetFontSize(.15)
	rotSliderXLabel.SetPositionY(window.ScaleY(-rotSliderX.WorldScale().Y() + (rotSliderX.WorldScale().Y() * .06)))
	rotSliderXLabel.SetText("X")
	rotSliderXLabel.OnResize(func(_, _, _, _ int32) {
		rotSliderXLabel.SetPositionY(window.ScaleY(-rotSliderX.WorldScale().Y() + (rotSliderX.WorldScale().Y() * .06)))
	})
	rotSliderX.AddChild(rotSliderXLabel)

	rotSliderYLabel := gfx.NewLabel()
	rotSliderYLabel.SetColor(gfx.Purple)
	rotSliderYLabel.SetFontSize(.15)
	rotSliderYLabel.SetPositionY(window.ScaleY(-rotSliderY.WorldScale().Y() + (rotSliderY.WorldScale().Y() * .06)))
	rotSliderYLabel.SetText("Y")
	rotSliderYLabel.OnResize(func(_, _, _, _ int32) {
		rotSliderYLabel.SetPositionY(window.ScaleY(-rotSliderY.WorldScale().Y() + (rotSliderY.WorldScale().Y() * .06)))
	})
	rotSliderY.AddChild(rotSliderYLabel)

	rotSliderZLabel := gfx.NewLabel()
	rotSliderZLabel.SetColor(gfx.Purple)
	rotSliderZLabel.SetFontSize(.15)
	rotSliderZLabel.SetPositionY(window.ScaleY(-rotSliderZ.WorldScale().Y() + (rotSliderZ.WorldScale().Y() * .06)))
	rotSliderZLabel.SetText("Z")
	rotSliderZLabel.OnResize(func(_, _, _, _ int32) {
		rotSliderZLabel.SetPositionY(window.ScaleY(-rotSliderZ.WorldScale().Y() + (rotSliderZ.WorldScale().Y() * .06)))
	})
	rotSliderZ.AddChild(rotSliderZLabel)

	return panel
}

func newLightControl(window *gfx.Window, obj *gfx.Shape3D, lightNumber int, rgba color.RGBA) *gfx.View {
	panel := gfx.NewView()

	lighting := obj.Lighting()
	lightCount := 0
	var light *gfx.DirectionalLight
	if lighting != nil {
		if quadLighting, ok := lighting.(*gfx.QuadDirectionalLighting); ok {
			lightCount = int(quadLighting.LightCount)
			light = &quadLighting.Lights[lightNumber-1]
		}
	}
	if lighting == nil || lightCount == 0 {
		panel.SetScale(mgl32.Vec3{})
		return panel
	}

	panel.SetAnchor(gfx.MiddleRight)
	panel.SetMarginRight(.1)
	panel.SetScale(mgl32.Vec3{0.3, .2})
	panel.SetFillColor(gfx.Opacity(rgba, 0.3))
	panel.SetBorderColor(rgba)
	panel.SetBorderThickness(0.02)

	dirSliderX := gfx.NewSlider(gfx.Vertical, true)
	dirSliderY := gfx.NewSlider(gfx.Vertical, true)
	dirSliderZ := gfx.NewSlider(gfx.Vertical, true)

	panel.AddChild(dirSliderX)
	panel.AddChild(dirSliderY)
	panel.AddChild(dirSliderZ)

	dirSliderX.SetWindow(window)
	dirSliderX.SetScaleX(.3333)
	dirSliderX.SetScaleY(.9)
	dirSliderX.SetAnchor(gfx.MiddleLeft)
	dirSliderX.SetFillColor(gfx.Transparent)
	dirSliderX.Rail().SetColor(gfx.Darken(rgba, .6))
	dirSliderX.Button().SetFillColor(rgba)
	dirSliderX.Button().SetMouseEnterFillColor(gfx.Darken(rgba, .3))
	dirSliderX.SetValue(.5)

	dirSliderY.SetWindow(window)
	dirSliderY.SetScaleX(.3333)
	dirSliderY.SetScaleY(.9)
	dirSliderY.SetAnchor(gfx.MiddleLeft)
	dirSliderY.SetMarginLeft(dirSliderX.WorldScale().X() * 2)
	dirSliderY.SetFillColor(gfx.Transparent)
	dirSliderY.Rail().SetColor(gfx.Darken(rgba, .6))
	dirSliderY.Button().SetFillColor(rgba)
	dirSliderY.Button().SetMouseEnterFillColor(gfx.Darken(rgba, .3))
	dirSliderY.SetValue(.5)

	dirSliderZ.SetWindow(window)
	dirSliderZ.SetScaleX(.3333)
	dirSliderZ.SetScaleY(.9)
	dirSliderZ.SetAnchor(gfx.MiddleLeft)
	dirSliderZ.SetMarginLeft(dirSliderX.WorldScale().X() * 4)
	dirSliderZ.SetFillColor(gfx.Transparent)
	dirSliderZ.Rail().SetColor(gfx.Darken(rgba, .6))
	dirSliderZ.Button().SetFillColor(rgba)
	dirSliderZ.Button().SetMouseEnterFillColor(gfx.Darken(rgba, .3))
	dirSliderZ.SetValue(.5)

	dirSliderX.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		light.Lock()
		light.Direction[0] = (value * 20) - 10
		light.Unlock()
	})
	dirSliderY.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		light.Lock()
		light.Direction[1] = (value * 20) - 10
		light.Unlock()
	})
	dirSliderZ.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		light.Lock()
		light.Direction[2] = (value * 20) - 10
		light.Unlock()
	})

	dirControlLabel := gfx.NewLabel()
	dirControlLabel.SetText("Light Direction")
	dirControlLabel.SetColor(rgba)
	dirControlLabel.SetFontSize(.1)
	dirControlLabel.SetAnchor(gfx.TopCenter)
	dirControlLabel.SetMarginTop(.02)
	panel.AddChild(dirControlLabel)

	dirSliderXLabel := gfx.NewLabel()
	dirSliderXLabel.SetColor(rgba)
	dirSliderXLabel.SetFontSize(.15)
	dirSliderXLabel.SetPositionY(window.ScaleY(-dirSliderX.WorldScale().Y() + (dirSliderX.WorldScale().Y() * .06)))
	dirSliderXLabel.SetText("X")
	dirSliderXLabel.OnResize(func(_, _, _, _ int32) {
		dirSliderXLabel.SetPositionY(window.ScaleY(-dirSliderX.WorldScale().Y() + (dirSliderX.WorldScale().Y() * .06)))
	})
	dirSliderX.AddChild(dirSliderXLabel)

	dirSliderYLabel := gfx.NewLabel()
	dirSliderYLabel.SetColor(rgba)
	dirSliderYLabel.SetFontSize(.15)
	dirSliderYLabel.SetPositionY(window.ScaleY(-dirSliderY.WorldScale().Y() + (dirSliderY.WorldScale().Y() * .06)))
	dirSliderYLabel.SetText("Y")
	dirSliderYLabel.OnResize(func(_, _, _, _ int32) {
		dirSliderYLabel.SetPositionY(window.ScaleY(-dirSliderY.WorldScale().Y() + (dirSliderY.WorldScale().Y() * .06)))
	})
	dirSliderY.AddChild(dirSliderYLabel)

	dirSliderZLabel := gfx.NewLabel()
	dirSliderZLabel.SetColor(rgba)
	dirSliderZLabel.SetFontSize(.15)
	dirSliderZLabel.SetPositionY(window.ScaleY(-dirSliderZ.WorldScale().Y() + (dirSliderZ.WorldScale().Y() * .06)))
	dirSliderZLabel.SetText("Z")
	dirSliderZLabel.OnResize(func(_, _, _, _ int32) {
		dirSliderZLabel.SetPositionY(window.ScaleY(-dirSliderZ.WorldScale().Y() + (dirSliderZ.WorldScale().Y() * .06)))
	})
	dirSliderZ.AddChild(dirSliderZLabel)

	return panel
}

func newCameraControl(obj *gfx.Shape3D) *gfx.View {
	panel := gfx.NewView()

	objCamera := obj.Camera()
	var camera *gfx.BasicCamera
	if objCamera != nil {
		if basicCamera, ok := objCamera.(*gfx.BasicCamera); ok {
			camera = basicCamera
		}
	}
	if camera == nil {
		panel.SetScale(mgl32.Vec3{})
		return panel
	}

	panel.SetAnchor(gfx.BottomCenter)
	panel.SetMarginBottom(.02)
	panel.SetScale(mgl32.Vec3{0.3, .1})
	panel.SetFillColor(gfx.Opacity(gfx.LightGray, 0.3))
	panel.SetBorderColor(gfx.LightGray)
	panel.SetBorderThickness(0.02)

	zoomSlider := gfx.NewSlider(gfx.Horizontal, true)
	zoomSlider.Button().SetScaleX(.081111)
	zoomSlider.Button().SetScaleY(.55)
	zoomSlider.SetScaleX(1)
	zoomSlider.SetScaleY(.9)
	zoomSlider.SetFillColor(gfx.Transparent)
	zoomSlider.Rail().SetColor(gfx.Darken(gfx.LightGray, .6))
	zoomSlider.Button().SetFillColor(gfx.LightGray)
	zoomSlider.Button().SetMouseEnterFillColor(gfx.Darken(gfx.LightGray, .3))
	zoomSlider.SetValue(.02)
	zoomSlider.OnValueChanging(func(sender gfx.WindowObject, value float32) {
		camera.Lock()
		camera.Properties.Position[2] = value * 50
		camera.Unlock()
	})

	zoomControlLabel := gfx.NewLabel()
	zoomControlLabel.SetText("Zoom")
	zoomControlLabel.SetColor(gfx.LightGray)
	zoomControlLabel.SetFontSize(.25)
	zoomControlLabel.SetAnchor(gfx.BottomCenter)
	zoomControlLabel.SetMarginTop(.02)

	panel.AddChild(zoomSlider)
	panel.AddChild(zoomControlLabel)

	return panel
}

func NewViewer(window *gfx.Window, obj *gfx.Shape3D) *gfx.View {
	view := gfx.NewView()
	view.SetMaintainAspectRatio(false)

	view.AddChild(obj)
	view.AddChild(newRotationControl(window, obj))

	lc1 := newLightControl(window, obj, 1, gfx.Blue)
	lc2 := newLightControl(window, obj, 2, gfx.Orange)
	lc1.SetAnchor(gfx.TopRight)
	lc1.SetMarginTop(.2)
	lc2.SetAnchor(gfx.BottomRight)
	lc2.SetMarginBottom(.2)
	view.AddChildren(lc1, lc2)

	view.AddChild(newCameraControl(obj))

	return view
}
