package view

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tonybillings/gfx"
	"github.com/tonybillings/gfx/examples/3d/in_mem/textures"
	"math"
	"time"
)

/******************************************************************************
 1. Create structs that implement Model/Mesh/Face

 Embedding the "Base" structs means we only have to define funcs
 for the combination of vertex attributes that the model will support.
 For example, it can contain just vertex position and color
 information or, like in the example that follows, it can be meant
 to hold texture coordinates instead of hard-coded colors. The gfx.Model
 interface also embeds the gfx.Asset interface; i.e., Models are Assets.
 gfx.ModelBase takes care of that concern for us as it embeds gfx.AssetBase.
******************************************************************************/

type Model struct {
	gfx.ModelBase
	vertices []float32
	uvs      []float32
	meshes   []*Mesh
}

type Mesh struct { // Mesh's must also implement gfx.Transform...
	gfx.MeshBase // ...and this takes care of that concern for us
	faces        []*Face
}

type Face struct {
	gfx.FaceBase
	vertexIndices []int
	uvIndices     []int
	material      gfx.Material
}

/******************************************************************************
 2. Create a struct that implements Material

 Here, we're creating a shader-bindable struct, which just means that it
 contains fields that can be "bound" to uniform variables/buffers defined in
 a shader program. These bindings make it easy to send data from system memory
 (RAM) to graphics card memory (VRAM). Fields meant to be bound must adhere to
 a set of rules defined in detail elsewhere, but for now know that they must
 be exported, be of a certain type, and be named exactly as they are in the
 shader, sans the "u_" and "a_" prefix for uniform variables and vertex
 attributes, respectively.
******************************************************************************/

type BasicMaterial struct {
	gfx.MaterialBase
	Time       float32             // binds to float
	WaveFactor float32             // binds to float
	DiffuseMap gfx.Texture         // binds to sampler2D
	Properties *MaterialProperties // pointers to structs are bound as UBOs
}

type MaterialProperties struct { // will be bound as a Uniform Buffer Object, std140 layout
	Ambient mgl32.Vec4 // binds to vec4
	Diffuse mgl32.Vec4 // binds to vec4
} // ensure the fields are 2/4/16-byte aligned as per std140...careful using Vec3!

/******************************************************************************
 3. Define the shader program
******************************************************************************/

var vertShader = `#version 410 core
// Vertex attributes can come in any order but must be  
// named a_Position, a_Color, a_UV, a_Normal, a_Tangent, or a_Bitangent
in vec3 a_Position;
in vec2 a_UV;

out vec2 UV;

void main() {
	UV = a_UV;
  	gl_Position = vec4(a_Position, 1.0);
}
`

var fragShader = `#version 410 core
in vec2 UV;

out vec4 FragColor;

// Uniforms (that you plan to bind to) must begin with u_
uniform sampler2D u_DiffuseMap; 
uniform float u_Time;
uniform float u_WaveFactor;

// For whole-struct binding (UBOs), use std140 layout.  The 
// given name must match the name of the bound struct field,
// unless that name is Properties, in which case the given 
// name here must match the name of the struct's type. In 
// our example, the bound struct field is named Properties,
// so the name here should be BasicMaterial since that's what
// we named the material struct defined above.
layout (std140) uniform BasicMaterial {  
    vec4 Ambient;
    vec4 Diffuse;
}; // you can also assign an instance name here, if you like

void main() {
  	vec4 diffuseMap = texture(u_DiffuseMap, UV);
	FragColor = Ambient + Diffuse * texture(u_DiffuseMap, vec2(UV.x + sin(UV.y * 100.0 + u_Time) / u_WaveFactor, UV.y));
}
`

/******************************************************************************
 4. Finish implementing Model/Mesh/Face
******************************************************************************/

