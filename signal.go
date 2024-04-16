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
 Signal
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

	filters      []Filter
	dataFiltered []float64

	transformers       []Transformer
	dataTransformed    []float64
	minTransformedData float64
	maxTransformedData float64

	fftEnabled     bool
	fftTransformer *FastFourierTransformer
	freqSpectrum   []float64
}

/******************************************************************************
 SignalLine
******************************************************************************/

type SignalLine struct {
	View
	Signal

	vertices    []float32
	vertexCount int32

	vao uint32
	vbo uint32

	shader Shader

	colorUniformLoc     int32
	thicknessUniformLoc int32

	thickness uint

	label *Label

	inspector            *SignalInspector
	inspectorKey         glfw.Key
	inspectorActive      atomic.Bool
	inspectKeyRegistered atomic.Bool

	dataExportKey           glfw.Key
	dataExportKeyRegistered atomic.Bool

	stateChanged atomic.Bool
}

/******************************************************************************
 SignalGroup
******************************************************************************/

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

	dataExportKey           glfw.Key
	dataExportKeyRegistered atomic.Bool

	stateChanged atomic.Bool
}

/******************************************************************************
 SignalInspector
******************************************************************************/

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

	fontSize float32

	stateChanged atomic.Bool
}

/******************************************************************************
 SignalControl
******************************************************************************/

type SignalControl interface {
	*SignalLine | *SignalGroup
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (l *SignalLine) Init() (ok bool) {
	if l.Initialized() {
		return true
	}

	if ok = l.View.Init(); !ok {
		return
	}

	l.initVertices()
	l.initVertexVao()

	if l.dataExportKey != glfw.KeyUnknown {
		l.EnableDataExportKey(l.dataExportKey)
	}

	if l.inspectorKey != glfw.KeyUnknown {
		l.EnableInspector()
	}
	l.inspector.Init()

	return l.label.Init()
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
	l.label.Update(deltaTime)

	return true
}

func (l *SignalLine) Close() {
	l.View.Close()
	l.label.Close()
	l.inspector.Close()
	l.closeVertexVao()
}

func (g *SignalGroup) Init() (ok bool) {
	if g.Initialized() {
		return true
	}

	g.initLayout()

	if ok = g.View.Init(); !ok {
		return
	}

	if g.dataExportKey != glfw.KeyUnknown {
		g.EnableDataExportKey(g.dataExportKey)
	}

	if g.inspectorKey != glfw.KeyUnknown {
		g.EnableInspector()
	}

	return g.inspector.Init()
}

func (g *SignalGroup) Update(deltaTime int64) (ok bool) {
	if !g.View.Update(deltaTime) {
		return false
	}

	if g.stateChanged.Load() {
		g.stateChanged.Store(false)
		g.updateSignalLayout()
	}

	g.inspector.Update(deltaTime)

	return true
}

func (g *SignalGroup) Close() {
	g.View.Close()
	g.inspector.Close()
}

func (g *SignalGroup) SetEnabled(enabled bool) Object {
	if g.enabled.Load() != enabled {
		g.stateChanged.Store(true)
	}

	g.WindowObjectBase.SetEnabled(enabled)
	return g
}

func (i *SignalInspector) Init() (ok bool) {
	if i.Initialized() {
		return true
	}

	i.initLayout()
	i.RefreshLayout()

	return i.WindowObjectBase.Init()
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
		if signal.fftEnabled {
			idx := int(xScale * float32(signal.dataSize-1))
			i.minValue.SetText("")
			i.maxValue.SetText("")
			i.avg.SetText(fmt.Sprintf("Freq: %.6f", signal.freqSpectrum[idx])) // hijacking this label
			i.std.SetText("")
			i.deltaAvg.SetText(fmt.Sprintf("Mag: %.6f", signal.dataTransformed[idx])) // hijacking this label
			i.deltaStd.SetText("")
			i.sample.SetText("")
		} else {
			sampleIdx := int(xScale * float32(signal.BufferSize()-1))
			i.minValue.SetText(fmt.Sprintf("Min:  %.6f", signal.MinValue()))
			i.maxValue.SetText(fmt.Sprintf("Max:  %.6f", signal.MaxValue()))
			i.avg.SetText(fmt.Sprintf("Avg:  %.6f", signal.Average()))
			i.std.SetText(fmt.Sprintf("Std:  %.6f", signal.StdDev()))
			i.deltaAvg.SetText(fmt.Sprintf("ΔAvg: %.6f", signal.DeltaAverage()))
			i.deltaStd.SetText(fmt.Sprintf("ΔStd: %.6f", signal.DeltaStdDev(false)))
			i.sample.SetText(fmt.Sprintf("%f", signal.dataTransformed[sampleIdx]))
		}
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

func (i *SignalInspector) Close() {
	i.bounds.Close()
	i.panel.Close()
}

/******************************************************************************
 DrawableObject Implementation
******************************************************************************/

func (l *SignalLine) Draw(deltaTime int64) (ok bool) {
	if !l.visible.Load() {
		return false
	}

	l.fill.Draw(deltaTime)

	l.shader.Activate()

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

	l.border.Draw(deltaTime)
	l.label.Draw(deltaTime)
	l.drawChildren(deltaTime)
	l.inspector.Draw(deltaTime)

	return true
}

func (g *SignalGroup) Draw(deltaTime int64) (ok bool) {
	if !g.View.Draw(deltaTime) {
		return false
	}

	g.inspector.Draw(deltaTime)

	return g.inspector.Draw(deltaTime)
}

func (i *SignalInspector) Draw(deltaTime int64) (ok bool) {
	if !i.visible.Load() {
		return false
	}

	i.panel.Draw(deltaTime)

	return i.WindowObjectBase.Draw(deltaTime)
}

/******************************************************************************
 Resizer Implementation
******************************************************************************/

func (l *SignalLine) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	l.View.Resize(oldWidth, oldHeight, newWidth, newHeight)
	l.label.Resize(oldWidth, oldHeight, newWidth, newHeight)
	l.inspector.Resize(oldWidth, oldHeight, newWidth, newHeight)
	l.initVertices()
}

/******************************************************************************
 Resizer Implementation
******************************************************************************/

func (g *SignalGroup) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	g.inspector.Resize(oldWidth, oldHeight, newWidth, newHeight)
	g.View.Resize(oldWidth, oldHeight, newWidth, newHeight)
}

