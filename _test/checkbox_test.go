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

func TestCheckBoxClickAndAnchoring(t *testing.T) {
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

	chk := gfx.NewCheckBox()
	chk.SetColor(gfx.Magenta)
	chk.SetScale(mgl32.Vec3{.05, .05})
	chk.SetAnchor(gfx.TopRight)
	chk.SetMargin(gfx.Margin{Top: .1, Right: .1})
	chk.OnCheckedChanged(func(sender gfx.WindowObject, checked bool) {
		if checked {
			win.SetClearColor(gfx.Blue)
		} else {
			win.SetClearColor(gfx.Green)
		}
	})
	chk.SetMouseSurface(mockWin)
	chk.SetWindow(win)

	posX := (1.0 - chk.Margin().Right) - chk.Width()*0.5
	posY := (1.0 - chk.Margin().Top) - chk.Height()*0.5

	validator := _test.NewSceneValidator(win)
	validator.AddPixelSampler(func() (x, y float32) { return 0, 0 }, _test.BackgroundColor, "center screen")
	validator.AddPixelSampler(func() (x, y float32) { return posX, posY }, _test.BackgroundColor, "inside the checkbox")

	win.AddObjects(chk, validator)
	win.Init(ctx, cancelFunc)
	<-win.ReadyChan()
	_test.SleepAFewFrames()

	_test.ValidateScene(t, validator)

	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the color change

	_test.SimulateMouseClick(mockWin, posX, posY) // check the checkbox
	validator.Samplers[0].ExpectedColor = gfx.Blue
	validator.Samplers[1].ExpectedColor = gfx.Magenta // the "checkmark" should be visible
	_test.ValidateScene(t, validator)

	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the color change

	_test.SimulateMouseClick(mockWin, posX, posY) // uncheck the checkbox
	validator.Samplers[0].ExpectedColor = gfx.Green
	validator.Samplers[1].ExpectedColor = _test.BackgroundColor // "checkmark" should be invisible again
	_test.ValidateScene(t, validator)

	win.Close()
	gfx.Close()
}
