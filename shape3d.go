package gfx

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

const (
	defaultShape3DName = "Shape3D"
)

/******************************************************************************
 Shape3D
******************************************************************************/

type Shape3D struct {
	WindowObjectBase

	viewport *Viewport
	camera   Camera
	lighting any

	modelAsset    Model
	modelInstance *modelInstance
	modelRenderer *modelRenderer

	cameraChanged   bool
	lightingChanged bool

	viewportBak [4]int32
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (s *Shape3D) Init() (ok bool) {
	if s.Initialized() {
		return true
	}

	s.initViewport()
	s.initModel()

	return s.WindowObjectBase.Init()
}

func (s *Shape3D) Close() {
	if !s.Initialized() {
		return
	}

	if s.modelInstance != nil {
		s.modelInstance.close()
	}

	if s.modelRenderer != nil {
		s.modelRenderer.close()
	}

	s.WindowObjectBase.Close()
}

/******************************************************************************
 DrawableObject Implementation
******************************************************************************/

func (s *Shape3D) Draw(deltaTime int64) (ok bool) {
	if !s.DrawableObjectBase.Draw(deltaTime) {
		return false
	}

	s.beginDraw()
	s.updateScene()
	s.draw()
	s.endDraw()

	return s.WindowObjectBase.drawChildren(deltaTime)
}

/******************************************************************************
 Resizer Implementation
******************************************************************************/

func (s *Shape3D) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	if s.viewport != nil {
		s.viewport.SetWindowSize(newWidth, newHeight)
	}

	s.WindowObjectBase.Resize(oldWidth, oldHeight, newWidth, newHeight)
}

/******************************************************************************
 Shape3D Functions
******************************************************************************/

func (s *Shape3D) initViewport() {
	if s.viewport == nil {
		s.viewport = NewViewport(s.window.Width(), s.window.Height())
	}
}

func (s *Shape3D) initModel() {
	s.modelInstance = newModelInstance(s.modelAsset, s)
	s.modelRenderer = newModelRenderer(s.modelInstance)
}

func (s *Shape3D) beginDraw() {
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.GetIntegerv(gl.VIEWPORT, &s.viewportBak[0])
}

func (s *Shape3D) updateScene() {
	s.stateMutex.Lock()

	gl.Viewport(s.viewport.Get())

	if s.cameraChanged {
		s.cameraChanged = false
		s.modelRenderer.setCamera(s.camera)
	}

	if s.lightingChanged {
		s.lightingChanged = false
		s.modelRenderer.setLighting(s.lighting)
	}

	s.stateMutex.Unlock()
}

func (s *Shape3D) draw() {
	s.modelRenderer.render()
}

func (s *Shape3D) endDraw() {
	gl.Viewport(s.viewportBak[0], s.viewportBak[1], s.viewportBak[2], s.viewportBak[3])

	gl.Disable(gl.BLEND)
	gl.Disable(gl.DEPTH_TEST)

	gl.BindVertexArray(0)

	gl.UseProgram(0)
}

func (s *Shape3D) Viewport() *Viewport {
	s.stateMutex.Lock()
	vp := s.viewport
	s.stateMutex.Unlock()
	return vp
}

func (s *Shape3D) SetViewport(viewport *Viewport) *Shape3D {
	s.stateMutex.Lock()
	s.viewport = viewport
	s.stateMutex.Unlock()
	return s
}

func (s *Shape3D) Camera() Camera {
	s.stateMutex.Lock()
	cam := s.camera
	s.stateMutex.Unlock()
	return cam
}

func (s *Shape3D) SetCamera(camera Camera) *Shape3D {
	s.stateMutex.Lock()
	s.camera = camera
	s.cameraChanged = true
	s.stateMutex.Unlock()
	return s
}

func (s *Shape3D) Lighting() any {
	s.stateMutex.Lock()
	lighting := s.lighting
	s.stateMutex.Unlock()
	return lighting
}

func (s *Shape3D) SetLighting(lighting any) *Shape3D {
	s.stateMutex.Lock()
	s.lighting = lighting
	s.lightingChanged = true
	s.stateMutex.Unlock()
	return s
}

func (s *Shape3D) SetModel(model Model) *Shape3D {
	s.stateMutex.Lock()
	s.modelAsset = model
	s.stateMutex.Unlock()
	return s
}

func (s *Shape3D) Meshes() []*meshInstance {
	if s.modelInstance == nil {
		return nil
	}
	s.stateMutex.Lock()
	meshes := s.modelInstance.Meshes()
	s.stateMutex.Unlock()
	return meshes
}

/******************************************************************************
 New Shape3D Function
******************************************************************************/

func NewShape3D() *Shape3D {
	m := &Shape3D{
		WindowObjectBase: *NewWindowObject(nil),
	}

	m.SetName(defaultShape3DName)
	return m
}
