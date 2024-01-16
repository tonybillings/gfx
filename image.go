package gfx

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/png" // required to decode PNGs
	"io"
	"os"
	"sync"

	"github.com/go-gl/gl/v4.1-core/gl"
)

var (
	loadedImages      = make(map[string]uint32)
	loadedImagesMutex = sync.Mutex{}
)

func loadImage(pathToPng string) uint32 {
	loadedImagesMutex.Lock()
	defer loadedImagesMutex.Unlock()

	if i, ok := loadedImages[pathToPng]; ok {
		return i
	}

	var imgReader io.Reader
	var imgFile *os.File
	var err error
	if fileExists(pathToPng) {
		imgFile, err = os.Open(pathToPng)
		if err != nil {
			panic(fmt.Errorf("failed to open image: %w", err))
		}
		imgReader = imgFile
	} else {
		imgReader = GetAssetReader(pathToPng)
		if imgReader == nil {
			panic("failed to open image (does not exist on file system or was not provided as an asset")
		}
	}

	img, _, err := image.Decode(imgReader)
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

	var tex uint32
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(nrgba.Rect.Size().X),
		int32(nrgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(nrgba.Pix))

	gl.BindTexture(gl.TEXTURE_2D, 0)

	loadedImages[pathToPng] = tex
	return tex
}

func createImage(rgba color.RGBA) uint32 {
	data := []uint8{
		rgba.R, rgba.G, rgba.B, rgba.A,
		rgba.R, rgba.G, rgba.B, rgba.A,
		rgba.R, rgba.G, rgba.B, rgba.A,
		rgba.R, rgba.G, rgba.B, rgba.A,
	}

	var tex uint32
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 2, 2, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
	gl.BindTexture(gl.TEXTURE_2D, 0)

	return tex
}
