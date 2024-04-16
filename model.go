package gfx

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	cameraUboBindPoint   = 5
	materialUboBindPoint = 6
	lightingUboBindPoint = 7
)

/******************************************************************************
 Model
******************************************************************************/

type Model interface {
	Asset

	Vertices() []float32
	Colors() []float32
	UVs() []float32
	Normals() []float32
	Tangents() []float32
	Bitangents() []float32

	Meshes() []Mesh
}

/******************************************************************************
 Mesh
******************************************************************************/

type Mesh interface {
	Transform

	Name() string
	Faces() []Face
}

/******************************************************************************
 Face
******************************************************************************/

type Face interface {
	VertexIndices() []int
	ColorIndices() []int
	UvIndices() []int
	NormalIndices() []int
	TangentIndices() []int
	BitangentIndices() []int

	AttachedMaterial() Material
}

/******************************************************************************
 ModelBase
******************************************************************************/

type ModelBase struct {
	AssetBase
}

func (m *ModelBase) Vertices() []float32 {
	return nil
}

func (m *ModelBase) Colors() []float32 {
	return nil
}

func (m *ModelBase) UVs() []float32 {
	return nil
}

func (m *ModelBase) Normals() []float32 {
	return nil
}

func (m *ModelBase) Tangents() []float32 {
	return nil
}

func (m *ModelBase) Bitangents() []float32 {
	return nil
}

/******************************************************************************
 MeshBase
******************************************************************************/

type MeshBase struct {
	ObjectTransform
}

func (m *MeshBase) Name() string {
	return ""
}

/******************************************************************************
 FaceBase
******************************************************************************/

type FaceBase struct{}

func (f *FaceBase) VertexIndices() []int {
	return nil
}

func (f *FaceBase) ColorIndices() []int {
	return nil
}

func (f *FaceBase) UvIndices() []int {
	return nil
}

func (f *FaceBase) NormalIndices() []int {
	return nil
}

func (f *FaceBase) TangentIndices() []int {
	return nil
}

func (f *FaceBase) BitangentIndices() []int {
	return nil
}

func (f *FaceBase) AttachedMaterial() Material {
	return nil
}

/******************************************************************************
 modelRenderer
******************************************************************************/

type modelRenderer struct {
	model *modelInstance

	activeCameraBinder *ShaderBinder
	cameraBinders      map[Camera]*ShaderBinder

	activeLightingBinder *ShaderBinder
	lightingBinders      map[any]*ShaderBinder
}

func (r *modelRenderer) setCamera(camera Camera) {
	if c, ok := r.cameraBinders[camera]; ok {
		r.activeCameraBinder = c
	} else {
		r.activeCameraBinder = newShaderBinder(r.model.shaders, camera, func() uint32 { return cameraUboBindPoint })
		r.activeCameraBinder.Init()
		r.cameraBinders[camera] = r.activeCameraBinder
	}
}

func (r *modelRenderer) setLighting(lighting any) {
	if b, ok := r.lightingBinders[lighting]; ok {
		r.activeLightingBinder = b
	} else {
		r.activeLightingBinder = newShaderBinder(r.model.shaders, lighting, func() uint32 { return lightingUboBindPoint })
		r.activeLightingBinder.Init()
		r.lightingBinders[lighting] = r.activeLightingBinder
	}
}

func (r *modelRenderer) drawFaces() {
	for _, mesh := range r.model.meshes {
		mesh.updateBindings()
		for _, group := range mesh.faceGroups {
			group.materialBinding.Update()
			gl.BindVertexArray(group.vao)
			gl.DrawArrays(gl.TRIANGLES, 0, group.vertexCount)
		}
	}
}

func (r *modelRenderer) render() {
	if r.activeCameraBinder != nil {
		r.activeCameraBinder.Update()
	}
	if r.activeLightingBinder != nil {
		r.activeLightingBinder.Update()
	}
	r.drawFaces()
}

func newModelRenderer(model *modelInstance) *modelRenderer {
	return &modelRenderer{
		model:           model,
		cameraBinders:   make(map[Camera]*ShaderBinder),
		lightingBinders: make(map[any]*ShaderBinder),
	}
}

/******************************************************************************
 modelInstance
******************************************************************************/

type modelInstance struct {
	model        Model
	meshes       []*meshInstance
	shaders      map[uint32]Shader
	bindingPoint uint32
}

func (m *modelInstance) getBindingPoint() uint32 {
	return m.bindingPoint
}

func (m *modelInstance) close() {
	for _, mesh := range m.meshes {
		mesh.close()
	}
}

