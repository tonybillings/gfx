package gfx

import "sync"

/******************************************************************************
 Viewport
******************************************************************************/

type Viewport struct {
	xNorm, yNorm float32
	wNorm, hNorm float32

	x, y int32
	w, h int32

	winW, winH int32

	stateMutex sync.Mutex
}

/******************************************************************************
 Viewport Functions
******************************************************************************/

func (v *Viewport) update() {
	v.x = int32(v.xNorm * float32(v.winW))
	v.y = int32(v.yNorm * float32(v.winH))
	v.w = int32(v.wNorm * float32(v.winW))
	v.h = int32(v.hNorm * float32(v.winH))
}

func (v *Viewport) Set(x, y, width, height float32) *Viewport {
	v.stateMutex.Lock()
	v.xNorm = x
	v.yNorm = y
	v.wNorm = width
	v.hNorm = height
	v.update()
	v.stateMutex.Unlock()
	return v
}

func (v *Viewport) SetWindowSize(width, height int32) *Viewport {
	v.stateMutex.Lock()
	v.winW = width
	v.winH = height
	v.update()
	v.stateMutex.Unlock()
	return v
}

func (v *Viewport) Get() (x, y, width, height int32) {
	v.stateMutex.Lock()
	x = v.x
	y = v.y
	width = v.w
	height = v.h
	v.stateMutex.Unlock()
	return
}

/******************************************************************************
 New Viewport Function
******************************************************************************/

func NewViewport(windowWidth, windowHeight int32) *Viewport {
	v := &Viewport{
		xNorm: 0.0,
		yNorm: 0.0,
		wNorm: 1.0,
		hNorm: 1.0,
	}
	v.SetWindowSize(windowWidth, windowHeight)
	return v
}
