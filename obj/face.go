package obj

import "github.com/tonybillings/gfx"

/******************************************************************************
 Face
******************************************************************************/

type Face struct {
	gfx.FaceBase

	usemtl string

	vertices   []int
	normals    []int
	uvs        []int
	tangents   []int
	bitangents []int

	material *BasicMaterial
}

/******************************************************************************
 gfx.Face Implementation
******************************************************************************/

func (f *Face) VertexIndices() []int {
	return f.vertices
}

func (f *Face) NormalIndices() []int {
	return f.normals
}

func (f *Face) UvIndices() []int {
	return f.uvs
}

func (f *Face) TangentIndices() []int {
	return f.tangents
}

func (f *Face) BitangentIndices() []int {
	return f.bitangents
}

func (f *Face) AttachedMaterial() gfx.Material {
	return f.material
}

/******************************************************************************
 gfx.Initer Implementation
******************************************************************************/

func (f *Face) Init() bool {
	if f.material != nil {
		return f.material.Init()
	}

	return true
}

/******************************************************************************
 gfx.Closer Implementation
******************************************************************************/

func (f *Face) Close() {
	if f.material != nil {
		f.material.Close()
	}
}
