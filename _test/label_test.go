package _test

import (
	"context"
	"testing"
	"time"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/_test"
)

const (
	winWidth1  = 1000
	winHeight1 = 1000
)

func TestLabelSetText(t *testing.T) {
	_test.Begin()
	defer _test.End()

	ctx, cancelFunc := context.WithCancel(context.Background())

	go func() { // worker thread...
		win := gfx.NewWindow().
			SetTitle(_test.WindowTitle).
			SetWidth(winWidth1).
			SetHeight(winHeight1)

		lbl := gfx.NewLabel()
		lbl.SetFontSize(1)
		lbl.SetColor(gfx.Magenta)
		lbl.SetText("O")

		validator := _test.NewSceneValidator(t, win)
		validator.AddPixelSampler(func() (x, y float32) { return 0, 0 }, _test.BackgroundColor, "center screen")
		validator.AddPixelSampler(func() (x, y float32) { return 0, .6 }, gfx.Magenta, "top center")
		validator.AddPixelSampler(func() (x, y float32) { return .5, 0 }, gfx.Magenta, "middle right")
		validator.AddPixelSampler(func() (x, y float32) { return 0, -.5 }, gfx.Magenta, "bottom center")
		validator.AddPixelSampler(func() (x, y float32) { return -.5, 0 }, gfx.Magenta, "middle left")

		win.AddObjects(lbl, validator)

		gfx.InitWindowAsync(win)
		<-win.ReadyChan()
		_test.SleepNFrames(5)

		validator.Validate()

		time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the initial label text

		lbl.SetText("X")
		validator.Samplers[0].ExpectedColor = gfx.Magenta
		for i := 1; i < len(validator.Samplers); i++ {
			validator.Samplers[i].ExpectedColor = _test.BackgroundColor
		}

		// May only need to sleep a couple frames, depending on system performance, target framerate, and label
		// scaling.  The larger the size of the label at the time it is initially rendered, the longer it will
		// take to render.  Again, just for the first time, as it's a texture that is created and stored in
		// VRAM at a size that's based on the label's scale at the time of creation...after that, it is treated
		// as a diffuse texture applied to a Shape2D instance, which means subsequent changes to the scaling
		// or font size of the label can happen without that performance hit.  If label caching is enabled (the
		// default), the next request to create a texture for a label with the same properties will be
		// responded to with a previously generated texture stored in the cache.  Changing the text of the
		// label always results in the creation of a new texture if caching is disabled or a suitable texture
		// does not yet exist in the cache.
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the text change

		cancelFunc()
	}()

	gfx.Run(ctx, cancelFunc)
}

func TestLabelAnchoring(t *testing.T) {
	_test.Begin()
	defer _test.End()

	ctx, cancelFunc := context.WithCancel(context.Background())

	go func() { // worker thread...
		win := gfx.NewWindow().
			SetTitle(_test.WindowTitle).
			SetWidth(winWidth1).
			SetHeight(winHeight1)

		lbl := gfx.NewLabel()
		lbl.SetFontSize(.1)
		lbl.SetAnchor(gfx.Center)
		lbl.SetScaleX(.1)
		lbl.SetColor(gfx.Magenta)
		lbl.SetText("X")

		validator := _test.NewSceneValidator(t, win)
		validator.AddPixelSampler(func() (x, y float32) { return 0, 0 }, gfx.Magenta, "center of the X")

		win.AddObjects(lbl, validator)

		gfx.InitWindowAsync(win)
		<-win.ReadyChan()
		_test.SleepNFrames(5)

		validator.Validate()

		time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the initial label position

		lbl.SetAnchor(gfx.TopLeft)
		lbl.RefreshLayout() // needed if changing the anchor after initialization
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return -1 + lbl.HalfWidth(), 1 - lbl.HalfHeight() }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the position change

		lbl.SetAnchor(gfx.TopCenter)
		lbl.RefreshLayout()
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return 0, 1 - lbl.HalfHeight() }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		lbl.SetAnchor(gfx.TopRight)
		lbl.RefreshLayout()
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return 1 - lbl.HalfWidth(), 1 - lbl.HalfHeight() }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		lbl.SetAnchor(gfx.MiddleRight)
		lbl.RefreshLayout()
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return 1 - lbl.HalfWidth(), 0 }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		lbl.SetAnchor(gfx.BottomRight)
		lbl.RefreshLayout()
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return 1 - lbl.HalfWidth(), -1 + lbl.HalfHeight() }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		lbl.SetAnchor(gfx.BottomCenter)
		lbl.RefreshLayout()
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return 0, -1 + lbl.HalfHeight() }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		lbl.SetAnchor(gfx.BottomLeft)
		lbl.RefreshLayout()
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return -1 + lbl.HalfWidth(), -1 + lbl.HalfHeight() }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		lbl.SetAnchor(gfx.MiddleLeft)
		lbl.RefreshLayout()
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return -1 + lbl.HalfWidth(), 0 }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		cancelFunc()
	}()

	gfx.Run(ctx, cancelFunc)
}

