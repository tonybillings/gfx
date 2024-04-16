package gfx

/******************************************************************************
 Anchor
******************************************************************************/

type Anchor int

const (
	NoAnchor Anchor = iota
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
 Alignment
******************************************************************************/

type Alignment int

const (
	Centered Alignment = iota
	Left
	Right
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

/******************************************************************************
 Resizer
******************************************************************************/

type Resizer interface {
	Resize(oldWidth, oldHeight, newWidth, newHeight int32)
}
