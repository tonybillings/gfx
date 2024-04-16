package gfx

/******************************************************************************
 QualityLevel
******************************************************************************/

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
