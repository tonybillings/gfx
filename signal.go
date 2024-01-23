package gfx

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"math"
	"sync"
	"sync/atomic"
)

/******************************************************************************
 Interfaces
******************************************************************************/

type SignalControl interface {
	*SignalLine | *SignalGroup
}

/******************************************************************************
 Types
******************************************************************************/

type Signal struct {
	label      string
	data       []float64
	dataIdx    int
	dataSize   int
	minData    float64
	maxData    float64
	deltas     []float64
	deltasSize int
	dataMutex  sync.Mutex
}

type SignalLine struct {
	View
	Signal

	vao uint32
	vbo uint32

	shader uint32

	colorUniformLoc     int32
	thicknessUniformLoc int32

	thickness uint

	label *Label

	inspector            *SignalInspector
	inspectorKey         glfw.Key
	inspectorActive      atomic.Bool
	inspectKeyRegistered atomic.Bool
}

type SignalGroup struct {
	View

	defaultSampleCount int
	defaultColors      []color.RGBA
	defaultColorIdx    int
	defaultThickness   uint

	inspector            *SignalInspector
	inspectorKey         glfw.Key
	inspectorActive      atomic.Bool
	inspectKeyRegistered atomic.Bool
}

type SignalInspector struct {
	WindowObjectBase

	bounds *BoundingBox

	panel    *View
	minValue *Label
	maxValue *Label
	avg      *Label
	std      *Label
	deltaAvg *Label
	deltaStd *Label
	sample   *Label

	alignment Alignment
	margin    float32
	fontSize  float32
}

/******************************************************************************
 Signal
******************************************************************************/

func (s *Signal) Lock() {
	s.dataMutex.Lock()
}

func (s *Signal) Unlock() {
	s.dataMutex.Unlock()
}

func (s *Signal) Label() string {
	return s.label
}

func (s *Signal) Data() []float64 {
	return s.data
}

func (s *Signal) AddSamples(data []float64) {
	s.dataMutex.Lock()

	for _, d := range data {
		s.data[s.dataIdx] = d
		s.dataIdx = (s.dataIdx + 1) % s.dataSize
		if d < s.minData {
			s.minData = d
		}
		if d > s.maxData {
			s.maxData = d
		}
	}

	s.dataMutex.Unlock()
}

func (s *Signal) Average() float64 {
	s.dataMutex.Lock()

	sum := 0.0
	for _, value := range s.data {
		sum += value
	}

	s.dataMutex.Unlock()
	mean := sum / float64(s.dataSize)
	return mean
}

func (s *Signal) DeltaAverage() float64 {
	s.dataMutex.Lock()

	for i := 1; i < s.dataSize; i++ {
		if i == s.dataIdx {
			continue
		}
		s.deltas[i-1] = s.data[i] - s.data[i-1]
	}
	sum := 0.0
	for _, value := range s.deltas {
		sum += value
	}

	s.dataMutex.Unlock()
	mean := sum / float64(s.deltasSize)
	return mean
}

func (s *Signal) StdDev() float64 {
	s.dataMutex.Lock()

	sum := 0.0
	for _, value := range s.data {
		sum += value
	}
	mean := sum / float64(s.dataSize)

	var sumOfSqrDiff float64
	for _, value := range s.data {
		diff := value - mean
		sumOfSqrDiff += diff * diff
	}

	s.dataMutex.Unlock()
	return math.Sqrt(sumOfSqrDiff / float64(s.dataSize))
}

func (s *Signal) DeltaStdDev(calculateDeltas bool) float64 {
	s.dataMutex.Lock()

	deltaAvg := 0.0
	if calculateDeltas {
		for i := 1; i < s.dataSize; i++ {
			if i == s.dataIdx {
				continue
			}
			s.deltas[i-1] = s.data[i] - s.data[i-1]
		}
		sum := 0.0
		for _, value := range s.deltas {
			sum += value
		}
		deltaAvg = sum / float64(s.deltasSize)
	} else {
		var sum float64
		for _, delta := range s.deltas {
			sum += delta
		}
		deltaAvg = sum / float64(s.deltasSize)
	}

	if s.deltasSize == 0 {
		s.dataMutex.Unlock()
		return 0
	}

	var sumOfSqrDiff float64
	for _, delta := range s.deltas {
		diff := delta - deltaAvg
		sumOfSqrDiff += diff * diff
	}

	s.dataMutex.Unlock()
	return math.Sqrt(sumOfSqrDiff / float64(s.deltasSize))
}

