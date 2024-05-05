package _test

import (
	"context"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"math"
	"reflect"
	"runtime"
	"testing"
	"time"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/_test"
	"tonysoft.com/gfx/obj"
	"unsafe"
)

const (
	winWidth1  = 1000
	winHeight1 = 1000
)

type TestMaterial struct {
	gfx.MaterialBase
	Ambient    mgl32.Vec4
	Diffuse    [4]float32
	SpecPow    float32
	LightCount uint32
	ExtraProps *ExtraProps // will bind as a UBO
}

type ExtraProps struct {
	ExtraProp1 mgl32.Vec3
	ExtraProp2 float32
}

type ShaderBindingTestObject struct {
	gfx.ObjectBase
	shader  gfx.Shader
	binding *gfx.ShaderBinding

	ambientLoc    int32
	diffuseLoc    int32
	specPowLoc    int32
	lightCountLoc int32

	uboName    uint32
	uboDataPtr unsafe.Pointer
	outUboData []float32

	InMaterial    *TestMaterial
	OutAmbient    [4]float32
	OutDiffuse    [4]float32
	OutSpecPow    float32
	OutLightCount uint32
	OutExtraProps ExtraProps
}

func (o *ShaderBindingTestObject) Init() (ok bool) {
	o.binding = gfx.NewShaderBinding(o.shader, o.InMaterial, nil)
	ok = o.binding.Init()

	o.ambientLoc = o.shader.GetUniformLocation("u_Ambient")
	o.diffuseLoc = o.shader.GetUniformLocation("u_Diffuse")
	o.specPowLoc = o.shader.GetUniformLocation("u_SpecPow")
	o.lightCountLoc = o.shader.GetUniformLocation("u_LightCount")

	o.uboName = o.binding.UboNames()[0]
	gl.BindBuffer(gl.UNIFORM_BUFFER, o.uboName)
	o.uboDataPtr = gl.MapBuffer(gl.UNIFORM_BUFFER, gl.READ_ONLY)
	if o.uboDataPtr == nil {
		panic("failed to map buffer")
	}
	gl.UnmapBuffer(gl.UNIFORM_BUFFER)
	o.outUboData = make([]float32, 4)

	return
}

func (o *ShaderBindingTestObject) Update(deltaTime int64) (ok bool) {
	o.binding.Update(deltaTime)

	gl.GetUniformfv(o.shader.GlName(), o.ambientLoc, &o.OutAmbient[0])
	gl.GetUniformfv(o.shader.GlName(), o.diffuseLoc, &o.OutDiffuse[0])
	gl.GetUniformfv(o.shader.GlName(), o.specPowLoc, &o.OutSpecPow)
	gl.GetUniformuiv(o.shader.GlName(), o.lightCountLoc, &o.OutLightCount)

	gl.BindBuffer(gl.UNIFORM_BUFFER, o.uboName)
	o.uboDataPtr = gl.MapBuffer(gl.UNIFORM_BUFFER, gl.READ_ONLY)
	reflect.Copy(reflect.ValueOf(o.outUboData), reflect.ValueOf((*[4]float32)(o.uboDataPtr)[:4]))
	gl.UnmapBuffer(gl.UNIFORM_BUFFER)

	o.OutExtraProps.ExtraProp1[0] = o.outUboData[0]
	o.OutExtraProps.ExtraProp1[1] = o.outUboData[1]
	o.OutExtraProps.ExtraProp1[2] = o.outUboData[2]
	o.OutExtraProps.ExtraProp2 = o.outUboData[3]

	return true
}

func (o *ShaderBindingTestObject) Close() {
	o.binding.Close()
}

func newShaderBindingTestObject(shader gfx.Shader) *ShaderBindingTestObject {
	return &ShaderBindingTestObject{
		shader: shader,
		InMaterial: &TestMaterial{
			ExtraProps: &ExtraProps{},
		},
	}
}

