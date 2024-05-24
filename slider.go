package gfx

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"sync"
)

const (
	defaultSliderName = "Slider"
)

/******************************************************************************
 Slider
******************************************************************************/

type Slider struct {
	View

	rail   *Shape2D
	button *Button

	orientation Orientation
	circular    bool

	value float32

	valueChanging bool

	valueChangedDispatcher  chan float32
	valueChangingDispatcher chan float32
	onValueChangingHandlers []func(WindowObject, float32)
	onValueChangedHandlers  []func(WindowObject, float32)
	eventHandlersMutex      sync.Mutex
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (s *Slider) Init() (ok bool) {
	if s.Initialized() {
		return true
	}

	s.rail.SetWindow(s.window)
	s.button.SetWindow(s.window)

	if !s.rail.Init() || !s.button.Init() {
		return false
	}

	s.initLayout()
	s.initDispatchers()

	if ok = s.View.Init(); !ok {
		return
	}

	s.SetValue(s.value)
	return true
}

func (s *Slider) Update(deltaTime int64) (ok bool) {
	if !s.View.Update(deltaTime) {
		return false
	}

	if !s.rail.Update(deltaTime) || !s.button.Update(deltaTime) {
		return false
	}

	return true
}

func (s *Slider) Close() {
	if !s.Initialized() {
		return
	}

	if s.valueChangingDispatcher != nil {
		close(s.valueChangingDispatcher)
	}

	if s.valueChangedDispatcher != nil {
		close(s.valueChangedDispatcher)
	}

	s.rail.Close()
	s.button.Close()
	s.View.Close()
}

/******************************************************************************
 DrawableObject Implementation
******************************************************************************/

func (s *Slider) Draw(deltaTime int64) (ok bool) {
	if !s.View.Draw(deltaTime) {
		return false
	}

	if !s.rail.Draw(deltaTime) || !s.button.Draw(deltaTime) {
		return false
	}

	return true
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (s *Slider) SetWindow(window *Window) WindowObject {
	s.View.SetWindow(window)
	s.rail.SetWindow(window)
	s.button.SetWindow(window)
	return s
}

/******************************************************************************
 Resizer Implementation
******************************************************************************/

func (s *Slider) Resize(newWidth, newHeight int) {
	s.rail.Resize(newWidth, newHeight)
	s.button.Resize(newWidth, newHeight)
	s.View.Resize(newWidth, newHeight)
	s.SetValue(s.Value())
}

/******************************************************************************
 Slider Functions
******************************************************************************/

func (s *Slider) defaultLayout() {
	s.fill.SetParent(s)
	s.border.SetParent(s)
	s.rail.SetParent(s)
	s.button.SetParent(s)

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
		mouse := s.button.bounds.MouseSurface().Mouse()
		var value float32
		if s.orientation == Vertical {
			y := mouse.Y - s.WorldPosition().Y()
			railEnd := s.Window().ScaleY(s.rail.WorldScale().Y())

			if y < -railEnd {
				y = -railEnd
			} else if y > railEnd {
				y = railEnd
			}

			s.button.SetPositionY(y)
			value = (y + railEnd) / (railEnd * 2.0)
		} else {
			x := mouse.X - s.WorldPosition().X()
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

		select {
		case s.valueChangingDispatcher <- value:
		default:
		}
	})

	s.button.bounds.OnPMouseUp(func(_ WindowObject, _ *MouseState) {
		s.stateMutex.Lock()
		if s.valueChanging {
			s.valueChanging = false
			value := s.value
			s.stateMutex.Unlock()

			select {
			case s.valueChangedDispatcher <- value:
			default:
			}
		} else {
			s.stateMutex.Unlock()
		}
	})
}

func (s *Slider) initDispatchers() {
	s.valueChangingDispatcher = make(chan float32, 1024)
	s.valueChangedDispatcher = make(chan float32, 64)

	go s.handleValueChanging()
	go s.handleValueChanged()
}

func (s *Slider) handleValueChanging() {
	for {
		select {
		case value, ok := <-s.valueChangingDispatcher:
			if !ok {
				return
			}

			s.eventHandlersMutex.Lock()
			for _, handler := range s.onValueChangingHandlers {
				handler(s, value)
			}
			s.eventHandlersMutex.Unlock()
		}
	}
}

func (s *Slider) handleValueChanged() {
	for {
		select {
		case value, ok := <-s.valueChangedDispatcher:
			if !ok {
				return
			}

			s.eventHandlersMutex.Lock()
			for _, handler := range s.onValueChangedHandlers {
				handler(s, value)
			}
			s.eventHandlersMutex.Unlock()
		}
	}
}

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

func (s *Slider) Rail() *Shape2D {
	return s.rail
}

func (s *Slider) OnValueChanging(handler func(sender WindowObject, value float32)) *Slider {
	s.eventHandlersMutex.Lock()
	s.onValueChangingHandlers = append(s.onValueChangingHandlers, handler)
	s.eventHandlersMutex.Unlock()
	return s
}

func (s *Slider) OnValueChanged(handler func(sender WindowObject, value float32)) *Slider {
	s.eventHandlersMutex.Lock()
	s.onValueChangedHandlers = append(s.onValueChangedHandlers, handler)
	s.eventHandlersMutex.Unlock()
	return s
}

func (s *Slider) SetMouseSurface(surface MouseSurface) {
	s.button.SetMouseSurface(surface)
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
	s.defaultLayout()

	return s
}
