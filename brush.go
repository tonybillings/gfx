package gfx

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"sync"
)

const (
	defaultBasicBrushName = "BasicBrush"
)

/******************************************************************************
 Brush
******************************************************************************/

// Brush objects are used to paint a Canvas by the end-user by clicking and
// dragging across the canvas.  Brushes should have no effect if not enabled.
type Brush interface {
	WindowObject

	// Canvas shall return the canvas to which this brush is assigned.
	Canvas() *Canvas
	SetCanvas(*Canvas) Brush
}

/******************************************************************************
 BasicBrush
******************************************************************************/

type BasicBrush struct {
	WindowObjectBase

	size      float32
	brushHead BrushHeadType
	canvas    *Canvas
	drawing   bool

	canvasBuffer  []uint8
	canvasBackup  []uint8
	undoRequested bool

	stateMutex sync.Mutex
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (b *BasicBrush) Update(deltaTime int64) (ok bool) {
	if !b.ObjectBase.Update(deltaTime) {
		return false
	}

	b.stateMutex.Lock()
	if b.canvas != nil {
		if b.undoRequested && b.canvasBackup != nil {
			b.undoRequested = false
			b.undo()
			b.stateMutex.Unlock()
			return true
		}

		if mouse := b.canvas.Mouse(); mouse != nil && mouse.PrimaryDown {
			if !b.drawing {
				b.drawing = true
				b.backupCanvas()
			}

			switch b.brushHead {
			case RoundBrushHead:
				b.updateCanvasRoundHead(mouse)
			case SquareBrushHead:
				b.updateCanvasSquareHead(mouse)
			}
		} else if mouse != nil && !mouse.PrimaryDown {
			if b.drawing {
				b.drawing = false
			}
		}
	}
	b.stateMutex.Unlock()

	return true
}

/******************************************************************************
 Brush Implementation
******************************************************************************/

func (b *BasicBrush) Canvas() (canvas *Canvas) {
	b.stateMutex.Lock()
	canvas = b.canvas
	b.stateMutex.Unlock()
	return
}

func (b *BasicBrush) SetCanvas(canvas *Canvas) Brush {
	b.stateMutex.Lock()
	b.canvas = canvas
	b.stateMutex.Unlock()
	return b
}

/******************************************************************************
 BasicBrush Functions
******************************************************************************/

func (b *BasicBrush) backupCanvas() {
	surface := b.canvas.Surface()

	if b.canvasBackup == nil {
		width := surface.Width()
		height := surface.Height()
		b.canvasBackup = make([]uint8, width*height*4)
	}

	gl.BindTexture(gl.TEXTURE_2D, surface.GlName())
	gl.GetTexImage(gl.TEXTURE_2D, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&b.canvasBackup[0]))
}

func (b *BasicBrush) updateCanvasRoundHead(mouse *MouseState) {
	surface := b.canvas.Surface()
	width := surface.Width()
	height := surface.Height()

	tx := int((mouse.X + 1) / 2 * float32(width))
	ty := int((mouse.Y + 1) / 2 * float32(height))

	if b.canvasBuffer == nil {
		b.canvasBuffer = make([]uint8, width*height*4)
	}
	gl.BindTexture(gl.TEXTURE_2D, surface.GlName())
	gl.GetTexImage(gl.TEXTURE_2D, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&b.canvasBuffer[0]))

	textureRadius := int(b.size * (float32(width) * 0.5))
	textureColor := FloatArrayToRgba(b.color)

	for i := -textureRadius; i <= textureRadius; i++ {
		for j := -textureRadius; j <= textureRadius; j++ {
			if i*i+j*j <= textureRadius*textureRadius {
				px := tx + i
				py := ty + j
				if px >= 0 && px < width && py >= 0 && py < height {
					index := (py*width + px) * 4
					b.canvasBuffer[index] = textureColor.R
					b.canvasBuffer[index+1] = textureColor.G
					b.canvasBuffer[index+2] = textureColor.B
					b.canvasBuffer[index+3] = textureColor.A
				}
			}
		}
	}

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(b.canvasBuffer))
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func (b *BasicBrush) updateCanvasSquareHead(mouse *MouseState) {
	surface := b.canvas.Surface()
	width := surface.Width()
	height := surface.Height()

	tx := int((mouse.X + 1) / 2 * float32(width))
	ty := int((mouse.Y + 1) / 2 * float32(height))

	if b.canvasBuffer == nil {
		b.canvasBuffer = make([]uint8, width*height*4)
	}
	gl.BindTexture(gl.TEXTURE_2D, surface.GlName())
	gl.GetTexImage(gl.TEXTURE_2D, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&b.canvasBuffer[0]))

	textureHalfSize := int(b.size * (float32(width) * 0.5))
	textureColor := FloatArrayToRgba(b.color)

	for i := -textureHalfSize; i <= textureHalfSize; i++ {
		for j := -textureHalfSize; j <= textureHalfSize; j++ {
			px := tx + i
			py := ty + j
			if px >= 0 && px < width && py >= 0 && py < height {
				index := (py*width + px) * 4
				b.canvasBuffer[index] = textureColor.R
				b.canvasBuffer[index+1] = textureColor.G
				b.canvasBuffer[index+2] = textureColor.B
				b.canvasBuffer[index+3] = textureColor.A
			}
		}
	}

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(b.canvasBuffer))
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func (b *BasicBrush) undo() {
	surface := b.canvas.Surface()
	width := surface.Width()
	height := surface.Height()
	gl.BindTexture(gl.TEXTURE_2D, surface.GlName())
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(b.canvasBackup))
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func (b *BasicBrush) Size() (size float32) {
	b.stateMutex.Lock()
	size = b.size
	b.stateMutex.Unlock()
	return
}

func (b *BasicBrush) SetSize(size float32) *BasicBrush {
	b.stateMutex.Lock()
	b.size = size
	b.stateMutex.Unlock()
	return b
}

func (b *BasicBrush) BrushHead() (brushHead BrushHeadType) {
	b.stateMutex.Lock()
	brushHead = b.brushHead
	b.stateMutex.Unlock()
	return
}

func (b *BasicBrush) SetBrushHead(brushHead BrushHeadType) *BasicBrush {
	b.stateMutex.Lock()
	b.brushHead = brushHead
	b.stateMutex.Unlock()
	return b
}

func (b *BasicBrush) Undo() {
	b.stateMutex.Lock()
	b.undoRequested = true
	b.stateMutex.Unlock()
}

/******************************************************************************
 New BasicBrush Function
******************************************************************************/

func NewBasicBrush() *BasicBrush {
	b := &BasicBrush{
		WindowObjectBase: *NewWindowObject(),
		size:             1,
		brushHead:        RoundBrushHead,
	}

	b.SetName(defaultBasicBrushName)

	return b
}

/******************************************************************************
 BrushHeadType
******************************************************************************/

type BrushHeadType int

const (
	RoundBrushHead BrushHeadType = iota
	SquareBrushHead
)
