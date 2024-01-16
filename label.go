package gfx

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
)

const (
	defaultLabelName = "Label"
	NoLabel          = ""
)

type Label struct {
	WindowObjectBase

	vao uint32
	vbo uint32

	shader uint32

	text         string
	fontFamily   string
	fontSize     float32
	cacheEnabled bool

	alignment TextAlignment

	texture           uint32
	textureWidth      int
	textureHeight     int
	textureUniformLoc int32
}

func (l *Label) initVertices() {
	l.stateMutex.Lock()
	l.vertices = []float32{
		-1.0, 1.0, 0.0, 0.0,
		1.0, 1.0, 1.0, 0.0,
		1.0, -1.0, 1.0, 1.0,
		-1.0, -1.0, 0.0, 1.0,
	}
	l.vertexCount = int32(len(l.vertices) / 2)
	l.stateMutex.Unlock()
}

func (l *Label) initGl() {
	l.shader = GetShaderProgram(TextShaderProgram)
	l.textureUniformLoc = gl.GetUniformLocation(l.shader, gl.Str("textTexture\x00"))

	gl.GenVertexArrays(1, &l.vao)
	gl.GenBuffers(1, &l.vbo)

	gl.BindVertexArray(l.vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, l.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(l.vertices)*sizeOfFloat32, gl.Ptr(l.vertices), gl.STATIC_DRAW)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(0, 2, gl.FLOAT, false, 4*4, 0)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, 4*4, 2*4)

	gl.GenTextures(1, &l.texture)
	gl.BindTexture(gl.TEXTURE_2D, l.texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func (l *Label) uninitGl() {
	gl.DeleteTextures(1, &l.texture)
}

func (l *Label) Init(window *Window) (ok bool) {
	if !l.WindowObjectBase.Init(window) {
		return false
	}

	l.initVertices()
	l.initGl()
	l.initialized.Store(true)

	return true
}

func (l *Label) Draw(_ int64) (ok bool) {
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

	worldPos := l.WorldPosition()
	worldScale := l.WorldScale()

	l.stateMutex.Lock()
	if l.stateChanged.Load() {
		l.stateChanged.Store(false)

		winWidth, winHeight := l.window.win.GetSize()
		img := rasterizeText(winWidth, winHeight, l.text, l.fontFamily, worldScale[1]*l.fontSize,
			worldPos,
			FloatArrayToRgba(l.color),
			l.alignment,
			l.cacheEnabled)
		l.stateMutex.Unlock()

		l.textureWidth = img.Rect.Dx()
		l.textureHeight = img.Rect.Dy()

		gl.BindTexture(gl.TEXTURE_2D, l.texture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(l.textureWidth), int32(l.textureHeight), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
	} else {
		l.stateMutex.Unlock()
	}

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, l.texture)
	gl.Uniform1i(l.textureUniformLoc, 0)

	gl.BindVertexArray(l.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, l.vbo)

	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	gl.Disable(gl.BLEND)

	gl.BindVertexArray(0)
	gl.UseProgram(0)

	return true
}

func (l *Label) SetColor(rgba color.RGBA) WindowObject {
	l.WindowObjectBase.SetColor(rgba)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) SetOrigin(origin mgl32.Vec3) WindowObject {
	l.WindowObjectBase.SetOrigin(origin)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) SetPosition(position mgl32.Vec3) WindowObject {
	l.WindowObjectBase.SetPosition(position)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) SetPositionX(x float32) WindowObject {
	l.WindowObjectBase.SetPositionX(x)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) SetPositionY(y float32) WindowObject {
	l.WindowObjectBase.SetPositionY(y)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) SetPositionZ(z float32) WindowObject {
	l.WindowObjectBase.SetPositionZ(z)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) SetRotation(rotation mgl32.Vec3) WindowObject {
	l.WindowObjectBase.SetRotation(rotation)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) SetRotationX(x float32) WindowObject {
	l.WindowObjectBase.SetRotationX(x)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) SetRotationY(y float32) WindowObject {
	l.WindowObjectBase.SetRotationY(y)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) SetRotationZ(z float32) WindowObject {
	l.WindowObjectBase.SetRotationZ(z)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) SetScale(scale mgl32.Vec3) WindowObject {
	l.WindowObjectBase.SetScale(scale)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) SetScaleX(x float32) WindowObject {
	l.WindowObjectBase.SetScaleX(x)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) SetScaleY(y float32) WindowObject {
	l.WindowObjectBase.SetScaleY(y)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) SetScaleZ(z float32) WindowObject {
	l.WindowObjectBase.SetScaleZ(z)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) MaintainAspectRatio(maintainAspectRatio bool) WindowObject {
	l.WindowObjectBase.MaintainAspectRatio(maintainAspectRatio)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) Text() string {
	l.stateMutex.Lock()
	text := l.text
	l.stateMutex.Unlock()
	return text
}

func (l *Label) SetText(text string) *Label {
	l.stateMutex.Lock()
	l.text = text
	l.stateChanged.Store(true)
	l.stateMutex.Unlock()
	return l
}

func (l *Label) FontFamily() string {
	l.stateMutex.Lock()
	font := l.fontFamily
	l.stateMutex.Unlock()
	return font
}

func (l *Label) SetFontFamily(font string) *Label {
	l.stateMutex.Lock()
	l.fontFamily = font
	l.stateChanged.Store(true)
	l.stateMutex.Unlock()
	return l
}

func (l *Label) SetFontSize(size float32) *Label {
	l.stateMutex.Lock()
	l.fontSize = size
	l.stateChanged.Store(true)
	l.stateMutex.Unlock()
	return l
}

func (l *Label) Alignment() TextAlignment {
	l.stateMutex.Lock()
	alignment := l.alignment
	l.stateMutex.Unlock()
	return alignment
}

func (l *Label) SetAlignment(alignment TextAlignment) *Label {
	l.stateMutex.Lock()
	l.alignment = alignment
	l.stateChanged.Store(true)
	l.stateMutex.Unlock()
	return l
}

func (l *Label) CacheEnabled() bool {
	l.stateMutex.Lock()
	enabled := l.cacheEnabled
	l.stateMutex.Unlock()
	return enabled
}

func (l *Label) SetCacheEnabled(enabled bool) *Label {
	l.stateMutex.Lock()
	l.cacheEnabled = enabled
	l.stateChanged.Store(true)
	l.stateMutex.Unlock()
	return l
}

func NewLabel() *Label {
	tl := &Label{
		WindowObjectBase: *NewObject(nil),
		fontFamily:       DefaultFont,
		fontSize:         1.0,
		alignment:        Center,
		cacheEnabled:     true,
	}

	tl.SetName(defaultLabelName)
	return tl
}