func TestShaderCompilation(t *testing.T) {
	_test.Begin()
	defer _test.End()

	ctx, cancelFunc := context.WithCancel(context.Background())

	win := gfx.NewWindow().
		SetTitle("Shader Compilation").
		SetWidth(_test.WindowWidth).
		SetHeight(_test.WindowHeight)

	gfx.InitWindowAsync(win)

	vertShader := `#version 410 core
in vec3 a_Position;
void main() {
  	gl_Position = vec4(a_Position, 1.0);
}
`

	fragShader := `#version 410 core
out vec4 FragColor;
void main() {
	FragColor = vec4(1,1,1,1);
}
`

	go func() {
		<-win.ReadyChan()
		shader := gfx.NewBasicShader("TestShader", vertShader, fragShader)
		win.InitObject(shader)
		win.CloseObject(shader)
		cancelFunc()
	}()

	gfx.Run(ctx, cancelFunc)
}

func TestShaderBinding(t *testing.T) {
	startRoutineCount := runtime.NumGoroutine()

	_test.Begin()

	ctx, cancelFunc := context.WithCancel(context.Background())

	vertShader := `#version 410 core
in vec3 a_Position;
out vec3 TestColor;
uniform vec4 u_Ambient;
uniform vec4 u_Diffuse;
uniform float u_SpecPow;
uniform uint u_LightCount;

layout (std140) uniform ExtraProps {
    vec4    ExtraProp1;
    float    ExtraProp2;
} u_Props;

void main() {
	// ignore the math, just ensuring they don't get optimized away
	TestColor = vec3(u_Ambient + u_Diffuse * u_SpecPow * u_LightCount * u_Props.ExtraProp1 * u_Props.ExtraProp2);
  	gl_Position = vec4(a_Position, 1.0);
}
`

	fragShader := `#version 410 core
in vec3 TestColor;
out vec4 FragColor;
void main() {
	FragColor = vec4(TestColor.rgb, 1.0);
}
`

	shader := gfx.NewBasicShader("TestShader", vertShader, fragShader)
	testObj := newShaderBindingTestObject(shader)

	go func() {
		win := gfx.NewWindow().
			SetTitle("Shader Compilation").
			SetWidth(_test.WindowWidth).
			SetHeight(_test.WindowHeight)

		gfx.InitWindowAsync(win)
		<-win.ReadyChan()

		win.InitObject(shader)
		win.AddObjects(testObj)

		// *Optional.  Allow the object to get initialized and updated during the Update() cycle
		time.Sleep(100 * time.Millisecond)

		testObj.InMaterial.Lock() // ensure changes aren't made while sending data to VRAM
		testObj.InMaterial.Ambient = mgl32.Vec4{1, 2, 3, 4}
		testObj.InMaterial.Diffuse = [4]float32{5, 6, 7, 8}
		testObj.InMaterial.SpecPow = 32.5
		testObj.InMaterial.LightCount = 3
		testObj.InMaterial.ExtraProps.ExtraProp1 = mgl32.Vec3{11, 22, 33}
		testObj.InMaterial.ExtraProps.ExtraProp2 = 44.4
		testObj.InMaterial.Unlock()

		_test.SleepACoupleFrames() // allow binding (RAM/VRAM IO) to occur during Update() cycle

		cancelFunc()
	}()

	gfx.Run(ctx, cancelFunc)
	_test.End()

	endRoutineCount := runtime.NumGoroutine()
	if endRoutineCount != startRoutineCount {
		t.Logf("Starting routine count: %d", startRoutineCount)
		t.Logf("Ending routine count: %d", endRoutineCount)
		t.Error("routine leak")
	}

	if testObj.OutAmbient[0] != 1 || testObj.OutAmbient[1] != 2 || testObj.OutAmbient[2] != 3 || testObj.OutAmbient[3] != 4 {
		t.Errorf("unexpected ambient value: expected [1 2 3 4], got %v", testObj.OutAmbient)
	}

	if testObj.OutDiffuse[0] != 5 || testObj.OutDiffuse[1] != 6 || testObj.OutDiffuse[2] != 7 || testObj.OutDiffuse[3] != 8 {
		t.Errorf("unexpected diffuse value: expected [5 6 7 8], got %v", testObj.OutDiffuse)
	}

	if testObj.OutSpecPow < 32.4 || testObj.OutSpecPow > 32.6 {
		t.Errorf("unexpected specular power value: expected 32.5, got %f", testObj.OutSpecPow)
	}

	if testObj.OutLightCount != 3 {
		t.Errorf("unexpected light count value: expected 3, got %d", testObj.OutLightCount)
	}

	if testObj.OutExtraProps.ExtraProp1[0] != 11 || testObj.OutExtraProps.ExtraProp1[1] != 22 ||
		testObj.OutExtraProps.ExtraProp1[2] != 33 {
		t.Errorf("unexpected ExtraProp1 value: expected [11 22 33], got %v", testObj.OutExtraProps.ExtraProp1)
	}

	if testObj.OutExtraProps.ExtraProp2 < 44.3 || testObj.OutExtraProps.ExtraProp2 > 44.5 {
		t.Errorf("unexpected ExtraProp2 value: expected 44.4, got %f", testObj.OutExtraProps.ExtraProp2)
	}
}

