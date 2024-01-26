package gfx

/******************************************************************************
 Alignment
******************************************************************************/

type Alignment int

const (
	NoAlignment Alignment = iota
	TopLeft
	MiddleLeft
	BottomLeft
	TopCenter
	Center
	BottomCenter
	TopRight
	MiddleRight
	BottomRight
)

/******************************************************************************
 Orientation
******************************************************************************/

type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

/******************************************************************************
 Margin
******************************************************************************/

type Margin struct {
	Top, Right, Bottom, Left float32
}
