package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"tonysoft.com/gfx"
)

var (
	windowTitle       = "GFX Test"
	windowWidth       = 1900
	windowHeight      = 1000
	targetFramerate   = 60   // decrease to assign more CPU time to sample acquisition efforts
	signalSampleCount = 2000 // the number of samples considered when calculating min/max/avg/std, etc
	backgroundColor   = gfx.Black
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

	win := gfx.NewWindow().SetTitle(windowTitle).
		SetWidth(windowWidth).
		SetHeight(windowHeight).
		SetTargetFramerate(targetFramerate).
		SetClearColor(backgroundColor)
	defer win.Close()

	ctx, cancelFunc := context.WithCancel(context.Background())

	win.AddObjects(NewTestView(ctx, win))

	win.EnableQuitButton(cancelFunc)
	win.EnableFullscreenButton()

	go waitForInterruptSignal(ctx, cancelFunc)
	win.Init(ctx, cancelFunc)

	<-ctx.Done()
}
