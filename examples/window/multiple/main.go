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

	win1 := gfx.NewWindow().
		SetTitle("Red Window (win1)").
		SetPosition(windowWidth, windowHeight). // not doing anything precise here, just approximating...
		SetWidth(windowWidth).
		SetHeight(windowHeight).
		SetClearColor(gfx.Red)

	win2 := gfx.NewWindow().
		SetTitle("Green Window (win2)").
		SetPosition(windowWidth+windowWidth, windowHeight). // ...so that the windows start out separated, etc
		SetWidth(windowWidth).
		SetHeight(windowHeight).
		SetClearColor(gfx.Green).
		SetSecondary(true)

	gfx.InitWindowAsync(win1)
	gfx.InitWindowAsync(win2)

	btn1 := gfx.NewButton()
	btn1.SetText("Spawn a purple window!")
	btn1.SetScaleX(0.5).SetScaleY(0.2)
	btn1.SetFillColor(gfx.White)
	btn1.SetTextColor(gfx.Gray)
	btn1.SetMouseEnterFillColor(gfx.Darken(gfx.White, 0.1))
	btn1.SetFontSize(.2)
	btn1.SetBorderColor(gfx.Gray)
	btn1.SetBorderThickness(0.02)
	btn1.OnClick(func(sender gfx.WindowObject, mouseState *gfx.MouseState) {
		newWin := gfx.NewWindow().
			SetTitle("I am a secondary window!").
			SetWidth(windowWidth).
			SetHeight(windowHeight).
			SetClearColor(gfx.Purple).
			SetSecondary(true)
		gfx.InitWindowAsync(newWin)

		help := gfx.NewLabel()
		help.SetText("(closing me does only that)")
		help.SetColor(gfx.White)
		help.SetFontSize(0.05)
		newWin.AddObject(help)
	})
	win2.AddObjects(btn1)

	btn2 := gfx.NewButton()
	btn2.SetText("Spawn an orange window!")
	btn2.SetScaleX(0.5).SetScaleY(0.2)
	btn2.SetFillColor(gfx.White)
	btn2.SetMouseEnterFillColor(gfx.Darken(gfx.White, 0.1))
	btn2.SetTextColor(gfx.Gray)
	btn2.SetFontSize(.2)
	btn2.SetBorderColor(gfx.Gray)
	btn2.SetBorderThickness(0.02)
	btn2.OnClick(func(sender gfx.WindowObject, mouseState *gfx.MouseState) {
		newWin := gfx.NewWindow().
			SetTitle("I am a primary window!"). // ...but closing this one will close the app!
			SetWidth(windowWidth).
			SetHeight(windowHeight).
			SetClearColor(gfx.Orange)
		gfx.InitWindowAsync(newWin)

		help := gfx.NewLabel()
		help.SetText("(closing me closes the app!)")
		help.SetColor(gfx.White)
		help.SetFontSize(0.05)
		newWin.AddObject(help)
	})
	win1.AddObjects(btn2)

	ctx, cancelFunc := context.WithCancel(context.Background())
	go waitForInterruptSignal(ctx, cancelFunc)
	gfx.Run(ctx, cancelFunc)
}