/******************************************************************************
 Resizer Implementation
******************************************************************************/

func (i *SignalInspector) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	i.WindowObjectBase.Resize(oldWidth, oldHeight, newWidth, newHeight)
	i.panel.Resize(oldWidth, oldHeight, newWidth, newHeight)
	i.stateChanged.Store(true)
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (l *SignalLine) SetColor(rgba color.RGBA) WindowObject {
	l.WindowObjectBase.SetColor(rgba)
	return l
}

func (i *SignalInspector) SetMaintainAspectRatio(maintainAspectRatio bool) WindowObject {
	i.bounds.SetMaintainAspectRatio(maintainAspectRatio)
	return i
}

func (l *SignalLine) SetWindow(window *Window) WindowObject {
	l.View.SetWindow(window)
	l.label.SetWindow(window)
	l.inspector.SetWindow(window)
	return l
}

func (g *SignalGroup) SetWindow(window *Window) WindowObject {
	g.View.SetWindow(window)
	g.stateMutex.Lock()
	for _, s := range g.children {
		s.SetWindow(window)
	}
	g.stateMutex.Unlock()
	g.inspector.SetWindow(window)
	return g
}

func (i *SignalInspector) SetWindow(window *Window) WindowObject {
	i.WindowObjectBase.SetWindow(window)
	i.panel.SetWindow(window)
	i.bounds.SetWindow(window)
	i.minValue.SetWindow(window)
	i.maxValue.SetWindow(window)
	i.avg.SetWindow(window)
	i.std.SetWindow(window)
	i.deltaAvg.SetWindow(window)
	i.deltaStd.SetWindow(window)
	i.sample.SetWindow(window)
	return i
}