func TestLabelAlignment(t *testing.T) {
	_test.Begin()
	defer _test.End()

	ctx, cancelFunc := context.WithCancel(context.Background())

	go func() { // worker thread...
		win := gfx.NewWindow().
			SetTitle(_test.WindowTitle).
			SetWidth(winWidth1).
			SetHeight(winHeight1)

		lbl := gfx.NewLabel()
		lbl.SetFontSize(.1)
		lbl.SetAnchor(gfx.TopCenter)
		lbl.SetAlignment(gfx.Left)
		lbl.SetColor(gfx.Magenta)
		lbl.SetText("X")
		win.AddObjects(lbl)

		// This time we let the horizontal scale be 1.0 (stretches to the entire
		// width of the parent, which in this case is the window) and use text
		// alignment to place the label horizontally.  This means that we need
		// use a different method for accounting for the width of the rendered
		// text (the letter X in this test).  We'll just use hard-coded value...
		halfWidth := float32(.06)

		validator := _test.NewSceneValidator(t, win)
		validator.AddPixelSampler(func() (x, y float32) { return -1.0 + halfWidth, 1.0 - lbl.HalfHeight() }, gfx.Magenta, "center of the X")
		win.AddObjects(validator)

		gfx.InitWindowAsync(win)
		<-win.ReadyChan()
		_test.SleepNFrames(5)

		validator.Validate()

		time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the initial label position

		lbl.SetAlignment(gfx.Centered)
		lbl.RefreshLayout() // needed if changing the anchor after initialization
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return 0, 1 - lbl.HalfHeight() }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the position change

		lbl.SetAlignment(gfx.Right)
		lbl.RefreshLayout()
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return 1 - halfWidth, 1 - lbl.HalfHeight() }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		lbl.SetAnchor(gfx.Center)

		lbl.SetAlignment(gfx.Left)
		lbl.RefreshLayout()
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return -1 + halfWidth, 0 }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		lbl.SetAlignment(gfx.Centered)
		lbl.RefreshLayout()
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return 0, 0 }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		lbl.SetAlignment(gfx.Right)
		lbl.RefreshLayout()
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return 1 - halfWidth, 0 }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		lbl.SetAnchor(gfx.BottomCenter)

		lbl.SetAlignment(gfx.Left)
		lbl.RefreshLayout()
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return -1 + halfWidth, -1 + lbl.HalfHeight() }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		lbl.SetAlignment(gfx.Centered)
		lbl.RefreshLayout()
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return 0, -1 + lbl.HalfHeight() }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		lbl.SetAlignment(gfx.Right)
		lbl.RefreshLayout()
		validator.Samplers[0].GetPixelPosFunc = func() (float32, float32) { return 1 - halfWidth, -1 + lbl.HalfHeight() }
		_test.SleepNFrames(5)
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		cancelFunc()
	}()

	gfx.Run(ctx, cancelFunc)
}
