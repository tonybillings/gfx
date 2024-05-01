package _test

import "tonysoft.com/gfx"

const (
	WindowTitle     = "Test"
	WindowWidth     = 1900
	WindowHeight    = 1000
	TargetFramerate = 200 // recommend a refresh rate of 120Hz with this value; *see note...
	VSyncEnabled    = true
)

var (
	BackgroundColor = gfx.Black
)

// *Some tests will begin to fail once the target framerate is too
// high, depending on system performance and current V-Sync settings
// (the higher the refresh rate the better, when V-Sync is enabled).
// Although the actual rendered framerate will not exceed the refresh
// rate when V-Sync is enabled (regardless of the target framerate
// set above) that target rate is still used in sleep calculations
// throughout the accompanying tests.  If these tests fail on your
// system, try lowering the target framerate...
