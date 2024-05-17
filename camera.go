package gfx

import (
	"github.com/go-gl/mathgl/mgl32"
	"sync"
)

/******************************************************************************
 Camera
******************************************************************************/

// Camera Use cameras to apply the same view-projection matrix to the world
// matrix of each Shape3D object to which the camera is assigned, ensuring
// they are all rendered from the same perspective.  Must be added to a
// window to ensure its properties get updated over time.
type Camera interface {
	Object
	sync.Locker

	// Location shall return the camera's coordinates in world space,
	// using a 4-byte-aligned float array.
	Location() mgl32.Vec4

	// View shall return the current view matrix for the camera.
	View() mgl32.Mat4

	// Projection shall return the current projection matrix for the camera.
	Projection() mgl32.Mat4

	// ViewProjection shall return the current combined view-projection matrix
	// for the camera.  Passing this matrix to the shader rather than computing
	// it there may increase performance.
	ViewProjection() mgl32.Mat4
}

/******************************************************************************
 CameraBase
******************************************************************************/

type CameraBase struct {
	ObjectBase
	stateMutex sync.Mutex
}

/******************************************************************************
 sync.Locker Implementation
******************************************************************************/

func (c *CameraBase) Lock() {
	c.stateMutex.Lock()
}

func (c *CameraBase) Unlock() {
	c.stateMutex.Unlock()
}

/******************************************************************************
 BasicCamera
******************************************************************************/

type BasicCamera struct {
	CameraBase

	verticalFoV float32
	near, far   float32
	projection  mgl32.Mat4

	Properties *BasicCameraProperties
}

type BasicCameraProperties struct {
	Position    mgl32.Vec4
	Target      mgl32.Vec4
	Up          mgl32.Vec4
	ViewProjMat mgl32.Mat4
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (c *BasicCamera) Update(_ int64) (ok bool) {
	c.stateMutex.Lock()
	c.Properties.ViewProjMat = c.projection.Mul4(mgl32.LookAtV(
		mgl32.Vec3{c.Properties.Position[0], c.Properties.Position[1], c.Properties.Position[2]},
		mgl32.Vec3{c.Properties.Target[0], c.Properties.Target[1], c.Properties.Target[2]},
		mgl32.Vec3{c.Properties.Up[0], c.Properties.Up[1], c.Properties.Up[2]}))
	c.stateMutex.Unlock()
	return true
}

/******************************************************************************
 Resizer Implementation
******************************************************************************/

func (c *BasicCamera) Resize(_, _, newWidth, newHeight int32) {
	c.stateMutex.Lock()
	c.projection = mgl32.Perspective(mgl32.DegToRad(c.verticalFoV), float32(newWidth)/float32(newHeight), c.near, c.far)
	c.stateMutex.Unlock()
}

/******************************************************************************
 Camera Implementation
******************************************************************************/

func (c *BasicCamera) Location() (loc mgl32.Vec4) {
	c.stateMutex.Lock()
	loc = c.Properties.Position
	c.stateMutex.Unlock()
	return
}

func (c *BasicCamera) View() (view mgl32.Mat4) {
	c.stateMutex.Lock()
	view = mgl32.LookAtV(mgl32.Vec3{c.Properties.Position[0], c.Properties.Position[1], c.Properties.Position[2]},
		mgl32.Vec3{c.Properties.Target[0], c.Properties.Target[1], c.Properties.Target[2]},
		mgl32.Vec3{c.Properties.Up[0], c.Properties.Up[1], c.Properties.Up[2]})
	c.stateMutex.Unlock()
	return
}

func (c *BasicCamera) Projection() (proj mgl32.Mat4) {
	c.stateMutex.Lock()
	proj = c.projection
	c.stateMutex.Unlock()
	return
}

func (c *BasicCamera) ViewProjection() (viewProj mgl32.Mat4) {
	c.stateMutex.Lock()
	viewProj = c.projection.Mul4(mgl32.LookAtV(
		mgl32.Vec3{c.Properties.Position[0], c.Properties.Position[1], c.Properties.Position[2]},
		mgl32.Vec3{c.Properties.Target[0], c.Properties.Target[1], c.Properties.Target[2]},
		mgl32.Vec3{c.Properties.Up[0], c.Properties.Up[1], c.Properties.Up[2]}))
	c.stateMutex.Unlock()
	return
}

/******************************************************************************
 BasicCamera Functions
******************************************************************************/

func (c *BasicCamera) SetProjection(verticalFoV, aspectRatio, near, far float32) mgl32.Mat4 {
	c.stateMutex.Lock()
	c.verticalFoV = verticalFoV
	c.near = near
	c.far = far
	proj := mgl32.Perspective(mgl32.DegToRad(verticalFoV), aspectRatio, near, far)
	c.projection = proj
	c.stateMutex.Unlock()
	return proj
}

/******************************************************************************
 New BasicCamera Function
******************************************************************************/

func NewCamera() *BasicCamera {
	c := &BasicCamera{
		Properties: &BasicCameraProperties{
			Position: mgl32.Vec4{0, 0, -1},
			Target:   mgl32.Vec4{0, 0, 0},
			Up:       mgl32.Vec4{0, 1, 0},
		},
	}

	c.SetProjection(45.0, 16.0/9.0, 0.1, 1000.0)
	return c
}
