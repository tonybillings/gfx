package main

import (
	"context"
	"github.com/tonybillings/gfx"
	"github.com/tonybillings/gfx/examples/3d/in_mem/view"
	"os"
	"os/signal"
	"syscall"
)

const (
	windowTitle     = "3D Test (in memory)"
	windowWidth     = 1000
	windowHeight    = 1000
	targetFramerate = 60
	vSyncEnabled    = true
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

	win := gfx.NewWindow().SetTitle(windowTitle).
		SetWidth(windowWidth).
		SetHeight(windowHeight).
		SetClearColor(backgroundColor)

	win.AddObjects(view.NewQuadView(win))

	win.EnableQuitKey()
	win.EnableFullscreenKey()

	gfx.InitWindowAsync(win)

	ctx, cancelFunc := context.WithCancel(context.Background())
	go waitForInterruptSignal(ctx, cancelFunc)
	gfx.Run(ctx, cancelFunc)
}
