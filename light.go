package gfx

import (
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"sync"
)

type Light struct {
	direction mgl32.Vec3
	color     mgl32.Vec3

	stateMutex sync.Mutex
}

func (l *Light) Direction() mgl32.Vec3 {
	l.stateMutex.Lock()
	direction := mgl32.Vec3{l.direction[0], l.direction[1], l.direction[2]}
	l.stateMutex.Unlock()
	return direction
}

func (l *Light) SetDirection(direction mgl32.Vec3) *Light {
	direction = direction.Normalize()
	l.stateMutex.Lock()
	l.direction[0] = direction[0]
	l.direction[1] = direction[1]
	l.direction[2] = direction[2]
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

func NewLight() *Light {
	return &Light{
		direction: mgl32.Vec3{0, -1, 0},
		color:     mgl32.Vec3{1, 1, 1},
	}
}
