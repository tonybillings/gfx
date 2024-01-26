package gfx

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
)

const (
	defaultSliderName = "Slider"
)

/******************************************************************************
 Slider
******************************************************************************/

type Slider struct {
	View

	rail   *Shape
	button *Button

	orientation Orientation
	circular    bool

	value float32

	valueChanging   bool
	onValueChanging []func(WindowObject, float32)
	onValueChanged  []func(WindowObject, float32)
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (s *Slider) defaultLayout() {
	s.fill.SetColor(Gray)

	if s.orientation == Vertical {
		s.SetScale(mgl32.Vec3{0.1, 0.2})
		s.rail.SetScale(mgl32.Vec3{0.05, 0.7})

		if s.circular {
			s.button.SetScale(mgl32.Vec3{0.3, 0.15})
		} else {
			s.button.SetScale(mgl32.Vec3{0.4, 0.1})
		}
	} else {
		s.SetScale(mgl32.Vec3{0.2, 0.1})
		s.rail.SetScale(mgl32.Vec3{0.7, 0.05})

		if s.circular {
			s.button.SetScale(mgl32.Vec3{0.15, 0.3})
		} else {
			s.button.SetScale(mgl32.Vec3{0.1, 0.4})
		}
	}

	s.rail.SetColor(DarkGray)
	s.button.SetColor(color.RGBA{R: 30, G: 30, B: 30, A: 255})
	s.button.SetMouseEnterFillColor(Lighten(s.button.FillColor(), 0.8))
}

func (s *Slider) initLayout() {
	s.button.bounds.OnMouseEnter(func(_ WindowObject, _ *MouseState) {
		if s.orientation == Vertical {
			s.Window().glwin.SetCursor(glfw.CreateStandardCursor(glfw.VResizeCursor))
		} else {
			s.Window().glwin.SetCursor(glfw.CreateStandardCursor(glfw.HResizeCursor))
		}
	})

	s.button.bounds.OnMouseLeave(func(_ WindowObject, _ *MouseState) {
		s.Window().glwin.SetCursor(glfw.CreateStandardCursor(glfw.ArrowCursor))
	})

	s.button.OnDepressed(func(_ WindowObject, _ *MouseState) {
		winMouse := s.Window().Mouse()
		var value float32
		if s.orientation == Vertical {
			y := winMouse.Y - s.WorldPosition().Y()
			railEnd := s.Window().ScaleY(s.rail.WorldScale().Y())

			if y < -railEnd {
				y = -railEnd
			} else if y > railEnd {
				y = railEnd
			}

			s.button.SetPositionY(y)
			value = (y + railEnd) / (railEnd * 2.0)
		} else {
			x := winMouse.X - s.WorldPosition().X()
			railEnd := s.Window().ScaleX(s.rail.WorldScale().X())

			if x < -railEnd {
				x = -railEnd
			} else if x > railEnd {
				x = railEnd
			}

			s.button.SetPositionX(x)
			value = (x + railEnd) / (railEnd * 2.0)
		}

		s.stateMutex.Lock()
		s.value = value
		s.valueChanging = true
		s.stateMutex.Unlock()

		for _, f := range s.onValueChanging {
			f(s, value)
		}
	})

	s.button.bounds.OnPMouseUp(func(_ WindowObject, _ *MouseState) {
		s.stateMutex.Lock()
		if s.valueChanging {
			s.valueChanging = false
			value := s.value
			s.stateMutex.Unlock()

			for _, f := range s.onValueChanged {
				f(s, value)
			}
		} else {
			s.stateMutex.Unlock()
		}
	})
}

func (s *Slider) Init(window *Window) (ok bool) {
	if !s.rail.Init(window) || !s.button.Init(window) {
		return false
	}

	if !s.View.Init(window) {
		return false
	}

	s.initLayout()
	s.initialized.Store(true)
	s.SetValue(s.value)

	return true
}

func (s *Slider) Update(deltaTime int64) (ok bool) {
	if !s.enabled.Load() {
		return false
	}

	if !s.rail.Update(deltaTime) || !s.button.Update(deltaTime) {
		return false
	}

	return s.View.Update(deltaTime)
}

func (s *Slider) Draw(deltaTime int64) (ok bool) {
	if !s.visible.Load() {
		return false
	}

	if !s.View.Draw(deltaTime) {
		return false
	}

	if !s.rail.Draw(deltaTime) || !s.button.Draw(deltaTime) {
		return false
	}

	return true
}

func (s *Slider) Close() {
	s.rail.Close()
	s.button.Close()
	s.View.Close()
}

func (s *Slider) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	s.rail.Resize(oldWidth, oldHeight, newWidth, newHeight)
	s.button.Resize(oldWidth, oldHeight, newWidth, newHeight)
	s.View.Resize(oldWidth, oldHeight, newWidth, newHeight)
	s.SetValue(s.Value())
}

func (s *Slider) SetWindow(window *Window) WindowObject {
	s.View.SetWindow(window)
	s.rail.SetWindow(window)
	s.button.SetWindow(window)
	return s
}

/******************************************************************************
 Slider Functions
******************************************************************************/

func (s *Slider) Value() float32 {
	s.stateMutex.Lock()
	v := s.value
	s.stateMutex.Unlock()
	return v
}

func (s *Slider) SetValue(value float32) {
	if value < 0 {
		value = 0
	} else if value > 1 {
		value = 1
	}

	s.stateMutex.Lock()
	s.value = value
	s.stateMutex.Unlock()

	if !s.initialized.Load() {
		return
	}

	if s.orientation == Vertical {
		railEnd := s.Window().ScaleY(s.rail.WorldScale().Y())
		y := -railEnd + (value * railEnd * 2.0)
		s.button.SetPositionY(y)
	} else {
		railEnd := s.Window().ScaleX(s.rail.WorldScale().X())
		x := -railEnd + (value * railEnd * 2.0)
		s.button.SetPositionX(x)
	}
}

func (s *Slider) Button() *Button {
	return s.button
}

func (s *Slider) Rail() *Shape {
	return s.rail
}

func (s *Slider) OnValueChanging(handler func(sender WindowObject, value float32)) *Slider {
	s.onValueChanging = append(s.onValueChanging, handler)
	return s
}

func (s *Slider) OnValueChanged(handler func(sender WindowObject, value float32)) *Slider {
	s.onValueChanged = append(s.onValueChanged, handler)
	return s
}

/******************************************************************************
 New Slider Function
******************************************************************************/

func NewSlider(orientation Orientation, isButtonCircular ...bool) *Slider {
	circular := len(isButtonCircular) > 0 && isButtonCircular[0]

	s := &Slider{
		View:        *NewView(),
		rail:        NewQuad(),
		button:      NewButton(circular),
		orientation: orientation,
		circular:    circular,
	}

	s.SetName(defaultSliderName)
	s.fill.SetParent(s)
	s.border.SetParent(s)
	s.rail.SetParent(s)
	s.button.SetParent(s)

	s.defaultLayout()

	return s
}