func TestColorShader(t *testing.T) {
	_test.Begin()
	defer _test.End()

	ctx, cancelFunc := context.WithCancel(context.Background())

	win := gfx.NewWindow().
		SetTitle(_test.WindowTitle).
		SetWidth(winWidth1).
		SetHeight(winHeight1)

	shader := gfx.NewBasicShader("test_shader", _test.ColorVertShader, _test.ColorFragShader)
	win.Assets().Add(shader)

	model := _test.NewColoredQuad()
	model.Meshes()[0].Faces()[0].AttachedMaterial().AttachShader(shader)

	quad := gfx.NewShape3D()
	quad.SetModel(model)

	vp := gfx.NewViewport(win.Width(), win.Height())
	vp.Set(-1, -1, 2, 2)
	quad.SetViewport(vp)

	validator := _test.NewSceneValidator(t, win)

	win.AddObjects(quad, validator)
	gfx.InitWindowAsync(win)

	go func() {
		<-win.ReadyChan()

		validator.AddPixelSampler(func() (float32, float32) { return 0, 0 }, gfx.Magenta, "center screen")
		validator.AddPixelSampler(func() (float32, float32) { return -.5, .5 }, gfx.Magenta, "upper-left quadrant")
		validator.AddPixelSampler(func() (float32, float32) { return .5, .5 }, gfx.Magenta, "upper-right quadrant")
		validator.AddPixelSampler(func() (float32, float32) { return -.5, -.5 }, gfx.Magenta, "bottom-left quadrant")
		validator.AddPixelSampler(func() (float32, float32) { return .5, -.5 }, gfx.Magenta, "bottom-right quadrant")

		_test.SleepAFewFrames()
		validator.Validate()

		time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the initial color

		win.CloseObject(quad)

		model = _test.NewColoredQuad([]color.RGBA{gfx.Green, gfx.Green, gfx.Green, gfx.Green}...)
		model.Meshes()[0].Faces()[0].AttachedMaterial().AttachShader(shader)
		quad.SetModel(model)

		win.InitObject(quad)

		for _, sampler := range validator.Samplers {
			sampler.ExpectedColor = gfx.Green
		}

		_test.SleepAFewFrames()
		validator.Validate()

		time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the color change

		cancelFunc()
	}()

	gfx.Run(ctx, cancelFunc)
}