func (m *modelInstance) getLayout() VertexAttributeLayout {
	if len(m.model.Bitangents()) > 0 && len(m.model.Tangents()) > 0 &&
		len(m.model.UVs()) > 0 && len(m.model.Normals()) > 0 && len(m.model.Vertices()) > 0 {
		return PositionNormalUvTangentsVaoLayout
	} else if len(m.model.UVs()) > 0 && len(m.model.Normals()) > 0 && len(m.model.Vertices()) > 0 {
		return PositionNormalUvVaoLayout
	} else if len(m.model.UVs()) > 0 && len(m.model.Vertices()) > 0 {
		return PositionUvVaoLayout
	} else if len(m.model.Colors()) > 0 && len(m.model.Vertices()) > 0 {
		return PositionColorVaoLayout
	} else if len(m.model.Vertices()) > 0 {
		return PositionOnlyVaoLayout
	}
	panic("unsupported vertex attribute layout")
}

func (m *modelInstance) Meshes() []*meshInstance {
	return m.meshes
}

func newModelInstance(model Model, parentTransform Transform) *modelInstance {
	if model == nil {
		panic("model cannot be nil")
	}

	if !model.Initialized() {
		if ok := model.Init(); !ok {
			panic("failed to initialize model")
		}
	}

	meshes := model.Meshes()
	if len(meshes) == 0 {
		panic("model must have at least one mesh")
	}

	instance := &modelInstance{}
	instance.model = model
	instance.meshes = make([]*meshInstance, len(meshes))
	instance.shaders = make(map[uint32]Shader)
	instance.bindingPoint = 5 // allow other bindings to take 0-4

	for i, mesh := range meshes {
		meshInst := newMeshInstance(mesh, parentTransform, instance)
		instance.meshes[i] = meshInst
		for shaderName, shader := range meshInst.shaders {
			instance.shaders[shaderName] = shader
		}
	}

	return instance
}

/******************************************************************************
 meshInstance
******************************************************************************/

type meshInstance struct {
	ObjectTransform

	parent *modelInstance

	name string

	faces      []*faceInstance
	faceGroups []*faceRenderGroup

	shaders map[uint32]Shader
	binder  *ShaderBinder

	WorldMat mgl32.Mat4
}

func (m *meshInstance) initFaces(mesh Mesh) {
	m.shaders = make(map[uint32]Shader)

	faces := mesh.Faces()
	if len(faces) == 0 {
		panic("mesh must have at least one face")
	}

	m.faces = make([]*faceInstance, len(faces))
	for j, face := range faces {
		faceInst := newFaceInstance(face)
		m.faces[j] = faceInst
		m.shaders[faceInst.shaderName] = faceInst.material.AttachedShader()
	}
}

