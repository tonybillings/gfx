package _test

import (
	"context"
	"github.com/go-gl/mathgl/mgl32"
	"testing"
	"time"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/_test"
)

const (
	winWidth1  = 1000
	winHeight1 = 1000
)

func TestSliderSetValue(t *testing.T) {
	_test.Begin()
	defer _test.End()

	ctx, cancelFunc := context.WithCancel(context.Background())

	go func() { // worker thread...
		win := gfx.NewWindow().
			SetTitle(_test.WindowTitle).
			SetWidth(winWidth1).
			SetHeight(winHeight1)

		ms := gfx.MouseState{
			X: -1,
			Y: -1,
		}
		mockWin := gfx.NewWindow()
		mockWin.OverrideMouseState(&ms)

		slider := gfx.NewSlider(gfx.Vertical)
		slider.SetFillColor(gfx.Gray)
		slider.Button().SetFillColor(gfx.Magenta)
		slider.SetScale(mgl32.Vec3{.2, .4})
		slider.SetValue(0)
		slider.SetMouseSurface(mockWin)

		validator := _test.NewSceneValidator(win)

		win.AddObjects(slider, validator)

		gfx.InitWindowAsync(win)
		<-win.ReadyChan()

		time.Sleep(200 * time.Millisecond) // *optional; allow us to see the slider's initial state

		validator.AddPixelSampler(func() (x, y float32) { return .01, -.28 }, gfx.Magenta, "inside slider button")
		_test.ValidateScene(t, validator)

		slider.SetValue(.5)                            // set to 50%
		_test.SleepAFewFrames()                        // allow the slider to render at the new value
		validator.Samplers[0].ExpectedColor = gfx.Gray // previous slider button position
		validator.AddPixelSampler(func() (x, y float32) { return .01, 0 }, gfx.Magenta, "inside slider button")
		_test.ValidateScene(t, validator)

		time.Sleep(200 * time.Millisecond) // *optional; allow us to see the change in value

		slider.SetValue(1) // set to 100%
		_test.SleepAFewFrames()
		validator.Samplers[1].ExpectedColor = gfx.Gray // previous slider button position
		validator.AddPixelSampler(func() (x, y float32) { return .01, .29 }, gfx.Magenta, "inside slider button")
		_test.ValidateScene(t, validator)

		time.Sleep(200 * time.Millisecond)

		cancelFunc()
	}()

	gfx.Run(ctx, cancelFunc)
}

func TestSliderValueChanged(t *testing.T) {
	_test.Begin()
	defer _test.End()

	ctx, cancelFunc := context.WithCancel(context.Background())

	go func() { // worker thread...
		win := gfx.NewWindow().
			SetTitle(_test.WindowTitle).
			SetWidth(winWidth1).
			SetHeight(winHeight1)

		mockWin := gfx.NewWindow()

		slider := gfx.NewSlider(gfx.Vertical)
		slider.SetFillColor(gfx.Gray)
		slider.Button().SetFillColor(gfx.Magenta)
		slider.SetScale(mgl32.Vec3{.2, .4})
		slider.SetValue(0)
		slider.OnValueChanged(func(sender gfx.WindowObject, value float32) {
			clearColor := gfx.Blue
			clearColor.B = uint8(value * float32(clearColor.B))
			win.SetClearColor(clearColor)
		})
		slider.SetMouseSurface(mockWin)

		validator := _test.NewSceneValidator(win)
		validator.AddPixelSampler(func() (x, y float32) { return .01, -.28 }, gfx.Magenta, "inside slider button")

		win.AddObjects(slider, validator)

		gfx.InitWindowAsync(win)
		<-win.ReadyChan()

		_test.ValidateScene(t, validator)

		_test.SimulateMouseClickAndDrag(mockWin, .01, -.28, .01, -.15, 10) // slide to ~23%
		validator.Samplers[0].ExpectedColor = gfx.Gray                     // previous slider button position
		validator.AddPixelSampler(func() (x, y float32) { return .01, -.15 }, gfx.Magenta, "inside slider button")
		_test.ValidateScene(t, validator)

		_test.SimulateMouseClickAndDrag(mockWin, .01, -.15, .01, 0, 10) // slide to 50%
		validator.Samplers[1].ExpectedColor = gfx.Gray                  // previous slider button position
		validator.AddPixelSampler(func() (x, y float32) { return .01, 0 }, gfx.Magenta, "inside slider button")
		validator.AddPixelSampler(func() (x, y float32) { return .99, .99 }, gfx.Darken(gfx.Blue, .5), "upper-right corner of window")
		_test.ValidateScene(t, validator)

		_test.SimulateMouseClickAndDrag(mockWin, .01, 0, .01, .29, 10) // slide to 100%
		validator.Samplers[2].ExpectedColor = gfx.Gray                 // previous slider button position
		validator.Samplers[3].ExpectedColor = gfx.Blue                 // upper-right corner of window
		validator.AddPixelSampler(func() (x, y float32) { return .01, .29 }, gfx.Magenta, "inside slider button")
		_test.ValidateScene(t, validator)

		cancelFunc()
	}()

	gfx.Run(ctx, cancelFunc)
}

