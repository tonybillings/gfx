package _test

import (
	"context"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tonybillings/gfx"
	"github.com/tonybillings/gfx/_test"
	"testing"
	"time"
)

const (
	winWidth1  = 1000
	winHeight1 = 1000
)

func TestCheckBoxClickAndAnchoring(t *testing.T) {
	_test.Begin()
	defer _test.End()

	ctx, cancelFunc := context.WithCancel(context.Background())

	go func() { // worker thread...
		win := gfx.NewWindow().SetTitle(_test.WindowTitle).
			SetWidth(winWidth1).
			SetHeight(winHeight1).
			SetClearColor(_test.BackgroundColor)

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

		validator := _test.NewSceneValidator(t, win)
		validator.AddPixelSampler(func() (x, y float32) { return 0, 0 }, _test.BackgroundColor, "center screen")
		validator.AddPixelSampler(func() (x, y float32) { return posX, posY }, _test.BackgroundColor, "inside the checkbox")

		win.AddObjects(chk, validator)

		mockMouse := _test.NewMockMouse(mockWin)

		gfx.InitWindowAsync(win)
		<-win.ReadyChan()

		validator.Validate()

		time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the color change

		mockMouse.Click(posX, posY) // check the checkbox
		validator.Samplers[0].ExpectedColor = gfx.Blue
		validator.Samplers[1].ExpectedColor = gfx.Magenta // the "checkmark" should be visible
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		mockMouse.Click(posX, posY) // uncheck the checkbox
		validator.Samplers[0].ExpectedColor = gfx.Green
		validator.Samplers[1].ExpectedColor = _test.BackgroundColor // "checkmark" should be invisible again
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		cancelFunc()
	}()

	gfx.Run(ctx, cancelFunc)
}
