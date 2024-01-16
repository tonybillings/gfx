package gfx

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"math"
	"sync"
)

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
	WindowObjectBase
	Signal

	vao uint32
	vbo uint32

	shader uint32

	colorUniformLoc     int32
	thicknessUniformLoc int32

	thickness uint

	label *Label
}

type SignalGroup struct {
	WindowObjectBase

	defaultSampleCount int
	defaultColors      []color.RGBA
	defaultColorIdx    int
	defaultThickness   uint
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

/******************************************************************************
 SignalLine
******************************************************************************/

func (l *SignalLine) initVertices() {
	l.vertices = make([]float32, l.vertexCount*2)
	step := 2.0 / float32(l.vertexCount)
	for i := int32(0); i < l.vertexCount; i++ {
		l.vertices[i*2] = -1 + float32(i)*step
		l.vertices[i*2+1] = l.position[1]
	}
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

func (l *SignalLine) Init(window *Window) (ok bool) {
	if !l.WindowObjectBase.Init(window) {
		return false
	}

	l.initVertices()
	l.initGl()
	l.initialized.Store(true)

	return l.label.Init(window)
}

func (l *SignalLine) Update(deltaTime int64) (ok bool) {
	if !l.WindowObjectBase.Update(deltaTime) {
		return false
	}

	if l.stateChanged.Load() {
		l.stateChanged.Store(false)
		l.updateVertices()
	}

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

	gl.UseProgram(l.shader)

	l.stateMutex.Lock()
	gl.Uniform1f(l.thicknessUniformLoc, float32(l.thickness))
	gl.Uniform4fv(l.colorUniformLoc, 1, &l.color[0])
	l.stateMutex.Unlock()

	gl.BindVertexArray(l.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, l.vbo)

	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(l.vertices)*sizeOfFloat32, gl.Ptr(l.vertices))

	gl.DrawArrays(gl.LINE_STRIP_ADJACENCY, 0, l.vertexCount)

	gl.BindVertexArray(0)
	gl.UseProgram(0)

	return l.label.Draw(deltaTime)
}

func (l *SignalLine) Close() {
	l.label.Close()
	l.WindowObjectBase.Close()
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

/******************************************************************************
 SignalGroup
******************************************************************************/

func (g *SignalGroup) updateSignalLayout() {
	g.stateMutex.Lock()
	signalCount := float32(len(g.children))
	scale := g.scale[1] / signalCount
	step := 2.0 * scale
	halfStep := step * .5
	for i, c := range g.children {
		if s, ok := c.(*SignalLine); ok {
			s.position[1] = 1.0 - (float32(i) * step) - halfStep
			s.scale[1] = scale
			s.label.SetPosition(mgl32.Vec3{-1, s.position[1]})
			s.label.SetScaleY(scale * .33)
			s.stateChanged.Store(true)
		}
	}
	g.stateMutex.Unlock()
}

func (g *SignalGroup) Update(deltaTime int64) (ok bool) {
	if g.enabled.Load() == true && g.stateChanged.Load() {
		g.stateChanged.Store(false)
		g.updateSignalLayout()
	}

	if !g.WindowObjectBase.Update(deltaTime) {
		return false
	}

	return true
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
	g.updateSignalLayout()
	return newSignal
}

func (g *SignalGroup) Add(signal *SignalLine) *SignalLine {
	g.stateMutex.Lock()
	g.children = append(g.Children(), signal)
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

func (g *SignalGroup) Signals() []*SignalLine {
	signals := make([]*SignalLine, len(g.children))
	for i, c := range g.children {
		if s, ok := c.(*SignalLine); ok {
			signals[i] = s
		}
	}
	return signals
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
	lbl.SetPosition(mgl32.Vec3{-1})

	sl := &SignalLine{
		WindowObjectBase: WindowObjectBase{
			color:       RgbaToFloatArray(White),
			position:    lbl.Position(),
			rotation:    [3]float32{0, 0, 0},
			scale:       [3]float32{1, 1, 1},
			vertices:    make([]float32, sampleCount*2),
			vertexCount: int32(sampleCount),
			children:    make([]WindowObject, 0),
		},
		Signal:    *NewSignal(label, sampleCount),
		thickness: 1,
		label:     lbl,
	}

	sl.name.Store(&label)
	sl.enabled.Store(true)
	sl.visible.Store(true)
	return sl
}

func NewSignalGroup(defaultSampleCount int, defaultLineThickness int, colors ...color.RGBA) *SignalGroup {
	if len(colors) == 0 {
		colors = DefaultColors
	}

	sg := &SignalGroup{
		WindowObjectBase:   *NewObject(nil),
		defaultSampleCount: defaultSampleCount,
		defaultThickness:   uint(defaultLineThickness),
		defaultColors:      colors,
	}

	return sg
}
