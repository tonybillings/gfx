package gfx

import (
	"sync/atomic"
)

/******************************************************************************
 Service
******************************************************************************/

type Service interface {
	Object

	Window() *Window
	SetWindow(*Window) Service

	Protected() bool
	SetProtected(bool) Service
}

/******************************************************************************
 ServiceBase
******************************************************************************/

type ServiceBase struct {
	ObjectBase
	window    *Window
	protected atomic.Bool
}

/******************************************************************************
 Service Implementation
******************************************************************************/

func (s *ServiceBase) Window() *Window {
	return s.window
}

func (s *ServiceBase) SetWindow(window *Window) Service {
	s.window = window
	return s
}

func (s *ServiceBase) Protected() bool {
	return s.protected.Load()
}

func (s *ServiceBase) SetProtected(protected bool) Service {
	s.protected.Store(protected)
	return s
}
