package gfx

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

/******************************************************************************
 Keyboard Input
******************************************************************************/

type KeyEventHandler struct {
	Key      glfw.Key
	Action   glfw.Action
	Callback func(*Window, glfw.Key, glfw.Action)
}

/******************************************************************************
 Mouse Input
******************************************************************************/

type MouseState struct {
	X, Y                       float32
	PrimaryDown, SecondaryDown bool
	ButtonsSwapped             bool
}

func (s *MouseState) Update(button glfw.MouseButton, action glfw.Action) {
	switch button {
	case glfw.MouseButtonLeft:
		switch action {
		case glfw.Press:
			if s.ButtonsSwapped {
				s.SecondaryDown = true
			} else {
				s.PrimaryDown = true
			}
		case glfw.Release:
			if s.ButtonsSwapped {
				s.SecondaryDown = false
			} else {
				s.PrimaryDown = false
			}
		}
	case glfw.MouseButtonRight:
		switch action {
		case glfw.Press:
			if s.ButtonsSwapped {
				s.PrimaryDown = true
			} else {
				s.SecondaryDown = true
			}
		case glfw.Release:
			if s.ButtonsSwapped {
				s.PrimaryDown = false
			} else {
				s.SecondaryDown = false
			}
		}
	}
}