func (s *Signal) MinValue() float64 {
	return s.minData
}

func (s *Signal) MaxValue() float64 {
	return s.maxData
}

func (s *Signal) BufferSize() int {
	return s.dataSize
}

/******************************************************************************
 SignalLine
******************************************************************************/

func (l *SignalLine) initVertices() {
	l.vertices = make([]float32, l.vertexCount*2)
	step := (l.WorldScale().X() * 2.0) / float32(l.vertexCount)
	for i := int32(0); i < l.vertexCount; i++ {
		l.vertices[i*2] = l.position[0] + (float32(i) * step)
		l.vertices[i*2+1] = l.position[1]
	}
}

func (l *SignalLine) updateVertices() {
	l.Signal.Lock()
	minV := l.Signal.MinValue()
	maxV := l.Signal.MaxValue()
	data := l.Signal.Data()
	for i, s := range data {
		minMaxRange := float32(maxV) - float32(minV)
		sNorm := (float32(s) - float32(minV)) / minMaxRange
		if minMaxRange == 0.0 {
			sNorm = 0
		}
		l.vertices[i*2+1] = l.position[1] + (((2.0 * sNorm) - 1.0) * l.scale[1])
	}
	l.Signal.Unlock()
}

func (l *SignalLine) initGl() {
	l.shader = GetShaderProgram(SignalShaderProgram)
	l.thicknessUniformLoc = gl.GetUniformLocation(l.shader, gl.Str("thickness\x00"))
	l.colorUniformLoc = gl.GetUniformLocation(l.shader, gl.Str("color\x00"))

	_, height := l.window.win.GetSize()
	pixelHeight := 2.0 / float32(height)
	stepUniformPos := gl.GetUniformLocation(l.shader, gl.Str("step\x00"))
	gl.UseProgram(l.shader)
	gl.Uniform1f(stepUniformPos, pixelHeight)
	gl.UseProgram(0)

	gl.GenVertexArrays(1, &l.vao)
	gl.GenBuffers(1, &l.vbo)

	gl.BindVertexArray(l.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, l.vbo)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, nil)

	gl.BufferData(gl.ARRAY_BUFFER, len(l.vertices)*sizeOfFloat32, gl.Ptr(l.vertices), gl.DYNAMIC_DRAW)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func (l *SignalLine) uninitGl() {
	if !l.Initialized() {
		return
	}
	l.SetInitialized(false)
	l.stateMutex.Lock()

	gl.BindVertexArray(l.vao)
	gl.DisableVertexAttribArray(0)
	gl.DisableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	gl.DeleteBuffers(1, &l.vbo)
	gl.DeleteVertexArrays(1, &l.vao)

	l.stateMutex.Unlock()
}

func (l *SignalLine) initLayout(window *Window) {
	l.View.Init(window)
	l.fill.MaintainAspectRatio(false)
	l.border.MaintainAspectRatio(false)
	l.inspector.MaintainAspectRatio(false)
	l.fill.SetPositionX(l.WorldScale().X())
	l.border.SetPositionX(l.WorldScale().X())
}

func (l *SignalLine) Init(window *Window) (ok bool) {
	if !l.View.Init(window) {
		return false
	}

	l.initVertices()
	l.initGl()
	l.initLayout(window)
	l.initialized.Store(true)

	if l.inspectorKey != glfw.KeyUnknown {
		l.EnableInspector()
	}
	l.inspector.Init(window)

	return l.label.Init(window)
}

func (l *SignalLine) Update(deltaTime int64) (ok bool) {
	if !l.View.Update(deltaTime) {
		return false
	}

	if l.stateChanged.Load() {
		l.stateChanged.Store(false)
		l.updateVertices()
	}

	l.inspector.Update(deltaTime)
	return l.label.Update(deltaTime)
}

