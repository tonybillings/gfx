package obj

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"io"
	"strings"
	"sync/atomic"
	"tonysoft.com/gfx"
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

func (l *MaterialLibrary) loadFromSlice(slice []byte) {
	reader := bufio.NewReader(bytes.NewReader(slice))
	l.loadFromReader(reader)
}

func (l *MaterialLibrary) loadFromFile(name string) (ok bool) {
	reader, closeFunc := gfx.Assets.GetReader(name)
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

func (l *MaterialLibrary) loadFromReader(reader *bufio.Reader) {
	lineNumber := 0
	currentMat := NewMaterial()

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

		switch fields[0] {
		case "newmtl":
			if currentMat.name != "" {
				l.materials[currentMat.name] = currentMat
				currentMat = NewMaterial()
			}

			if name, err := parseString(fields[1:]); err != nil {
				panic(fmt.Errorf("MTL parse newmtl error: line %d: %w", lineNumber, err))
			} else {
				currentMat.name = name
			}
		case "Ka":
			if value, err := parseVec4(fields[1:]); err != nil {
				panic(fmt.Errorf("MTL Ka parsing error: line %d: %w", lineNumber, err))
			} else {
				currentMat.Properties.Ambient = value
			}
		case "Kd":
			if value, err := parseVec4(fields[1:]); err != nil {
				panic(fmt.Errorf("MTL Kd parsing error: line %d: %w", lineNumber, err))
			} else {
				currentMat.Properties.Diffuse = value
			}
		case "Ks":
			if value, err := parseVec4(fields[1:]); err != nil {
				panic(fmt.Errorf("MTL Ks parsing error: line %d: %w", lineNumber, err))
			} else {
				currentMat.Properties.Specular = value
			}
		case "Ns":
			if value, err := parseFloat(fields[1:]); err != nil {
				panic(fmt.Errorf("MTL Ns parsing error: line %d: %w", lineNumber, err))
			} else {
				currentMat.Properties.Shininess = value
			}
		case "Ke":
			if value, err := parseVec4(fields[1:]); err != nil {
				panic(fmt.Errorf("MTL Ke parsing error: line %d: %w", lineNumber, err))
			} else {
				currentMat.Properties.Emissive = value
			}
		case "Tr":
			if value, err := parseFloat(fields[1:]); err != nil {
				panic(fmt.Errorf("MTL Tr parsing error: line %d: %w", lineNumber, err))
			} else {
				currentMat.Properties.Transparency = value
			}
		case "map_Kd":
			if value, err := parseString(fields[1:]); err != nil {
				panic(fmt.Errorf("MTL map_Kd parsing error: line %d: %w", lineNumber, err))
			} else {
				currentMat.mapKd = value
				currentMat.DiffuseMap = gfx.NewTexture2D(value, value)
				currentMat.textures = append(currentMat.textures, currentMat.DiffuseMap)
			}
		case "map_Ks":
			if value, err := parseString(fields[1:]); err != nil {
				panic(fmt.Errorf("MTL map_Ks parsing error: line %d: %w", lineNumber, err))
			} else {
				currentMat.mapKs = value
				currentMat.SpecularMap = gfx.NewTexture2D(value, value)
				currentMat.textures = append(currentMat.textures, currentMat.SpecularMap)
			}
		case "norm", "map_Kn":
			if value, err := parseString(fields[1:]); err != nil {
				panic(fmt.Errorf("MTL norm/map_Kn parsing error: line %d: %w", lineNumber, err))
			} else {
				currentMat.mapNorm = value
				currentMat.NormalMap = gfx.NewTexture2D(value, value)
				currentMat.textures = append(currentMat.textures, currentMat.NormalMap)
			}
		}
	}

	l.materials[currentMat.name] = currentMat

	for _, mat := range l.materials {
		if mat.DiffuseMap == nil {
			mat.DiffuseMap = gfx.NewTexture2D("", gfx.White)
			mat.textures = append(mat.textures, mat.DiffuseMap)
		}

		if mat.NormalMap == nil {
			mat.NormalMap = gfx.NewTexture2D("", gfx.White)
			mat.textures = append(mat.textures, mat.NormalMap)
		}

		if mat.SpecularMap == nil {
			mat.SpecularMap = gfx.NewTexture2D("", gfx.White)
			mat.textures = append(mat.textures, mat.SpecularMap)
		}
	}
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