func TestSliderValueChanging(t *testing.T) {
	_test.Begin()
	defer _test.End()

	ctx, cancelFunc := context.WithCancel(context.Background())

	go func() { // worker thread...
		win := gfx.NewWindow().
			SetTitle(_test.WindowTitle).
			SetWidth(winWidth1).
			SetHeight(winHeight1)

		mockWin := gfx.NewWindow()

		slider := gfx.NewSlider(gfx.Vertical)
		slider.SetFillColor(gfx.Gray)
		slider.Button().SetFillColor(gfx.Magenta)
		slider.SetScale(mgl32.Vec3{.2, .4})
		slider.SetValue(0)
		slider.OnValueChanging(func(sender gfx.WindowObject, value float32) {
			clearColor := gfx.Green
			clearColor.G = uint8(value * float32(clearColor.G))
			win.SetClearColor(clearColor)
		})
		slider.SetMouseSurface(mockWin)

		validator := _test.NewSceneValidator(win)
		validator.AddPixelSampler(func() (x, y float32) { return .01, -.28 }, gfx.Magenta, "inside slider button")

		win.AddObjects(slider, validator)

		gfx.InitWindowAsync(win)
		<-win.ReadyChan()

		_test.ValidateScene(t, validator)

		_test.SimulateMouseClickAndDrag(mockWin, .01, -.28, .01, -.15, 20) // slide to ~23%
		validator.Samplers[0].ExpectedColor = gfx.Gray                     // previous slider button position
		validator.AddPixelSampler(func() (x, y float32) { return .01, -.15 }, gfx.Magenta, "inside slider button")
		_test.ValidateScene(t, validator)

		_test.SimulateMouseClickAndDrag(mockWin, .01, -.15, .01, 0, 20) // slide to 50%
		validator.Samplers[1].ExpectedColor = gfx.Gray                  // previous slider button position
		validator.AddPixelSampler(func() (x, y float32) { return .01, 0 }, gfx.Magenta, "inside slider button")
		validator.AddPixelSampler(func() (x, y float32) { return .99, .99 }, gfx.Darken(gfx.Green, .5), "upper-right corner of window")
		_test.ValidateScene(t, validator)

		_test.SimulateMouseClickAndDrag(mockWin, .01, 0, .01, .29, 20) // slide to 100%
		validator.Samplers[2].ExpectedColor = gfx.Gray                 // previous slider button position
		validator.Samplers[3].ExpectedColor = gfx.Green                // upper-right corner of window
		validator.AddPixelSampler(func() (x, y float32) { return .01, .29 }, gfx.Magenta, "inside slider button")
		_test.ValidateScene(t, validator)

		cancelFunc()
	}()

	gfx.Run(ctx, cancelFunc)
}

func TestAnchoredSliderValueChanged(t *testing.T) {
	_test.Begin()
	defer _test.End()

	ctx, cancelFunc := context.WithCancel(context.Background())

	go func() { // worker thread...
		win := gfx.NewWindow().
			SetTitle(_test.WindowTitle).
			SetWidth(winWidth1).
			SetHeight(winHeight1)

		mockWin := gfx.NewWindow()

		slider := gfx.NewSlider(gfx.Vertical)
		slider.SetFillColor(gfx.Gray)
		slider.Button().SetFillColor(gfx.Magenta)
		slider.SetScale(mgl32.Vec3{.2, .4})
		slider.SetAnchor(gfx.BottomLeft)
		slider.SetValue(0)
		slider.OnValueChanged(func(sender gfx.WindowObject, value float32) {
			clearColor := gfx.Blue
			clearColor.B = uint8(value * float32(clearColor.B))
			win.SetClearColor(clearColor)
		})
		slider.SetMouseSurface(mockWin)

		posX := float32(-1.0 + (.2*.5)*2.0) // adding half the width to the left screen edge
		posY := float32(-1.0 + (.4*.5)*2.0) // adding half the height to the bottom screen edge
		// ...and in both cases, we don't have to worry about adjusting for MAR
		// (Maintain Aspect Ratio), as the current ratio is 1:1.  But we do multiply by
		// 2.0 because this is dealing with the screen space range of [-1,1].

		validator := _test.NewSceneValidator(win)
		validator.AddPixelSampler(func() (x, y float32) { return posX + .015, posY + -.28 }, gfx.Magenta, "inside slider button")

		win.AddObjects(slider, validator)

		gfx.InitWindowAsync(win)
		<-win.ReadyChan()

		_test.ValidateScene(t, validator)

		_test.SimulateMouseClickAndDrag(mockWin, posX+.01, posY+-.28, posX+.01, posY+-.15, 10) // slide to ~23%
		validator.Samplers[0].ExpectedColor = gfx.Gray                                         // previous slider button position
		validator.AddPixelSampler(func() (x, y float32) { return posX + .015, posY + -.15 }, gfx.Magenta, "inside slider button")
		_test.ValidateScene(t, validator)

		_test.SimulateMouseClickAndDrag(mockWin, posX+.01, posY+-.15, posX+.01, posY+0, 10) // slide to 50%
		validator.Samplers[1].ExpectedColor = gfx.Gray                                      // previous slider button position
		validator.AddPixelSampler(func() (x, y float32) { return posX + .015, posY + 0 }, gfx.Magenta, "inside slider button")
		validator.AddPixelSampler(func() (x, y float32) { return .99, .99 }, gfx.Darken(gfx.Blue, .5), "upper-right corner of window")
		_test.ValidateScene(t, validator)

		_test.SimulateMouseClickAndDrag(mockWin, posX+.01, posY+0, posX+.01, posY+.29, 10) // slide to 100%
		validator.Samplers[2].ExpectedColor = gfx.Gray                                     // previous slider button position
		validator.Samplers[3].ExpectedColor = gfx.Blue                                     // upper-right corner of window
		validator.AddPixelSampler(func() (x, y float32) { return posX + .015, posY + .29 }, gfx.Magenta, "inside slider button")
		_test.ValidateScene(t, validator)

		cancelFunc()
	}()

	gfx.Run(ctx, cancelFunc)
}

