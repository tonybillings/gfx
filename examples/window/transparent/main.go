package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"tonysoft.com/gfx"
)

var (
	windowWidth  = 1000
	windowHeight = 1000
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

	win := gfx.NewWindow().
		SetTitle("You should see what is behind me!").
		SetWidth(windowWidth).
		SetHeight(windowHeight).
		SetClearColor(gfx.Black).
		SetOpacity(64). // this is how we make the window transparent
		EnableQuitKey()

	gfx.InitWindowAsync(win)

	// Since we're not using embedded files or absolute paths,
	// you must execute this program using the run.sh script so
	// that the relative path to the texture is correct.  Running
	// within an IDE like GoLand can also work but only if the
	// working directory is set to this location and not the
	// root directory of the project/module.
	texture := gfx.NewTexture2D("emoji.png", "emoji.png")
	win.Assets().Add(texture)

	quad := gfx.NewQuad()
	quad.SetTexture(texture)

	win.AddObjects(quad)

	ctx, cancelFunc := context.WithCancel(context.Background())
	go waitForInterruptSignal(ctx, cancelFunc)
	gfx.Run(ctx, cancelFunc)
}
