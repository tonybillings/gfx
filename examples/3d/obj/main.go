package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/examples/3d/obj/view"
)

const (
	windowTitle     = "3D Test (OBJ/MTL)"
	windowWidth     = 1900
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

	win.AddObjects(view.NewCubeView(win))

	win.EnableQuitKey()
	win.EnableFullscreenKey()

	gfx.InitWindowAsync(win)

	ctx, cancelFunc := context.WithCancel(context.Background())
	go waitForInterruptSignal(ctx, cancelFunc)
	gfx.Run(ctx, cancelFunc)
}
