package gfx

import (
	"sync/atomic"
)

/******************************************************************************
 Service
******************************************************************************/

// Service objects are initialized/updated before any other type of Object,
// are meant to have a life cycle that lives beyond the other objects added
// to a window, and can be used by other objects for various purposes.  The
// AssetLibrary type is an example implementation of Service.
type Service interface {
	Object

	// Window shall return the Window to which this service was assigned.
	Window() *Window
	SetWindow(*Window) Service

	// Protected shall return true if this service has been marked
	// for protection from closure/removal after being added to a
	// Window. It is used, for example, to allow consumers to freely
	// call DisposeAllServices() on the Window without concern that it
	// will affect the default services that a Window comes with.
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
