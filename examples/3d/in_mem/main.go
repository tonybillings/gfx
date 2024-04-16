package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/examples/3d/in_mem/view"
)

var (
	windowTitle     = "3D Test"
	windowWidth     = 1000
	windowHeight    = 1000
	targetFramerate = 60
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

	win := gfx.NewWindow().SetTitle(windowTitle).
		SetWidth(windowWidth).
		SetHeight(windowHeight).
		SetTargetFramerate(targetFramerate).
		SetClearColor(backgroundColor)
	defer win.Close()

	ctx, cancelFunc := context.WithCancel(context.Background())

	win.AddObjects(view.NewQuadView(win))

	win.EnableQuitKey(cancelFunc)
	win.EnableFullscreenKey()

	go waitForInterruptSignal(ctx, cancelFunc)
	win.Init(ctx, cancelFunc)

	<-ctx.Done()
}
