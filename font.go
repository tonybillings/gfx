package gfx

import (
	_ "embed"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	DefaultFont = "default"
	SquareFont  = "square"
	AnitaFont   = "anita"
	fontsDir    = "fonts"
)

var (
	//go:embed fonts/default.ttf
	defaultFontTtf string

	//go:embed fonts/square.ttf
	squareFontTtf string

	//go:embed fonts/anita.ttf
	anitaFontTtf string
)

var (
	loadedFonts      map[string]*truetype.Font
	loadedFontsMutex = sync.Mutex{}
	renderedLabels   = make(map[string]*image.RGBA)
)

func initFonts() error {
	defaultFont, err := truetype.Parse([]byte(defaultFontTtf))
	if err != nil {
		return err
	}

	squareFont, err := truetype.Parse([]byte(squareFontTtf))
	if err != nil {
		return err
	}

	anitaFont, err := truetype.Parse([]byte(anitaFontTtf))
	if err != nil {
		return err
	}

	loadedFonts = make(map[string]*truetype.Font)
	loadedFonts[DefaultFont] = defaultFont
	loadedFonts[SquareFont] = squareFont
	loadedFonts[AnitaFont] = anitaFont

	return loadFonts()
}

func loadFonts() error {
	info, err := os.Stat(fontsDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return nil
	}

	files, err := os.ReadDir(fontsDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Type().IsRegular() && strings.ToLower(filepath.Ext(file.Name())) == ".ttf" {
			fullPath := filepath.Join(fontsDir, file.Name())

			data, e := os.ReadFile(fullPath)
			if e != nil {
				fmt.Printf("Could not read font file: %v\n", e)
				return e
			}

			f, e := truetype.Parse(data)
			if e != nil {
				fmt.Printf("Could not parse font file: %v\n", e)
				return e
			}

			extension := filepath.Ext(file.Name())
			fileName := file.Name()[0 : len(file.Name())-len(extension)]
			loadedFontsMutex.Lock()
			loadedFonts[fileName] = f
			loadedFontsMutex.Unlock()
		}
	}

	for name, asset := range loadedAssets {
		if strings.ToLower(filepath.Ext(name)) == ".ttf" {
			f, e := truetype.Parse(asset)
			if e != nil {
				fmt.Printf("Could not parse font file: %v\n", e)
				return e
			}

			loadedFontsMutex.Lock()
			loadedFonts[name] = f
			loadedFontsMutex.Unlock()
		}
	}

	return nil
}

func getFontOrDefault(fontName string) *truetype.Font {
	loadedFontsMutex.Lock()
	defer loadedFontsMutex.Unlock()

	if f, ok := loadedFonts[fontName]; ok {
		return f
	} else {
		return loadedFonts[DefaultFont]
	}
}

func rasterizeText(windowWidth, windowHeight int, text string, fontFamily string, fontSize float32,
	position mgl32.Vec3, rgba color.RGBA, alignment Alignment, useCache bool) *image.RGBA {

	id := fmt.Sprintf("%d%d%s%s%f%v%v%d", windowWidth, windowHeight, text, fontFamily, fontSize, position, rgba, alignment)

	var img *image.RGBA
	var ok bool
	if useCache {
		if img, ok = renderedLabels[id]; ok {
			return img
		}
		img = image.NewRGBA(image.Rect(0, 0, windowWidth, windowHeight))
		renderedLabels[id] = img
	} else {
		img = image.NewRGBA(image.Rect(0, 0, windowWidth, windowHeight))
	}

	textFont := getFontOrDefault(fontFamily)
	ctx := freetype.NewContext()
	ctx.SetFont(textFont)

	ctx.SetDPI(72.0)

	absFontSize := float64(windowHeight) * float64(fontSize)
	if windowHeight > windowWidth {
		absFontSize *= float64(windowWidth) / float64(windowHeight)
	}

	fpFontSize := fixed.Int26_6(absFontSize)
	ctx.SetFontSize(absFontSize)

	ctx.SetClip(img.Bounds())
	ctx.SetDst(img)
	ctx.SetSrc(image.NewUniform(rgba))

	textHeightInPixels := measureTextHeight(textFont, text, fpFontSize)
	textHeightNormalized := ((textHeightInPixels) / float32(windowHeight)) * 2.0

	textWidthNormalized := (measureTextWidth(textFont, text, fpFontSize) / float32(windowWidth)) * 2.0

	switch alignment {
	case TopLeft:
		position[1] -= textHeightNormalized
	case MiddleLeft:
		position[1] -= textHeightNormalized * .5
	case BottomLeft:
		break
	case TopCenter:
		position[0] -= textWidthNormalized * .5
	case Center:
		position[0] -= textWidthNormalized * .5
		position[1] -= textHeightNormalized * .5
	case BottomCenter:
		position[0] -= textWidthNormalized * .5
		position[1] -= textHeightNormalized
	case TopRight:
		position[1] -= textHeightNormalized
		position[0] -= textWidthNormalized
	case MiddleRight:
		position[1] -= textHeightNormalized * .5
		position[0] -= textWidthNormalized
	case BottomRight:
		position[0] -= textWidthNormalized
	}

	posX := ((position[0] + 1.0) * .5) * float32(windowWidth)
	posY := (1.0 - ((position[1] + 1.0) * .5)) * float32(windowHeight)
	pt := freetype.Pt(int(posX), int(posY))
	ctx.SetHinting(font.HintingVertical)

	_, _ = ctx.DrawString(text, pt)

	return img
}

func measureTextWidth(f *truetype.Font, text string, scale fixed.Int26_6) float32 {
	var totalWidth fixed.Int26_6
	for _, runeValue := range text {
		glyphIndex := f.Index(runeValue)
		hMetric := f.HMetric(scale, glyphIndex)
		totalWidth += hMetric.AdvanceWidth
	}
	return float32(totalWidth)
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
