package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/examples/ui/view"
)

const (
	windowTitle       = "UI Test"
	windowWidth       = 1900 // try changing the width/height!
	windowHeight      = 1000
	targetFramerate   = 60 // decrease to assign more CPU time to sample acquisition efforts
	vSyncEnabled      = true
	signalSampleCount = 2000 // number of samples considered when calculating min/max/avg/std, etc
)

var (
	backgroundColor = gfx.Black
)

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func waitForInterruptSignal(ctx context.Context, cancelFunc context.CancelFunc) {
	sigIntChan := make(chan os.Signal, 1)
	signal.Notify(sigIntChan, syscall.SIGINT)

	select {
	case <-ctx.Done():
		return
	case <-sigIntChan:
		cancelFunc()
		return
	}
}

func main() {
	panicOnErr(gfx.Init())
	defer gfx.Close()

	gfx.SetTargetFramerate(targetFramerate)
	gfx.SetVSyncEnabled(vSyncEnabled)

	win := gfx.NewWindow().
		SetTitle(windowTitle).
		SetWidth(windowWidth).
		SetHeight(windowHeight).
		SetClearColor(backgroundColor)

	ctx, cancelFunc := context.WithCancel(context.Background())

	win.AddObjects(gfx.NewTabGroup(
		view.NewSignalsView(ctx, win, signalSampleCount),
		view.NewSignalInspectorView(ctx, win, signalSampleCount),
		view.NewSliderView(ctx, win, signalSampleCount),
		view.NewRadioButtonView(ctx, win, signalSampleCount),
		view.NewCheckBoxView(ctx, win, signalSampleCount),
		view.NewButtonView(win)))

	win.EnableQuitKey()
	win.EnableFullscreenKey()

	gfx.InitWindowAsync(win)

	go waitForInterruptSignal(ctx, cancelFunc)
	gfx.Run(ctx, cancelFunc)
}