func (m *meshInstance) initFaceGroups(mesh Mesh) {
	model := m.parent.model
	vertices := model.Vertices()
	colors := model.Colors()
	uvs := model.UVs()
	normals := model.Normals()
	tangents := model.Tangents()
	bitangents := model.Bitangents()

	layout := m.parent.getLayout()
	faces := mesh.Faces()

	group := &faceRenderGroup{
		model:    m.parent,
		layout:   layout,
		material: m.faces[0].material,
	}
	for _, face := range faces {
		material := face.AttachedMaterial()
		if material != group.material {
			group.vertexCount = int32(group.faceCount * 3)
			m.faceGroups = append(m.faceGroups, group)
			group = &faceRenderGroup{
				model:    m.parent,
				layout:   layout,
				material: material,
			}
		}

		switch layout {
		case PositionOnlyVaoLayout:
			vertIndices := face.VertexIndices()
			for i := 0; i < 3; i++ {
				vertexIdx := vertIndices[i] * 3
				group.buffer = append(group.buffer, vertices[vertexIdx:vertexIdx+3]...)
			}
		case PositionColorVaoLayout:
			vertIndices := face.VertexIndices()
			colIndices := face.ColorIndices()
			for i := 0; i < 3; i++ {
				vertexIdx := vertIndices[i] * 3
				group.buffer = append(group.buffer, vertices[vertexIdx:vertexIdx+3]...)

				colIdx := colIndices[i] * 3
				group.buffer = append(group.buffer, colors[colIdx:colIdx+3]...)
			}
		case PositionUvVaoLayout:
			vertIndices := face.VertexIndices()
			uvIndices := face.UvIndices()
			for i := 0; i < 3; i++ {
				vertexIdx := vertIndices[i] * 3
				group.buffer = append(group.buffer, vertices[vertexIdx:vertexIdx+3]...)

				uvIdx := uvIndices[i] * 2
				group.buffer = append(group.buffer, uvs[uvIdx:uvIdx+2]...)
			}
		case PositionNormalUvVaoLayout:
			vertIndices := face.VertexIndices()
			normIndices := face.NormalIndices()
			uvIndices := face.UvIndices()
			for i := 0; i < 3; i++ {
				vertexIdx := vertIndices[i] * 3
				group.buffer = append(group.buffer, vertices[vertexIdx:vertexIdx+3]...)

				normalIdx := normIndices[i] * 3
				group.buffer = append(group.buffer, normals[normalIdx:normalIdx+3]...)

				uvIdx := uvIndices[i] * 2
				group.buffer = append(group.buffer, uvs[uvIdx:uvIdx+2]...)
			}
		case PositionNormalUvTangentsVaoLayout:
			vertIndices := face.VertexIndices()
			normIndices := face.NormalIndices()
			uvIndices := face.UvIndices()
			tanIndices := face.TangentIndices()
			bitanIndices := face.BitangentIndices()
			for i := 0; i < 3; i++ {
				vertexIdx := vertIndices[i] * 3
				group.buffer = append(group.buffer, vertices[vertexIdx:vertexIdx+3]...)

				normalIdx := normIndices[i] * 3
				group.buffer = append(group.buffer, normals[normalIdx:normalIdx+3]...)

				uvIdx := uvIndices[i] * 2
				group.buffer = append(group.buffer, uvs[uvIdx:uvIdx+2]...)

				tanIdx := tanIndices[i] * 3
				group.buffer = append(group.buffer, tangents[tanIdx:tanIdx+3]...)

				bitanIdx := bitanIndices[i] * 3
				group.buffer = append(group.buffer, bitangents[bitanIdx:bitanIdx+3]...)
			}
		}

		group.faceCount++
	}

	if group.faceCount > 0 {
		group.vertexCount = int32(group.faceCount * 3)
		m.faceGroups = append(m.faceGroups, group)
	}

	for _, group = range m.faceGroups {
		group.init()
	}
}

func (m *meshInstance) initBindings() {
	m.binder = newShaderBinder(m.shaders, m, nil)
	m.binder.Init()
}

func (m *meshInstance) updateBindings() {
	m.WorldMat = m.WorldMatrix()
	m.binder.Update()
}

func (m *meshInstance) close() {
	for _, g := range m.faceGroups {
		g.close()
	}
	m.binder.Close()
}

func (m *meshInstance) Name() string {
	return m.name
}

func (m *meshInstance) Faces() []*faceInstance {
	return m.faces
}

func newMeshInstance(mesh Mesh, parentTransform Transform, parentModel *modelInstance) *meshInstance {
	instance := &meshInstance{
		parent: parentModel,
		name:   mesh.Name(),
	}

	instance.SetParentTransform(parentTransform)
	instance.SetOrigin(mesh.Origin())
	instance.SetPosition(mesh.Position())
	instance.SetRotation(mesh.Rotation())
	instance.SetScale(mesh.Scale())
	if rot := mesh.RotationQuat(); rot != mgl32.QuatIdent() {
		instance.SetRotationQuat(rot)
	}

	instance.initFaces(mesh)
	instance.initFaceGroups(mesh)
	instance.initBindings()

	return instance
}

/******************************************************************************
 faceInstance
******************************************************************************/

type faceInstance struct {
	material   Material
	shaderName uint32
}

func (f *faceInstance) Material() Material {
	return f.material
}

func newFaceInstance(face Face) *faceInstance {
	material := face.AttachedMaterial()
	if material == nil {
		panic("face must have an attached material")
	}
	if shader := material.AttachedShader(); shader == nil {
		panic("material must have an attached shader")
	} else {
		return &faceInstance{
			material:   material,
			shaderName: shader.GlName(),
		}
	}
}

/******************************************************************************
 faceRenderGroup
******************************************************************************/

type faceRenderGroup struct {
	model           *modelInstance
	material        Material
	shader          Shader
	layout          VertexAttributeLayout
	buffer          []float32
	faceCount       int
	vertexCount     int32
	materialBinding *ShaderBinding
	vao             uint32
	closeFunc       func()
}

func (g *faceRenderGroup) init() {
	g.shader = g.material.AttachedShader()
	g.materialBinding = newShaderBinding(g.shader.GlName(), g.material, func() uint32 { return materialUboBindPoint })
	g.materialBinding.Init()
	g.vao, g.closeFunc = newVertexArrayObject(g.layout, g.shader, g.buffer)
}

func (g *faceRenderGroup) close() {
	if g.closeFunc != nil {
		g.closeFunc()
	}
	g.materialBinding.Close()
}
