package _test

import (
	"context"
	"github.com/go-gl/mathgl/mgl32"
	"strconv"
	"testing"
	"time"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/_test"
)

const (
	winWidth1  = 1000
	winHeight1 = 1000
)

func TestButtonClickAndAnchoring(t *testing.T) {
	_test.PanicOnErr(gfx.Init())

	win := gfx.NewWindow().SetTitle(_test.WindowTitle).
		SetWidth(winWidth1).
		SetHeight(winHeight1).
		SetTargetFramerate(_test.TargetFramerate).
		SetClearColor(_test.BackgroundColor)
	ctx, cancelFunc := context.WithCancel(context.Background())

	win.EnableQuitKey(cancelFunc)
	win.EnableFullscreenKey()

	mockWin := gfx.NewWindow()

	btn := gfx.NewButton()
	btn.SetFillColor(gfx.Magenta)
	btn.SetScale(mgl32.Vec3{.25, .25})
	btn.SetAnchor(gfx.BottomRight)
	btn.OnClick(func(sender gfx.WindowObject, mouseState *gfx.MouseState) {
		anchor := gfx.BottomRight
		if a, ok := btn.Tag().(gfx.Anchor); ok { // not necessary to use Tag for storing anchor, but just for example...
			anchor = a
		}
		switch anchor { // with each click, rotate the anchor to the next corner, clock-wise rotation
		case gfx.BottomRight:
			anchor = gfx.BottomLeft
		case gfx.BottomLeft:
			anchor = gfx.TopLeft
		case gfx.TopLeft:
			anchor = gfx.TopRight
		case gfx.TopRight:
			anchor = gfx.BottomRight
		}
		btn.SetTag(anchor)
		btn.SetAnchor(anchor) // ...this is why using Tag isn't necessary, as we could use [Set]Anchor()
		btn.RefreshLayout()
	})
	btn.SetMouseSurface(mockWin)

	validator := _test.NewSceneValidator(win)
	validator.AddPixelSampler(func() (x, y float32) { return 0, 0 }, _test.BackgroundColor, "center screen")
	validator.AddPixelSampler(func() (x, y float32) { return .85, -.85 }, gfx.Magenta, "inside the button, bottom-right corner")

	win.AddObjects(btn, validator)
	win.Init(ctx, cancelFunc)
	<-win.ReadyChan()
	_test.SleepAFewFrames()

	_test.ValidateScene(t, validator)

	time.Sleep(200 * time.Millisecond) // *optional; give us some time to see the position change

	_test.SimulateMouseClick(mockWin, .85, -.85)
	validator.Samplers[1].ExpectedColor = _test.BackgroundColor
	validator.AddPixelSampler(func() (x, y float32) { return -.85, -.85 }, gfx.Magenta, "inside the button, bottom-left corner")
	_test.ValidateScene(t, validator)

	time.Sleep(200 * time.Millisecond) // *optional; give us some time to see the position change

	_test.SimulateMouseClick(mockWin, -.85, -.85)
	validator.Samplers[2].ExpectedColor = _test.BackgroundColor
	validator.AddPixelSampler(func() (x, y float32) { return -.85, .85 }, gfx.Magenta, "inside the button, top-left corner")
	_test.ValidateScene(t, validator)

	time.Sleep(200 * time.Millisecond) // *optional; give us some time to see the position change

	_test.SimulateMouseClick(mockWin, -.85, .85)
	validator.Samplers[3].ExpectedColor = _test.BackgroundColor
	validator.AddPixelSampler(func() (x, y float32) { return .85, .85 }, gfx.Magenta, "inside the button, top-right corner")
	_test.ValidateScene(t, validator)

	time.Sleep(200 * time.Millisecond) // *optional; give us some time to see the position change

	win.Close()
	gfx.Close()
}

func TestButtonEnableDisable(t *testing.T) {
	_test.PanicOnErr(gfx.Init())

	win := gfx.NewWindow().SetTitle(_test.WindowTitle).
		SetWidth(winWidth1).
		SetHeight(winHeight1).
		SetTargetFramerate(_test.TargetFramerate).
		SetClearColor(_test.BackgroundColor)
	ctx, cancelFunc := context.WithCancel(context.Background())

	win.EnableQuitKey(cancelFunc)
	win.EnableFullscreenKey()

	mockWin := gfx.NewWindow()

	btn := gfx.NewButton()
	btn.SetScale(mgl32.Vec3{.5, .25})
	btn.SetBorderThickness(.2)
	btn.SetBorderColor(gfx.Magenta)
	btn.SetDisabledBorderColor(gfx.Gray)
	btn.SetDisabledTextColor(gfx.Gray)
	btn.SetTextColor(gfx.Magenta)
	btn.Label().SetAnchor(gfx.TopCenter)
	btn.Label().SetMarginTop(.02)
	btn.SetText("0")
	btn.OnClick(func(sender gfx.WindowObject, mouseState *gfx.MouseState) {
		if count, err := strconv.Atoi(btn.Text()); err != nil {
			t.Error(err)
			t.FailNow()
		} else {
			count++
			btn.SetText(strconv.Itoa(count))
		}
	})
	btn.SetMouseSurface(mockWin)

	validator := _test.NewSceneValidator(win)
	validator.AddPixelSampler(func() (x, y float32) { return 0, 0 }, _test.BackgroundColor, "center button/text")
	validator.AddPixelSampler(func() (x, y float32) { return -.498, 0 }, gfx.Magenta, "button left border")

	win.AddObjects(btn, validator)
	win.Init(ctx, cancelFunc)
	<-win.ReadyChan()
	_test.SleepAFewFrames()

	_test.ValidateScene(t, validator)

	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the initial button text

	_test.SimulateMouseClick(mockWin, 0, 0)
	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the text change
	if btn.Text() != "1" {
		t.Errorf("unexpected button text, expected %s, got %s", "1", btn.Text())
	}

	validator.Samplers[0].ExpectedColor = gfx.Magenta
	_test.ValidateScene(t, validator)

	_test.SimulateMouseClick(mockWin, 0, 0)
	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the text change
	if btn.Text() != "2" {
		t.Errorf("unexpected button text, expected %s, got %s", "2", btn.Text())
	}

	validator.Samplers[0].ExpectedColor = gfx.Magenta
	_test.ValidateScene(t, validator)

	btn.SetEnabled(false) // clicking the button should not result in any change now

	_test.SimulateMouseClick(mockWin, 0, 0)
	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the text *NOT* change
	if btn.Text() != "2" {
		t.Errorf("unexpected button text, expected %s, got %s", "2", btn.Text())
	}

	validator.Samplers[0].ExpectedColor = gfx.Gray
	validator.Samplers[1].ExpectedColor = gfx.Gray
	_test.ValidateScene(t, validator)

	btn.SetEnabled(true) // allow clicking once again

	_test.SimulateMouseClick(mockWin, 0, 0)
	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the text change
	if btn.Text() != "3" {
		t.Errorf("unexpected button text, expected %s, got %s", "3", btn.Text())
	}

	validator.Samplers[0].ExpectedColor = gfx.Magenta
	validator.Samplers[1].ExpectedColor = gfx.Magenta
	_test.ValidateScene(t, validator)

	win.Close()
	gfx.Close()
}