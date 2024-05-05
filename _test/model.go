package _test

import (
	"image/color"
	"tonysoft.com/gfx"
)

/******************************************************************************
 Model/Mesh/Face/Material
******************************************************************************/

type Model struct {
	gfx.ModelBase

	vertices   []float32
	colors     []float32
	uvs        []float32
	normals    []float32
	tangents   []float32
	bitangents []float32

	meshes []*Mesh
}

type Mesh struct {
	gfx.MeshBase
	faces []*Face
}

type Face struct {
	gfx.FaceBase

	vertices   []int
	colors     []int
	uvs        []int
	normals    []int
	tangents   []int
	bitangents []int

	material gfx.Material
}

type Material struct {
	gfx.MaterialBase
	DiffuseMap gfx.Texture
}

/******************************************************************************
 gfx.Model/gfx.Mesh/gfx.Face Implementation
******************************************************************************/

func (m *Model) Vertices() []float32 {
	return m.vertices
}

func (m *Model) Colors() []float32 {
	return m.colors
}

func (m *Model) UVs() []float32 {
	return m.uvs
}

func (m *Model) Normals() []float32 {
	return m.normals
}

func (m *Model) Tangents() []float32 {
	return m.tangents
}

func (m *Model) Bitangents() []float32 {
	return m.bitangents
}

func (m *Model) Meshes() []gfx.Mesh {
	meshes := make([]gfx.Mesh, len(m.meshes))
	for i, mesh := range m.meshes {
		meshes[i] = mesh
	}
	return meshes
}

func (m *Mesh) Faces() []gfx.Face {
	faces := make([]gfx.Face, len(m.faces))
	for i, f := range m.faces {
		faces[i] = f
	}
	return faces
}

func (f *Face) VertexIndices() []int {
	return f.vertices
}

func (f *Face) ColorIndices() []int {
	return f.colors
}

func (f *Face) UvIndices() []int {
	return f.uvs
}

func (f *Face) NormalIndices() []int {
	return f.normals
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
 New Model Functions
******************************************************************************/

func NewColoredQuad(rgba ...color.RGBA) *Model {
	for len(rgba) < 4 {
		rgba = append(rgba, gfx.Magenta)
	}

	material := &Material{}

	face0 := &Face{
		vertices: []int{0, 1, 2},
		colors:   []int{0, 1, 2},
		material: material,
	}
	face1 := &Face{
		vertices: []int{2, 3, 0},
		colors:   []int{2, 3, 0},
		material: material,
	}

	mesh := &Mesh{
		faces: []*Face{face0, face1},
	}

	model := &Model{
		vertices: []float32{
			0, 0, 0,
			1, 0, 0,
			1, 1, 0,
			0, 1, 0,
		},
		colors: []float32{
			float32(rgba[0].R) / 255.0, float32(rgba[0].G) / 255.0, float32(rgba[0].B) / 255.0,
			float32(rgba[1].R) / 255.0, float32(rgba[1].G) / 255.0, float32(rgba[1].B) / 255.0,
			float32(rgba[2].R) / 255.0, float32(rgba[2].G) / 255.0, float32(rgba[2].B) / 255.0,
			float32(rgba[3].R) / 255.0, float32(rgba[3].G) / 255.0, float32(rgba[3].B) / 255.0,
		},
		meshes: []*Mesh{mesh},
	}

	return model
}

func NewTexturedQuad() *Model {
	material := &Material{}

	face0 := &Face{
		vertices: []int{0, 1, 2},
		uvs:      []int{0, 1, 2},
		material: material,
	}
	face1 := &Face{
		vertices: []int{2, 3, 0},
		uvs:      []int{2, 3, 0},
		material: material,
	}

	mesh := &Mesh{
		faces: []*Face{face0, face1},
	}

	model := &Model{
		vertices: []float32{
			0, 0, 0,
			1, 0, 0,
			1, 1, 0,
			0, 1, 0,
		},
		uvs: []float32{
			0, 0,
			1, 0,
			1, 1,
			0, 1,
		},
		meshes: []*Mesh{mesh},
	}

	return model
}
