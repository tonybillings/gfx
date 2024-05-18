package gfx

import (
	"bytes"
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"image"
	"image/png"
	"os"
	"path"
	"sync"
	"time"
)

const (
	defaultCanvasName = "Canvas"
)

/******************************************************************************
 Canvas
******************************************************************************/

// Canvas can be used with a Brush to allow end-users the ability to paint
// to a Texture2D-based surface, which can then be exported to a PNG.  A
// canvas is just a type of View, which means you can change the background
// (FillColor) and, optionally, configure a frame (Border) with a specific
// thickness/color, but that frame will not be part of the exported PNG.
type Canvas struct {
	View

	bounds BoundingObject

	surface    *Shape2D
	surfaceTex *Texture2D

	mouse *MouseState

	clearRequested  bool
	exportRequested bool
	exportDirectory string

	stateMutex sync.Mutex
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (c *Canvas) Init() (ok bool) {
	if c.Initialized() {
		return true
	}

	if !c.initSurfaceTexture() {
		return false
	}

	if !c.initSurface() {
		return false
	}

	c.RefreshLayout()

	return c.View.Init()
}

func (c *Canvas) Update(deltaTime int64) (ok bool) {
	if !c.View.Update(deltaTime) {
		return false
	}

	c.surface.Update(deltaTime)
	c.bounds.Update(deltaTime)

	return true
}

func (c *Canvas) Close() {
	if !c.Initialized() {
		return
	}

	c.surface.Close()
	c.surfaceTex.Close()
	c.bounds.Close()
	c.View.Close()
}

/******************************************************************************
 DrawableObject Implementation
******************************************************************************/

func (c *Canvas) Draw(deltaTime int64) (ok bool) {
	if !c.visible.Load() || !c.initialized.Load() {
		return false
	}

	c.fill.Draw(deltaTime)

	c.stateMutex.Lock()

	if c.clearRequested {
		c.clearRequested = false
		c.clear()
	}

	c.surface.Draw(deltaTime)

	if c.exportRequested {
		c.exportRequested = false
		c.export()
	}

	c.stateMutex.Unlock()

	c.border.Draw(deltaTime)

	return c.drawChildren(deltaTime)
}

/******************************************************************************
 Resizer Implementation
******************************************************************************/

func (c *Canvas) Resize(newWidth, newHeight int) {
	c.View.Resize(newWidth, newHeight)
	c.surface.Resize(newWidth, newHeight)
	c.bounds.Resize(newWidth, newHeight)
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (c *Canvas) SetWindow(window *Window) WindowObject {
	c.View.SetWindow(window)
	c.surface.SetWindow(window)
	c.bounds.SetWindow(window)
	return c
}

func (c *Canvas) SetMaintainAspectRatio(maintainAspectRatio bool) WindowObject {
	c.View.SetMaintainAspectRatio(maintainAspectRatio)
	c.bounds.SetMaintainAspectRatio(maintainAspectRatio)
	return c
}

/******************************************************************************
 Canvas Functions
******************************************************************************/

func (c *Canvas) defaultLayout() {
	c.surface.SetParent(c)
	c.bounds.SetParent(c)
	c.View.defaultLayout()
}

func (c *Canvas) initSurface() (ok bool) {
	c.surface.SetWindow(c.window)
	c.surface.SetMaintainAspectRatio(c.maintainAspectRatio)
	c.surface.SetTexture(c.surfaceTex)
	if !c.surface.Init() {
		return false
	}

	c.bounds.SetWindow(c.window)
	c.bounds.SetMaintainAspectRatio(c.maintainAspectRatio)
	if !c.bounds.Init() {
		return false
	}

	return true
}

func (c *Canvas) initSurfaceTexture() (ok bool) {
	win := c.Window()
	c.surfaceTex = NewTexture2D("surface_texture", Transparent, NewTextureConfig(LowestQuality))
	c.surfaceTex.SetSize(int(win.ScaleX(c.WorldScale().X())*float32(win.Width())), int(win.ScaleY(c.WorldScale().Y())*float32(win.Height())))
	return c.surfaceTex.Init()
}

func (c *Canvas) clear() {
	buffer := make([]uint8, c.surfaceTex.Width()*c.surfaceTex.Height()*4)
	gl.BindTexture(gl.TEXTURE_2D, c.surfaceTex.GlName())
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(c.surfaceTex.Width()), int32(c.surfaceTex.Height()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(buffer))
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func (c *Canvas) export() {
	width := c.surfaceTex.Width()
	height := c.surfaceTex.Height()
	bufferLen := width * height * 4
	buffer := make([]uint8, bufferLen)
	gl.BindTexture(gl.TEXTURE_2D, c.surfaceTex.GlName())
	gl.GetTexImage(gl.TEXTURE_2D, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&buffer[0]))
	gl.BindTexture(gl.TEXTURE_2D, 0)

	go func() {
		bgColor := c.FillColor()
		for i := 0; i < bufferLen; i += 4 {
			if buffer[i+3] == 0 {
				buffer[i] = bgColor.R
				buffer[i+1] = bgColor.G
				buffer[i+2] = bgColor.B
				buffer[i+3] = bgColor.A
			}
		}

		img := image.NewRGBA(image.Rect(0, 0, width, height))

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				i := y*width*4 + x*4
				j := (height-y-1)*width*4 + x*4
				img.Pix[i+0] = buffer[j+0]
				img.Pix[i+1] = buffer[j+1]
				img.Pix[i+2] = buffer[j+2]
				img.Pix[i+3] = buffer[j+3]
			}
		}

		var resultBuffer bytes.Buffer
		if err := png.Encode(&resultBuffer, img); err != nil {
			panic(err)
		}

		filename := path.Join(c.exportDirectory, fmt.Sprintf("canvas_export_%d.png", time.Now().UnixMilli()))
		if err := os.WriteFile(filename, resultBuffer.Bytes(), 0660); err != nil {
			fmt.Println("error: failed to export canvas") // TODO: refactor as part of error handling revamping
		}
	}()
}

func (c *Canvas) Surface() Texture {
	return c.surface.Texture()
}

func (c *Canvas) UpdateSurface(surface Texture) {
	c.surface.SetTexture(surface)
}

func (c *Canvas) Mouse() *MouseState {
	return c.bounds.LocalMouse()
}

func (c *Canvas) Clear() {
	c.stateMutex.Lock()
	c.clearRequested = true
	c.stateMutex.Unlock()
}

func (c *Canvas) Export(toDirectory ...string) {
	dir := ""
	if len(toDirectory) > 0 {
		dir = toDirectory[0]
	}

	c.stateMutex.Lock()
	c.exportRequested = true
	c.exportDirectory = dir
	c.stateMutex.Unlock()
}

/******************************************************************************
 New Canvas Function
******************************************************************************/

func NewCanvas() *Canvas {
	c := &Canvas{
		View:    *NewView(),
		surface: NewQuad(),
		bounds:  NewBoundingBox(),
	}

	c.SetName(defaultCanvasName)
	c.defaultLayout()

	return c
}