func TestAnchoredHorizontalSliderValueChanged(t *testing.T) {
	_test.Begin()
	defer _test.End()

	ctx, cancelFunc := context.WithCancel(context.Background())

	go func() { // worker thread...
		win := gfx.NewWindow().
			SetTitle(_test.WindowTitle).
			SetWidth(winWidth1).
			SetHeight(winHeight1)

		mockWin := gfx.NewWindow()

		slider := gfx.NewSlider(gfx.Horizontal)
		slider.SetFillColor(gfx.Gray)
		slider.Button().SetFillColor(gfx.Magenta)
		slider.SetScale(mgl32.Vec3{.4, .2})
		slider.SetAnchor(gfx.BottomRight)
		slider.SetValue(0)
		slider.OnValueChanged(func(sender gfx.WindowObject, value float32) {
			clearColor := gfx.Red
			clearColor.R = uint8(value * float32(clearColor.R))
			win.SetClearColor(clearColor)
		})
		slider.SetMouseSurface(mockWin)

		posX := float32(1.0 - (.4*.5)*2.0)  // adding half the width to the right screen edge
		posY := float32(-1.0 + (.2*.5)*2.0) // adding half the height to the bottom screen edge
		// ...and in both cases, we don't have to worry about adjusting for MAR
		// (Maintain Aspect Ratio), as the current ratio is 1:1.  But we do multiply by
		// 2.0 because this is dealing with the screen space range of [-1,1].

		validator := _test.NewSceneValidator(win)
		validator.AddPixelSampler(func() (x, y float32) { return posX + -.28, posY + .015 }, gfx.Magenta, "inside slider button")

		win.AddObjects(slider, validator)

		gfx.InitWindowAsync(win)
		<-win.ReadyChan()

		_test.ValidateScene(t, validator)

		_test.SimulateMouseClickAndDrag(mockWin, posX+-.28, posY+.01, posX+-.15, posY+.01, 10) // slide to ~23%
		validator.Samplers[0].ExpectedColor = gfx.Gray                                         // previous slider button position
		validator.AddPixelSampler(func() (x, y float32) { return posX + -.15, posY + .015 }, gfx.Magenta, "inside slider button")
		_test.ValidateScene(t, validator)

		_test.SimulateMouseClickAndDrag(mockWin, posX+-.15, posY+.01, posX+0, posY+.01, 10) // slide to 50%
		validator.Samplers[1].ExpectedColor = gfx.Gray                                      // previous slider button position
		validator.AddPixelSampler(func() (x, y float32) { return posX + 0, posY + .015 }, gfx.Magenta, "inside slider button")
		validator.AddPixelSampler(func() (x, y float32) { return .99, .99 }, gfx.Darken(gfx.Red, .5), "upper-right corner of window")
		_test.ValidateScene(t, validator)

		_test.SimulateMouseClickAndDrag(mockWin, posX+0, posY+.01, posX+.29, posY+.01, 10) // slide to 100%
		validator.Samplers[2].ExpectedColor = gfx.Gray                                     // previous slider button position
		validator.Samplers[3].ExpectedColor = gfx.Red                                      // upper-right corner of window
		validator.AddPixelSampler(func() (x, y float32) { return posX + .29, posY + .015 }, gfx.Magenta, "inside slider button")
		_test.ValidateScene(t, validator)

		cancelFunc()
	}()

	gfx.Run(ctx, cancelFunc)
}
