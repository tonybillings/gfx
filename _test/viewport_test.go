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

func TestNewViewport(t *testing.T) {
	windowWidth, windowHeight := 1920, 1080
	v := gfx.NewViewport(windowWidth, windowHeight)

	expectedX, expectedY := int32(0), int32(0)
	expectedW, expectedH := int32(windowWidth), int32(windowHeight)

	x, y, w, h := v.Get()
	if x != expectedX || y != expectedY || w != expectedW || h != expectedH {
		t.Errorf("unexpected viewport: expected (%d, %d, %d, %d), got (%d, %d, %d, %d)", expectedX, expectedY, expectedW, expectedH, x, y, w, h)
	}
}

func TestViewportSet(t *testing.T) {
	v := gfx.NewViewport(800, 600)

	v.Set(0.1, 0.1, 0.5, 0.5)

	expectedX, expectedY := int32(80), int32(60)
	expectedW, expectedH := int32(400), int32(300)

	x, y, w, h := v.Get()
	if x != expectedX || y != expectedY || w != expectedW || h != expectedH {
		t.Errorf("unexpected viewport: expected (%d, %d, %d, %d), got (%d, %d, %d, %d)", expectedX, expectedY, expectedW, expectedH, x, y, w, h)
	}
}

func TestViewportSetWindowSize(t *testing.T) {
	v := gfx.NewViewport(800, 600)

	v.Set(0.1, 0.1, 0.5, 0.5)
	v.SetWindowSize(1600, 1200)

	expectedX, expectedY := int32(160), int32(120)
	expectedW, expectedH := int32(800), int32(600)

	x, y, w, h := v.Get()
	if x != expectedX || y != expectedY || w != expectedW || h != expectedH {
		t.Errorf("unexpected viewport: expected (%d, %d, %d, %d), got (%d, %d, %d, %d)", expectedX, expectedY, expectedW, expectedH, x, y, w, h)
	}
}

func TestViewportGet(t *testing.T) {
	v := gfx.NewViewport(1024, 768)

	x, y, w, h := v.Get()

	if x != 0 || y != 0 || w != 1024 || h != 768 {
		t.Errorf("unexpected viewport: expected (0, 0, 1024, 768), got (%d, %d, %d, %d)", x, y, w, h)
	}
}

func TestViewportRendering(t *testing.T) {
	_test.PanicOnErr(gfx.Init())

	win := gfx.NewWindow().SetTitle(_test.WindowTitle).
		SetWidth(winWidth1).
		SetHeight(winHeight1).
		SetTargetFramerate(_test.TargetFramerate).
		SetClearColor(_test.BackgroundColor)
	ctx, cancelFunc := context.WithCancel(context.Background())

	win.EnableQuitKey(cancelFunc)
	win.EnableFullscreenKey()

	quad := gfx.NewQuad()
	quad.SetColor(gfx.Magenta)
	quad.SetScale(mgl32.Vec3{.5, .5})
	quad.SetAnchor(gfx.TopRight)

	validator := _test.NewSceneValidator(win)
	validator.AddPixelSampler(func() (x, y float32) { return -.5, -.5 }, _test.BackgroundColor, "lower-left quadrant")

	win.AddObjects(quad, validator)
	win.Init(ctx, cancelFunc)
	<-win.ReadyChan()
	_test.SleepAFewFrames()

	_test.ValidateScene(t, validator)

	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the color change

	vp := gfx.NewViewport(winWidth1, winHeight1)
	vp.Set(-1, -1, 2, 2)
	quad.SetViewport(vp)

	validator.Samplers[0].ExpectedColor = gfx.Magenta
	_test.ValidateScene(t, validator)

	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the color change

	vp.Set(0, 0, 1, 1)

	validator.Samplers[0].ExpectedColor = _test.BackgroundColor
	_test.ValidateScene(t, validator)

	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the color change

	win.Close()
	gfx.Close()
}
