package gfx

import (
	"fmt"
	"github.com/g3n/engine/loader/obj"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	defaultModelName = "Model"
)

/******************************************************************************
 Model
******************************************************************************/

type Model struct {
	WindowObjectBase

	vao uint32
	vbo uint32

	shader        uint32
	shaderNoLight uint32

	objFilename string
	mtlFilename string

	worldUniformLoc    int32
	viewProjUniformLoc int32
	textureUniformLoc  int32
	ambientUniformLoc  int32
	diffuseUniformLoc  int32

	worldUniformLoc2    int32
	viewProjUniformLoc2 int32
	textureUniformLoc2  int32
	ambientUniformLoc2  int32
	diffuseUniformLoc2  int32

	specularUniformLoc   int32
	shininessUniformLoc  int32
	lightDirUniformLoc   int32
	lightColorUniformLoc int32
	viewPosUniformLoc    int32

	texture   Texture
	indices   []uint32
	materials []*obj.Material

	camera   *Camera
	lights   []*Light
	viewport [4]float32
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (m *Model) uninitGl() {
	if !m.Initialized() {
		return
	}
	m.SetInitialized(false)

	m.stateMutex.Lock()
	gl.BindVertexArray(m.vao)
	gl.DisableVertexAttribArray(0)
	gl.DisableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	gl.DeleteBuffers(1, &m.vbo)
	gl.DeleteVertexArrays(1, &m.vao)
	m.stateMutex.Unlock()

	m.uninitTexture()
}

func (m *Model) initGl() {
	m.shader = GetShaderProgram(ModelShaderProgram)
	m.shaderNoLight = GetShaderProgram(ModelNoLightShaderProgram)

	m.worldUniformLoc = gl.GetUniformLocation(m.shader, gl.Str("world\x00"))
	m.viewProjUniformLoc = gl.GetUniformLocation(m.shader, gl.Str("viewProj\x00"))
	m.textureUniformLoc = gl.GetUniformLocation(m.shader, gl.Str("tex2D\x00"))
	m.ambientUniformLoc = gl.GetUniformLocation(m.shader, gl.Str("material.ambient\x00"))
	m.diffuseUniformLoc = gl.GetUniformLocation(m.shader, gl.Str("material.diffuse\x00"))

	m.worldUniformLoc2 = gl.GetUniformLocation(m.shaderNoLight, gl.Str("world\x00"))
	m.viewProjUniformLoc2 = gl.GetUniformLocation(m.shaderNoLight, gl.Str("viewProj\x00"))
	m.textureUniformLoc2 = gl.GetUniformLocation(m.shaderNoLight, gl.Str("tex2D\x00"))
	m.ambientUniformLoc2 = gl.GetUniformLocation(m.shaderNoLight, gl.Str("material.ambient\x00"))
	m.diffuseUniformLoc2 = gl.GetUniformLocation(m.shaderNoLight, gl.Str("material.diffuse\x00"))

	m.lightDirUniformLoc = gl.GetUniformLocation(m.shader, gl.Str("lightDir\x00"))
	m.lightColorUniformLoc = gl.GetUniformLocation(m.shader, gl.Str("lightColor\x00"))
	m.specularUniformLoc = gl.GetUniformLocation(m.shader, gl.Str("material.specular\x00"))
	m.shininessUniformLoc = gl.GetUniformLocation(m.shader, gl.Str("material.shininess\x00"))
	m.viewPosUniformLoc = gl.GetUniformLocation(m.shader, gl.Str("viewPos\x00"))

	gl.GenVertexArrays(1, &m.vao)
	gl.GenBuffers(1, &m.vbo)

	gl.BindVertexArray(m.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)

	stride := int32(8 * sizeOfFloat32)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, stride, 0)

	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, stride, uintptr(3*sizeOfFloat32))

	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, stride, uintptr(6*sizeOfFloat32))

	gl.BufferData(gl.ARRAY_BUFFER, len(m.vertices)*sizeOfFloat32, gl.Ptr(m.vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func (m *Model) Init(window *Window) (ok bool) {
	if !m.WindowObjectBase.Init(window) {
		return false
	}

	m.loadOBJ()
	m.initTexture()
	m.initGl()
	m.initialized.Store(true)

	return true
}

func (m *Model) Draw(deltaTime int64) (ok bool) {
	if !m.visible.Load() || m.closed.Load() {
		return false
	}

	if m.closing.Load() {
		m.uninitGl()
		m.closed.Store(true)
		m.closing.Store(false)
		return false
	}

	worldMat := m.WorldMatrix()
	viewMat := m.camera.View()
	projMat := m.camera.Projection()
	viewProjMat := projMat.Mul4(viewMat)

	gl.ActiveTexture(gl.TEXTURE0)

	if m.texture != nil {
		gl.BindTexture(gl.TEXTURE_2D, m.texture.name())
	}

	m.stateMutex.Lock()

	if len(m.lights) > 0 {
		gl.UseProgram(m.shader)

		light := m.lights[0]
		lightDir := light.Direction().Normalize()
		lightColor := light.Color()
		viewPos := m.camera.Position()

		gl.Uniform3fv(m.lightDirUniformLoc, 1, &lightDir[0])
		gl.Uniform3fv(m.lightColorUniformLoc, 1, &lightColor[0])
		gl.Uniform3fv(m.viewPosUniformLoc, 1, &viewPos[0])

		if len(m.materials) > 0 {
			mat := m.materials[0]
			gl.Uniform3fv(m.ambientUniformLoc, 1, &mat.Ambient.R)
			gl.Uniform3fv(m.diffuseUniformLoc, 1, &mat.Diffuse.R)
			gl.Uniform3fv(m.specularUniformLoc, 1, &mat.Specular.R)
			gl.Uniform1f(m.shininessUniformLoc, mat.Shininess)
		}

		gl.UniformMatrix4fv(m.worldUniformLoc, 1, false, &worldMat[0])
		gl.UniformMatrix4fv(m.viewProjUniformLoc, 1, false, &viewProjMat[0])
		gl.Uniform1i(m.textureUniformLoc, 0)
	} else {
		gl.UseProgram(m.shaderNoLight)

		if len(m.materials) > 0 {
			mat := m.materials[0]
			gl.Uniform3fv(m.ambientUniformLoc2, 1, &mat.Ambient.R)
			gl.Uniform3fv(m.diffuseUniformLoc2, 1, &mat.Diffuse.R)
		}

		gl.UniformMatrix4fv(m.worldUniformLoc2, 1, false, &worldMat[0])
		gl.UniformMatrix4fv(m.viewProjUniformLoc2, 1, false, &viewProjMat[0])
		gl.Uniform1i(m.textureUniformLoc2, 0)
	}

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	winWidth := m.window.Width()
	winHeight := m.window.Height()
	gl.Viewport(
		int32(m.viewport[0]*float32(winWidth)),
		int32(m.viewport[1]*float32(winHeight)),
		int32(m.viewport[2]*float32(winWidth)),
		int32(m.viewport[3]*float32(winHeight)))

	m.stateMutex.Unlock()

	gl.BindVertexArray(m.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	gl.Enable(gl.DEPTH_TEST)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.DrawArrays(gl.TRIANGLES, 0, m.vertexCount)

	gl.Viewport(0, 0, m.window.Width(), m.window.Height())

	gl.Disable(gl.BLEND)
	gl.Disable(gl.DEPTH_TEST)

	gl.BindVertexArray(0)
	gl.UseProgram(0)

	return m.WindowObjectBase.drawChildren(deltaTime)
}

/******************************************************************************
 Model Functions
******************************************************************************/

func (m *Model) loadOBJ() {
	if m.objFilename == "" {
		return
	}

	m.stateMutex.Lock()

	var dec *obj.Decoder
	var err error

	if fileExists(m.objFilename) && fileExists(m.mtlFilename) {
		dec, err = obj.Decode(m.objFilename, m.mtlFilename)
		if err != nil {
			panic(fmt.Errorf("error decoding OBJ/MTL files: %w", err))
		}
	} else if fileExists(m.objFilename) {
		dec, err = obj.Decode(m.objFilename, "")
		if err != nil {
			panic(fmt.Errorf("error decoding OBJ file: %w", err))
		}
	} else {
		objAsset := GetAssetReader(m.objFilename)
		mtlAsset := GetAssetReader(m.mtlFilename)

		if objAsset != nil && mtlAsset != nil {
			dec, err = obj.DecodeReader(objAsset, mtlAsset)
			if err != nil {
				panic(fmt.Errorf("error decoding OBJ/MTL files: %w", err))
			}
		} else if objAsset != nil {
			dec, err = obj.DecodeReader(objAsset, nil)
			if err != nil {
				panic(fmt.Errorf("error decoding OBJ file: %w", err))
			}
		} else {
			panic("a valid OBJ file must be provided")
		}
	}

	for _, object := range dec.Objects {
		for _, face := range object.Faces {
			for i := 0; i < len(face.Vertices); i++ {
				vertexStartIdx := face.Vertices[i] * 3
				x := dec.Vertices[vertexStartIdx]
				y := dec.Vertices[vertexStartIdx+1]
				z := dec.Vertices[vertexStartIdx+2]
				m.vertices = append(m.vertices, x, y, z)

				normalStartIdx := face.Normals[i] * 3
				nx := dec.Normals[normalStartIdx]
				ny := dec.Normals[normalStartIdx+1]
				nz := dec.Normals[normalStartIdx+2]
				m.vertices = append(m.vertices, nx, ny, nz)

				uvStartIdx := face.Uvs[i] * 2
				u := dec.Uvs[uvStartIdx]
				v := dec.Uvs[uvStartIdx+1]
				m.vertices = append(m.vertices, u, v)
			}
		}
	}

	m.vertexCount = int32(len(m.vertices) / 8)
	for _, mat := range dec.Materials {
		m.materials = append(m.materials, mat)
	}

	m.stateMutex.Unlock()
}

func (m *Model) initTexture() {
	m.stateMutex.Lock()
	if m.texture != nil {
		m.texture.init()
	}
	m.stateMutex.Unlock()
}

func (m *Model) uninitTexture() {
	m.stateMutex.Lock()
	if m.texture != nil {
		m.texture.close()
	}
	m.stateMutex.Unlock()
}

func (m *Model) Viewport() mgl32.Vec4 {
	m.stateMutex.Lock()
	vp := mgl32.Vec4{m.viewport[0], m.viewport[1], m.viewport[2], m.viewport[3]}
	m.stateMutex.Unlock()
	return vp
}

func (m *Model) SetViewport(viewport mgl32.Vec4) *Model {
	m.stateMutex.Lock()
	m.viewport[0] = viewport[0]
	m.viewport[1] = viewport[1]
	m.viewport[2] = viewport[2]
	m.viewport[3] = viewport[3]
	m.stateChanged.Store(true)
	m.stateMutex.Unlock()
	return m
}

func (m *Model) OBJ() string {
	m.stateMutex.Lock()
	filename := m.objFilename
	m.stateMutex.Unlock()
	return filename
}

func (m *Model) SetOBJ(pathToObj string) *Model {
	m.stateMutex.Lock()
	m.objFilename = pathToObj
	m.stateChanged.Store(true)
	m.stateMutex.Unlock()
	return m
}

func (m *Model) MTL() string {
	m.stateMutex.Lock()
	filename := m.mtlFilename
	m.stateMutex.Unlock()
	return filename
}

func (m *Model) SetMTL(pathToMtl string) *Model {
	m.stateMutex.Lock()
	m.mtlFilename = pathToMtl
	m.stateChanged.Store(true)
	m.stateMutex.Unlock()
	return m
}

func (m *Model) SetTexture(texture Texture) *Model {
	m.stateMutex.Lock()
	m.texture = texture
	m.stateChanged.Store(true)
	m.stateMutex.Unlock()
	return m
}

func (m *Model) WorldMatrix() mgl32.Mat4 {
	tran := m.WorldPosition()
	rot := m.WorldRotation()
	scale := m.WorldScale()

	var rotateMat mgl32.Mat4
	if rot.Len() > 0.0001 {
		rotateMat = mgl32.HomogRotate3D(rot.Len(), rot.Normalize())
	} else {
		rotateMat = mgl32.Ident4()
	}

	translateMat := mgl32.Translate3D(tran.X(), tran.Y(), tran.Z())
	scaleMat := mgl32.Scale3D(scale.X(), scale.Y(), scale.Z())
	a := translateMat.Mul4(rotateMat.Mul4(scaleMat))
	return a
}

func (m *Model) Camera() *Camera {
	m.stateMutex.Lock()
	cam := m.camera
	m.stateMutex.Unlock()
	return cam
}

func (m *Model) SetCamera(camera *Camera) *Model {
	m.stateMutex.Lock()
	m.camera = camera
	m.stateMutex.Unlock()
	return m
}

func (m *Model) AddLight(light *Light) *Model {
	m.stateMutex.Lock()
	m.lights = append(m.lights, light)
	m.stateMutex.Unlock()
	return m
}

func (m *Model) GetLight(index int) *Light {
	m.stateMutex.Lock()
	if index < 0 || index >= len(m.lights) {
		m.stateMutex.Unlock()
		return nil
	}
	light := m.lights[index]
	m.stateMutex.Unlock()
	return light
}

func (m *Model) RemoveLights() *Model {
	m.stateMutex.Lock()
	m.lights = make([]*Light, 0)
	m.stateMutex.Unlock()
	return m
}

func (m *Model) AddMaterial(material *obj.Material) *Model {
	m.stateMutex.Lock()
	m.materials = append(m.materials, material)
	m.stateMutex.Unlock()
	return m
}

func (m *Model) RemoveMaterials() *Model {
	m.stateMutex.Lock()
	m.materials = make([]*obj.Material, 0)
	m.stateMutex.Unlock()
	return m
}

/******************************************************************************
 New Model Function
******************************************************************************/

func NewModel() *Model {
	m := &Model{
		WindowObjectBase: *NewObject(nil),
		camera:           NewCamera(),
		lights:           make([]*Light, 0),
		viewport:         mgl32.Vec4{0, 0, 1, 1},
		materials:        make([]*obj.Material, 0),
	}

	m.SetName(defaultModelName)
	return m
}
