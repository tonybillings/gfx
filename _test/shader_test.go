package _test

import (
	"context"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"reflect"
	"runtime"
	"testing"
	"time"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/_test"
	"unsafe"
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