func TestTextureShader(t *testing.T) {
	_test.Begin()
	defer _test.End()

	ctx, cancelFunc := context.WithCancel(context.Background())

	win := gfx.NewWindow().
		SetTitle(_test.WindowTitle).
		SetWidth(winWidth1).
		SetHeight(winHeight1)

	shader := gfx.NewBasicShader("test_shader", _test.TextureVertShader, _test.TextureFragShader)
	win.Assets().Add(shader)

	textureMagenta := gfx.NewTexture2D("test_texture_magenta", gfx.Magenta)
	textureGreen := gfx.NewTexture2D("test_texture_green", gfx.Green)
	win.Assets().Add(textureMagenta)
	win.Assets().Add(textureGreen)

	model := _test.NewTexturedQuad()

	material := model.Meshes()[0].Faces()[0].AttachedMaterial().(*_test.Material)
	material.AttachShader(shader)
	material.DiffuseMap = textureMagenta

	quad := gfx.NewShape3D()
	quad.SetModel(model)

	vp := gfx.NewViewport(win.Width(), win.Height())
	vp.Set(-1, -1, 2, 2)
	quad.SetViewport(vp)

	validator := _test.NewSceneValidator(t, win)

	win.AddObjects(quad, validator)
	gfx.InitWindowAsync(win)

	go func() {
		<-win.ReadyChan()

		validator.AddPixelSampler(func() (float32, float32) { return 0, 0 }, gfx.Magenta, "center screen")
		validator.AddPixelSampler(func() (float32, float32) { return -.5, .5 }, gfx.Magenta, "upper-left quadrant")
		validator.AddPixelSampler(func() (float32, float32) { return .5, .5 }, gfx.Magenta, "upper-right quadrant")
		validator.AddPixelSampler(func() (float32, float32) { return -.5, -.5 }, gfx.Magenta, "bottom-left quadrant")
		validator.AddPixelSampler(func() (float32, float32) { return .5, -.5 }, gfx.Magenta, "bottom-right quadrant")

		_test.SleepAFewFrames()
		validator.Validate()

		time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the initial color

		win.CloseObject(quad) // changing a material map or shader requires re-initialization
		material.DiffuseMap = textureGreen
		win.InitObject(quad)

		for _, sampler := range validator.Samplers {
			sampler.ExpectedColor = gfx.Green
		}

		_test.SleepAFewFrames()
		validator.Validate()

		time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the color change

		cancelFunc()
	}()

	gfx.Run(ctx, cancelFunc)
}

