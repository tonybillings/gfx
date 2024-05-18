package gfx

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"image"
	"image/color"
	"image/draw"
	_ "image/png" // required to decode PNGs
)

/******************************************************************************
 Texture
******************************************************************************/

// Texture assets represent texture maps that can be sampled by shaders by
// binding them to sampler2D variables.
type Texture interface {
	GlAsset

	// Width shall return the width of the texture, in pixels.
	Width() int

	// Height shall return the height of the texture, in pixels.
	Height() int
}

/******************************************************************************
 Texture2D
******************************************************************************/

type Texture2D struct {
	AssetBase

	width  int
	height int

	glName        uint32
	uWrapMode     int32
	vWrapMode     int32
	minFilterMode int32
	magFilterMode int32
	useMipMaps    bool
}

/******************************************************************************
 Asset Implementation
******************************************************************************/

func (t *Texture2D) Init() bool {
	if t.Initialized() {
		return true
	}

	switch source := t.source.(type) {
	case []byte:
		t.createFromSlice(source)
	case string:
		t.createFromFile(source)
	case color.RGBA:
		t.createFromColor(source)
	case *image.RGBA:
		t.createFromImage(source)
	case *image.NRGBA:
		t.createFromImage(source)
	default:
		panic("unexpected error: source type is not supported")
	}

	return t.AssetBase.Init()
}

func (t *Texture2D) Close() {
	if !t.Initialized() {
		return
	}

	gl.DeleteTextures(1, &t.glName)
	t.glName = 0

	t.AssetBase.Close()
}

/******************************************************************************
 Texture Implementation
******************************************************************************/

func (t *Texture2D) GlName() uint32 {
	return t.glName
}

func (t *Texture2D) Width() int {
	return t.width
}

func (t *Texture2D) Height() int {
	return t.height
}

/******************************************************************************
 Texture2D Functions
******************************************************************************/

func (t *Texture2D) createFromSlice(slice []byte) {
	reader := bufio.NewReader(bytes.NewReader(slice))
	t.createFromReader(reader)
}

func (t *Texture2D) createFromFile(name string) {
	reader, closeFunc := t.getSourceReader(name)
	defer closeFunc()
	t.createFromReader(reader)
}

func (t *Texture2D) createFromReader(reader *bufio.Reader) {
	if reader == nil {
		panic(fmt.Errorf("reader cannot be nil"))
	}

	var img image.Image
	var err error
	if img, _, err = image.Decode(reader); err != nil {
		panic(fmt.Errorf("decode image error: %w", err))
	}

	var nrgba *image.NRGBA
	if src, ok := img.(*image.NRGBA); ok {
		nrgba = src
	} else {
		nrgba = image.NewNRGBA(img.Bounds())
		draw.Draw(nrgba, nrgba.Bounds(), img, image.Point{}, draw.Src)
	}

	flipped := t.flipImage(nrgba)
	t.createFromImage(flipped)
}

func (t *Texture2D) createFromColor(rgba color.RGBA) {
	data := make([]uint8, t.width*t.height*4)
	for i := 0; i < t.width*t.height*4; i += 4 {
		data[i] = rgba.R
		data[i+1] = rgba.G
		data[i+2] = rgba.B
		data[i+3] = rgba.A
	}

	var name uint32
	gl.GenTextures(1, &name)
	gl.BindTexture(gl.TEXTURE_2D, name)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(t.width), int32(t.height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
	gl.BindTexture(gl.TEXTURE_2D, 0)

	t.glName = name
}

func (t *Texture2D) createFromImage(img image.Image) {
	var name uint32
	gl.GenTextures(1, &name)
	gl.BindTexture(gl.TEXTURE_2D, name)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, t.uWrapMode)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, t.vWrapMode)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, t.minFilterMode)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, t.magFilterMode)

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

	if t.useMipMaps {
		gl.GenerateMipmap(gl.TEXTURE_2D)
	}

	gl.BindTexture(gl.TEXTURE_2D, 0)
	t.glName = name
	t.width = img.Bounds().Size().X
	t.height = img.Bounds().Size().Y
}

