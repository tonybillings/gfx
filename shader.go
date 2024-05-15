package gfx

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tonybillings/gfx/shaders"
	"golang.org/x/image/math/f32"
	"os"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

const (
	// SignalShader Used by SignalLine to render the line itself, incorporating
	// a geometry shader for line thickness.
	SignalShader = "_shader_signal"

	// BlurXShader Used by Shape2D when blur is enabled to render the horizontally
	// blurred shape.
	BlurXShader = "_shader_blur_x"

	// BlurYShader Used by Shape2D when blur is enabled to render the vertically
	// blurred shape.
	BlurYShader = "_shader_blur_y"

	// TextureShader Used by Shape2D when blur is enabled to render the final,
	// horizontally/vertically blurred shape, but can also be used to render any
	// texture using normalized device/screen coordinates for the vertex data.
	TextureShader = "_shader_texture"

	// Shape2DShader Used by Shape2D to render a textured, two-dimensional
	// shape.
	Shape2DShader = "_shader_shape2d"

	// Shape2DNoTextureShader Used by Shape2D to render a single-colored,
	// two-dimensional shape.
	Shape2DNoTextureShader = "_shader_shape2d_no_texture"

	// Shape3DShader Can be used by Shape3D to render a textured Model with
	// support for: ambient/diffuse/specular/emissive/transparent lighting,
	// directional lights, and diffuse/normal/specular maps.  Expects the Model
	// vertex buffer to have the PositionNormalUvTangentsVaoLayout.
	Shape3DShader = "_shader_shape3d"

	// Shape3DNoNormalSpecularMapsShader Can be used by Shape3D to render a
	// textured Model with support for: ambient/diffuse/specular/emissive/transparent
	// lighting, directional lights, and diffuse maps.  Expects the Model vertex
	// buffer to have the PositionNormalUvVaoLayout.
	Shape3DNoNormalSpecularMapsShader = "_shader_shape3d_no_norm_spec"

	// Shape3DNoLightsShader Can be used by Shape3D to render a
	// textured Model with support for: diffuse/emissive/transparent
	// lighting and diffuse maps.  Expects the Model vertex buffer to have
	// the PositionUvVaoLayout.
	Shape3DNoLightsShader = "_shader_shape3d_no_lights"
)

/******************************************************************************
 Shader
******************************************************************************/

type Shader interface {
	GlAsset
	Activate()
	GetAttribLocation(name string) int32
	GetUniformLocation(name string) int32
	GetUniformBlockIndex(name string) uint32
}

/******************************************************************************
 ShaderBase
******************************************************************************/

type ShaderBase struct {
	AssetBase
	glName uint32
}

func (s *ShaderBase) GlName() uint32 {
	return s.glName
}

func (s *ShaderBase) Activate() {
	gl.UseProgram(s.glName)
}

func (s *ShaderBase) GetAttribLocation(name string) int32 {
	return gl.GetAttribLocation(s.glName, gl.Str(name+"\x00"))
}

func (s *ShaderBase) GetUniformLocation(name string) int32 {
	return gl.GetUniformLocation(s.glName, gl.Str(name+"\x00"))
}

func (s *ShaderBase) GetUniformBlockIndex(name string) uint32 {
	return gl.GetUniformBlockIndex(s.glName, gl.Str(name+"\x00"))
}

/******************************************************************************
 BasicShader
******************************************************************************/

type BasicShader struct {
	ShaderBase

	vsSource any
	gsSource any
	fsSource any
}

/******************************************************************************
 Asset Implementation
******************************************************************************/

func (s *BasicShader) Init() (ok bool) {
	if s.Initialized() {
		return true
	}

	var vsShader, fsShader, gsShader uint32
	switch s.vsSource.(type) {
	case []byte:
		vsShader = s.loadShaderFromSlice(gl.VERTEX_SHADER, s.vsSource.([]byte))
		fsShader = s.loadShaderFromSlice(gl.FRAGMENT_SHADER, s.fsSource.([]byte))
		if s.gsSource != nil {
			gsShader = s.loadShaderFromSlice(gl.GEOMETRY_SHADER, s.gsSource.([]byte))
		}
	case string:
		vsShader = s.loadShaderFromFile(gl.VERTEX_SHADER, s.vsSource.(string))
		if vsShader == 0 {
			vsShader = s.loadShaderFromString(gl.VERTEX_SHADER, s.vsSource.(string))
		}

		fsShader = s.loadShaderFromFile(gl.FRAGMENT_SHADER, s.fsSource.(string))
		if fsShader == 0 {
			fsShader = s.loadShaderFromString(gl.FRAGMENT_SHADER, s.fsSource.(string))
		}

		if s.gsSource != nil {
			gsShader = s.loadShaderFromFile(gl.GEOMETRY_SHADER, s.gsSource.(string))
			if gsShader == 0 {
				gsShader = s.loadShaderFromString(gl.GEOMETRY_SHADER, s.gsSource.(string))
			}
		}
	default:
		panic("unexpected error: source type is not supported")
	}

	s.load(vsShader, fsShader, gsShader)

	gl.DeleteShader(vsShader)
	gl.DeleteShader(fsShader)
	gl.DeleteShader(gsShader)

	return s.AssetBase.Init()
}