func (l *SignalLine) Draw(deltaTime int64) (ok bool) {
	if !l.visible.Load() || l.closed.Load() {
		return false
	}

	if l.closing.Load() {
		l.uninitGl()
		l.closed.Store(true)
		l.closing.Store(false)
		return false
	}

	l.View.Draw(deltaTime)

	gl.UseProgram(l.shader)

	l.stateMutex.Lock()
	gl.Uniform1f(l.thicknessUniformLoc, float32(l.thickness))
	gl.Uniform4fv(l.colorUniformLoc, 1, &l.color[0])
	l.stateMutex.Unlock()

	gl.BindVertexArray(l.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, l.vbo)

	gl.Viewport(0, 0, l.window.Width(), l.window.Height())

	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(l.vertices)*sizeOfFloat32, gl.Ptr(l.vertices))

	gl.DrawArrays(gl.LINE_STRIP_ADJACENCY, 0, l.vertexCount)

	gl.BindVertexArray(0)
	gl.UseProgram(0)

	if !l.label.Draw(deltaTime) {
		return false
	}

	l.inspector.Draw(deltaTime)

	return true
}

func (l *SignalLine) Close() {
	l.View.Close()
	l.label.Close()
	l.inspector.Close()
}

func (l *SignalLine) Thickness() uint {
	l.stateMutex.Lock()
	thickness := l.thickness
	l.stateMutex.Unlock()
	return thickness
}

func (l *SignalLine) SetThickness(thickness uint) *SignalLine {
	if thickness < 1 {
		thickness = 1
	}
	l.stateMutex.Lock()
	l.thickness = thickness
	l.stateMutex.Unlock()
	return l
}

func (l *SignalLine) SetColor(rgba color.RGBA) WindowObject {
	l.WindowObjectBase.SetColor(rgba)
	return l
}

func (l *SignalLine) AddSamples(data []float64) {
	l.Signal.AddSamples(data)
	l.stateChanged.Store(true)
}

func (l *SignalLine) Label() *Label {
	return l.label
}

func (l *SignalLine) EnableInspector(inspectKey ...glfw.Key) {
	l.inspectorKey = glfw.KeyLeftControl
	if len(inspectKey) > 0 {
		l.inspectorKey = inspectKey[0]
	}

	if l.inspectKeyRegistered.Load() || !l.initialized.Load() {
		return
	}
	l.inspectKeyRegistered.Store(true)

	l.window.AddKeyEventHandler(l.inspectorKey, glfw.Press, func(_ *Window, _ glfw.Key, _ glfw.Action) {
		if l.enabled.Load() && !l.inspectorActive.Load() {
			l.inspectorActive.Store(true)
			l.inspector.SetEnabled(true)
			l.inspector.SetVisibility(true)
		}
	})
	l.window.AddKeyEventHandler(l.inspectorKey, glfw.Release, func(_ *Window, _ glfw.Key, _ glfw.Action) {
		if l.enabled.Load() && l.inspectorActive.Load() {
			l.inspectorActive.Store(false)
			l.inspector.SetEnabled(false)
			l.inspector.SetVisibility(false)
		}
	})
}

func (l *SignalLine) SetInspectorAlignment(alignment Alignment) {
	l.inspector.SetAlignment(alignment)
}

func (l *SignalLine) SetInspectorMargin(margin float32) *SignalLine {
	l.inspector.SetMargin(margin)
	return l
}

func (l *SignalLine) SetInspectorFontSize(size float32) *SignalLine {
	l.inspector.SetFontSize(size)
	return l
}

func (l *SignalLine) SetInspectorPanelSize(size float32) *SignalLine {
	l.inspector.SetPanelSize(size)
	return l
}

func (l *SignalLine) InspectorPanel() *View {
	return l.inspector.Panel()
}

func (l *SignalLine) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	l.View.Resize(oldWidth, oldHeight, newWidth, newHeight)
	l.label.Resize(oldWidth, oldHeight, newWidth, newHeight)
	l.inspector.Resize(oldWidth, oldHeight, newWidth, newHeight)
}

/******************************************************************************
 SignalGroup
******************************************************************************/

func (g *SignalGroup) initLayout(window *Window) {
	g.View.Init(window)
	g.fill.MaintainAspectRatio(false)
	g.border.MaintainAspectRatio(false)
	g.inspector.MaintainAspectRatio(false)
	g.fill.SetPositionX(g.WorldScale().X())
	g.border.SetPositionX(g.WorldScale().X())
}

func (g *SignalGroup) updateSignalLayout() {
	g.stateMutex.Lock()
	signalCount := float32(len(g.children))
	scale := g.scale[1] / signalCount
	step := 2.0 * scale
	halfStep := step * 0.5
	for i, c := range g.children {
		if s, ok := c.(*SignalLine); ok {
			s.position[0] = g.position[0]
			s.position[1] = (g.position[1] + g.scale[1]) - (float32(i) * step) - halfStep
			s.scale[1] = scale
			s.label.SetParent(g)
			s.label.SetPositionY(s.position[1] - g.position[1])
			s.label.SetScaleY(scale * 0.33)
			s.stateChanged.Store(true)
		}
	}
	g.stateMutex.Unlock()
}