/******************************************************************************
 sync.Locker Implementation
******************************************************************************/

func (s *Signal) Lock() {
	s.dataMutex.Lock()
}

func (s *Signal) Unlock() {
	s.dataMutex.Unlock()
}

/******************************************************************************
 Signal Functions
******************************************************************************/

func (s *Signal) Label() string {
	return s.label
}

func (s *Signal) Data() []float64 {
	return s.data
}

func (s *Signal) FilteredData() []float64 {
	return s.dataFiltered
}

func (s *Signal) AddSamples(data []float64) {
	s.dataMutex.Lock()

	for _, d := range data {
		s.data[s.dataIdx] = d

		s.dataFiltered[s.dataIdx] = d
		filterIdx := 0
		for _, f := range s.filters {
			if !f.Enabled() {
				continue
			}
			if filterIdx == 0 {
				s.dataFiltered[s.dataIdx] = f.Apply(s.dataIdx, s.data)
			} else {
				s.dataFiltered[s.dataIdx] = f.Apply(s.dataIdx, s.dataFiltered)
			}
			filterIdx++
		}

		s.dataIdx = (s.dataIdx + 1) % s.dataSize
		if d < s.minData {
			s.minData = d
		}
		if d > s.maxData {
			s.maxData = d
		}
	}

	transformed := false
	if s.fftEnabled {
		s.freqSpectrum = s.fftTransformer.Transform(s.dataTransformed, s.dataFiltered)
		transformed = true
	} else {
		for i, t := range s.transformers {
			if i == 0 {
				t.Transform(s.dataTransformed, s.dataFiltered)
			} else {
				t.Transform(s.dataTransformed, s.dataTransformed)
			}
			transformed = true
		}
	}

	if transformed {
		for _, d := range s.dataTransformed {
			if d < s.minTransformedData {
				s.minTransformedData = d
			}
			if d > s.maxTransformedData {
				s.maxTransformedData = d
			}
		}
	} else {
		s.dataTransformed = s.dataFiltered
		s.minTransformedData = s.minData
		s.maxTransformedData = s.maxData
	}

	s.dataMutex.Unlock()
}