func (s *BasicShader) Close() {
	if !s.Initialized() {
		return
	}

	gl.DeleteProgram(s.glName)
	s.glName = 0

	s.AssetBase.Close()
}

/******************************************************************************
 BasicShader Functions
******************************************************************************/

func (s *BasicShader) loadShaderFromSlice(shaderType uint32, slice []byte) uint32 {
	if len(slice) == 0 {
		return 0
	}
	return s.loadShaderFromString(shaderType, string(slice))
}

func (s *BasicShader) loadShaderFromFile(shaderType uint32, name string) uint32 {
	reader, closeFunc := s.getSourceReader(name)
	defer closeFunc()
	if reader == nil {
		return 0
	}

	shaderBytes := make([]byte, reader.Size())

	sb := strings.Builder{}
	for {
		if n, err := reader.Read(shaderBytes); n > 0 && err == nil {
			sb.Write(shaderBytes[:n])
		} else {
			if err != nil {
				panic(fmt.Errorf("shader file read error: %w", err))
			}
			break
		}
	}
	if sb.Len() == 0 {
		panic("unable to read shader file into memory")
	}

	return s.loadShaderFromString(shaderType, sb.String())
}

func (s *BasicShader) loadShaderFromString(shaderType uint32, code string) uint32 {
	shader := gl.CreateShader(shaderType)
	cString, free := gl.Strs(code + "\x00")
	gl.ShaderSource(shader, 1, cString, nil)
	free()
	gl.CompileShader(shader)
	if err := checkShaderError(shader); err != nil {
		panic(fmt.Errorf("%w\nsource: \n\n%s", err, code))
	}
	return shader
}

func (s *BasicShader) load(vsShader, fsShader, gsShader uint32) {
	shaderProgram := gl.CreateProgram()

	gl.AttachShader(shaderProgram, vsShader)
	gl.AttachShader(shaderProgram, fsShader)
	if gsShader > 0 {
		gl.AttachShader(shaderProgram, gsShader)
	}

	gl.LinkProgram(shaderProgram)
	if err := checkProgramError(shaderProgram); err != nil {
		panic(fmt.Errorf("link shader program error: %w", err))
	}

	s.glName = shaderProgram
}

/******************************************************************************
 New BasicShader Function
******************************************************************************/

func NewBasicShader[T ShaderSource](name string, vertexShaderSource, fragmentShaderSource T, geometryShaderSource ...T) *BasicShader {
	var gsSource any
	if len(geometryShaderSource) > 0 {
		gsSource = geometryShaderSource[0]
	}

	return &BasicShader{
		ShaderBase: ShaderBase{
			AssetBase: AssetBase{
				name: name,
			},
		},
		vsSource: vertexShaderSource,
		gsSource: gsSource,
		fsSource: fragmentShaderSource,
	}
}

/******************************************************************************
 Initialize Default Shaders
******************************************************************************/

func newDefaultShader(assetName string, filenames ...string) *BasicShader {
	fs := shaders.Assets
	var vsFile, fsFile, gsFile []byte
	var err error

	if len(filenames) == 0 {
		panic("must supply one to three filenames")
	}

	if len(filenames) == 1 {
		filename := filenames[0]

		if vsFile, err = fs.ReadFile(filename + "_vert.glsl"); err != nil {
			panic(fmt.Errorf("vertex shader read error: %w", err))
		}
		if fsFile, err = fs.ReadFile(filename + "_frag.glsl"); err != nil {
			panic(fmt.Errorf("fragment shader read error: %w", err))
		}
		if gsFile, err = fs.ReadFile(filename + "_geom.glsl"); err != nil && !errors.Is(err, os.ErrNotExist) {
			panic(fmt.Errorf("geometry shader read error: %w", err))
		}
	} else {
		if vsFile, err = fs.ReadFile(filenames[0] + "_vert.glsl"); err != nil {
			panic(fmt.Errorf("vertex shader read error: %w", err))
		}
		if fsFile, err = fs.ReadFile(filenames[1] + "_frag.glsl"); err != nil {
			panic(fmt.Errorf("fragment shader read error: %w", err))
		}

		if len(filenames) == 3 {
			if gsFile, err = fs.ReadFile(filenames[2] + "_geom.glsl"); err != nil {
				panic(fmt.Errorf("geometry shader read error: %w", err))
			}
		}
	}

	shader := NewBasicShader(assetName, vsFile, fsFile, gsFile)
	shader.SetProtected(true)
	return shader
}