func (g *SignalGroup) Init(window *Window) (ok bool) {
	if !g.View.Init(window) {
		return false
	}

	g.initLayout(window)
	g.initialized.Store(true)

	if g.inspectorKey != glfw.KeyUnknown {
		g.EnableInspector()
	}
	return g.inspector.Init(window)
}

func (g *SignalGroup) Update(deltaTime int64) (ok bool) {
	if g.enabled.Load() && g.stateChanged.Load() {
		g.stateChanged.Store(false)
		g.updateSignalLayout()
	}

	if !g.View.Update(deltaTime) {
		return false
	}

	g.inspector.Update(deltaTime)
	return true
}

func (g *SignalGroup) Draw(deltaTime int64) (ok bool) {
	if !g.View.Draw(deltaTime) {
		return false
	}

	g.inspector.Draw(deltaTime)

	return g.inspector.Draw(deltaTime)
}

func (g *SignalGroup) Close() {
	g.View.Close()
	g.inspector.Close()
}

func (g *SignalGroup) New(label string) *SignalLine {
	g.stateMutex.Lock()
	newSignal := NewSignalLine(label, g.defaultSampleCount)
	newSignal.
		SetThickness(g.defaultThickness).
		SetColor(g.defaultColors[g.defaultColorIdx])

	g.children = append(g.children, newSignal)
	g.defaultColorIdx = (g.defaultColorIdx + 1) % len(g.defaultColors)
	g.stateMutex.Unlock()
	newSignal.SetParent(g)
	g.updateSignalLayout()
	return newSignal
}

func (g *SignalGroup) Add(signal *SignalLine) *SignalLine {
	g.stateMutex.Lock()
	g.children = append(g.children, signal)
	g.stateChanged.Store(true)
	g.stateMutex.Unlock()
	g.updateSignalLayout()
	return signal
}

func (g *SignalGroup) RemoveAt(index int) {
	g.stateMutex.Lock()
	if index < 0 || index > len(g.children)-1 {
		g.stateMutex.Unlock()
		return
	}
	g.children = append(g.children[:index], g.children[index+1:]...)
	g.stateChanged.Store(true)
	g.stateMutex.Unlock()
}

func (g *SignalGroup) Remove(signal *SignalLine) *SignalLine {
	if signal == nil {
		return nil
	}

	g.stateMutex.Lock()
	removeAt := -1
	for i, s := range g.children {
		if s == signal {
			removeAt = i
			break
		}
	}
	if removeAt != -1 {
		g.children = append(g.children[:removeAt], g.children[removeAt+1:]...)
		g.stateMutex.Unlock()
		return signal
	}

	g.stateChanged.Store(true)
	g.stateMutex.Unlock()
	return nil
}

func (g *SignalGroup) SetEnabled(enabled bool) WindowObject {
	if g.enabled.Load() != enabled {
		g.stateChanged.Store(true)
	}

	g.WindowObjectBase.SetEnabled(enabled)
	return g
}

func (g *SignalGroup) EnableInspector(inspectKey ...glfw.Key) {
	g.inspectorKey = glfw.KeyLeftControl
	if len(inspectKey) > 0 {
		g.inspectorKey = inspectKey[0]
	}

	if g.inspectKeyRegistered.Load() || !g.initialized.Load() {
		return
	}
	g.inspectKeyRegistered.Store(true)

	g.window.AddKeyEventHandler(g.inspectorKey, glfw.Press, func(_ *Window, _ glfw.Key, _ glfw.Action) {
		if g.enabled.Load() && !g.inspectorActive.Load() {
			g.inspectorActive.Store(true)
			g.inspector.SetEnabled(true)
			g.inspector.SetVisibility(true)
		}
	})
	g.window.AddKeyEventHandler(g.inspectorKey, glfw.Release, func(_ *Window, _ glfw.Key, _ glfw.Action) {
		if g.enabled.Load() && g.inspectorActive.Load() {
			g.inspectorActive.Store(false)
			g.inspector.SetEnabled(false)
			g.inspector.SetVisibility(false)
		}
	})
}

func (g *SignalGroup) SetInspectorAlignment(alignment Alignment) {
	g.inspector.SetAlignment(alignment)
}

