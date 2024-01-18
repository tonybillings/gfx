package main

import (
	"context"
	"github.com/go-gl/glfw/v3.3/glfw"
	"os"
	"os/signal"
	"syscall"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/examples/ui/view"
)

var (
	windowTitle     = "GFX Test"
	windowWidth     = 1900 // try changing the width/height!
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

	win.AddObjects(view.NewButtonView(win))

	win.AddKeyEventHandler(glfw.KeyEscape, glfw.Press, func(_ *gfx.Window, _ glfw.Key, _ glfw.Action) {
		cancelFunc()
	})

	go waitForInterruptSignal(ctx, cancelFunc)
	win.Init(ctx, cancelFunc)

	<-ctx.Done()
}