func addDefaultShaders(lib *AssetLibrary) {
	const prefix = "_shader_"
	const pfxLen = len(prefix)

	lib.Add(newDefaultShader(SignalShader, SignalShader[pfxLen:]))
	lib.Add(newDefaultShader(BlurXShader, TextureShader[pfxLen:], BlurXShader[pfxLen:]))
	lib.Add(newDefaultShader(BlurYShader, TextureShader[pfxLen:], BlurYShader[pfxLen:]))
	lib.Add(newDefaultShader(TextureShader, TextureShader[pfxLen:]))
	lib.Add(newDefaultShader(Shape2DShader, Shape2DShader[pfxLen:]))
	lib.Add(newDefaultShader(Shape2DNoTextureShader, Shape2DNoTextureShader[pfxLen:]))
	lib.Add(newDefaultShader(Shape3DShader, Shape3DShader[pfxLen:]))
	lib.Add(newDefaultShader(Shape3DNoNormalSpecularMapsShader, Shape3DNoNormalSpecularMapsShader[pfxLen:]))
	lib.Add(newDefaultShader(Shape3DNoLightsShader, Shape3DNoLightsShader[pfxLen:]))
}

/******************************************************************************
 Error Checking
******************************************************************************/

func checkShaderError(shader uint32) error {
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return fmt.Errorf("compile shader error: %s", log)
	}
	return nil
}

func checkProgramError(program uint32) error {
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return fmt.Errorf("link program error: %s", log)
	}
	return nil
}

/******************************************************************************
 ShaderBinding
******************************************************************************/

type ShaderBinding struct {
	ObjectBase

	shader              Shader
	shaderName          uint32
	boundStruct         any
	getBindingPointFunc func() uint32

	bindingPoint uint32
	textureCount uint32
	uboNames     []uint32

	updateFunc  func()
	updateFuncs []func()

	closeFunc  func()
	closeFuncs []func()
}

/******************************************************************************
 ShaderBinding Functions
******************************************************************************/

func (b *ShaderBinding) activate() {
	gl.UseProgram(b.shaderName)
}

func (b *ShaderBinding) bindStruct(uniformName string, nestedStruct reflect.Value) {
	blockIndex := gl.GetUniformBlockIndex(b.shaderName, gl.Str(uniformName+"\x00"))
	if blockIndex == gl.INVALID_INDEX {
		return
	}

	bindingPoint := uint32(0)
	if b.getBindingPointFunc == nil {
		bindingPoint = b.bindingPoint
		b.bindingPoint++
	} else {
		bindingPoint = b.getBindingPointFunc()
	}

	ptr := unsafe.Pointer(nestedStruct.Elem().UnsafeAddr())
	bufferSize := int(nestedStruct.Elem().Type().Size())

	var ubo uint32
	gl.GenBuffers(1, &ubo)
	gl.BindBuffer(gl.UNIFORM_BUFFER, ubo)
	gl.BufferData(gl.UNIFORM_BUFFER, bufferSize, nil, gl.DYNAMIC_DRAW)

	b.uboNames = append(b.uboNames, ubo)

	b.updateFuncs = append(b.updateFuncs, func() {
		gl.BindBuffer(gl.UNIFORM_BUFFER, ubo)
		gl.BufferSubData(gl.UNIFORM_BUFFER, 0, bufferSize, ptr)
		gl.BindBufferBase(gl.UNIFORM_BUFFER, bindingPoint, ubo)
		gl.UniformBlockBinding(b.shaderName, blockIndex, bindingPoint)
	})

	b.closeFuncs = append(b.closeFuncs, func() {
		gl.DeleteBuffers(1, &ubo)
	})
}

func (b *ShaderBinding) bindStructFields(struct_ reflect.Value, uniformNamePrefix string) {
	for i := 0; i < struct_.NumField(); i++ {
		field := struct_.Type().Field(i)
		if !field.IsExported() {
			continue
		}
		b.bindField(struct_.Field(i), field.Name, uniformNamePrefix)
	}
}