func (t *Texture2D) flipImage(nrgba *image.NRGBA) *image.NRGBA {
	flipped := image.NewNRGBA(nrgba.Bounds())
	for y := 0; y < nrgba.Bounds().Dy(); y++ {
		for x := 0; x < nrgba.Bounds().Dx(); x++ {
			srcX := x
			srcY := y
			dstX := x
			dstY := nrgba.Bounds().Dy() - y - 1
			flipped.Set(dstX, dstY, nrgba.At(srcX, srcY))
		}
	}
	return flipped
}

func (t *Texture2D) SetSize(width, height int) {
	t.width = width
	t.height = height
}

/******************************************************************************
 New Texture2D Function
******************************************************************************/

func NewTexture2D[T TextureSource](name string, source T, config ...*TextureConfig) *Texture2D {
	if len(config) == 0 {
		config = append(config, NewTextureConfig(HighestQuality))
	}

	cfg := config[0]
	useMipMaps, minFilterMode, magFilterMode := cfg.GetFilterConfig()

	return &Texture2D{
		AssetBase: AssetBase{
			name:   name,
			source: source,
		},
		uWrapMode:     int32(cfg.UWrapMode),
		vWrapMode:     int32(cfg.VWrapMode),
		minFilterMode: minFilterMode,
		magFilterMode: magFilterMode,
		useMipMaps:    useMipMaps,
		width:         2,
		height:        2,
	}
}

/******************************************************************************
 TextureWrapMode
******************************************************************************/

type TextureWrapMode int32

const (
	Repeat TextureWrapMode = gl.REPEAT
	Clamp  TextureWrapMode = gl.CLAMP_TO_EDGE
	Mirror TextureWrapMode = gl.MIRRORED_REPEAT
)

/******************************************************************************
 TextureConfig
******************************************************************************/

type TextureConfig struct {
	FilterQuality QualityLevel
	UWrapMode     TextureWrapMode
	VWrapMode     TextureWrapMode
}

func (c *TextureConfig) GetFilterConfig() (useMipMaps bool, minFilterMode, magFilterMode int32) {
	switch c.FilterQuality {
	case LowestQuality:
		minFilterMode = gl.NEAREST
		magFilterMode = gl.NEAREST
		useMipMaps = false
	case VeryLowQuality:
		minFilterMode = gl.LINEAR
		magFilterMode = gl.NEAREST
		useMipMaps = false
	case LowQuality:
		minFilterMode = gl.LINEAR
		magFilterMode = gl.LINEAR
		useMipMaps = false
	case MediumQuality:
		minFilterMode = gl.NEAREST_MIPMAP_NEAREST
		magFilterMode = gl.LINEAR
		useMipMaps = true
	case HighQuality:
		minFilterMode = gl.LINEAR_MIPMAP_NEAREST
		magFilterMode = gl.LINEAR
		useMipMaps = true
	case VeryHighQuality:
		minFilterMode = gl.NEAREST_MIPMAP_LINEAR
		magFilterMode = gl.LINEAR
		useMipMaps = true
	case HighestQuality:
		minFilterMode = gl.LINEAR_MIPMAP_LINEAR
		magFilterMode = gl.LINEAR
		useMipMaps = true
	default:
		minFilterMode = gl.NEAREST
		magFilterMode = gl.NEAREST
		useMipMaps = false
	}
	return
}

func NewTextureConfig(filterQuality QualityLevel, wrapMode ...TextureWrapMode) *TextureConfig {
	switch len(wrapMode) {
	case 1:
		return &TextureConfig{
			FilterQuality: filterQuality,
			UWrapMode:     wrapMode[0],
			VWrapMode:     wrapMode[0],
		}
	case 2:
		return &TextureConfig{
			FilterQuality: filterQuality,
			UWrapMode:     wrapMode[0],
			VWrapMode:     wrapMode[1],
		}
	default:
		return &TextureConfig{
			FilterQuality: filterQuality,
			UWrapMode:     Repeat,
			VWrapMode:     Repeat,
		}
	}
}
