package gfx

import (
	_ "embed"
	"fmt"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/tonybillings/gfx/fonts"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"io"
)

const (
	DefaultFont = "_font_default"
	SquareFont  = "_font_square"
	AnitaFont   = "_font_anita"
)

/******************************************************************************
 Font
******************************************************************************/

type Font interface {
	Asset
	TTF() *truetype.Font
}

/******************************************************************************
 TrueTypeFont
******************************************************************************/

type TrueTypeFont struct {
	AssetBase
	ttf *truetype.Font
}

/******************************************************************************
 Asset Implementation
******************************************************************************/

func (f *TrueTypeFont) Init() bool {
	if f.Initialized() {
		return true
	}

	switch source := f.source.(type) {
	case []byte:
		f.loadFromSlice(source)
	case string:
		f.loadFromFile(source)
	default:
		panic("unexpected error: source type is not supported")
	}

	return f.AssetBase.Init()
}

func (f *TrueTypeFont) Close() {
	if !f.Initialized() {
		return
	}

	f.ttf = nil

	f.AssetBase.Close()
}

/******************************************************************************
 Font Functions
******************************************************************************/

func (f *TrueTypeFont) loadFromSlice(slice []byte) {
	if ttf, err := truetype.Parse(slice); err != nil {
		panic(fmt.Errorf("TTF parsing error: %w", err))
	} else {
		f.ttf = ttf
	}
}

func (f *TrueTypeFont) loadFromFile(name string) {
	reader, closeFunc := f.getSourceReader(name)
	defer closeFunc()

	if reader == nil {
		panic("font file not found")
	}

	buffer := make([]byte, reader.Size())
	fontBytes := make([]byte, 0)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			fontBytes = append(fontBytes, buffer[:n]...)
		} else {
			if err != nil && err != io.EOF {
				panic(fmt.Errorf("font file read error: %w", err))
			}
			break
		}
	}

	f.loadFromSlice(fontBytes)
}

func (f *TrueTypeFont) TTF() *truetype.Font {
	return f.ttf
}

/******************************************************************************
 New Font Function
******************************************************************************/

func NewFont[T FontSource](name string, source T) *TrueTypeFont {
	return &TrueTypeFont{
		AssetBase: AssetBase{
			name:   name,
			source: source,
		},
	}
}

/******************************************************************************
 Initialize Default Fonts
******************************************************************************/

func newDefaultFont(assetName string, filename string) *TrueTypeFont {
	fs := fonts.Assets

	if fontFile, err := fs.ReadFile(filename + ".ttf"); err != nil {
		panic(fmt.Errorf("font file read error: %w", err))
	} else {
		ttf := NewFont(assetName, fontFile)
		ttf.SetProtected(true)
		return ttf
	}
}

func addDefaultFonts(lib *AssetLibrary) {
	const prefix = "_font_"
	const pfxLen = len(prefix)

	lib.Add(newDefaultFont(DefaultFont, DefaultFont[pfxLen:]))
	lib.Add(newDefaultFont(SquareFont, SquareFont[pfxLen:]))
	lib.Add(newDefaultFont(AnitaFont, AnitaFont[pfxLen:]))
}

/******************************************************************************
 Rasterization Functions
******************************************************************************/

func textToTexture(text string, ttf *truetype.Font, alignment Alignment, rgba color.RGBA,
	windowWidth, windowHeight int, scaleX, scaleY float32, maintainAspectRatio, cacheEnabled bool, cache map[string]*Texture2D) *Texture2D {

	scale := [2]float32{}
	if maintainAspectRatio {
		switch {
		case windowWidth > windowHeight:
			scale[0] = scaleX * (float32(windowHeight) / float32(windowWidth))
			scale[1] = scaleY
		case windowHeight > windowWidth:
			scale[0] = scaleX
			scale[1] = scaleY * (float32(windowWidth) / float32(windowHeight))
		default:
			scale[0] = scaleX
			scale[1] = scaleY
		}
	} else {
		scale[0] = scaleX
		scale[1] = scaleY
	}

	id := fmt.Sprintf("%s%v%v%v%f%f", text, ttf.Name(1), alignment, rgba, scale[0], scale[1])

	if cacheEnabled {
		if cached := cache[id]; cached != nil {
			return cached
		}
	}

	absFontSize := float64(scale[1] * float32(windowHeight))

	img := image.NewRGBA(image.Rect(0, 0, int(scale[0]*float32(windowWidth)*1.005), int(scale[1]*float32(windowHeight)*1.005)))

	ctx := freetype.NewContext()
	ctx.SetFont(ttf)
	ctx.SetFontSize(absFontSize)
	ctx.SetDPI(72.0)
	ctx.SetHinting(font.HintingVertical)
	ctx.SetClip(img.Bounds())
	ctx.SetSrc(image.NewUniform(rgba))
	ctx.SetDst(img)

	fpFontSize := fixed.Int26_6(scale[1] * float32(windowHeight))
	textWidth := measureTextWidth(ttf, text, fpFontSize)
	textHeight := measureTextHeight(ttf, text, fpFontSize)

	shapeWidth := float32(img.Bounds().Size().X) * 0.995
	shapeHeight := float32(img.Bounds().Size().Y) * 0.995

	hSpacing := 0
	vSpacing := int((shapeHeight + textHeight) * 0.5)
	switch alignment {
	case Centered:
		hSpacing = int((shapeWidth - float32(textWidth)) * 0.5)
	case Right:
		hSpacing = int(shapeWidth - float32(textWidth))
	}
	pt := freetype.Pt(hSpacing, vSpacing)

	_, _ = ctx.DrawString(text, pt)

	texture := NewTexture2D(id, img)

	if cacheEnabled {
		cache[id] = texture
	}

	return texture
}

func measureTextWidth(f *truetype.Font, text string, scale fixed.Int26_6) float64 {
	var totalWidth fixed.Int26_6
	for _, runeValue := range text {
		glyphIndex := f.Index(runeValue)
		hMetric := f.HMetric(scale, glyphIndex)
		totalWidth += hMetric.AdvanceWidth
	}

	return float64(totalWidth)
}

func measureTextHeight(f *truetype.Font, text string, scale fixed.Int26_6) float32 {
	var maxTop, minBottom fixed.Int26_6
	for _, runeValue := range text {
		glyphIndex := f.Index(runeValue)
		vMetric := f.VMetric(scale, glyphIndex)
		top := vMetric.TopSideBearing
		bottom := vMetric.TopSideBearing - vMetric.AdvanceHeight

		if top > maxTop {
			maxTop = top
		}
		if bottom < minBottom {
			minBottom = bottom
		}
	}

	if minBottom < 0 {
		return float32(maxTop-(minBottom)*2) * .5
	} else {
		return float32(maxTop - minBottom*2)
	}
}
