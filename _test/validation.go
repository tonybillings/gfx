package _test

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/google/uuid"
	"image/color"
	"sync/atomic"
	"testing"
	"tonysoft.com/gfx"
)

/******************************************************************************
 PixelSampler
******************************************************************************/

type PixelSampler struct {
	window *gfx.Window

	GetPixelPosFunc func() (posX, posY float32)
	Description     string
	ExpectedColor   color.RGBA
	ActualColor     color.RGBA
	IsValid         bool
}

func (s *PixelSampler) Sample() (isValid bool) {
	posX, posY := s.GetPixelPosFunc()

	x := int32(float32(s.window.Width()) * ((posX + 1.0) / 2.0))
	y := int32(float32(s.window.Height()) * ((posY + 1.0) / 2.0))

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

/******************************************************************************
 SceneValidator
******************************************************************************/

type SceneValidator struct {
	gfx.DrawableObjectBase

	t         *testing.T
	window    *gfx.Window
	validated atomic.Bool

	Samplers []*PixelSampler
	Errors   []error
}

func (v *SceneValidator) validate() {
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
		v.validate()
		v.validated.Store(true)
		return true
	}

	return true
}

func (v *SceneValidator) Reset() {
	v.validated.Store(false)
}

func (v *SceneValidator) AddPixelSampler(getPixelPositionFunc func() (posX, posY float32), expectedColor color.RGBA, description ...string) {
	desc := "(none provided)"
	if len(description) > 0 {
		desc = description[0]
	}

	sampler := &PixelSampler{}
	sampler.window = v.window
	sampler.GetPixelPosFunc = getPixelPositionFunc
	sampler.ExpectedColor = expectedColor
	sampler.Description = desc

	v.Samplers = append(v.Samplers, sampler)
}

func (v *SceneValidator) Validate() {
	v.Reset()
	SleepNFrames(5)
	for _, e := range v.Errors {
		v.t.Error(e)
	}
	if len(v.Errors) > 0 {
		v.t.FailNow()
	}
}

func NewSceneValidator(t *testing.T, window *gfx.Window) *SceneValidator {
	v := &SceneValidator{
		t:      t,
		window: window,
	}
	v.SetName(uuid.New().String())
	return v
}
