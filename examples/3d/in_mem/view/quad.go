package view

import (
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"time"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/examples/3d/in_mem/textures"
)

type Model struct {
	gfx.ModelBase
	vertices []float32
	normals  []float32
	uvs      []float32
	meshes   []*Mesh
}

type Mesh struct {
	gfx.MeshBase
	faces []*Face
}

type Face struct {
	gfx.FaceBase
	vertexIndices []int
	normalIndices []int
	uvIndices     []int
	material      gfx.Material
}

type BasicMaterial struct {
	gfx.MaterialBase
	Time       float32
	WaveFactor float32
	DiffuseMap gfx.Texture
	Properties *MaterialProperties
}

type MaterialProperties struct {
	Ambient mgl32.Vec4
	Diffuse mgl32.Vec4
}

var vertShader = `#version 410 core
in vec3 a_Position;
in vec3 a_Normal;
in vec2 a_UV;

out vec3 Normal;
out vec2 UV;

void main() {
	Normal = a_Normal;
	UV = a_UV;
  	gl_Position = vec4(a_Position, 1.0);
}
`

var fragShader = `#version 410 core
in vec3 Normal;
in vec2 UV;

out vec4 FragColor;

uniform sampler2D u_DiffuseMap;
uniform sampler2D u_DiffuseMap2;
uniform float u_Time;
uniform float u_WaveFactor;

layout (std140) uniform BasicMaterial {
    vec4 Ambient;
    vec4 Diffuse;
};

void main() {
  	vec4 diffuseMap = texture(u_DiffuseMap, UV);
	FragColor = Diffuse * texture(u_DiffuseMap, vec2(UV.x + sin(UV.y * 100.0 + u_Time) / u_WaveFactor, UV.y));
}
`

func (m *Model) Vertices() []float32 {
	return m.vertices
}

func (m *Model) Normals() []float32 {
	return m.normals
}

func (m *Model) UVs() []float32 {
	return m.uvs
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
	return f.vertexIndices
}

func (f *Face) NormalIndices() []int {
	return f.normalIndices
}

func (f *Face) UvIndices() []int {
	return f.uvIndices
}

func (f *Face) AttachedMaterial() gfx.Material {
	return f.material
}

func NewQuadView(window *gfx.Window) gfx.WindowObject {
	gfx.Assets.AddEmbeddedFiles(textures.Assets)

	shader := gfx.NewBasicShader("test", vertShader, fragShader)
	gfx.Assets.Add(shader)

	cubetxt2d := gfx.NewTexture2D("cube.t2d", "cube.png")
	gfx.Assets.Add(cubetxt2d)

	material := &BasicMaterial{
		Properties: &MaterialProperties{
			Ambient: mgl32.Vec4{0, 0, .5, 0},
			Diffuse: mgl32.Vec4{0.0001, .5, .5, 1},
		},
		DiffuseMap: cubetxt2d,
	}
	material.AttachShader(shader)

	face0 := &Face{
		vertexIndices: []int{0, 1, 2},
		normalIndices: []int{1, 1, 0},
		uvIndices:     []int{0, 1, 2},
		material:      material,
	}
	face1 := &Face{
		vertexIndices: []int{2, 3, 0},
		normalIndices: []int{0, 0, 1},
		uvIndices:     []int{2, 3, 0},
		material:      material,
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
		normals: []float32{
			0, -1, 0,
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

	quad := gfx.NewShape3D()
	quad.SetModel(model)

	vp := gfx.NewViewport(window.Width(), window.Height())
	vp.Set(-1, -1, 2, 2)
	quad.SetViewport(vp)

	go func() {
		for {
			material.Lock()
			material.Properties.Ambient[2] += 0.001
			material.Properties.Diffuse[0] += 0.001
			material.Time += .004
			material.WaveFactor = float32(1000 - (math.Sin(float64(material.Time*2)) * 999))
			material.Unlock()
			time.Sleep(4 * time.Millisecond)
		}
	}()

	return quad
}
