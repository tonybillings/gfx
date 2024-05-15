package _test

import (
	"github.com/tonybillings/gfx"
	"os"
)

/******************************************************************************
 MockMouse
******************************************************************************/

type MockMouse struct {
	window    *gfx.Window
	sleepTime int
}

func (m *MockMouse) Click(x, y float32) {
	ms := gfx.MouseState{
		X: x,
		Y: y,
	}
	m.window.OverrideMouseState(&ms)
	SleepNFrames(m.sleepTime) // wait for the mouse over event to be triggered/handled
	ms.PrimaryDown = true
	m.window.OverrideMouseState(&ms)
	SleepNFrames(m.sleepTime) // wait for the mouse down event to be triggered/handled
	ms.PrimaryDown = false
	m.window.OverrideMouseState(&ms)
	SleepNFrames(m.sleepTime) // wait for the mouse up event to be triggered/handled
}

func (m *MockMouse) ClickAndDrag(startX, startY, endX, endY, steps float32) {
	ms := gfx.MouseState{
		X: startX,
		Y: startY,
	}

	m.window.OverrideMouseState(&ms)
	SleepNFrames(m.sleepTime)

	deltaX := endX - startX
	deltaY := endY - startY
	stepX := deltaX / steps
	stepY := deltaY / steps

	for i := 0; i < int(steps); i++ {
		ms.X += stepX
		ms.Y += stepY
		ms.PrimaryDown = true
		m.window.OverrideMouseState(&ms)
		SleepNFrames(m.sleepTime)
	}

	ms.PrimaryDown = false
	m.window.OverrideMouseState(&ms)
	SleepNFrames(m.sleepTime)
}

func NewMockMouse(window *gfx.Window) *MockMouse {
	m := &MockMouse{
		window: window,
	}

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		m.sleepTime = 6
	} else {
		m.sleepTime = 2
	}

	return m
}
