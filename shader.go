package gfx

import (
	_ "embed"
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	SignalShaderProgram        = "signal"
	TextShaderProgram          = "text"
	ShapeShaderProgram         = "shape"
	ShapeTexturedShaderProgram = "shape_tex"
	BlurXShaderProgram         = "blur_x"
	BlurYShaderProgram         = "blur_y"
	TextureShaderProgram       = "texture"
	ModelShaderProgram         = "model"
	ModelNoLightShaderProgram  = "model_nolit"
)

var (
	//go:embed shaders/signal.vs
	signalShaderVs string

	//go:embed shaders/signal.gs
	signalShaderGs string

	//go:embed shaders/signal.fs
	signalShaderFs string

	//go:embed shaders/text.vs
	textShaderVs string

	//go:embed shaders/text.fs
	textShaderFs string

	//go:embed shaders/shape.vs
	shapeShaderVs string

	//go:embed shaders/shape.fs
	shapeShaderFs string

	//go:embed shaders/shape_tex.vs
	shapeTexShaderVs string

	//go:embed shaders/shape_tex.fs
	shapeTexShaderFs string

	//go:embed shaders/blur_x.fs
	blurXShaderFs string

	//go:embed shaders/blur_y.fs
	blurYShaderFs string

	//go:embed shaders/texture.vs
	textureShaderVs string

	//go:embed shaders/texture.fs
	textureShaderFs string

	//go:embed shaders/model.vs
	modelShaderVs string

	//go:embed shaders/model.fs
	modelShaderFs string

	//go:embed shaders/model_nolit.vs
	modelNoLightShaderVs string

	//go:embed shaders/model_nolit.fs
	modelNoLightShaderFs string
)

var (
	_shaders           map[string]uint32
	shadersMutex       sync.Mutex
	shadersInitialized atomic.Bool
)

func initShaders() error {
	shadersMutex.Lock()
	defer shadersMutex.Unlock()

	if shadersInitialized.Load() {
		return nil
	}

	signalProg, err := CreateShaderProgram(signalShaderVs, signalShaderGs, signalShaderFs)
	if err != nil {
		return err
	}

	textProg, err := CreateShaderProgram(textShaderVs, "", textShaderFs)
	if err != nil {
		return err
	}

	shapeProg, err := CreateShaderProgram(shapeShaderVs, "", shapeShaderFs)
	if err != nil {
		return err
	}

	shapeTexProg, err := CreateShaderProgram(shapeTexShaderVs, "", shapeTexShaderFs)
	if err != nil {
		return err
	}

	blurXProg, err := CreateShaderProgram(textureShaderVs, "", blurXShaderFs)
	if err != nil {
		return err
	}

	blurYProg, err := CreateShaderProgram(textureShaderVs, "", blurYShaderFs)
	if err != nil {
		return err
	}

	textureProg, err := CreateShaderProgram(textureShaderVs, "", textureShaderFs)
	if err != nil {
		return err
	}

	modelProg, err := CreateShaderProgram(modelShaderVs, "", modelShaderFs)
	if err != nil {
		return err
	}

	modelNoLightProg, err := CreateShaderProgram(modelNoLightShaderVs, "", modelNoLightShaderFs)
	if err != nil {
		return err
	}

	_shaders = make(map[string]uint32)
	_shaders[SignalShaderProgram] = signalProg
	_shaders[TextShaderProgram] = textProg
	_shaders[ShapeShaderProgram] = shapeProg
	_shaders[ShapeTexturedShaderProgram] = shapeTexProg
	_shaders[BlurXShaderProgram] = blurXProg
	_shaders[BlurYShaderProgram] = blurYProg
	_shaders[TextureShaderProgram] = textureProg
	_shaders[ModelShaderProgram] = modelProg
	_shaders[ModelNoLightShaderProgram] = modelNoLightProg

	shadersInitialized.Store(true)
	return nil
}

func checkShaderError(shader uint32) error {
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return fmt.Errorf("failed to compile shader: %s", log)
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

		return fmt.Errorf("failed to link program: %s", log)
	}
	return nil
}

func CreateShaderProgram(vertexSource, geometrySource, fragmentSource string) (uint32, error) {
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	cstr, free := gl.Strs(vertexSource + "\x00")
	gl.ShaderSource(vertexShader, 1, cstr, nil)
	free()
	gl.CompileShader(vertexShader)
	if err := checkShaderError(vertexShader); err != nil {
		return 0, err
	}

	var geometryShader uint32
	if geometrySource != "" {
		geometryShader = gl.CreateShader(gl.GEOMETRY_SHADER)
		cstr, free = gl.Strs(geometrySource + "\x00")
		gl.ShaderSource(geometryShader, 1, cstr, nil)
		free()
		gl.CompileShader(geometryShader)
		if err := checkShaderError(geometryShader); err != nil {
			return 0, err
		}
	}

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	cstr, free = gl.Strs(fragmentSource + "\x00")
	gl.ShaderSource(fragmentShader, 1, cstr, nil)
	free()
	gl.CompileShader(fragmentShader)
	if err := checkShaderError(fragmentShader); err != nil {
		return 0, err
	}

	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, vertexShader)
	if geometrySource != "" {
		gl.AttachShader(shaderProgram, geometryShader)
	}
	gl.AttachShader(shaderProgram, fragmentShader)
	gl.LinkProgram(shaderProgram)
	if err := checkProgramError(shaderProgram); err != nil {
		return 0, err
	}

	gl.DeleteShader(vertexShader)
	if geometrySource != "" {
		gl.DeleteShader(geometryShader)
	}
	gl.DeleteShader(fragmentShader)

	return shaderProgram, nil
}

func GetShaderProgram(name string) uint32 {
	if !shadersInitialized.Load() {
		return 0
	}

	shadersMutex.Lock()
	if s, ok := _shaders[name]; ok {
		shadersMutex.Unlock()
		return s
	}
	shadersMutex.Unlock()
	return 0
}
