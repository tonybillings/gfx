package gfx

/******************************************************************************
 Anchor
******************************************************************************/

// Anchor specifies how the object should be positioned with respect to its
// parent, or the window if it has none.  Note that when an anchor is used,
// the Position property of objects is ignored and Margin should be used
// instead for offset adjustments.
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
 TextAlignment
******************************************************************************/

// TextAlignment specifies how the text within Label objects should be positioned
// with respect to the "frame" of the Label, which is based on its scale.
type TextAlignment int

const (
	Centered TextAlignment = iota
	Left
	Right
)

/******************************************************************************
 Orientation
******************************************************************************/

// Orientation specifies how certain controls should be oriented/arranged,
// such as whether a Slider object's button will slide along a horizontally
// or vertically aligned rail, etc.
type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

/******************************************************************************
 Margin
******************************************************************************/

// Margin contains the position offsets used when objects are anchored, in the
// normalized device/screen space range of [-1,1] for both axes.
type Margin struct {
	Top, Right, Bottom, Left float32
}

/******************************************************************************
 Bounds
******************************************************************************/

// Bounds represents the boundaries or "frame" in which the associated object
// is rendered, if not the full width/height of the window.  Objects that
// become children of other objects will have their bounds adjusted, such
// that the scale and positioning of the parent determines the maximum bounds
// for its children, affecting their scaling/positioning/anchoring, etc.
// Values are in the normalized device/screen space range of [-1,1] for both
// axes.
type Bounds struct {
	Top, Right, Bottom, Left float32
}

/******************************************************************************
 Resizer
******************************************************************************/

// Resizer should be implemented by WindowObject types that need to readjust
// themselves, recompute variables, etc, whenever the window's size has
// changed or when going from windowed to fullscreen mode or vice-versa.
type Resizer interface {
	Resize(newWidth, newHeight int)
}
