package gfx

import (
	"image/color"
)

const (
	defaultLabelName = "Label"
	NoLabel          = ""
)

/******************************************************************************
 Label
******************************************************************************/

type Label struct {
	View

	text       string
	fontFamily string
	alignment  Alignment

	cacheEnabled bool

	texture  Texture
	textView *View
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (l *Label) Init(window *Window) (ok bool) {
	if !l.View.Init(window) {
		return false
	}

	return l.textView.Init(window)
}

func (l *Label) Update(deltaTime int64) (ok bool) {
	if !l.View.Update(deltaTime) {
		return false
	}

	if l.stateChanged.Load() {
		l.stateChanged.Store(false)
		scale := l.WorldScale()
		l.stateMutex.Lock()
		img := rasterizeText(
			l.text,
			l.fontFamily,
			l.alignment,
			FloatArrayToRgba(l.color),
			int(l.window.Width()), int(l.window.Height()),
			scale.X(),
			scale.Y(),
			l.maintainAspectRatio,
			l.cacheEnabled)
		l.texture = NewTexture(img)
		l.texture.init()
		l.stateMutex.Unlock()
		l.textView.SetTexture(l.texture)
	}

	return l.textView.Update(deltaTime)
}

func (l *Label) Draw(deltaTime int64) (ok bool) {
	if !l.View.Draw(deltaTime) {
		return false
	}

	return l.textView.Draw(deltaTime)
}

func (l *Label) Close() {
	l.View.Close()
	l.textView.Close()
	if l.texture != nil {
		l.texture.close()
	}
}

func (l *Label) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	l.stateChanged.Store(true)
	l.View.Resize(oldWidth, oldHeight, newWidth, newHeight)
	l.textView.Resize(oldWidth, oldHeight, newWidth, newHeight)
}

func (l *Label) RefreshLayout() {
	l.WindowObjectBase.RefreshLayout()
	l.textView.RefreshLayout()
}

func (l *Label) SetFillColor(rgba color.RGBA) *Label {
	l.fill.SetColor(rgba)
	return l
}

func (l *Label) SetColor(rgba color.RGBA) WindowObject {
	l.WindowObjectBase.SetColor(rgba)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) Margin() *Margin {
	l.stateMutex.Lock()
	m := l.textView.margin
	l.stateMutex.Unlock()
	return &m
}

func (l *Label) SetMargin(margin Margin) WindowObject {
	l.stateMutex.Lock()
	l.textView.margin = margin
	l.stateMutex.Unlock()
	return l
}

func (l *Label) SetMarginTop(margin float32) WindowObject {
	l.stateMutex.Lock()
	l.textView.margin.Top = margin
	l.stateMutex.Unlock()
	return l
}

func (l *Label) SetMarginRight(margin float32) WindowObject {
	l.stateMutex.Lock()
	l.textView.margin.Right = margin
	l.stateMutex.Unlock()
	return l
}

func (l *Label) SetMarginBottom(margin float32) WindowObject {
	l.stateMutex.Lock()
	l.textView.margin.Bottom = margin
	l.stateMutex.Unlock()
	return l
}

func (l *Label) SetMarginLeft(margin float32) WindowObject {
	l.stateMutex.Lock()
	l.textView.margin.Left = margin
	l.stateMutex.Unlock()
	return l
}

/******************************************************************************
 Label Functions
******************************************************************************/

func (l *Label) defaultLayout() {
	l.SetColor(White)
	l.SetFillColor(color.RGBA{})

	l.textView.SetFillColor(White)
	l.textView.SetAnchor(Center)
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
	l.scale[1] = size
	l.stateMutex.Unlock()
	return l
}

func (l *Label) Alignment() Alignment {
	l.stateMutex.Lock()
	alignment := l.alignment
	l.stateMutex.Unlock()
	return alignment
}

func (l *Label) SetAlignment(alignment Alignment) *Label {
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

func (l *Label) SetMaintainAspectRatio(maintainAspectRatio bool) WindowObject {
	l.View.SetMaintainAspectRatio(maintainAspectRatio)
	l.textView.SetMaintainAspectRatio(maintainAspectRatio)
	return l
}

func (l *Label) SetPaddingTop(padding float32) *Label {
	l.stateMutex.Lock()
	l.textView.margin.Top = padding
	l.stateMutex.Unlock()
	return l
}

func (l *Label) SetPaddingRight(padding float32) *Label {
	l.stateMutex.Lock()
	l.textView.margin.Right = padding
	l.stateMutex.Unlock()
	return l
}

func (l *Label) SetPaddingBottom(padding float32) *Label {
	l.stateMutex.Lock()
	l.textView.margin.Bottom = padding
	l.stateMutex.Unlock()
	return l
}

func (l *Label) SetPaddingLeft(padding float32) *Label {
	l.stateMutex.Lock()
	l.textView.margin.Left = padding
	l.stateMutex.Unlock()
	return l
}

/******************************************************************************
 New Label Function
******************************************************************************/

func NewLabel() *Label {
	l := &Label{
		View:         *NewView(),
		fontFamily:   DefaultFont,
		cacheEnabled: true,
		textView:     NewView(),
	}

	l.SetName(defaultLabelName)
	l.fill.SetParent(l)
	l.border.SetParent(l)
	l.textView.SetParent(l)
	l.defaultLayout()

	return l
}
