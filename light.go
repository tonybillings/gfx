package gfx

import (
	"github.com/go-gl/mathgl/mgl32"
	"sync"
)

/******************************************************************************
 Light
******************************************************************************/

type Light interface {
	sync.Locker

	Name() string
	SetName(string)

	Enabled() bool
	SetEnabled(bool)
}

/******************************************************************************
 LightBase
******************************************************************************/

type LightBase struct {
	name       string
	enabled    bool
	stateMutex sync.Mutex
}

/******************************************************************************
 Light Implementation
******************************************************************************/

func (l *LightBase) Name() string {
	l.stateMutex.Lock()
	name := l.name
	l.stateMutex.Unlock()
	return name
}

func (l *LightBase) SetName(name string) {
	l.stateMutex.Lock()
	l.name = name
	l.stateMutex.Unlock()
}

func (l *LightBase) Enabled() bool {
	l.stateMutex.Lock()
	enabled := l.enabled
	l.stateMutex.Unlock()
	return enabled
}

func (l *LightBase) SetEnabled(enabled bool) {
	l.stateMutex.Lock()
	l.enabled = enabled
	l.stateMutex.Unlock()
}

/******************************************************************************
 sync.Locker Implementation
******************************************************************************/

func (l *LightBase) Lock() {
	l.stateMutex.Lock()
}

func (l *LightBase) Unlock() {
	l.stateMutex.Unlock()
}

/******************************************************************************
 DirectionalLight
******************************************************************************/

type DirectionalLight struct {
	LightBase

	Color     mgl32.Vec3
	Direction mgl32.Vec3
}

/******************************************************************************
 New DirectionalLight Function
******************************************************************************/

func NewDirectionalLight() *DirectionalLight {
	return &DirectionalLight{
		Color:     mgl32.Vec3{1, 1, 1},
		Direction: mgl32.Vec3{0, -1, 0},
	}
}

/******************************************************************************
 QuadDirectionalLighting
******************************************************************************/

type QuadDirectionalLighting struct {
	stateMutex sync.Mutex

	Lights     [4]DirectionalLight
	LightCount int32
}

func (l *QuadDirectionalLighting) Lock() {
	l.stateMutex.Lock()
}

func (l *QuadDirectionalLighting) Unlock() {
	l.stateMutex.Unlock()
}

func NewQuadDirectionalLighting() *QuadDirectionalLighting {
	return &QuadDirectionalLighting{}
}
