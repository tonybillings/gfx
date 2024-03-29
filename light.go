package gfx

import (
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"sync"
)

/******************************************************************************
 Light
******************************************************************************/

type Light struct {
	direction mgl32.Vec3
	color     mgl32.Vec3

	stateMutex sync.Mutex
}

/******************************************************************************
 Light Functions
******************************************************************************/

func (l *Light) Direction() mgl32.Vec3 {
	l.stateMutex.Lock()
	direction := l.direction
	l.stateMutex.Unlock()
	return direction
}

func (l *Light) SetDirection(direction mgl32.Vec3) *Light {
	l.stateMutex.Lock()
	l.direction = direction
	l.stateMutex.Unlock()
	return l
}

func (l *Light) SetDirectionX(x float32) *Light {
	l.stateMutex.Lock()
	l.direction[0] = x
	l.stateMutex.Unlock()
	return l
}

func (l *Light) SetDirectionY(y float32) *Light {
	l.stateMutex.Lock()
	l.direction[1] = y
	l.stateMutex.Unlock()
	return l
}

func (l *Light) SetDirectionZ(z float32) *Light {
	l.stateMutex.Lock()
	l.direction[2] = z
	l.stateMutex.Unlock()
	return l
}

func (l *Light) Color() mgl32.Vec3 {
	l.stateMutex.Lock()
	c := mgl32.Vec3{l.color[0], l.color[1], l.color[2]}
	l.stateMutex.Unlock()
	return c
}

func (l *Light) SetColor(rgb color.RGBA) *Light {
	l.stateMutex.Lock()
	l.color[0] = float32(rgb.R) / 255.0
	l.color[1] = float32(rgb.G) / 255.0
	l.color[2] = float32(rgb.B) / 255.0
	l.stateMutex.Unlock()
	return l
}

/******************************************************************************
 New Light Function
******************************************************************************/

func NewLight() *Light {
	return &Light{
		direction: mgl32.Vec3{0, -1, 0},
		color:     mgl32.Vec3{1, 1, 1},
	}
}
