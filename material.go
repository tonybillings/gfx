package gfx

import (
	"sync"
)

/******************************************************************************
 Material
******************************************************************************/

// Material instances are used in the rendering of faces, providing various
// properties that will be used by the attached shader.  Before these
// properties are changed, it is recommended to call Lock() on the material
// to ensure the shader is not being sent their values while they are being
// changed; and of course, call Unlock() when finished changing them.
type Material interface {
	Asset
	sync.Locker

	// AttachedShader shall return the shader that this material will bind
	// to, sending it the property values and texture maps that constitute
	// the material.
	AttachedShader() Shader
	AttachShader(shader Shader)
}

/******************************************************************************
 MaterialBase
******************************************************************************/

type MaterialBase struct {
	AssetBase

	shader     Shader
	stateMutex sync.Mutex
}

/******************************************************************************
 Material Implementation
******************************************************************************/

func (m *MaterialBase) AttachShader(shader Shader) {
	m.stateMutex.Lock()
	m.shader = shader
	m.stateMutex.Unlock()
}

func (m *MaterialBase) AttachedShader() Shader {
	m.stateMutex.Lock()
	shader := m.shader
	m.stateMutex.Unlock()
	return shader
}

/******************************************************************************
 sync.Locker Implementation
******************************************************************************/

func (m *MaterialBase) Lock() {
	m.stateMutex.Lock()
}

func (m *MaterialBase) Unlock() {
	m.stateMutex.Unlock()
}
