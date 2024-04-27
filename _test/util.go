package _test

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/google/uuid"
	"image/color"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
	"tonysoft.com/gfx"
)

type SceneValidator struct {
	gfx.DrawableObjectBase

	window    *gfx.Window
	validated atomic.Bool

	Samplers []*PixelSampler
	Errors   []error
}

type PixelSampler struct {
	window               *gfx.Window
	getPixelPositionFunc func() (posX, posY float32)

	PosX, PosY    float32
	Description   string
	ExpectedColor color.RGBA
	ActualColor   color.RGBA
	IsValid       bool
}

func (s *PixelSampler) Sample() (isValid bool) {
	s.PosX, s.PosY = s.getPixelPositionFunc()

	x := int32(float32(s.window.Width()) * ((s.PosX + 1.0) / 2.0))
	y := int32(float32(s.window.Height()) * ((s.PosY + 1.0) / 2.0))

	var pixel [4]uint8
	gl.ReadPixels(x, y, 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixel[:]))

	s.ActualColor.R = pixel[0]
	s.ActualColor.G = pixel[1]
	s.ActualColor.B = pixel[2]
	s.ActualColor.A = pixel[3]

	s.IsValid = s.ActualColor == s.ExpectedColor
	return s.IsValid
}

func (s *PixelSampler) String() string {
	if s.IsValid {
		return fmt.Sprintf("the sample passed validation")
	} else {
		return fmt.Sprintf("the sample failed validation: expected %v, got %v: description: %s",
			s.ExpectedColor, s.ActualColor, s.Description)
	}
}

func (v *SceneValidator) AddPixelSampler(getPixelPositionFunc func() (posX, posY float32), expectedColor color.RGBA, description ...string) {
	desc := "(none provided)"
	if len(description) > 0 {
		desc = description[0]
	}

	sampler := &PixelSampler{}
	sampler.window = v.window
	sampler.getPixelPositionFunc = getPixelPositionFunc
	sampler.ExpectedColor = expectedColor
	sampler.Description = desc

	v.Samplers = append(v.Samplers, sampler)
}

func (v *SceneValidator) Validate() {
	v.Errors = make([]error, 0)
	for _, s := range v.Samplers {
		if isValid := s.Sample(); !isValid {
			v.Errors = append(v.Errors, fmt.Errorf(s.String()))
		}
	}
	if len(v.Errors) == 0 {
		v.Errors = nil
	}
}

func (v *SceneValidator) Draw(_ int64) (ok bool) {
	if !v.validated.Load() {
		v.Validate()
		v.validated.Store(true)
		return true
	}

	return true
}

func (v *SceneValidator) Reset() {
	v.validated.Store(false)
}

func NewSceneValidator(window *gfx.Window) *SceneValidator {
	v := &SceneValidator{
		window: window,
	}
	v.SetName(uuid.New().String())
	return v
}

func ValidateScene(t *testing.T, v *SceneValidator) {
	v.Reset()
	SleepNFrames(5)
	for _, e := range v.Errors {
		t.Error(e)
	}
	if len(v.Errors) > 0 {
		t.FailNow()
	}
}

func SimulateMouseClick(window *gfx.Window, x, y float32) {
	ms := gfx.MouseState{
		X: x,
		Y: y,
	}
	window.OverrideMouseState(&ms)
	SleepAFewFrames() // wait for the mouse over event to be triggered/handled
	ms.PrimaryDown = true
	window.OverrideMouseState(&ms)
	SleepAFewFrames() // wait for the mouse down event to be triggered/handled
	ms.PrimaryDown = false
	window.OverrideMouseState(&ms)
	SleepAFewFrames() // wait for the mouse up event to be triggered/handled
}

func SimulateMouseClickAndDrag(window *gfx.Window, startX, startY, endX, endY, steps float32) {
	ms := gfx.MouseState{
		X: startX,
		Y: startY,
	}

	window.OverrideMouseState(&ms)
	SleepACoupleFrames()

	deltaX := endX - startX
	deltaY := endY - startY
	stepX := deltaX / steps
	stepY := deltaY / steps

	for i := 0; i < int(steps); i++ {
		ms.X += stepX
		ms.Y += stepY
		ms.PrimaryDown = true
		window.OverrideMouseState(&ms)
		SleepACoupleFrames()
	}

	ms.PrimaryDown = false
	window.OverrideMouseState(&ms)
	SleepACoupleFrames()
}

type OddValueFilter struct {
	gfx.FilterBase
}

func (f *OddValueFilter) Apply(index int, input []float64) (output float64) {
	if int(input[index])%2 == 0 {
		return input[index]
	} else {
		return input[index] + 1
	}
}

func NewOddValueFilter() *OddValueFilter {
	f := &OddValueFilter{}
	f.SetEnabled(true)
	return f
}

type PlusNTransformer struct {
	gfx.TransformerBase
}

func (t *PlusNTransformer) Transform(dst, src []float64) []float64 {
	n := rand.Float64() * 10
	nArr := make([]float64, len(src))
	for i, v := range src {
		dst[i] = v + n
		nArr[i] = n
		n = rand.Float64() * 10
	}
	return nArr
}

func NewPlusNTransformer() *PlusNTransformer {
	t := &PlusNTransformer{}
	t.SetEnabled(true)
	return t
}

type Model struct {
	gfx.ModelBase
	vertices []float32
	uvs      []float32
	meshes   []*Mesh
}

type Mesh struct {
	gfx.MeshBase
	faces []*Face
}

type Face struct {
	gfx.FaceBase
	vertexIndices []int
	uvIndices     []int
	material      gfx.Material
}

type TestMaterial struct {
	gfx.MaterialBase
	DiffuseMap gfx.Texture
}

var VertShader = `#version 410 core
in vec3 a_Position;
in vec2 a_UV;
out vec2 UV;
void main() {
	UV = a_UV;
  	gl_Position = vec4(a_Position, 1.0);
}
`

var FragShader = `#version 410 core
in vec2 UV;
out vec4 FragColor;
uniform sampler2D u_DiffuseMap; 
void main() {
	FragColor = texture(u_DiffuseMap, UV);
}
`

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

func NewTestModel() *Model {
	material := &TestMaterial{}

	face0 := &Face{
		vertexIndices: []int{0, 1, 2},
		uvIndices:     []int{0, 1, 2},
		material:      material,
	}
	face1 := &Face{
		vertexIndices: []int{2, 3, 0},
		uvIndices:     []int{2, 3, 0},
		material:      material,
	}

	mesh := &Mesh{
		faces: []*Face{face0, face1},
	}

	model := &Model{
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

	return model
}

func PanicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func SleepACoupleFrames() {
	time.Sleep(time.Duration((1000/TargetFramerate)*2) * time.Millisecond)
}

func SleepAFewFrames() {
	time.Sleep(time.Duration((1000/TargetFramerate)*3) * time.Millisecond)
}

func SleepNFrames(n int) {
	time.Sleep(time.Duration((1000/TargetFramerate)*n) * time.Millisecond)
}
