package gfx

/******************************************************************************
 QualityLevel
******************************************************************************/

// QualityLevel is used to determine the visual quality vs performance balance
// when rendering certain objects, like textures.
type QualityLevel int

const (
	LowestQuality QualityLevel = iota
	VeryLowQuality
	LowQuality
	MediumQuality
	HighQuality
	VeryHighQuality
	HighestQuality
)