func (g *SignalGroup) SetInspectorMargin(margin float32) *SignalGroup {
	g.inspector.SetMargin(margin)
	return g
}

func (g *SignalGroup) SetInspectorFontSize(size float32) *SignalGroup {
	g.inspector.SetFontSize(size)
	return g
}

func (g *SignalGroup) SetInspectorPanelSize(size float32) *SignalGroup {
	g.inspector.SetPanelSize(size)
	return g
}

func (g *SignalGroup) InspectorPanel() *View {
	return g.inspector.Panel()
}

func (g *SignalGroup) Signals() []*SignalLine {
	signals := make([]*SignalLine, len(g.children))
	for i, c := range g.children {
		if s, ok := c.(*SignalLine); ok {
			signals[i] = s
		}
	}
	return signals
}

func (g *SignalGroup) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	g.inspector.Resize(oldWidth, oldHeight, newWidth, newHeight)
	g.View.Resize(oldWidth, oldHeight, newWidth, newHeight)
}

/******************************************************************************
 SignalInspector
******************************************************************************/

func (i *SignalInspector) initLayout(window *Window) {
	i.minValue.SetFontSize(i.fontSize)
	i.maxValue.SetFontSize(i.fontSize)
	i.avg.SetFontSize(i.fontSize)
	i.std.SetFontSize(i.fontSize)
	i.deltaAvg.SetFontSize(i.fontSize)
	i.deltaStd.SetFontSize(i.fontSize)
	i.sample.SetFontSize(i.fontSize)

	i.panel.AddChildren(i.minValue, i.maxValue, i.avg, i.std, i.deltaAvg, i.deltaStd, i.sample)

	i.panel.Init(window)

	i.bounds.SetPosition(mgl32.Vec3{i.WorldScale().X(), 0, 0})
	i.bounds.Init(window)

	i.updateLayout()
}

func (i *SignalInspector) updateLayout() {
	worldScaleX := i.WorldScale().X()
	worldScaleY := i.panel.WorldScale().Y()

	i.stateMutex.Lock()

	marginX := i.window.ScaleX(i.margin)
	marginY := i.window.ScaleY(i.margin)
	scaleX := i.window.ScaleX(i.panel.Scale().X())
	scaleY := i.window.ScaleY(i.panel.Scale().Y())

	switch i.alignment {
	case TopLeft:
		i.panel.SetPositionX((-1 + marginX) + scaleX)
		i.panel.SetPositionY((1 - marginY) - scaleY)
	case TopCenter:
		i.panel.SetPositionX(0)
		i.panel.SetPositionY((1 - marginY) - scaleY)
	case TopRight:
		i.panel.SetPositionX((1 - marginX) - scaleX)
		i.panel.SetPositionY((1 - marginY) - scaleY)
	case MiddleLeft:
		i.panel.SetPositionX((-1 + marginX) + scaleX)
		i.panel.SetPositionY(0)
	case Center:
		i.panel.SetPositionX(0)
		i.panel.SetPositionY(0)
	case MiddleRight:
		i.panel.SetPositionX((1 - marginX) - scaleX)
		i.panel.SetPositionY(0)
	case BottomLeft:
		i.panel.SetPositionX((-1 + marginX) + scaleX)
		i.panel.SetPositionY((-1 + marginY) + scaleY)
	case BottomCenter:
		i.panel.SetPositionX(0)
		i.panel.SetPositionY((-1 + marginY) + scaleY)
	case BottomRight:
		i.panel.SetPositionX((1 - marginX) - scaleX)
		i.panel.SetPositionY((-1 + marginY) + scaleY)
	}

	labelSpacing := float32(50.0)
	i.minValue.SetFontSize(i.fontSize).SetPosition(mgl32.Vec3{0, i.window.ScaleY(0.15 * i.fontSize * labelSpacing * worldScaleY), 0})
	i.maxValue.SetFontSize(i.fontSize).SetPosition(mgl32.Vec3{0, i.window.ScaleY(0.1 * i.fontSize * labelSpacing * worldScaleY), 0})
	i.avg.SetFontSize(i.fontSize).SetPosition(mgl32.Vec3{0, i.window.ScaleY(0.05 * i.fontSize * labelSpacing * worldScaleY), 0})
	i.std.SetFontSize(i.fontSize).SetPosition(mgl32.Vec3{0, i.window.ScaleY(0), 0})
	i.deltaAvg.SetFontSize(i.fontSize).SetPosition(mgl32.Vec3{0, i.window.ScaleY(-0.05 * i.fontSize * labelSpacing * worldScaleY), 0})
	i.deltaStd.SetFontSize(i.fontSize).SetPosition(mgl32.Vec3{0, i.window.ScaleY(-0.1 * i.fontSize * labelSpacing * worldScaleY), 0})
	i.sample.SetFontSize(i.fontSize).SetPosition(mgl32.Vec3{0, i.window.ScaleY(-0.15 * i.fontSize * labelSpacing * worldScaleY), 0})
	i.bounds.SetPosition(mgl32.Vec3{worldScaleX, 0, 0})

	i.stateMutex.Unlock()
}