func (m *Model) Vertices() []float32 {
	return m.vertices
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

func (f *Face) UvIndices() []int {
	return f.uvIndices
}

func (f *Face) AttachedMaterial() gfx.Material {
	return f.material
}

/******************************************************************************
 5. Create a fullscreen quad that will be textured and animated
******************************************************************************/

func NewQuadView(window *gfx.Window) gfx.WindowObject {
	// Create the Shader asset and add to AssetLibrary for life-cycle management
	shader := gfx.NewBasicShader("my_shader", vertShader, fragShader)
	window.Assets().Add(shader)

	// Create the texture from an on-disk PNG (this part is not very
	// "in mem[ory]" but textures can also be created from solid colors,
	// which can be helpful for use in shaders where the absence of
	// a proper texture must be handled gracefully/seamlessly, but here
	// we'll go for something more interesting to see).  First, we
	// load the PNG's raw binary data into RAM and store as an Asset:
	window.Assets().AddEmbeddedFiles(textures.Assets)
	// Next, we create a Texture asset using the filename of the PNG
	// asset as the source.  Because it exists in Window.Assets(), it
	// will be loaded from there (RAM); if it doesn't exist there,
	// then it will look for the PNG on the local file system (disk).
	// Texture assets, once initialized, are references to texture
	// data stored in graphics card memory (VRAM).
	emojiTexture := gfx.NewTexture2D("emoji.t2d", "emoji.png")
	// Finally, because Texture assets require participation in the
	// Init()/Close() life cycle, they should be managed by an
	// AssetLibrary, so here we'll just add to the default one:
	window.Assets().Add(emojiTexture)
	// Alternatively, you could manually initialize/close the texture
	// using the related InitObject()/CloseObject() functions the
	// Window struct provides (unless the object does nothing OpenGL-
	// related, you shouldn't directly invoke the life-cycle methods
	// Init()/Update()/Draw()/Close() that the object exposes).

	// Instantiate our Material, attach the diffuse map and shader.
	// Materials are also Assets but don't necessarily need to participate
	// in any life-cycle routine, as they are just data containers at
	// the very least (but "smart/dynamic" Materials could be imagined
	// that could require such participation).  So no need to add to
	// the Window's asset library...
	material := &BasicMaterial{
		Properties: &MaterialProperties{
			Ambient: mgl32.Vec4{.1, .1, .1, 1},
			Diffuse: mgl32.Vec4{0, .5, .5, 1},
		},
		DiffuseMap: emojiTexture,
	}
	material.AttachShader(shader)

	// Instantiate the two faces of the quad, CCW winding order
	face0 := &Face{
		vertexIndices: []int{0, 1, 2}, // like the OBJ format, faces contain only indices
		uvIndices:     []int{0, 1, 2},
		material:      material,
	}
	face1 := &Face{
		vertexIndices: []int{2, 3, 0},
		uvIndices:     []int{2, 3, 0},
		material:      material,
	}

	// Instantiate the mesh that will contain the faces
	mesh := &Mesh{
		faces: []*Face{face0, face1},
	}

	// Instantiate the Model.  Models are also Assets, but in this case
	// there is no need to add to the Window's asset library as there's
	// nothing to "initialize"	or "update", though you could of course
	// use that library simply to make the Asset available throughout the
	// application (but only for objects rendered on that Window instance).
	model := &Model{ // here we store the data referenced by the faces
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

	// Create the Shape3D object used to render the Model
	quad := gfx.NewShape3D()
	quad.SetModel(model) // and simply provide the Model asset

	// Normally, you would also provide a Camera for space
	// transformations, but in this example we'll deal strictly
	// with normalized device/screen-space coordinates.  Because
	// the vertices were defined in the positive X/Y quadrant
	// only, we can use a custom viewport to effectively stretch
	// and reposition the rendered quad such that it fills the
	// entire screen instead of just the upper-right quadrant.
	vp := gfx.NewViewport(window.Width(), window.Height())
	vp.Set(-1, -1, 2, 2)
	quad.SetViewport(vp)

	go func() {
		for {
			material.Lock() // gfx types that implement sync.Locker should be used like this...
			material.Properties.Ambient[2] += 0.001
			material.Properties.Diffuse[0] += 0.001
			material.Properties.Diffuse[1] -= 0.001
			material.Time += .004
			material.WaveFactor = float32(1000 - (math.Sin(float64(material.Time*2)) * 999))
			material.Unlock() // ...to ensure changes aren't made while sending data to VRAM
			time.Sleep(4 * time.Millisecond)
		}
	}()

	return quad
}
