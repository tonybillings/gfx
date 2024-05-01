package gfx

import (
	"github.com/golang/freetype/truetype"
	"image/color"
	"sync/atomic"
)

const (
	defaultLabelName = "Label"
)

/******************************************************************************
 Label
******************************************************************************/

type Label struct {
	View

	text      string
	font      Font
	ttf       *truetype.Font
	alignment Alignment

	cache        map[string]*Texture2D
	cacheEnabled bool

	texture  *Texture2D
	textView *View

	stateChanged atomic.Bool
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (l *Label) Init() (ok bool) {
	if l.Initialized() {
		return true
	}

	l.textView.fill.FlipUV(true)
	if ok = l.textView.Init(); !ok {
		return
	}

	l.initFont()
	l.initCache()

	return l.View.Init()
}

func (l *Label) Update(deltaTime int64) (ok bool) {
	if !l.View.Update(deltaTime) {
		return false
	}

	return l.textView.Update(deltaTime)
}

func (l *Label) Close() {
	l.View.Close()
	l.textView.Close()
	if l.texture != nil {
		l.texture.Close()
	}
}

/******************************************************************************
 DrawableObject Implementation
******************************************************************************/

func (l *Label) Draw(deltaTime int64) (ok bool) {
	if !l.View.Draw(deltaTime) {
		return false
	}

	if l.stateChanged.Load() {
		l.stateChanged.Store(false)
		scale := l.WorldScale()

		if l.ttf != nil {
			l.stateMutex.Lock()
			if l.texture != nil {
				l.texture.Close()
			}
			l.texture = textToTexture(
				l.text,
				l.ttf,
				l.alignment,
				FloatArrayToRgba(l.color),
				l.window.Width(), l.window.Height(),
				scale.X(),
				scale.Y(),
				l.maintainAspectRatio,
				l.cacheEnabled,
				l.cache)
			l.texture.Init()
			l.stateMutex.Unlock()
			l.textView.SetTexture(l.texture)
		}
	}

	return l.textView.Draw(deltaTime)
}

/******************************************************************************
 Resizer Implementation
******************************************************************************/

func (l *Label) Resize(newWidth, newHeight int) {
	l.stateChanged.Store(true)
	l.View.Resize(newWidth, newHeight)
	l.textView.Resize(newWidth, newHeight)
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (l *Label) SetColor(rgba color.RGBA) WindowObject {
	l.WindowObjectBase.SetColor(rgba)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) SetMaintainAspectRatio(maintainAspectRatio bool) WindowObject {
	l.View.SetMaintainAspectRatio(maintainAspectRatio)
	l.textView.SetMaintainAspectRatio(maintainAspectRatio)
	l.stateChanged.Store(true)
	return l
}

func (l *Label) RefreshLayout() {
	l.WindowObjectBase.RefreshLayout()
	l.textView.RefreshLayout()
}

func (l *Label) SetWindow(window *Window) WindowObject {
	l.View.SetWindow(window)
	l.textView.SetWindow(window)
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

func (l *Label) initFont() {
	if l.font == nil {
		l.font = l.window.getFontOrDefault(DefaultFont)
		l.ttf = l.font.TTF()
	}
}

func (l *Label) initCache() {
	if l.cacheEnabled {
		l.cache = l.Window().labelCache
	} else {
		l.cache = nil
	}
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

func (l *Label) Font() Font {
	l.stateMutex.Lock()
	ttfFont := l.font
	l.stateMutex.Unlock()
	return ttfFont
}

func (l *Label) SetFont(ttfFont Font) *Label {
	l.stateMutex.Lock()
	l.font = ttfFont
	l.ttf = ttfFont.TTF()
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
	if enabled {
		l.cache = l.Window().labelCache
	} else {
		l.cache = nil
	}
	l.cacheEnabled = enabled
	l.stateMutex.Unlock()
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

func (l *Label) SetFillColor(rgba color.RGBA) *Label {
	l.fill.SetColor(rgba)
	return l
}

/******************************************************************************
 New Label Function
******************************************************************************/

func NewLabel() *Label {
	l := &Label{
		View:     *NewView(),
		textView: NewView(),
	}

	l.SetName(defaultLabelName)
	l.fill.SetParent(l)
	l.border.SetParent(l)
	l.textView.SetParent(l)
	l.defaultLayout()

	return l
}
