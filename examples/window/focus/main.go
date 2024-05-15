package main

import (
	"context"
	"github.com/tonybillings/gfx"
	"github.com/tonybillings/gfx/examples/ui/view"
	"os"
	"os/signal"
	"syscall"
)

var (
	windowWidth       = 1000
	windowHeight      = 1000
	signalSampleCount = 2000
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

	ctx, cancelFunc := context.WithCancel(context.Background())

	win1 := gfx.NewWindow().
		SetTitle("I will pause when not focused!").
		SetPosition(windowWidth, windowHeight). // not doing anything precise here, just approximating...
		SetWidth(windowWidth).
		SetHeight(windowHeight).
		DisableOnBlur(true) // this is how to make the window pause when not focused

	win2 := gfx.NewWindow().
		SetTitle("But I will always keep running!").
		SetPosition(windowWidth+windowWidth, windowHeight). // ...so that the windows start out separated, etc
		SetWidth(windowWidth).
		SetHeight(windowHeight)

	win1.AddObjects(view.NewSignalsView(ctx, win1, signalSampleCount))
	win2.AddObjects(view.NewSignalsView(ctx, win2, signalSampleCount))

	gfx.InitWindowAsync(win1)
	gfx.InitWindowAsync(win2)

	go waitForInterruptSignal(ctx, cancelFunc)
	gfx.Run(ctx, cancelFunc)
}
