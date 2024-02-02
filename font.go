package gfx

import (
	_ "embed"
	"fmt"
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
	loadedFonts = make(map[string]*truetype.Font)

	if f, err := truetype.Parse([]byte(defaultFontTtf)); err == nil {
		loadedFonts[DefaultFont] = f
	} else {
		return err
	}

	if f, err := truetype.Parse([]byte(squareFontTtf)); err == nil {
		loadedFonts[SquareFont] = f
	} else {
		return err
	}

	if f, err := truetype.Parse([]byte(anitaFontTtf)); err == nil {
		loadedFonts[AnitaFont] = f
	} else {
		return err
	}

	return loadFonts()
}

func loadFonts() (err error) {
	var info os.FileInfo
	info, err = os.Stat(fontsDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return nil
	}

	var files []os.DirEntry
	files, err = os.ReadDir(fontsDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Type().IsRegular() && strings.ToLower(filepath.Ext(file.Name())) == ".ttf" {
			fullPath := filepath.Join(fontsDir, file.Name())

			var data []byte
			data, err = os.ReadFile(fullPath)
			if err != nil {
				return fmt.Errorf("error reading font file: %w", err)
			}

			var f *truetype.Font
			f, err = truetype.Parse(data)
			if err != nil {
				return fmt.Errorf("error parsing font file: %w", err)
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
			var f *truetype.Font
			f, err = truetype.Parse(asset)
			if err != nil {
				return fmt.Errorf("error parsing font file: %w", err)
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

func rasterizeText(text string, fontFamily string, alignment Alignment, rgba color.RGBA,
	windowWidth, windowHeight int, scaleX, scaleY float32, maintainAspectRatio, useCache bool) *image.RGBA {

	id := fmt.Sprintf("%s%s%v%v%d%d%f%f%v", text, fontFamily, alignment, rgba, windowWidth, windowHeight, scaleX, scaleY, maintainAspectRatio)

	if useCache {
		if img, ok := renderedLabels[id]; ok {
			return img
		}
	}

	textFont := getFontOrDefault(fontFamily)

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

	absFontSize := float64(scale[1] * float32(windowHeight))

	img := image.NewRGBA(image.Rect(0, 0, int(scale[0]*float32(windowWidth)), int(scale[1]*float32(windowHeight))))

	ctx := freetype.NewContext()
	ctx.SetFont(textFont)
	ctx.SetFontSize(absFontSize)
	ctx.SetDPI(72.0)
	ctx.SetHinting(font.HintingVertical)
	ctx.SetClip(img.Bounds())
	ctx.SetSrc(image.NewUniform(rgba))
	ctx.SetDst(img)

	fpFontSize := fixed.Int26_6(scale[1] * float32(windowHeight))
	textWidth := measureTextWidth(textFont, text, fpFontSize)
	textHeight := measureTextHeight(textFont, text, fpFontSize)

	shapeWidth := float32(img.Bounds().Size().X)
	shapeHeight := float32(img.Bounds().Size().Y)

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

	if useCache {
		renderedLabels[id] = img
	}

	return img
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
