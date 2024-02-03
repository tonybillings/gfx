package _test

import (
	"context"
	"runtime"
	"testing"
	"time"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/_test"
	"tonysoft.com/gfx/examples/ui/view"
)

func BenchmarkSignals(b *testing.B) {
	_test.PanicOnErr(gfx.Init())

	startRoutineCount := runtime.NumGoroutine()
	b.Logf("Starting routine count: %d", startRoutineCount)

	for i := 0; i < b.N; i++ {
		win := gfx.NewWindow().SetTitle(_test.WindowTitle).
			SetWidth(_test.WindowWidth).
			SetHeight(_test.WindowHeight).
			SetTargetFramerate(_test.TargetFramerate).
			SetClearColor(_test.BackgroundColor)
		ctx, cancelFunc := context.WithCancel(context.Background())

		win.AddObjects(gfx.NewTabGroup(
			view.NewSignalsView(ctx, win, _test.SignalSampleCount),
			view.NewSignalInspectorView(ctx, win, _test.SignalSampleCount),
			view.NewSliderView(ctx, win, _test.SignalSampleCount),
			view.NewRadioButtonView(ctx, win, _test.SignalSampleCount),
			view.NewCheckBoxView(ctx, win, _test.SignalSampleCount)))

		win.EnableQuitKey(cancelFunc)
		win.EnableFullscreenKey()

		win.Init(ctx, cancelFunc)

		select {
		case <-time.After(10 * time.Second):
			cancelFunc()
		case <-ctx.Done():
			break
		}

		time.Sleep(3 * time.Second) // wait for routines to return
	}

	endRoutineCount := runtime.NumGoroutine()
	b.Logf("Ending routine count: %d", endRoutineCount)

	if endRoutineCount != startRoutineCount {
		b.Log("Test failed: routine leak")
		b.Fail()
	}
}
