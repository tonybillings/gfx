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

func TestTabGroup(t *testing.T) {
	_test.PanicOnErr(gfx.Init())

	win := gfx.NewWindow().SetTitle(_test.WindowTitle).
		SetWidth(winWidth1).
		SetHeight(winHeight1).
		SetTargetFramerate(_test.TargetFramerate).
		SetClearColor(_test.BackgroundColor)
	ctx, cancelFunc := context.WithCancel(context.Background())

	win.EnableQuitKey(cancelFunc)
	win.EnableFullscreenKey()

	quad1 := gfx.NewQuad()
	quad1.SetColor(gfx.Red)
	quad1.SetMaintainAspectRatio(false)

	quad2 := gfx.NewQuad()
	quad2.SetColor(gfx.Green)
	quad2.SetMaintainAspectRatio(false)

	quad3 := gfx.NewQuad()
	quad3.SetColor(gfx.Blue)
	quad3.SetMaintainAspectRatio(false)

	tg := gfx.NewTabGroup(quad1, quad2, quad3)

	validator := _test.NewSceneValidator(win)
	validator.AddPixelSampler(func() (x, y float32) { return 0, 0 }, gfx.Red, "center screen")

	win.AddObjects(tg, validator)
	win.Init(ctx, cancelFunc)
	<-win.ReadyChan()
	_test.SleepAFewFrames()

	time.Sleep(200 * time.Millisecond) // *optional; give us some time to see the initial tab

	_test.ValidateScene(t, validator)

	tg.Next()
	time.Sleep(200 * time.Millisecond) // *optional; give us some time to see the active tab change
	validator.Samplers[0].ExpectedColor = gfx.Green
	_test.ValidateScene(t, validator)

	tg.Next()
	time.Sleep(200 * time.Millisecond)
	validator.Samplers[0].ExpectedColor = gfx.Blue
	_test.ValidateScene(t, validator)

	tg.Previous()
	time.Sleep(200 * time.Millisecond)
	validator.Samplers[0].ExpectedColor = gfx.Green
	_test.ValidateScene(t, validator)

	tg.Previous()
	time.Sleep(200 * time.Millisecond)
	validator.Samplers[0].ExpectedColor = gfx.Red
	_test.ValidateScene(t, validator)

	tg.Activate(1)
	time.Sleep(200 * time.Millisecond)
	validator.Samplers[0].ExpectedColor = gfx.Green
	_test.ValidateScene(t, validator)

	tg.Activate(2)
	time.Sleep(200 * time.Millisecond)
	validator.Samplers[0].ExpectedColor = gfx.Blue
	_test.ValidateScene(t, validator)

	win.Close()
	gfx.Close()
}
