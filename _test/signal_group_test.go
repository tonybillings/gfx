package _test

import (
	"context"
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"testing"
	"time"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/_test"
)

const (
	signalSampleCount = 2000
)

func newSignalGroup(ctx context.Context, window *gfx.Window, signalSampleCount int) gfx.WindowObject {
	sg := gfx.NewSignalGroup(signalSampleCount, 3, []color.RGBA{gfx.Green, gfx.Blue, gfx.Orange, gfx.Purple}...)
	sg.EnableInspector()
	sg.SetInspectorAnchor(gfx.TopCenter)
	sg.SetInspectorMargin(gfx.Margin{Top: .05})

	const signalCount = 20

	labels := []string{"Sine / Green", "Square / Blue", "Saw / Orange", "Random / Purple"}
	labelsIdx := 0

	for i := 0; i < signalCount; i++ {
		sg.New(labels[labelsIdx])
		labelsIdx = (labelsIdx + 1) % len(labels)
	}

	go func(ctx context.Context, sg *gfx.SignalGroup) {
		<-window.ReadyChan()

		x := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			sineValue := math.Sin(float64(x) * .025)
			for i := 0; i < signalCount; i += len(labels) {
				sg.Signals()[i].AddSamples([]float64{sineValue})
			}

			squareValue := 0.0
			if sineValue > 0 {
				squareValue = 1.0
			} else {
				squareValue = -1.0
			}
			for i := 1; i < signalCount; i += len(labels) {
				sg.Signals()[i].AddSamples([]float64{squareValue})
			}

			sawValue := 2.0*float64(x%100)/100.0 - 1.0
			for i := 2; i < signalCount; i += len(labels) {
				sg.Signals()[i].AddSamples([]float64{sawValue})
			}

			for i := 3; i < signalCount; i += len(labels) {
				sg.Signals()[i].AddSamples([]float64{rand.Float64()})
			}

			x++
			time.Sleep(4 * time.Millisecond)
		}
	}(ctx, sg)

	return sg
}

// BenchmarkSignalGroup Use this test to measure performance,
// detect resource leaks, etc.  Try using the program argument
// "-test.benchtime 30s" (without the quotes) to run for 30 seconds.
func BenchmarkSignalGroup(b *testing.B) {
	startRoutineCount := runtime.NumGoroutine()

	_test.Begin()

	mainCtx, mainCancelFunc := context.WithCancel(context.Background())

	go func() {
		for i := 0; i < b.N; i++ {
			win := gfx.NewWindow().
				SetTitle(_test.WindowTitle).
				SetWidth(_test.WindowWidth).
				SetHeight(_test.WindowHeight).
				SetSecondary(true)

			ctx, cancelFunc := context.WithCancel(context.Background())

			win.AddObjects(newSignalGroup(ctx, win, signalSampleCount))
			win.EnableQuitKey(cancelFunc)
			win.EnableFullscreenKey()

			gfx.InitWindowAsync(win)

			select {
			case <-time.After(3 * time.Second):
				cancelFunc()
				gfx.CloseWindowAsync(win)
			case <-ctx.Done():
				break
			}
		}

		mainCancelFunc()
	}()

	gfx.Run(mainCtx, mainCancelFunc)
	_test.End()

	endRoutineCount := runtime.NumGoroutine()
	if endRoutineCount != startRoutineCount {
		b.Logf("Starting routine count: %d", startRoutineCount)
		b.Logf("Ending routine count: %d", endRoutineCount)
		b.Error("routine leak")
	}
}