func (i *SignalInspector) Init(window *Window) (ok bool) {
	if !i.WindowObjectBase.Init(window) {
		return false
	}

	i.initLayout(window)
	i.initialized.Store(true)
	return true
}

func (i *SignalInspector) Update(deltaTime int64) (ok bool) {
	if !i.enabled.Load() {
		return true
	}

	if i.stateChanged.Load() {
		i.stateChanged.Store(false)
		i.updateLayout()
	}

	i.bounds.Update(deltaTime)
	i.panel.Update(deltaTime)

	mouse := i.bounds.LocalMouse()
	xScale := (mouse.X + 1.0) / 2.0

	updatePanel := func(signal *SignalLine) {
		sampleIdx := int(xScale * float32(signal.BufferSize()-1))
		i.minValue.SetText(fmt.Sprintf("Min:  %.6f", signal.MinValue()))
		i.maxValue.SetText(fmt.Sprintf("Max:  %.6f", signal.MaxValue()))
		i.avg.SetText(fmt.Sprintf("Avg:  %.6f", signal.Average()))
		i.std.SetText(fmt.Sprintf("Std:  %.6f", signal.StdDev()))
		i.deltaAvg.SetText(fmt.Sprintf("ΔAvg: %.6f", signal.DeltaAverage()))
		i.deltaStd.SetText(fmt.Sprintf("ΔStd: %.6f", signal.DeltaStdDev(false)))
		i.sample.SetText(fmt.Sprintf("%f", signal.Data()[sampleIdx]))
	}

	if i.bounds.MouseOver() {
		var signal *SignalLine
		if s, okay := i.Parent().(*SignalLine); okay {
			signal = s
		}
		if signal == nil {
			sg := i.Parent().(*SignalGroup)
			signalCount := len(sg.Children())
			if signalCount > 0 {
				idx := int(((mouse.Y + 1.0) / 2.0) * float32(signalCount))
				if idx == signalCount {
					idx--
				}
				signal = sg.Children()[(signalCount-1)-idx].(*SignalLine)
				i.SetVisibility(true)
				updatePanel(signal)
			} else {
				i.SetVisibility(false)
			}
		} else {
			i.SetVisibility(true)
			updatePanel(signal)
		}
	} else {
		i.SetVisibility(false)
	}

	return true
}

func (i *SignalInspector) Draw(deltaTime int64) (ok bool) {
	if !i.visible.Load() {
		return false
	}

	i.panel.Draw(deltaTime)

	return i.WindowObjectBase.Draw(deltaTime)
}

func (i *SignalInspector) Close() {
	i.bounds.Close()
	i.panel.Close()
}

func (i *SignalInspector) SetEnabled(isEnabled bool) WindowObject {
	i.enabled.Store(isEnabled)
	i.bounds.enabled.Store(isEnabled)
	i.panel.enabled.Store(isEnabled)
	return i
}

func (i *SignalInspector) SetVisibility(isVisible bool) WindowObject {
	i.visible.Store(isVisible)
	i.bounds.visible.Store(isVisible)
	i.panel.visible.Store(isVisible)
	return i
}

func (i *SignalInspector) SetAlignment(alignment Alignment) {
	i.stateMutex.Lock()
	i.alignment = alignment
	i.stateChanged.Store(true)
	i.stateMutex.Unlock()
}

func (i *SignalInspector) SetMargin(margin float32) *SignalInspector {
	i.stateMutex.Lock()
	i.margin = margin
	i.stateChanged.Store(true)
	i.stateMutex.Unlock()
	return i
}