func (b *ShaderBinding) bindField(field reflect.Value, name, uniformNamePrefix string) {
	const invalidLoc = -1

	if uniformNamePrefix == "" {
		uniformNamePrefix = "u_"
	}

	uniformName := uniformNamePrefix + name
	uniformLoc := gl.GetUniformLocation(b.shaderName, gl.Str(uniformName+"\x00"))
	fieldKind := field.Kind()
	if uniformLoc == invalidLoc {
		if fieldKind != reflect.Struct && fieldKind != reflect.Pointer && fieldKind != reflect.Interface &&
			(fieldKind != reflect.Array || field.Type().Elem().Kind() != reflect.Struct || field.Len() == 0) {
			return
		}
	}

	switch fieldKind {
	case reflect.Array:
		if field.Type().Elem().Kind() == reflect.Struct {
			for i := 0; i < field.Len(); i++ {
				b.bindStructFields(field.Index(i), fmt.Sprintf("u_%s[%d].", name, i))
			}
		} else {
			switch field.Interface().(type) {
			case mgl32.Mat4:
				ptr := &getPointer[mgl32.Mat4](field)[0]
				b.updateFuncs = append(b.updateFuncs, func() {
					gl.UniformMatrix4fv(uniformLoc, 1, false, ptr)
				})
			case f32.Mat4:
				ptr := &getPointer[f32.Mat4](field)[0]
				b.updateFuncs = append(b.updateFuncs, func() {
					gl.UniformMatrix4fv(uniformLoc, 1, false, ptr)
				})
			case [16]float32:
				ptr := &getPointer[[16]float32](field)[0]
				b.updateFuncs = append(b.updateFuncs, func() {
					gl.UniformMatrix4fv(uniformLoc, 1, false, ptr)
				})
			case mgl32.Mat3:
				ptr := &getPointer[mgl32.Mat3](field)[0]
				b.updateFuncs = append(b.updateFuncs, func() {
					gl.UniformMatrix3fv(uniformLoc, 1, false, ptr)
				})
			case f32.Mat3:
				ptr := &getPointer[f32.Mat3](field)[0]
				b.updateFuncs = append(b.updateFuncs, func() {
					gl.UniformMatrix3fv(uniformLoc, 1, false, ptr)
				})
			case [9]float32:
				ptr := &getPointer[[9]float32](field)[0]
				b.updateFuncs = append(b.updateFuncs, func() {
					gl.UniformMatrix3fv(uniformLoc, 1, false, ptr)
				})
			case mgl32.Vec4:
				ptr := &getPointer[mgl32.Vec4](field)[0]
				b.updateFuncs = append(b.updateFuncs, func() {
					gl.Uniform4fv(uniformLoc, 1, ptr)
				})
			case f32.Vec4:
				ptr := &getPointer[f32.Vec4](field)[0]
				b.updateFuncs = append(b.updateFuncs, func() {
					gl.Uniform4fv(uniformLoc, 1, ptr)
				})
			case [4]float32:
				ptr := &getPointer[[4]float32](field)[0]
				b.updateFuncs = append(b.updateFuncs, func() {
					gl.Uniform4fv(uniformLoc, 1, ptr)
				})
			case mgl32.Vec3:
				ptr := &getPointer[mgl32.Vec3](field)[0]
				b.updateFuncs = append(b.updateFuncs, func() {
					gl.Uniform3fv(uniformLoc, 1, ptr)
				})
			case f32.Vec3:
				ptr := &getPointer[f32.Vec3](field)[0]
				b.updateFuncs = append(b.updateFuncs, func() {
					gl.Uniform3fv(uniformLoc, 1, ptr)
				})
			case [3]float32:
				ptr := &getPointer[[3]float32](field)[0]
				b.updateFuncs = append(b.updateFuncs, func() {
					gl.Uniform3fv(uniformLoc, 1, ptr)
				})
			case mgl32.Vec2:
				ptr := &getPointer[mgl32.Vec2](field)[0]
				b.updateFuncs = append(b.updateFuncs, func() {
					gl.Uniform2fv(uniformLoc, 1, ptr)
				})
			case f32.Vec2:
				ptr := &getPointer[f32.Vec2](field)[0]
				b.updateFuncs = append(b.updateFuncs, func() {
					gl.Uniform2fv(uniformLoc, 1, ptr)
				})
			case [2]float32:
				ptr := &getPointer[[2]float32](field)[0]
				b.updateFuncs = append(b.updateFuncs, func() {
					gl.Uniform2fv(uniformLoc, 1, ptr)
				})
			}
		}
	case reflect.Float32:
		ptr := getPointer[float32](field)
		b.updateFuncs = append(b.updateFuncs, func() {
			gl.Uniform1f(uniformLoc, *ptr)
		})
	case reflect.Int32:
		ptr := getPointer[int32](field)
		b.updateFuncs = append(b.updateFuncs, func() {
			gl.Uniform1i(uniformLoc, *ptr)
		})
	case reflect.Uint32:
		ptr := getPointer[uint32](field)
		b.updateFuncs = append(b.updateFuncs, func() {
			gl.Uniform1ui(uniformLoc, *ptr)
		})
	case reflect.Struct:
		b.bindStructFields(field, fmt.Sprintf("u_%s.", name))
	case reflect.Pointer:
		if field.IsNil() {
			return
		}
		if name == "Properties" {
			boundStructName := fmt.Sprintf("%T", b.boundStruct)
			boundStructNameParts := strings.Split(boundStructName, ".")
			name = boundStructNameParts[len(boundStructNameParts)-1]
		}
		b.bindStruct(name, field)
	case reflect.Interface:
		if field.IsNil() {
			return
		}
		switch fieldAsType := field.Interface().(type) {
		case Texture:
			glName := fieldAsType.GlName()
			unit := b.textureCount
			b.updateFuncs = append(b.updateFuncs, func() {
				gl.ActiveTexture(gl.TEXTURE0 + unit)
				gl.BindTexture(gl.TEXTURE_2D, glName)
				gl.Uniform1i(uniformLoc, int32(unit))
			})
			b.textureCount++
		default:
			b.bindStructFields(field, fmt.Sprintf("u_%s.", name))
		}
	}
}

