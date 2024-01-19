package gfx

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"sync"
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
	x, y                float64
	leftDown, rightDown bool
	stateMutex          sync.Mutex
}

func (s *MouseState) Get() (x, y float64, leftIsDown, rightIsDown bool) {
	s.stateMutex.Lock()
	x = s.x
	y = s.y
	leftIsDown = s.leftDown
	rightIsDown = s.rightDown
	s.stateMutex.Unlock()
	return
}

func (s *MouseState) UpdatePosition(x, y float64) {
	s.stateMutex.Lock()
	s.x = x
	s.y = y
	s.stateMutex.Unlock()
}

func (s *MouseState) UpdateButton(buttonIndex int, isDown bool) {
	s.stateMutex.Lock()
	switch buttonIndex {
	case 0:
		s.leftDown = isDown
	case 1:
		s.rightDown = isDown
	default:
		panic("invalid mouse button index (expecting 0 or 1)")
	}
	s.stateMutex.Unlock()
}