func TestShape3DShaders(t *testing.T) {
	_test.Begin()
	defer _test.End()

	ctx, cancelFunc := context.WithCancel(context.Background())

	win := gfx.NewWindow().
		SetTitle(_test.WindowTitle).
		SetWidth(winWidth1).
		SetHeight(winHeight1)

	shaderNoLights := win.Assets().Get(gfx.Shape3DNoLightsShader).(gfx.Shader)
	shaderNoNormSpec := win.Assets().Get(gfx.Shape3DNoNormalSpecularMapsShader).(gfx.Shader)
	shaderDefault := win.Assets().Get(gfx.Shape3DShader).(gfx.Shader)

	textureGreen := gfx.NewTexture2D("green_texture", gfx.Green)
	textureWhite := gfx.NewTexture2D("white_texture", gfx.White)
	textureBlack := gfx.NewTexture2D("black_texture", gfx.Black)
	win.Assets().Add(textureGreen)
	win.Assets().Add(textureWhite)
	win.Assets().Add(textureBlack)

	textureNorm := gfx.NewTexture2D("normal_texture", gfx.DefaultNormalMapColor)
	win.Assets().Add(textureNorm)

	textureSpec := gfx.NewTexture2D("specular_texture", gfx.DefaultSpecularMapColor)
	win.Assets().Add(textureSpec)

	model := obj.NewModel("cube_model", "cube.obj")
	model.ComputeTangents(true)
	model.Load()
	win.Assets().Add(model)

	material := model.Meshes()[0].Faces()[0].AttachedMaterial().(*obj.BasicMaterial)
	material.AttachShader(shaderNoLights) // we'll start with the minimal Shape3D shader
	material.DiffuseMap = textureGreen    // the diffuse map is the only map used by this shader

	// Given that we created the following textures using the default
	// normal/specular colors, setting these maps using the code below would
	// be redundant as obj.BasicMaterial does this for us by default...will
	// be confirmed later when we switch to the default Shape3D shader, which
	// uses these maps.
	// material.NormalMap = textureNorm
	// material.SpecularMap = textureSpec

	camera := gfx.NewCamera()
	camera.SetProjection(45, win.AspectRatio(), .1, 1000)
	camera.Properties.Position = mgl32.Vec4{0, 0, 1.8}
	win.AddObjects(camera)

	cube := gfx.NewShape3D()
	cube.SetModel(model)
	cube.SetCamera(camera)

	validator := _test.NewSceneValidator(t, win)

	win.AddObjects(cube, validator)
	gfx.InitWindowAsync(win)

	go func() {
		<-win.ReadyChan()

		expectedColor1 := color.RGBA{G: 127, A: 255} // material.Diffuse(.5) * green_texture(255) + material.Emissive(0)

		validator.AddPixelSampler(func() (float32, float32) { return 0, 0 }, expectedColor1, "center screen")
		validator.AddPixelSampler(func() (float32, float32) { return -.8, .8 }, expectedColor1, "upper-left corner")
		validator.AddPixelSampler(func() (float32, float32) { return .8, .8 }, expectedColor1, "upper-right corner")
		validator.AddPixelSampler(func() (float32, float32) { return -.8, -.8 }, expectedColor1, "bottom-left corner")
		validator.AddPixelSampler(func() (float32, float32) { return .8, -.8 }, expectedColor1, "bottom-right corner")

		validator.Validate()

		time.Sleep(400 * time.Millisecond) // *optional; pause in between each change to give us time to observe

		// Rotate the cube, which will mean the pixel samplers at the corners
		// of the screen will no longer be sampling the cube but rather the
		// window's clear/background color. Additionally, since we're using
		// this minimal shader, rotation will not affect the final fragment
		// color for any view of the cube.
		cube.SetRotationX(math.Pi * .25)
		expectedColor2 := _test.BackgroundColor
		for i := 1; i < len(validator.Samplers); i++ {
			validator.Samplers[i].ExpectedColor = expectedColor2
		}
		_test.SleepAFewFrames()
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		material.Lock()
		material.Properties.Emissive[1] = .5 // add .5 green emissive light
		material.Unlock()
		expectedColor1.G = 255 // material.Diffuse(.5) * green_texture(255) + material.Emissive(.5)
		validator.Samplers[0].ExpectedColor = expectedColor1
		_test.SleepAFewFrames()
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		cube.SetRotationX(0) // reset the cube rotation, which means the corner samplers will be on the cube again
		for _, sampler := range validator.Samplers {
			sampler.ExpectedColor = expectedColor1
		}
		_test.SleepAFewFrames()
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		// With this shader, directional lights, ambient/specular/emissive
		// lighting, etc., is supported.  Changing a material map or the
		// attached shader requires re-initialization.
		material.AttachShader(shaderNoNormSpec)
		material.Lock()
		material.Properties.Emissive[1] = 0 // reset/clear emissive light
		material.Unlock()
		win.CloseObject(cube)
		win.InitObject(cube)
		cube.SetRotationX(0) // reset the cube rotation to simplify validation

		// Ambient light only modulates (does not add to) the base color, which
		// is the sampled diffuse map color modulated by the material's diffuse color.
		expectedColor1.R = 0  // material.Ambient(.2) * material.Diffuse(.5) * green_texture(0) + material.Emissive(0)
		expectedColor1.G = 25 // material.Ambient(.2) * material.Diffuse(.5) * green_texture(255) + material.Emissive(0)
		expectedColor1.B = 0  // material.Ambient(.2) * material.Diffuse(.5) * green_texture(0) + material.Emissive(0)
		for _, sampler := range validator.Samplers {
			sampler.ExpectedColor = expectedColor1
		}
		_test.SleepAFewFrames()
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		// Although we are now using a shader that supports lights, we haven't
		// added any yet.  Here, we use a lighting object that supports up to
		// four directional lights.
		lighting := gfx.NewQuadDirectionalLighting()
		lighting.Lights[0].Color = mgl32.Vec3{1, 1, 1}
		lighting.Lights[0].Direction[2] = -1
		lighting.LightCount = 1
		cube.SetLighting(lighting)

		// With white light aimed directly at the center-screen (maximum
		// diffuse/specular power at that location) and the material's
		// diffuse/specular color is at 50% white and ambient light at
		// 20% white, that comes to 127 for red/blue and 255 for green.
		// For the red/blue channels, they get their intensity only from
		// the directional specular light while the green channel gets
		// that same intensity plus what was sampled from the diffuse map.
		expectedColor1.R = 127
		expectedColor1.G = 255
		expectedColor1.B = 127
		validator.Samplers[0].ExpectedColor = expectedColor1

		// Things are different as you move away from the center, as the
		// diffuse/specular power of the directional light decreases as
		// the angle between the light direction and surface normal increases.
		expectedColor2 = color.RGBA{R: 0, G: 153, B: 0, A: 255}
		for i := 1; i < len(validator.Samplers); i++ {
			validator.Samplers[i].ExpectedColor = expectedColor2
		}

		_test.SleepAFewFrames()
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		// Now we attach the full-featured Shape3D shader, which
		// adds support for normal and specular maps.
		win.CloseObject(cube)
		material.AttachShader(shaderDefault)
		win.InitObject(cube)

		time.Sleep(400 * time.Millisecond)

		// The default specular map assigned to obj.BasicMaterial
		// has a medium intensity (i.e., the color value for every
		// pixel, every color channel, is 127 or 0.5).  That means
		// the center-screen pixel sampler must be adjusted to account
		// for the lower specular light.  The red/blue channels are
		// now only getting half of the directional specular light
		// and only half of that amount as per the material's specular
		// setting.  The green channel gets its color from the sampled
		// green texture (255 or 1.0), modulated down to 127 by the
		// material's diffuse setting (0.5) and further down to 25 by
		// the material's ambient setting (0.2), but then boosted back
		// up to 153 when lit by the diffuse component of the directional
		// light and then finally up to 216 when adding the same specular
		// light given to the red/blue channels.
		expectedColor1.R = 63
		expectedColor1.G = 216
		expectedColor1.B = 63
		validator.Samplers[0].ExpectedColor = expectedColor1
		_test.SleepAFewFrames()
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		// By using a specular map with maximum "shininess," that
		// center-screen pixel will go back to being as lit as it
		// was before using this shader.
		win.CloseObject(cube)
		material.SpecularMap = textureWhite
		win.InitObject(cube)
		expectedColor1.R = 127
		expectedColor1.G = 255
		expectedColor1.B = 127
		validator.Samplers[0].ExpectedColor = expectedColor1
		_test.SleepAFewFrames()
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		// The default normal map has every pixel set to a color that
		// represents no modulation to the normals provided in the mesh's
		// vertex buffer.  By replacing this map with the green texture,
		// the normals expressed in the map will be parallel to the
		// mesh face (in tangent space) and, more precisely, perfectly
		// aligned with the Y (or V) axis.  That effectively eliminates
		// the effect of the directional light.
		win.CloseObject(cube)
		material.NormalMap = textureGreen
		win.InitObject(cube)
		expectedColor1.R = 0
		expectedColor1.G = 25
		expectedColor1.B = 0
		for _, sample := range validator.Samplers {
			sample.ExpectedColor = expectedColor1
		}
		_test.SleepAFewFrames()
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		// Ensure using a normal map that we provide which was
		// initialized to the default color works as intended...
		win.CloseObject(cube)
		material.NormalMap = textureNorm
		win.InitObject(cube)
		expectedColor1.R = 127
		expectedColor1.G = 255
		expectedColor1.B = 127
		validator.Samplers[0].ExpectedColor = expectedColor1
		expectedColor2 = color.RGBA{R: 0, G: 153, B: 0, A: 255}
		for i := 1; i < len(validator.Samplers); i++ {
			validator.Samplers[i].ExpectedColor = expectedColor2
		}
		_test.SleepAFewFrames()
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		// Ensure using a specular map that we provide which was
		// initialized to the default color works as intended...
		win.CloseObject(cube)
		material.SpecularMap = textureSpec
		win.InitObject(cube)
		expectedColor1.R = 63
		expectedColor1.G = 216
		expectedColor1.B = 63
		validator.Samplers[0].ExpectedColor = expectedColor1
		expectedColor2 = color.RGBA{R: 0, G: 153, B: 0, A: 255}
		for i := 1; i < len(validator.Samplers); i++ {
			validator.Samplers[i].ExpectedColor = expectedColor2
		}
		_test.SleepAFewFrames()
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		// Time to go black...
		win.CloseObject(cube)
		material.SpecularMap = textureBlack                           // no specular light
		material.NormalMap = textureGreen                             // no direct light
		material.Properties.Ambient = gfx.RgbaToFloatArray(gfx.Black) // no ambient light
		win.InitObject(cube)

		expectedColor1 = gfx.Black
		for _, sample := range validator.Samplers {
			sample.ExpectedColor = expectedColor1
		}

		_test.SleepAFewFrames()
		validator.Validate()

		time.Sleep(400 * time.Millisecond)

		cancelFunc()
	}()

	gfx.Run(ctx, cancelFunc)
}
