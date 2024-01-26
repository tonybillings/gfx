package gfx

import (
	"github.com/go-gl/mathgl/mgl32"
	"sync"
)

/******************************************************************************
 Camera
******************************************************************************/

type Camera struct {
	position    mgl32.Vec3
	target      mgl32.Vec3
	up          mgl32.Vec3
	fovY        float32
	aspectRatio float32
	near        float32
	far         float32

	stateMutex sync.Mutex
}

/******************************************************************************
 Camera Functions
******************************************************************************/

func (c *Camera) View() mgl32.Mat4 {
	c.stateMutex.Lock()
	view := mgl32.LookAtV(c.position, c.target, c.up)
	c.stateMutex.Unlock()
	return view
}

func (c *Camera) Position() mgl32.Vec3 {
	c.stateMutex.Lock()
	pos := mgl32.Vec3{c.position[0], c.position[1], c.position[2]}
	c.stateMutex.Unlock()
	return pos
}

func (c *Camera) SetPosition(position mgl32.Vec3) *Camera {
	c.stateMutex.Lock()
	c.position = mgl32.Vec3{position[0], position[1], position[2]}
	c.stateMutex.Unlock()
	return c
}

func (c *Camera) Target() mgl32.Vec3 {
	c.stateMutex.Lock()
	target := mgl32.Vec3{c.target[0], c.target[1], c.target[2]}
	c.stateMutex.Unlock()
	return target
}

func (c *Camera) SetTarget(target mgl32.Vec3) *Camera {
	c.stateMutex.Lock()
	c.target = mgl32.Vec3{target[0], target[1], target[2]}
	c.stateMutex.Unlock()
	return c
}

func (c *Camera) Up() mgl32.Vec3 {
	c.stateMutex.Lock()
	up := mgl32.Vec3{c.up[0], c.up[1], c.up[2]}
	c.stateMutex.Unlock()
	return up
}

func (c *Camera) SetUp(up mgl32.Vec3) *Camera {
	c.stateMutex.Lock()
	c.up = mgl32.Vec3{up[0], up[1], up[2]}
	c.stateMutex.Unlock()
	return c
}

func (c *Camera) Projection() mgl32.Mat4 {
	c.stateMutex.Lock()
	proj := mgl32.Perspective(mgl32.DegToRad(c.fovY), c.aspectRatio, c.near, c.far)
	c.stateMutex.Unlock()
	return proj
}

func (c *Camera) SetProjection(fovY, aspectRatio, near, far float32) *Camera {
	c.stateMutex.Lock()
	c.fovY = fovY
	c.aspectRatio = aspectRatio
	c.near = near
	c.far = far
	c.stateMutex.Unlock()
	return c
}

/******************************************************************************
 New Camera Function
******************************************************************************/

func NewCamera() *Camera {
	return &Camera{
		position:    mgl32.Vec3{0, 0, -1},
		target:      mgl32.Vec3{0, 0, 0},
		up:          mgl32.Vec3{0, 1, 0},
		fovY:        45,
		aspectRatio: 16 / 9,
		near:        .5,
		far:         1000,
	}
}
