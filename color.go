package gfx

import "image/color"

var (
	White     = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	Black     = color.RGBA{A: 255}
	LightGray = color.RGBA{R: 192, G: 192, B: 192, A: 255}
	Gray      = color.RGBA{R: 128, G: 128, B: 128, A: 255}
	DarkGray  = color.RGBA{R: 64, G: 64, B: 64, A: 255}
	Red       = color.RGBA{R: 255, A: 255}
	Green     = color.RGBA{G: 255, A: 255}
	Blue      = color.RGBA{B: 255, A: 255}
	Magenta   = color.RGBA{R: 255, B: 255, A: 255}
	Yellow    = color.RGBA{R: 255, G: 255, A: 255}
	Purple    = color.RGBA{R: 100, G: 50, B: 130, A: 255}
	Orange    = color.RGBA{R: 230, G: 126, B: 34, A: 255}
	Teal      = color.RGBA{R: 118, G: 215, B: 196, A: 255}
	Maroon    = color.RGBA{R: 146, G: 43, B: 33, A: 255}
	SkyBlue   = color.RGBA{R: 174, G: 170, B: 241, A: 255}
)

var (
	DefaultColors = []color.RGBA{
		Red,
		Green,
		Blue,
		Magenta,
		Yellow,
		Purple,
		Orange,
		Teal,
		Maroon,
		SkyBlue,
	}
)

func RgbaToFloatArray(rgba color.RGBA) [4]float32 {
	return [4]float32{
		float32(rgba.R) / 255.0,
		float32(rgba.G) / 255.0,
		float32(rgba.B) / 255.0,
		float32(rgba.A) / 255.0}
}

func FloatArrayToRgba(array [4]float32) color.RGBA {
	return color.RGBA{
		R: uint8(array[0] * 255.0),
		G: uint8(array[1] * 255.0),
		B: uint8(array[2] * 255.0),
		A: uint8(array[3] * 255.0)}
}
