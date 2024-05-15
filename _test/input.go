package _test

import "github.com/tonybillings/gfx"

/******************************************************************************
 MockMouse
******************************************************************************/

type MockMouse struct {
	window *gfx.Window
}

func (m *MockMouse) Click(x, y float32) {
	ms := gfx.MouseState{
		X: x,
		Y: y,
	}
	m.window.OverrideMouseState(&ms)
	SleepAFewFrames() // wait for the mouse over event to be triggered/handled
	ms.PrimaryDown = true
	m.window.OverrideMouseState(&ms)
	SleepAFewFrames() // wait for the mouse down event to be triggered/handled
	ms.PrimaryDown = false
	m.window.OverrideMouseState(&ms)
	SleepAFewFrames() // wait for the mouse up event to be triggered/handled
}

func (m *MockMouse) ClickAndDrag(startX, startY, endX, endY, steps float32) {
	ms := gfx.MouseState{
		X: startX,
		Y: startY,
	}

	m.window.OverrideMouseState(&ms)
	SleepACoupleFrames()

	deltaX := endX - startX
	deltaY := endY - startY
	stepX := deltaX / steps
	stepY := deltaY / steps

	for i := 0; i < int(steps); i++ {
		ms.X += stepX
		ms.Y += stepY
		ms.PrimaryDown = true
		m.window.OverrideMouseState(&ms)
		SleepACoupleFrames()
	}

	ms.PrimaryDown = false
	m.window.OverrideMouseState(&ms)
	SleepAFewFrames()
}

func NewMockMouse(window *gfx.Window) *MockMouse {
	return &MockMouse{
		window: window,
	}
}
