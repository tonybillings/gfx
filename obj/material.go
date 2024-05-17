package obj

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tonybillings/gfx"
	"io"
	"strings"
	"sync/atomic"
)

/******************************************************************************
 Material
******************************************************************************/

type BasicMaterial struct {
	gfx.MaterialBase

	name    string
	mapKd   string
	mapKs   string
	mapNorm string

	textures []gfx.Texture

	Properties  *BasicMaterialProperties
	DiffuseMap  gfx.Texture
	SpecularMap gfx.Texture
	NormalMap   gfx.Texture
}

type BasicMaterialProperties struct {
	Ambient      mgl32.Vec4
	Diffuse      mgl32.Vec4
	Specular     mgl32.Vec4
	Emissive     mgl32.Vec4
	Shininess    float32
	Transparency float32
}

/******************************************************************************
 Asset Implementation
******************************************************************************/

func (m *BasicMaterial) Name() string {
	return m.name
}

func (m *BasicMaterial) Init() bool {
	if m.Initialized() {
		return true
	}

	for _, t := range m.textures {
		t.Init()
	}

	return m.AssetBase.Init()
}

func (m *BasicMaterial) Close() {
	if !m.Initialized() {
		return
	}

	for _, t := range m.textures {
		t.Close()
	}

	m.AssetBase.Close()
}

/******************************************************************************
 New Material Function
******************************************************************************/

func NewMaterial() *BasicMaterial {
	return &BasicMaterial{
		Properties: &BasicMaterialProperties{
			Ambient:      mgl32.Vec4{0.2, 0.2, 0.2},
			Diffuse:      mgl32.Vec4{0.5, 0.5, 0.5},
			Specular:     mgl32.Vec4{1.0, 1.0, 1.0},
			Emissive:     mgl32.Vec4{0.0, 0.0, 0.0},
			Shininess:    32.0,
			Transparency: 0.0,
		},
	}
}

/******************************************************************************
 MaterialLibrary
******************************************************************************/

type MaterialLibrary struct {
	gfx.AssetBase

	materials map[string]*BasicMaterial
	loaded    atomic.Bool
}

/******************************************************************************
 Asset Implementation
******************************************************************************/

func (l *MaterialLibrary) Init() bool {
	if l.Initialized() {
		return true
	}

	l.Load()

	if ok := l.initMaterials(); !ok {
		return false
	}

	return l.AssetBase.Init()
}

func (l *MaterialLibrary) Close() {
	if !l.Initialized() {
		return
	}

	l.closeMaterials()

	l.AssetBase.Close()
}

/******************************************************************************
 MaterialLibrary Functions
******************************************************************************/

func (l *MaterialLibrary) parseFields(fields []string, currentMat *BasicMaterial, lineNumber int) *BasicMaterial {
	switch fields[0] {
	case "newmtl":
		currentMat = l.parseNewmtl(fields, currentMat, lineNumber)
	case "Ka":
		l.parseKa(fields, currentMat, lineNumber)
	case "Kd":
		l.parseKd(fields, currentMat, lineNumber)
	case "Ks":
		l.parseKs(fields, currentMat, lineNumber)
	case "Ns":
		l.parseNs(fields, currentMat, lineNumber)
	case "Ke":
		l.parseKe(fields, currentMat, lineNumber)
	case "Tr":
		l.parseTr(fields, currentMat, lineNumber)
	case "map_Kd":
		l.parseMapKd(fields, currentMat, lineNumber)
	case "map_Ks":
		l.parseMapKs(fields, currentMat, lineNumber)
	case "norm", "map_Kn":
		l.parseMapKn(fields, currentMat, lineNumber)
	}

	return currentMat
}

func (l *MaterialLibrary) parseNewmtl(fields []string, currentMat *BasicMaterial, lineNumber int) *BasicMaterial {
	if currentMat.name != "" {
		l.materials[currentMat.name] = currentMat
		currentMat = NewMaterial()
		currentMat.SetSourceLibrary(l.SourceLibrary())
	}

	if name, err := parseString(fields[1:]); err != nil {
		panic(fmt.Errorf("MTL parse newmtl error: line %d: %w", lineNumber, err))
	} else {
		currentMat.name = name
	}

	return currentMat
}

func (l *MaterialLibrary) parseKa(fields []string, currentMat *BasicMaterial, lineNumber int) {
	if value, err := parseVec4(fields[1:]); err != nil {
		panic(fmt.Errorf("MTL Ka parsing error: line %d: %w", lineNumber, err))
	} else {
		currentMat.Properties.Ambient = value
	}
}

func (l *MaterialLibrary) parseKd(fields []string, currentMat *BasicMaterial, lineNumber int) {
	if value, err := parseVec4(fields[1:]); err != nil {
		panic(fmt.Errorf("MTL Kd parsing error: line %d: %w", lineNumber, err))
	} else {
		currentMat.Properties.Diffuse = value
	}
}

func (l *MaterialLibrary) parseKs(fields []string, currentMat *BasicMaterial, lineNumber int) {
	if value, err := parseVec4(fields[1:]); err != nil {
		panic(fmt.Errorf("MTL Ks parsing error: line %d: %w", lineNumber, err))
	} else {
		currentMat.Properties.Specular = value
	}
}

func (l *MaterialLibrary) parseNs(fields []string, currentMat *BasicMaterial, lineNumber int) {
	if value, err := parseFloat(fields[1:]); err != nil {
		panic(fmt.Errorf("MTL Ns parsing error: line %d: %w", lineNumber, err))
	} else {
		currentMat.Properties.Shininess = value
	}
}

