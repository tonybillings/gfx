package gfx

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

const (
	sizeOfFloat32 = 4 // byte count
)

type VertexAttributeLayout int

// Expected names for the vertex attributes defined in the shaders noted below:
const (
	PositionOnlyVaoLayout             VertexAttributeLayout = iota // a_Position
	PositionColorVaoLayout                                         // a_Position, a_Color
	PositionUvVaoLayout                                            // a_Position, a_UV
	PositionNormalUvVaoLayout                                      // a_Position, a_Normal, a_UV
	PositionNormalUvTangentsVaoLayout                              // a_Position, a_Normal, a_UV, a_Tangent, a_Bitangent
)

func newVertexArrayObject(layout VertexAttributeLayout, shader Shader, vertices []float32) (glName uint32, closeFunc func()) {
	if shader == nil || !shader.Initialized() {
		panic("shader cannot be nil or uninitialized")
	}

	if vertices == nil || len(vertices) == 0 {
		panic("vertices cannot be nil or zero-length")
	}

	vao, vbo := uint32(0), uint32(0)

	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	stride := int32(0)

	posLoc := uint32(shader.GetAttribLocation("a_Position"))
	colorLoc := uint32(shader.GetAttribLocation("a_Color"))
	uvLoc := uint32(shader.GetAttribLocation("a_UV"))
	normLoc := uint32(shader.GetAttribLocation("a_Normal"))
	tanLoc := uint32(shader.GetAttribLocation("a_Tangent"))
	bitanLoc := uint32(shader.GetAttribLocation("a_Bitangent"))

	switch layout {
	case PositionOnlyVaoLayout:
		gl.EnableVertexAttribArray(posLoc)
		gl.VertexAttribPointerWithOffset(posLoc, 3, gl.FLOAT, false, stride, 0)
	case PositionColorVaoLayout:
		stride = 24
		gl.EnableVertexAttribArray(posLoc)
		gl.VertexAttribPointerWithOffset(posLoc, 3, gl.FLOAT, false, stride, 0)
		gl.EnableVertexAttribArray(colorLoc)
		gl.VertexAttribPointerWithOffset(colorLoc, 3, gl.FLOAT, false, stride, uintptr(3*sizeOfFloat32))
	case PositionUvVaoLayout:
		stride = 20
		gl.EnableVertexAttribArray(posLoc)
		gl.VertexAttribPointerWithOffset(posLoc, 3, gl.FLOAT, false, stride, 0)
		gl.EnableVertexAttribArray(uvLoc)
		gl.VertexAttribPointerWithOffset(uvLoc, 2, gl.FLOAT, false, stride, uintptr(3*sizeOfFloat32))
	case PositionNormalUvVaoLayout:
		stride = 32
		gl.EnableVertexAttribArray(posLoc)
		gl.VertexAttribPointerWithOffset(posLoc, 3, gl.FLOAT, false, stride, 0)
		gl.EnableVertexAttribArray(normLoc)
		gl.VertexAttribPointerWithOffset(normLoc, 3, gl.FLOAT, false, stride, uintptr(3*sizeOfFloat32))
		gl.EnableVertexAttribArray(uvLoc)
		gl.VertexAttribPointerWithOffset(uvLoc, 2, gl.FLOAT, false, stride, uintptr(6*sizeOfFloat32))
	case PositionNormalUvTangentsVaoLayout:
		stride = 56
		gl.EnableVertexAttribArray(posLoc)
		gl.VertexAttribPointerWithOffset(posLoc, 3, gl.FLOAT, false, stride, 0)
		gl.EnableVertexAttribArray(normLoc)
		gl.VertexAttribPointerWithOffset(normLoc, 3, gl.FLOAT, false, stride, uintptr(3*sizeOfFloat32))
		gl.EnableVertexAttribArray(uvLoc)
		gl.VertexAttribPointerWithOffset(uvLoc, 2, gl.FLOAT, false, stride, uintptr(6*sizeOfFloat32))
		gl.EnableVertexAttribArray(tanLoc)
		gl.VertexAttribPointerWithOffset(tanLoc, 3, gl.FLOAT, false, stride, uintptr(8*sizeOfFloat32))
		gl.EnableVertexAttribArray(bitanLoc)
		gl.VertexAttribPointerWithOffset(bitanLoc, 3, gl.FLOAT, false, stride, uintptr(11*sizeOfFloat32))
	}

	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*sizeOfFloat32, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	glName = vao
	closeFunc = func() {
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.DeleteBuffers(1, &vbo)

		gl.BindVertexArray(0)
		gl.DeleteVertexArrays(1, &vao)
	}

	return
}