func (i *SignalInspector) SetFontSize(size float32) *SignalInspector {
	i.stateMutex.Lock()
	i.fontSize = size
	i.stateChanged.Store(true)
	i.stateMutex.Unlock()
	return i
}

func (i *SignalInspector) SetPanelSize(size float32) *SignalInspector {
	i.panel.SetScale(mgl32.Vec3{size, size, 1})
	i.stateChanged.Store(true)
	return i
}

func (i *SignalInspector) MaintainAspectRatio(maintainAspectRatio bool) WindowObject {
	i.bounds.MaintainAspectRatio(maintainAspectRatio)
	return i
}

func (i *SignalInspector) Panel() *View {
	return i.panel
}

func (i *SignalInspector) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	i.WindowObjectBase.Resize(oldWidth, oldHeight, newWidth, newHeight)
	i.panel.Resize(oldWidth, oldHeight, newWidth, newHeight)
	i.stateChanged.Store(true)
}

/******************************************************************************
 New Functions
******************************************************************************/

func NewSignal(label string, bufferSize int) *Signal {
	return &Signal{
		label:      label,
		data:       make([]float64, bufferSize),
		dataSize:   bufferSize,
		deltas:     make([]float64, bufferSize-1),
		deltasSize: bufferSize - 1,
	}
}

func NewSignalLine(label string, sampleCount int) *SignalLine {
	lbl := NewLabel().SetAlignment(MiddleLeft).SetText(label)

	sl := &SignalLine{
		View:   *NewView(),
		Signal: *NewSignal(label, sampleCount),
	}

	lbl.SetParent(sl)
	sl.fill.SetParent(sl)
	sl.border.SetParent(sl)
	sl.label = lbl
	sl.thickness = 1
	sl.name.Store(&label)
	sl.position[0] = -1
	sl.vertices = make([]float32, sampleCount*2)
	sl.vertexCount = int32(sampleCount)
	sl.inspectorKey = glfw.KeyUnknown
	sl.inspector = NewSignalInspector(sl)
	sl.inspector.MaintainAspectRatio(false)
	sl.enabled.Store(true)
	sl.visible.Store(true)
	return sl
}

func NewSignalGroup(defaultSampleCount int, defaultLineThickness int, colors ...color.RGBA) *SignalGroup {
	if len(colors) == 0 {
		colors = DefaultColors
	}

	sg := &SignalGroup{
		View:               *NewView(),
		defaultSampleCount: defaultSampleCount,
		defaultThickness:   uint(defaultLineThickness),
		defaultColors:      colors,
		inspectorKey:       glfw.KeyUnknown,
	}

	sg.inspectorKey = glfw.KeyUnknown
	sg.inspector = NewSignalInspector(sg)
	sg.inspector.MaintainAspectRatio(false)
	sg.fill.SetParent(sg)
	sg.border.SetParent(sg)
	sg.enabled.Store(true)
	sg.visible.Store(true)
	return sg
}

func NewSignalInspector[T SignalControl](parent T) *SignalInspector {
	if parent == nil {
		panic("parent cannot be nil")
	}

	si := &SignalInspector{
		WindowObjectBase: *NewObject(nil),
		bounds:           NewBoundingBox(),
		panel:            NewView(),
		minValue:         NewLabel(),
		maxValue:         NewLabel(),
		avg:              NewLabel(),
		std:              NewLabel(),
		deltaAvg:         NewLabel(),
		deltaStd:         NewLabel(),
		sample:           NewLabel(),
		alignment:        TopRight,
		margin:           0,
		fontSize:         0.1,
	}

	si.SetParent(any(parent).(WindowObject))
	si.bounds.SetParent(si)

	si.panel.SetScale(mgl32.Vec3{0.2, 0.2, 0})
	si.panel.SetFillColor(color.RGBA{R: 52, G: 53, B: 65, A: 255})
	si.panel.SetBorderColor(White)
	si.panel.SetBorderThickness(0.02)

	si.minValue.SetCacheEnabled(true).SetName("min")
	si.maxValue.SetCacheEnabled(true).SetName("max")
	si.avg.SetCacheEnabled(true).SetName("avg")
	si.std.SetCacheEnabled(true).SetName("std")
	si.deltaAvg.SetCacheEnabled(true).SetName("delta_avg")
	si.deltaStd.SetCacheEnabled(true).SetName("delta_std")
	si.sample.SetCacheEnabled(true).SetName("sample")

	si.enabled.Store(false)
	si.visible.Store(false)
	return si
}
