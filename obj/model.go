package obj

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/tonybillings/gfx"
	"io"
	"strings"
	"sync/atomic"
)

/******************************************************************************
 Model
******************************************************************************/

type Model struct {
	gfx.ModelBase

	mtllibs []string

	vertices   []float32
	normals    []float32
	uvs        []float32
	tangents   []float32
	bitangents []float32

	meshes       []*Mesh
	materialLibs []*MaterialLibrary

	defaultMaterial *BasicMaterial
	defaultShader   gfx.Shader

	computeTangentsOnLoad bool

	loaded atomic.Bool
}

/******************************************************************************
 gfx.Asset Implementation
******************************************************************************/

func (m *Model) Init() bool {
	if m.Initialized() {
		return true
	}

	m.Load()

	if ok := m.initMaterialLibraries(); !ok {
		return false
	}

	if ok := m.initMeshes(); !ok {
		return false
	}

	return m.AssetBase.Init()
}

func (m *Model) Close() {
	if !m.Initialized() {
		return
	}

	m.closeMeshes()
	m.closeMaterialLibraries()

	m.AssetBase.Close()
}

/******************************************************************************
 gfx.Model Implementation
******************************************************************************/

func (m *Model) Vertices() []float32 {
	return m.vertices
}

func (m *Model) Normals() []float32 {
	return m.normals
}