func (s *Signal) Average() float64 {
	s.dataMutex.Lock()

	sum := 0.0
	for _, value := range s.dataTransformed {
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
		s.deltas[i-1] = s.dataTransformed[i] - s.dataTransformed[i-1]
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
	for _, value := range s.dataTransformed {
		sum += value
	}
	mean := sum / float64(s.dataSize)

	var sumOfSqrDiff float64
	for _, value := range s.dataTransformed {
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
			s.deltas[i-1] = s.dataTransformed[i] - s.dataTransformed[i-1]
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
	s.dataMutex.Lock()
	value := s.minTransformedData
	s.dataMutex.Unlock()
	return value
}

func (s *Signal) MaxValue() float64 {
	s.dataMutex.Lock()
	value := s.maxTransformedData
	s.dataMutex.Unlock()
	return value
}

func (s *Signal) BufferSize() int {
	return s.dataSize
}

func (s *Signal) AddFilter(filter Filter) {
	if s.filters == nil {
		s.filters = make([]Filter, 0)
	}
	s.filters = append(s.filters, filter)
}

func (s *Signal) GetFilter(name string) Filter {
	if s.filters == nil {
		return nil
	}

	for _, f := range s.filters {
		if f.Name() == name {
			return f
		}
	}

	return nil
}

func (s *Signal) AddTransformer(transformer Transformer) {
	if s.transformers == nil {
		s.transformers = make([]Transformer, 0)
	}
	s.transformers = append(s.transformers, transformer)
}

func (s *Signal) GetTransformer(name string) Transformer {
	if s.transformers == nil {
		return nil
	}

	for _, t := range s.transformers {
		if t.Name() == name {
			return t
		}
	}

	return nil
}

func (s *Signal) FFTEnabled() bool {
	return s.fftEnabled
}

func (s *Signal) EnableFFT(sampleRate float64) {
	if s.fftTransformer == nil {
		s.fftTransformer = NewFastFourierTransformer(s.dataSize, sampleRate)
	}
	s.dataTransformed = make([]float64, s.dataSize)
	s.minTransformedData = 0
	s.maxTransformedData = 0
	s.fftEnabled = true
}

func (s *Signal) DisableFFT() {
	s.fftEnabled = false
}

func (s *Signal) SetFFTSampleRate(rate float64) {
	if s.fftTransformer == nil {
		s.fftTransformer = NewFastFourierTransformer(s.dataSize, rate)
	}
	s.fftTransformer.SetSampleRate(rate)
}

/******************************************************************************
 SignalLine Functions
******************************************************************************/

func (l *SignalLine) initVertices() {
	l.vertices = make([]float32, l.vertexCount*2)
	step := (l.WorldScale().X() * 2.0) / float32(l.vertexCount)
	posX := l.WorldPosition().X()
	posY := l.WorldPosition().Y()
	scaleX := l.WorldScale().X()
	for i := int32(0); i < l.vertexCount; i++ {
		l.vertices[i*2] = (posX + (float32(i) * step)) - scaleX
		l.vertices[i*2+1] = posY
	}
}

func (l *SignalLine) updateVertices() {
	worldY := l.WorldPosition().Y()
	l.Signal.Lock()
	minV := l.Signal.minTransformedData
	maxV := l.Signal.maxTransformedData
	scaleY := l.scale[1]
	data := l.Signal.dataTransformed

	for i, sample := range data {
		minMaxRange := float32(maxV) - float32(minV)
		sNorm := (float32(sample) - float32(minV)) / minMaxRange
		if minMaxRange == 0.0 {
			sNorm = 0
		}
		l.vertices[i*2+1] = worldY + (((2.0 * sNorm) - 1.0) * scaleY)
	}
	l.Signal.Unlock()
}

func (l *SignalLine) initVertexVao() {
	l.shader = Assets.Get(SignalShader).(Shader)

	l.thicknessUniformLoc = l.shader.GetUniformLocation("u_Thickness")
	l.colorUniformLoc = l.shader.GetUniformLocation("u_Color")

	_, height := l.window.glwin.GetSize()
	pixelHeight := 2.0 / float32(height)
	pixelHeightUniformPos := l.shader.GetUniformLocation("u_PixelHeight")

	l.shader.Activate()
	gl.Uniform1f(pixelHeightUniformPos, pixelHeight)
	gl.UseProgram(0)

	gl.GenVertexArrays(1, &l.vao)
	gl.GenBuffers(1, &l.vbo)

	gl.BindVertexArray(l.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, l.vbo)

	posLoc := l.shader.GetAttribLocation("a_Position")
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(uint32(posLoc), 2, gl.FLOAT, false, 0, nil)

	gl.BufferData(gl.ARRAY_BUFFER, len(l.vertices)*sizeOfFloat32, gl.Ptr(l.vertices), gl.DYNAMIC_DRAW)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func (l *SignalLine) closeVertexVao() {
	l.stateMutex.Lock()

	gl.BindVertexArray(0)
	gl.DeleteVertexArrays(1, &l.vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.DeleteBuffers(1, &l.vbo)

	l.stateMutex.Unlock()
}

func (l *SignalLine) defaultLayout() {
	l.label.SetAnchor(Center)
	l.label.SetAlignment(Left)
	l.label.SetFontSize(0.33)

	l.thickness = 1

	l.SetMaintainAspectRatio(false)
	l.fill.SetMaintainAspectRatio(false)
	l.border.SetMaintainAspectRatio(false)
	l.inspector.SetMaintainAspectRatio(false)
	l.label.SetMaintainAspectRatio(false)
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

func (l *SignalLine) SetInspectorAnchor(anchor Anchor) {
	l.inspector.panel.SetAnchor(anchor)
}

func (l *SignalLine) SetInspectorMargin(margin Margin) *SignalLine {
	l.inspector.panel.SetMargin(margin)
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

func (l *SignalLine) EnableDataExportKey(key glfw.Key) *SignalLine {
	l.dataExportKey = key
	if l.dataExportKeyRegistered.Load() || !l.initialized.Load() {
		return l
	}
	l.dataExportKeyRegistered.Store(true)

	l.window.AddKeyEventHandler(key, glfw.Press, func(window *Window, key glfw.Key, action glfw.Action) {
		_ = ExportSignalDataToCsv(l)
	})

	return l
}

/******************************************************************************
 SignalGroup Functions
******************************************************************************/

func (g *SignalGroup) initLayout() {
	g.fill.SetMaintainAspectRatio(false)
	g.border.SetMaintainAspectRatio(false)
	g.inspector.SetMaintainAspectRatio(false)
}

func (g *SignalGroup) updateSignalLayout() {
	g.stateMutex.Lock()
	signalCount := float32(len(g.children))
	scale := g.scale[1] / signalCount
	step := 2.0 * scale
	halfStep := step * 0.5
	for i, c := range g.children {
		if s, ok := c.(*SignalLine); ok {
			s.position[1] = g.scale[1] - (float32(i) * step) - halfStep
			s.scale[1] = scale
			s.stateChanged.Store(true)
		}
	}
	g.stateMutex.Unlock()
}

func (g *SignalGroup) New(label string) *SignalLine {
	g.stateMutex.Lock()
	newSignal := NewSignalLine(label, g.defaultSampleCount)
	newSignal.
		SetThickness(g.defaultThickness).
		SetColor(g.defaultColors[g.defaultColorIdx]).
		SetWindow(g.window).
		SetParent(g)

	g.children = append(g.children, newSignal)
	g.defaultColorIdx = (g.defaultColorIdx + 1) % len(g.defaultColors)
	g.stateMutex.Unlock()
	g.updateSignalLayout()
	return newSignal
}

func (g *SignalGroup) Add(signal *SignalLine) *SignalLine {
	signal.SetWindow(g.window)
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

func (g *SignalGroup) SetInspectorAnchor(anchor Anchor) {
	g.inspector.panel.SetAnchor(anchor)
}

func (g *SignalGroup) SetInspectorMargin(margin Margin) *SignalGroup {
	g.inspector.panel.SetMargin(margin)
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

func (g *SignalGroup) EnableDataExportKey(key glfw.Key) *SignalGroup {
	g.dataExportKey = key
	if g.dataExportKeyRegistered.Load() || !g.initialized.Load() {
		return g
	}
	g.dataExportKeyRegistered.Store(true)

	g.window.AddKeyEventHandler(key, glfw.Press, func(window *Window, key glfw.Key, action glfw.Action) {
		_ = ExportSignalGroupDataToCsv(g)
	})

	return g
}

/******************************************************************************
 SignalInspector Functions
******************************************************************************/

func (i *SignalInspector) defaultLayout() {
	i.panel.SetAnchor(TopRight)
	i.panel.SetScale(mgl32.Vec3{0.2, 0.2, 0})
	i.panel.SetFillColor(color.RGBA{R: 52, G: 53, B: 65, A: 255})
	i.panel.SetBorderColor(White)
	i.panel.SetBorderThickness(0.02)

	i.minValue.SetName("min")
	i.maxValue.SetName("max")
	i.avg.SetName("avg")
	i.std.SetName("std")
	i.deltaAvg.SetName("delta_avg")
	i.deltaStd.SetName("delta_std")
	i.sample.SetName("sample")
}

func (i *SignalInspector) initLayout() {
	i.panel.AddChildren(i.minValue, i.maxValue, i.avg, i.std, i.deltaAvg, i.deltaStd, i.sample)
	i.panel.Init()
	i.bounds.Init()
	i.updateLayout()
}

func (i *SignalInspector) updateLayout() {
	worldScaleY := i.panel.WorldScale().Y()

	i.RefreshLayout()
	i.stateMutex.Lock()

	i.minValue.SetFontSize(i.fontSize)
	i.maxValue.SetFontSize(i.fontSize)
	i.avg.SetFontSize(i.fontSize)
	i.std.SetFontSize(i.fontSize)
	i.deltaAvg.SetFontSize(i.fontSize)
	i.deltaStd.SetFontSize(i.fontSize)
	i.sample.SetFontSize(i.fontSize)

	labelSpacing := float32(50.0)
	i.minValue.SetPosition(mgl32.Vec3{0, i.window.ScaleY(0.15 * i.fontSize * labelSpacing * worldScaleY), 0})
	i.maxValue.SetPosition(mgl32.Vec3{0, i.window.ScaleY(0.1 * i.fontSize * labelSpacing * worldScaleY), 0})
	i.avg.SetPosition(mgl32.Vec3{0, i.window.ScaleY(0.05 * i.fontSize * labelSpacing * worldScaleY), 0})
	i.std.SetPosition(mgl32.Vec3{0, i.window.ScaleY(0), 0})
	i.deltaAvg.SetPosition(mgl32.Vec3{0, i.window.ScaleY(-0.05 * i.fontSize * labelSpacing * worldScaleY), 0})
	i.deltaStd.SetPosition(mgl32.Vec3{0, i.window.ScaleY(-0.1 * i.fontSize * labelSpacing * worldScaleY), 0})
	i.sample.SetPosition(mgl32.Vec3{0, i.window.ScaleY(-0.15 * i.fontSize * labelSpacing * worldScaleY), 0})

	i.stateMutex.Unlock()
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

func (i *SignalInspector) Panel() *View {
	return i.panel
}

/******************************************************************************
 New Functions
******************************************************************************/

func NewSignal(label string, bufferSize int) *Signal {
	return &Signal{
		label:           label,
		data:            make([]float64, bufferSize),
		dataSize:        bufferSize,
		deltas:          make([]float64, bufferSize-1),
		deltasSize:      bufferSize - 1,
		dataFiltered:    make([]float64, bufferSize),
		dataTransformed: make([]float64, bufferSize),
		freqSpectrum:    make([]float64, bufferSize),
	}
}

func NewSignalLine(label string, sampleCount int) *SignalLine {
	sl := &SignalLine{
		View:          *NewView(),
		Signal:        *NewSignal(label, sampleCount),
		label:         NewLabel().SetText(label),
		dataExportKey: glfw.KeyUnknown,
	}

	sl.fill.SetParent(sl)
	sl.border.SetParent(sl)
	sl.label.SetParent(sl)

	sl.name.Store(&label)

	sl.vertices = make([]float32, sampleCount*2)
	sl.vertexCount = int32(sampleCount)

	sl.inspectorKey = glfw.KeyUnknown
	sl.inspector = NewSignalInspector(sl)
	sl.inspector.SetMaintainAspectRatio(false)

	sl.defaultLayout()

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
		dataExportKey:      glfw.KeyUnknown,
	}

	sg.inspectorKey = glfw.KeyUnknown
	sg.inspector = NewSignalInspector(sg)
	sg.inspector.SetMaintainAspectRatio(false)
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
		WindowObjectBase: *NewWindowObject(nil),
		bounds:           NewBoundingBox(),
		panel:            NewView(),
		minValue:         NewLabel(),
		maxValue:         NewLabel(),
		avg:              NewLabel(),
		std:              NewLabel(),
		deltaAvg:         NewLabel(),
		deltaStd:         NewLabel(),
		sample:           NewLabel(),
		fontSize:         0.1,
	}

	si.SetParent(any(parent).(WindowObject))
	si.bounds.SetParent(si)

	si.defaultLayout()

	si.enabled.Store(false)
	si.visible.Store(false)
	return si
}
