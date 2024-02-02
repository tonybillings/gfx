package gfx

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/png" // required to decode PNGs
	"io"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
)

/******************************************************************************
 Texture
******************************************************************************/

type Texture interface {
	init()
	close()
	name() uint32
}

/******************************************************************************
 TextureSource
******************************************************************************/

type TextureSource interface {
	color.RGBA | string | *image.RGBA | *image.NRGBA
}

/******************************************************************************
 TextureBase
******************************************************************************/

type TextureBase[T TextureSource] struct {
	source T
	glName uint32
}

/******************************************************************************
 Texture Functions
******************************************************************************/

func (t *TextureBase[T]) createFromColor(rgba color.RGBA) uint32 {
	data := []uint8{
		rgba.R, rgba.G, rgba.B, rgba.A,
		rgba.R, rgba.G, rgba.B, rgba.A,
		rgba.R, rgba.G, rgba.B, rgba.A,
		rgba.R, rgba.G, rgba.B, rgba.A,
	}

	var name uint32
	gl.GenTextures(1, &name)
	gl.BindTexture(gl.TEXTURE_2D, name)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 2, 2, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
	gl.BindTexture(gl.TEXTURE_2D, 0)

	return name
}

func (t *TextureBase[T]) loadFromStorage(path string) uint32 {
	var imgReader io.Reader
	var imgFile *os.File
	var err error
	if fileExists(path) {
		imgFile, err = os.Open(path)
		if err != nil {
			panic(fmt.Errorf("failed to open image: %w", err))
		}
		imgReader = imgFile
	} else {
		imgReader = GetAssetReader(path)
		if imgReader == nil {
			panic("failed to open image (does not exist on file system or was not provided as an asset")
		}
	}

	var img image.Image
	img, _, err = image.Decode(imgReader)
	if err != nil {
		panic(fmt.Errorf("failed to decode image: %w", err))
	}

	if imgFile != nil {
		e := imgFile.Close()
		if e != nil {
			panic(fmt.Errorf("failed to close image: %w", e))
		}
	}

	var nrgba *image.NRGBA
	if src, ok := img.(*image.NRGBA); ok {
		nrgba = src
	} else {
		nrgba = image.NewNRGBA(img.Bounds())
		draw.Draw(nrgba, nrgba.Bounds(), img, image.Point{}, draw.Src)
	}

	name := t.loadFromMemory(img)
	return name
}

func (t *TextureBase[T]) loadFromMemory(img image.Image) uint32 {
	var tex uint32
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	switch s := any(img).(type) {
	case *image.RGBA:
		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			gl.RGBA,
			int32(s.Rect.Size().X),
			int32(s.Rect.Size().Y),
			0,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			gl.Ptr(s.Pix))
	case *image.NRGBA:
		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			gl.RGBA,
			int32(s.Rect.Size().X),
			int32(s.Rect.Size().Y),
			0,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			gl.Ptr(s.Pix))
	}

	gl.BindTexture(gl.TEXTURE_2D, 0)
	return tex
}

func (t *TextureBase[T]) init() {
	switch s := any(t.source).(type) {
	case color.RGBA:
		t.glName = t.createFromColor(s)
	case string:
		t.glName = t.loadFromStorage(s)
	case *image.RGBA:
		t.glName = t.loadFromMemory(s)
	case *image.NRGBA:
		t.glName = t.loadFromMemory(s)
	default:
		panic("unexpected error: source is invalid")
	}
}

func (t *TextureBase[T]) close() {
	if t.glName == 0 {
		return
	}

	gl.DeleteTextures(1, &t.glName)
}

func (t *TextureBase[T]) name() uint32 {
	return t.glName
}

/******************************************************************************
 New Texture Function
******************************************************************************/

func NewTexture[T TextureSource](source T) Texture {
	return &TextureBase[T]{
		source: source,
	}
}