func (b *ShaderBinding) initFuncs() {
	if locker, ok := b.boundStruct.(sync.Locker); ok {
		b.updateFunc = func() {
			locker.Lock()
			for _, updateFunc := range b.updateFuncs {
				updateFunc()
			}
			locker.Unlock()
		}
		b.closeFunc = func() {
			locker.Lock()
			for _, closeFunc := range b.closeFuncs {
				closeFunc()
			}
			locker.Unlock()
		}
	} else {
		b.updateFunc = func() {
			for _, updateFunc := range b.updateFuncs {
				updateFunc()
			}
		}
		b.closeFunc = func() {
			for _, closeFunc := range b.closeFuncs {
				closeFunc()
			}
		}
	}
}

func (b *ShaderBinding) Init() (ok bool) {
	if !b.shader.Initialized() {
		b.shader.Init()
	}
	b.shaderName = b.shader.GlName()

	b.bindStructFields(reflect.Indirect(reflect.ValueOf(b.boundStruct)), "")
	b.initFuncs()

	return b.ObjectBase.Init()
}

func (b *ShaderBinding) Update(_ int64) (ok bool) {
	b.activate()
	b.updateFunc()
	return true
}

func (b *ShaderBinding) Close() {
	b.closeFunc()
	b.ObjectBase.Close()
}

func (b *ShaderBinding) UboNames() []uint32 {
	return b.uboNames
}

func NewShaderBinding(shader Shader, boundStruct any, getBindingPointFunc func() uint32) *ShaderBinding {
	return &ShaderBinding{
		shader:              shader,
		boundStruct:         boundStruct,
		getBindingPointFunc: getBindingPointFunc,
	}
}

/******************************************************************************
 ShaderBinder
******************************************************************************/

type ShaderBinder struct {
	ObjectBase
	bindings    []*ShaderBinding
	boundStruct any
}

func (b *ShaderBinder) Init() (ok bool) {
	ok = true
	for _, binding := range b.bindings {
		ok = ok && binding.Init()
	}
	return ok && b.ObjectBase.Init()
}

func (b *ShaderBinder) Update(deltaTime int64) (ok bool) {
	ok = true
	for _, binding := range b.bindings {
		ok = ok && binding.Update(deltaTime)
	}
	return
}

func (b *ShaderBinder) Close() {
	for _, binding := range b.bindings {
		binding.Close()
	}
	b.ObjectBase.Close()
}

func NewShaderBinder(shaders map[uint32]Shader, boundStruct any, getBindingPointFunc func() uint32) *ShaderBinder {
	binder := &ShaderBinder{}
	for _, shader := range shaders {
		binding := NewShaderBinding(shader, boundStruct, getBindingPointFunc)
		binder.bindings = append(binder.bindings, binding)
	}
	return binder
}

/******************************************************************************
 Utility Functions
******************************************************************************/

func getPointer[T any](val reflect.Value) *T {
	return (*T)(unsafe.Pointer(val.Addr().Pointer()))
}
