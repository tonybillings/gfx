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

	// The first hint hides the window border / title bar; the second
	// allows the window to be maximized/resized; the third enables/disables
	// multisampling (4xMSAA).
	hints := gfx.NewWindowHints(true, false, false)

	// If no hints are provided, the default result is a decorated (not
	// borderless), fixed-sized window with MSAA disabled.
	win := gfx.NewWindow(hints).
		SetTitle("You should not see this text!").
		SetWidth(windowWidth).
		SetHeight(windowHeight).
		SetClearColor(gfx.Black).
		EnableQuitKey() // the quit key is ESC (escape)

	// Regardless of hints, you can set the window to run in fullscreen mode at any time via:
	// win.SetFullscreenEnabled(true)
	// You can also enable the fullscreen key, which is F11
	win.EnableFullscreenKey()

	gfx.InitWindowAsync(win)

	help := gfx.NewLabel()
	help.SetText("Press ESC to quit!")
	help.SetColor(gfx.Green)
	help.SetFontSize(0.1)

	win.AddObjects(help)

	ctx, cancelFunc := context.WithCancel(context.Background())
	go waitForInterruptSignal(ctx, cancelFunc)
	gfx.Run(ctx, cancelFunc)
}
