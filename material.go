package gfx

import (
	"sync"
)

/******************************************************************************
 Material
******************************************************************************/

type Material interface {
	Asset
	sync.Locker

	AttachShader(shader Shader)
	AttachedShader() Shader
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
