package _test

import (
	"fmt"
	"github.com/tonybillings/gfx"
	"os"
	"runtime"
	"time"
)

var (
	startRoutineCount int
)

func PanicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Begin() {
	// Outside of testing, don't explicitly do this;
	// rely on gfx.init() (the unexported init func)
	// to be called by the Go runtime for you and to
	// successfully lock to the main thread.
	runtime.LockOSThread()

	startRoutineCount = runtime.NumGoroutine()

	PanicOnErr(gfx.Init())

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		gfx.SetTargetFramerate(60)
	} else {
		gfx.SetTargetFramerate(TargetFramerate)
	}

	gfx.SetVSyncEnabled(VSyncEnabled)
}

func End() {
	gfx.Close()

	endRoutineCount := runtime.NumGoroutine()
	if endRoutineCount != startRoutineCount {
		panic(fmt.Errorf("routine leak detected: expected %d, got %d", startRoutineCount, endRoutineCount))
	}
}

func SleepACoupleFrames() {
	time.Sleep(time.Duration((1000/TargetFramerate)*2) * time.Millisecond)
}

func SleepAFewFrames() {
	time.Sleep(time.Duration((1000/TargetFramerate)*3) * time.Millisecond)
}

func SleepNFrames(n int) {
	time.Sleep(time.Duration((1000/TargetFramerate)*n) * time.Millisecond)
}