func (m *Model) UVs() []float32 {
	return m.uvs
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

/******************************************************************************
 Model Functions
******************************************************************************/

func (m *Model) parseFields(fields []string, currentMesh *Mesh, currentMat string, lineNumber int) (*Mesh, string) {
	switch fields[0] {
	case "mtllib":
		m.parseMtllib(fields, lineNumber)
	case "usemtl":
		currentMat = m.parseUsemtl(fields, currentMat, lineNumber)
	case "g":
		currentMesh = m.parseGroup(fields, currentMesh, lineNumber)
	case "v":
		m.parseVertex(fields, lineNumber)
	case "vn":
		m.parseVertexNormal(fields, lineNumber)
	case "vt":
		m.parseVertexTexture(fields, lineNumber)
	case "f":
		m.parseFace(fields, currentMesh, currentMat, lineNumber)
	}

	return currentMesh, currentMat
}

func (m *Model) parseMtllib(fields []string, lineNumber int) {
	if mat, err := parseString(fields[1:]); err != nil {
		panic(fmt.Errorf("OBJ parse mtllib error: line %d: %w", lineNumber, err))
	} else {
		m.mtllibs = append(m.mtllibs, mat)
	}
}

func (m *Model) parseUsemtl(fields []string, currentMat string, lineNumber int) string {
	if mat, err := parseString(fields[1:]); err != nil {
		panic(fmt.Errorf("OBJ parse usemtl error: line %d: %w", lineNumber, err))
	} else {
		currentMat = mat
	}
	return currentMat
}

func (m *Model) parseGroup(fields []string, currentMesh *Mesh, lineNumber int) *Mesh {
	if currentMesh.name != "" {
		m.meshes = append(m.meshes, currentMesh)
		currentMesh = NewMesh()
	}

	if group, err := parseString(fields[1:]); err != nil {
		panic(fmt.Errorf("OBJ parse group error: line %d: %w", lineNumber, err))
	} else {
		currentMesh.name = group
	}

	return currentMesh
}

func (m *Model) parseVertex(fields []string, lineNumber int) {
	if vertex, err := parseVec3(fields[1:]); err != nil {
		panic(fmt.Errorf("OBJ vertex parsing error: line %d: %w", lineNumber, err))
	} else {
		m.vertices = append(m.vertices, vertex[:]...)
	}
}

func (m *Model) parseVertexNormal(fields []string, lineNumber int) {
	if normal, err := parseVec3(fields[1:]); err != nil {
		panic(fmt.Errorf("OBJ vertex normal parsing error: line %d: %w", lineNumber, err))
	} else {
		m.normals = append(m.normals, normal[:]...)
	}
}

func (m *Model) parseVertexTexture(fields []string, lineNumber int) {
	if uv, err := parseVec2(fields[1:]); err != nil {
		panic(fmt.Errorf("OBJ vertex texture parsing error: line %d: %w", lineNumber, err))
	} else {
		m.uvs = append(m.uvs, uv[:]...)
	}
}

func (m *Model) parseFace(fields []string, currentMesh *Mesh, currentMat string, lineNumber int) {
	if face, err := parseFace(fields[1:]); err != nil {
		panic(fmt.Errorf("OBJ face parsing error: line %d: %w", lineNumber, err))
	} else {
		face.usemtl = currentMat
		currentMesh.faces = append(currentMesh.faces, face)
	}
}

func (m *Model) loadFromSlice(slice []byte) {
	reader := bufio.NewReader(bytes.NewReader(slice))
	m.loadFromReader(reader)
}

func (m *Model) loadFromFile(name string) (ok bool) {
	reader, closeFunc := getSourceReader(m, name)
	defer closeFunc()

	if reader == nil {
		return false
	}

	m.loadFromReader(reader)
	return true
}

func (m *Model) loadFromString(obj string) {
	reader := bufio.NewReader(strings.NewReader(obj))
	m.loadFromReader(reader)
}

func (m *Model) loadFromReader(reader *bufio.Reader) {
	currentMat := ""
	lineNumber := 0
	currentMesh := NewMesh()

	for {
		lineNumber++

		line, readErr := reader.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			panic(fmt.Errorf("file read error: %w", readErr))
		} else if readErr == io.EOF {
			break
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		currentMesh, currentMat = m.parseFields(fields, currentMesh, currentMat, lineNumber)
	}

	m.meshes = append(m.meshes, currentMesh)
}

func (m *Model) computeTangents() {
	for _, mesh := range m.meshes {
		for _, face := range mesh.faces {
			setFaceTangents(m, face)
		}
	}
}

func (m *Model) loadMaterialLibraries() {
	for _, mtllib := range m.mtllibs {
		mtl := NewMaterialLibrary(mtllib, mtllib)
		mtl.SetSourceLibrary(m.SourceLibrary())
		mtl.Load()
		m.materialLibs = append(m.materialLibs, mtl)
	}

	if m.defaultShader == nil {
		srcLib := m.SourceLibrary()
		if srcLib != nil {
			if defaultShader := srcLib.Get(gfx.Shape3DShader); defaultShader != nil {
				if shader, ok := defaultShader.(gfx.Shader); ok {
					m.defaultShader = shader
				}
			}
		}
	}

	if m.defaultMaterial == nil {
		m.defaultMaterial = NewMaterial()
		m.defaultMaterial.AttachShader(m.defaultShader)
	}
}

func (m *Model) initMaterialLibraries() bool {
	ok := true
	for _, mtllib := range m.materialLibs {
		ok = ok && mtllib.Init()
	}
	return ok && m.defaultMaterial.Init()
}

func (m *Model) getMaterial(name string) *BasicMaterial {
	for _, mtl := range m.materialLibs {
		if mat := mtl.Get(name); mat != nil {
			return mat
		}
	}
	return nil
}

func (m *Model) setMaterials() {
	for _, mesh := range m.meshes {
		for _, face := range mesh.faces {
			face.material = m.getMaterial(face.usemtl)
			if face.material == nil {
				face.material = m.defaultMaterial
			}

			if face.material.AttachedShader() == nil && m.defaultShader != nil {
				face.material.AttachShader(m.defaultShader)
			}
		}
	}
}

func (m *Model) closeMaterialLibraries() {
	for _, mtllib := range m.materialLibs {
		mtllib.Close()
	}

	if m.defaultMaterial != nil {
		m.defaultMaterial.Close()
	}
}

func (m *Model) initMeshes() bool {
	ok := true
	for _, mesh := range m.meshes {
		ok = ok && mesh.Init()
	}
	return ok
}

func (m *Model) closeMeshes() {
	for _, mesh := range m.meshes {
		mesh.Close()
	}
}

func (m *Model) SetDefaultShader(shader gfx.Shader) *Model {
	m.defaultShader = shader
	return m
}

func (m *Model) Load() {
	if m.loaded.Load() {
		return
	}

	switch source := m.Source().(type) {
	case []byte:
		m.loadFromSlice(source)
	case string:
		if ok := m.loadFromFile(source); !ok {
			m.loadFromString(source)
		}
	default:
		panic("unexpected error: source type is not supported")
	}

	if m.computeTangentsOnLoad {
		m.computeTangents()
	}

	m.loadMaterialLibraries()
	m.setMaterials()
	m.loaded.Store(true)
}

func (m *Model) ComputeTangents(computeOnLoad bool) {
	m.computeTangentsOnLoad = computeOnLoad
}

/******************************************************************************
 New Model Function
******************************************************************************/

func NewModel[T gfx.ModelSource](name string, source T) *Model {
	return &Model{
		ModelBase: gfx.ModelBase{
			AssetBase: *gfx.NewAssetBase(name, source),
		},
	}
}