func (l *MaterialLibrary) parseKe(fields []string, currentMat *BasicMaterial, lineNumber int) {
	if value, err := parseVec4(fields[1:]); err != nil {
		panic(fmt.Errorf("MTL Ke parsing error: line %d: %w", lineNumber, err))
	} else {
		currentMat.Properties.Emissive = value
	}
}

func (l *MaterialLibrary) parseTr(fields []string, currentMat *BasicMaterial, lineNumber int) {
	if value, err := parseFloat(fields[1:]); err != nil {
		panic(fmt.Errorf("MTL Tr parsing error: line %d: %w", lineNumber, err))
	} else {
		currentMat.Properties.Transparency = value
	}
}

func (l *MaterialLibrary) parseMapKd(fields []string, currentMat *BasicMaterial, lineNumber int) {
	if value, err := parseString(fields[1:]); err != nil {
		panic(fmt.Errorf("MTL map_Kd parsing error: line %d: %w", lineNumber, err))
	} else {
		currentMat.mapKd = value
		currentMat.DiffuseMap = gfx.NewTexture2D(value, value)
		currentMat.DiffuseMap.SetSourceLibrary(l.SourceLibrary())
		currentMat.textures = append(currentMat.textures, currentMat.DiffuseMap)
	}
}

func (l *MaterialLibrary) parseMapKs(fields []string, currentMat *BasicMaterial, lineNumber int) {
	if value, err := parseString(fields[1:]); err != nil {
		panic(fmt.Errorf("MTL map_Ks parsing error: line %d: %w", lineNumber, err))
	} else {
		currentMat.mapKs = value
		currentMat.SpecularMap = gfx.NewTexture2D(value, value)
		currentMat.SpecularMap.SetSourceLibrary(l.SourceLibrary())
		currentMat.textures = append(currentMat.textures, currentMat.SpecularMap)
	}
}

func (l *MaterialLibrary) parseMapKn(fields []string, currentMat *BasicMaterial, lineNumber int) {
	if value, err := parseString(fields[1:]); err != nil {
		panic(fmt.Errorf("MTL norm/map_Kn parsing error: line %d: %w", lineNumber, err))
	} else {
		currentMat.mapNorm = value
		currentMat.NormalMap = gfx.NewTexture2D(value, value)
		currentMat.NormalMap.SetSourceLibrary(l.SourceLibrary())
		currentMat.textures = append(currentMat.textures, currentMat.NormalMap)
	}
}

func (l *MaterialLibrary) loadFromSlice(slice []byte) {
	reader := bufio.NewReader(bytes.NewReader(slice))
	l.loadFromReader(reader)
}

func (l *MaterialLibrary) loadFromFile(name string) (ok bool) {
	reader, closeFunc := getSourceReader(l, name)
	defer closeFunc()

	if reader == nil {
		return false
	}

	l.loadFromReader(reader)
	return true
}

func (l *MaterialLibrary) loadFromString(mtl string) {
	reader := bufio.NewReader(strings.NewReader(mtl))
	l.loadFromReader(reader)
}

func (l *MaterialLibrary) loadTextures() {
	for _, mat := range l.materials {
		if mat.DiffuseMap == nil {
			mat.DiffuseMap = gfx.NewTexture2D("", gfx.White)
			mat.DiffuseMap.SetSourceLibrary(l.SourceLibrary())
			mat.textures = append(mat.textures, mat.DiffuseMap)
		}

		if mat.NormalMap == nil {
			mat.NormalMap = gfx.NewTexture2D("", gfx.DefaultNormalMapColor)
			mat.NormalMap.SetSourceLibrary(l.SourceLibrary())
			mat.textures = append(mat.textures, mat.NormalMap)
		}

		if mat.SpecularMap == nil {
			mat.SpecularMap = gfx.NewTexture2D("", gfx.DefaultSpecularMapColor)
			mat.SpecularMap.SetSourceLibrary(l.SourceLibrary())
			mat.textures = append(mat.textures, mat.SpecularMap)
		}
	}
}

func (l *MaterialLibrary) loadFromReader(reader *bufio.Reader) {
	lineNumber := 0
	currentMat := NewMaterial()
	currentMat.SetSourceLibrary(l.SourceLibrary())

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

		currentMat = l.parseFields(fields, currentMat, lineNumber)
	}

	l.materials[currentMat.name] = currentMat

	l.loadTextures()
}

func (l *MaterialLibrary) initMaterials() bool {
	ok := true
	for _, mat := range l.materials {
		ok = ok && mat.Init()
	}
	return ok
}

func (l *MaterialLibrary) closeMaterials() {
	for _, mat := range l.materials {
		mat.Close()
	}
}

func (l *MaterialLibrary) Get(name string) *BasicMaterial {
	return l.materials[name]
}

func (l *MaterialLibrary) GetNames() []string {
	names := make([]string, 0)
	for _, m := range l.materials {
		names = append(names, m.Name())
	}
	return names
}

func (l *MaterialLibrary) Load() {
	if l.loaded.Load() {
		return
	}

	switch source := l.Source().(type) {
	case []byte:
		l.loadFromSlice(source)
	case string:
		if ok := l.loadFromFile(source); !ok {
			l.loadFromString(source)
		}
	default:
		panic("unexpected error: source type is not supported")
	}

	l.loaded.Store(true)
}

/******************************************************************************
 New MaterialLibrary Function
******************************************************************************/

func NewMaterialLibrary[T gfx.MaterialLibrarySource](name string, source T) *MaterialLibrary {
	return &MaterialLibrary{
		AssetBase: *gfx.NewAssetBase(name, source),
		materials: make(map[string]*BasicMaterial),
	}
}
