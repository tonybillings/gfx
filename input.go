package gfx

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

/******************************************************************************
 Keyboard
******************************************************************************/

type KeyEvent struct {
	Key    glfw.Key
	Action glfw.Action
}

func (e *KeyEvent) Event() uint64 {
	return (uint64(e.Key) << 32) | uint64(e.Action)
}

type KeyEventHandler struct {
	Receiver any
	Key      glfw.Key
	Action   glfw.Action
	Callback func(*Window, glfw.Key, glfw.Action)
}

func (h *KeyEventHandler) Event() uint64 {
	return (uint64(h.Key) << 32) | uint64(h.Action)
}

/******************************************************************************
 Mouse
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

/******************************************************************************
 MouseSurface
******************************************************************************/

type MouseSurface interface {
	Mouse() *MouseState
	Width() int
	Height() int
}
