package gfx

import (
	"context"
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"math"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

/******************************************************************************
 Window
******************************************************************************/

type Window struct {
	glwin *glfw.Window

	title         atomic.Value
	width         atomic.Int32
	height        atomic.Int32
	lastWidth     atomic.Int32
	lastHeight    atomic.Int32
	lastPositionX atomic.Int32
	lastPositionY atomic.Int32

	fullscreen              atomic.Bool
	fullscreenButtonEnabled atomic.Bool
	fullscreenRequested     atomic.Bool
	windowRequested         atomic.Bool

	quitButtonEnabled atomic.Bool

	targetFramerate atomic.Uint32
	clearColor      mgl32.Vec4

	objects []WindowObject

	keyEventHandlers      []*KeyEventHandler
	keyEventHandlersMutex sync.Mutex
	inputEnabled          atomic.Bool

	tranTarget      WindowObject
	tranQuad        *Shape
	tranQuadShowing bool
	tranQuadHiding  bool
	tranQuadOpacity float64
	tranSpeed       float64

	mouseTrackingEnabled atomic.Bool
	mouseState           MouseState
	mouseStateMutex      sync.Mutex

	initialized atomic.Bool
	cancelFunc  context.CancelFunc
	stateMutex  sync.Mutex
}

/******************************************************************************
 Window Functions
******************************************************************************/

func (w *Window) clearScreen() {
	gl.ClearColor(w.clearColor[0], w.clearColor[1], w.clearColor[2], w.clearColor[3])
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (w *Window) enableFullscreenMode() {
	primaryMonitor := glfw.GetPrimaryMonitor()
	vidMode := primaryMonitor.GetVideoMode()

	x, y := w.glwin.GetPos()
	w.lastPositionX.Store(int32(x))
	w.lastPositionY.Store(int32(y))
	w.lastWidth.Store(w.width.Load())
	w.lastHeight.Store(w.height.Load())
	w.width.Store(int32(vidMode.Width))
	w.height.Store(int32(vidMode.Height))
	w.glwin.SetMonitor(primaryMonitor, 0, 0, vidMode.Width, vidMode.Height, vidMode.RefreshRate)
	w.fullscreen.Store(true)
}

func (w *Window) enableWindowedMode() {
	primaryMonitor := glfw.GetPrimaryMonitor()
	vidMode := primaryMonitor.GetVideoMode()

	w.glwin.SetMonitor(nil, int(w.lastPositionX.Load()), int(w.lastPositionY.Load()), int(w.lastWidth.Load()), int(w.lastHeight.Load()), vidMode.RefreshRate)
	w.width.Store(w.lastWidth.Load())
	w.height.Store(w.lastHeight.Load())
	w.lastWidth.Store(int32(vidMode.Width))
	w.lastHeight.Store(int32(vidMode.Height))
	w.fullscreen.Store(false)
}

func (w *Window) initAssets() error {
	if err := initFonts(); err != nil {
		return fmt.Errorf("error initializing fonts: %w", err)
	}

	if err := initShaders(); err != nil {
		return fmt.Errorf("error initializing shaders: %w", err)
	}

	return nil
}

func (w *Window) initTransitionQuad() {
	w.tranQuad = NewQuad()
	w.tranQuad.
		SetColor(w.ClearColor()).
		SetOpacity(0).
		MaintainAspectRatio(false)
	w.tranQuad.Init(w)
}

func (w *Window) initObjects() {
	for _, o := range w.objects {
		o.Init(w)
	}
}

func (w *Window) updateObjects(deltaTime int64) {
	for _, o := range w.objects {
		if !o.Closed() {
			o.Update(deltaTime)
		}
	}
}

func (w *Window) drawObjects(deltaTime int64) {
	for _, o := range w.objects {
		if !o.Closed() {
			o.Draw(deltaTime)
		}
	}
}

func (w *Window) closeObjects() {
	for _, o := range w.objects {
		o.Close()
	}
}

func (w *Window) resizeObjects(oldWidth, oldHeight, newWidth, newHeight int32) {
	for _, o := range w.objects {
		if !o.Closed() {
			o.Resize(oldWidth, oldHeight, newWidth, newHeight)
		}
	}
}

func (w *Window) tick(deltaTime int64) {
	if w.tranQuadShowing {
		w.tranQuadOpacity += w.tranSpeed
		w.tranQuad.SetOpacity(uint8(math.Min(255.0, w.tranQuadOpacity*255.0)))
		if w.tranQuad.Opacity() == 255 {
			w.tranQuadShowing = false
			w.tranQuadHiding = true
			w.tranQuadOpacity = 1

			for _, o := range w.objects {
				o.SetEnabled(false).SetVisibility(false)
			}

			w.tranTarget.SetEnabled(true).SetVisibility(true)
		}

		w.updateObjects(deltaTime)
		w.drawObjects(deltaTime)

		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		w.tranQuad.Draw(deltaTime)
	} else if w.tranQuadHiding {
		w.tranQuadOpacity -= w.tranSpeed
		w.tranQuad.SetOpacity(uint8(math.Max(0, w.tranQuadOpacity*255.0)))
		if w.tranQuad.Opacity() == 0 {
			w.tranQuadHiding = false
			w.tranQuadOpacity = 0
			gl.Disable(gl.BLEND)
		}

		w.tranTarget.Update(deltaTime)
		w.tranTarget.Draw(deltaTime)

		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		w.tranQuad.Draw(deltaTime)
	} else {
		w.updateObjects(deltaTime)
		w.drawObjects(deltaTime)
	}
}

func (w *Window) keyEventCallback(_ *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
	if !w.inputEnabled.Load() {
		return
	}

	w.keyEventHandlersMutex.Lock()
	if w.keyEventHandlers == nil {
		w.keyEventHandlersMutex.Unlock()
		return
	}

	for _, h := range w.keyEventHandlers {
		if h.Key == key && h.Action == action {
			h.Callback(w, key, action)
		}
	}

	w.keyEventHandlersMutex.Unlock()
}

func (w *Window) close() {
	w.stateMutex.Lock()
	w.closeObjects()
	w.stateMutex.Unlock()

	w.glwin.SetKeyCallback(nil)
	w.glwin.SetShouldClose(true)

	windowCount.Add(-1)
	if windowCount.Load() == 0 {
		glfw.Terminate()
	}
}

func (w *Window) Init(ctx context.Context, cancelFunc context.CancelFunc) {
	if w.initialized.Load() {
		cancelFunc()
		return
	}

	w.cancelFunc = cancelFunc

	go func() {
		window, err := newGlfwWindow(w.Title(), w.Width(), w.Height())
		if err != nil {
			cancelFunc()
			panic(err)
		}

		if err = w.initAssets(); err != nil {
			cancelFunc()
			panic(err)
		}

		w.initialized.Store(true)

		runtime.LockOSThread()

		w.stateMutex.Lock()
		w.glwin = window
		w.glwin.SetKeyCallback(w.keyEventCallback)
		w.initTransitionQuad()
		w.initObjects()
		w.stateMutex.Unlock()

		now := time.Now().UnixMicro()
		lastTick := now
		deltaTime := now
		drawInterval := int64(1000000 / w.targetFramerate.Load())

		for {
			if w.glwin.ShouldClose() {
				w.close()
				cancelFunc()
				return
			}
			select {
			case <-ctx.Done():
				w.close()
				return
			default:
			}

			now = time.Now().UnixMicro()
			deltaTime = now - lastTick
			if deltaTime < drawInterval {
				time.Sleep(time.Microsecond * time.Duration(drawInterval-deltaTime))
			}
			deltaTime = time.Now().UnixMicro() - lastTick
			lastTick = time.Now().UnixMicro()

			if w.fullscreenRequested.Load() {
				w.fullscreenRequested.Store(false)
				w.enableFullscreenMode()
				w.resizeObjects(w.lastWidth.Load(), w.lastHeight.Load(), w.width.Load(), w.height.Load())
			} else if w.windowRequested.Load() {
				w.windowRequested.Store(false)
				w.enableWindowedMode()
				w.resizeObjects(w.lastWidth.Load(), w.lastHeight.Load(), w.width.Load(), w.height.Load())
			}

			w.stateMutex.Lock()
			w.clearScreen()
			w.tick(deltaTime)
			w.glwin.SwapBuffers()
			w.stateMutex.Unlock()

			glfw.PollEvents()
		}
	}()
}

func (w *Window) Close() {
	if w.initialized.Load() {
		w.cancelFunc()
	}
}

func (w *Window) AddKeyEventHandler(key glfw.Key, action glfw.Action, callback func(window *Window, key glfw.Key, action glfw.Action)) *KeyEventHandler {
	w.keyEventHandlersMutex.Lock()
	if w.keyEventHandlers == nil {
		w.keyEventHandlers = make([]*KeyEventHandler, 0)
	}

	handler := &KeyEventHandler{
		Key:      key,
		Action:   action,
		Callback: callback,
	}

	w.keyEventHandlers = append(w.keyEventHandlers, handler)
	w.keyEventHandlersMutex.Unlock()
	return handler
}

func (w *Window) RemoveKeyEventHandler(handler *KeyEventHandler) {
	if w.keyEventHandlers == nil {
		return
	}

	w.keyEventHandlersMutex.Lock()
	index := -1
	for i, h := range w.keyEventHandlers {
		if h == handler {
			index = i
			break
		}
	}
	if index != -1 {
		w.keyEventHandlers = append(w.keyEventHandlers[:index], w.keyEventHandlers[index+1:]...)
	}
	w.keyEventHandlersMutex.Unlock()
}

func (w *Window) SetInputEnabled(enabled bool) *Window {
	w.inputEnabled.Store(enabled)
	return w
}

func (w *Window) SetFullscreenEnabled(enabled bool) *Window {
	if enabled {
		w.fullscreenRequested.Store(true)
	} else {
		w.windowRequested.Store(true)
	}
	return w
}

func (w *Window) EnableFullscreenKey() *Window {
	if !w.fullscreenButtonEnabled.Load() {
		w.fullscreenButtonEnabled.Store(true)

		w.AddKeyEventHandler(glfw.KeyF11, glfw.Press, func(_ *Window, _ glfw.Key, _ glfw.Action) {
			fullscreen := w.glwin.GetMonitor() != nil
			if fullscreen {
				w.enableWindowedMode()
			} else {
				w.enableFullscreenMode()
			}
			w.resizeObjects(w.lastWidth.Load(), w.lastHeight.Load(), w.width.Load(), w.height.Load())
		})
	}
	return w
}

func (w *Window) EnableQuitKey(cancelFuncs ...context.CancelFunc) *Window {
	if !w.quitButtonEnabled.Load() {
		w.quitButtonEnabled.Store(true)

		w.AddKeyEventHandler(glfw.KeyEscape, glfw.Press, func(_ *Window, _ glfw.Key, _ glfw.Action) {
			for _, cancelFunc := range cancelFuncs {
				cancelFunc()
			}
			time.Sleep(time.Second)
			os.Exit(0)
		})
	}
	return w
}

func (w *Window) ClearColor() color.RGBA {
	return FloatArrayToRgba(w.clearColor)
}

func (w *Window) SetClearColor(rgba color.RGBA) *Window {
	w.stateMutex.Lock()
	w.clearColor[0] = float32(rgba.R) / 255.0
	w.clearColor[1] = float32(rgba.G) / 255.0
	w.clearColor[2] = float32(rgba.B) / 255.0
	w.clearColor[3] = float32(rgba.A) / 255.0
	w.stateMutex.Unlock()
	return w
}

func (w *Window) Object(name string) WindowObject {
	for _, obj := range w.objects {
		if obj.Name() == name {
			return obj
		}
		o := obj.Child(name)
		if o != nil {
			return o
		}
	}
	return nil
}

func (w *Window) AddObjects(objects ...WindowObject) {
	w.objects = append(w.objects, objects...)
}

func (w *Window) RemoveObjectAt(index int) {
	if index < 0 || index > len(w.objects)-1 {
		return
	}
	w.objects = append(w.objects[:index], w.objects[index+1:]...)
}

func (w *Window) TransitionTo(name string, speed ...float64) {
	w.stateMutex.Lock()
	if len(speed) > 0 {
		w.tranSpeed = speed[0]
	} else {
		w.tranSpeed = 0.1
	}

	if w.tranQuadShowing || w.tranQuadHiding {
		w.stateMutex.Unlock()
		return
	}

	w.tranTarget = w.Object(name)
	w.tranQuadShowing = true
	w.stateMutex.Unlock()
}

func (w *Window) Title() string {
	if t, ok := w.title.Load().(string); ok {
		return t
	}
	return NoLabel
}

func (w *Window) SetTitle(title string) *Window {
	w.title.Store(title)

	if w.glwin != nil {
		w.stateMutex.Lock()
		w.glwin.SetTitle(w.title.Load().(string))
		w.stateMutex.Unlock()
	}

	return w
}

func (w *Window) Width() int32 {
	return w.width.Load()
}

func (w *Window) SetWidth(width int) *Window {
	w.width.Store(int32(width))
	if w.glwin != nil {
		w.stateMutex.Lock()
		w.glwin.SetSize(int(w.width.Load()), int(w.height.Load()))
		w.stateMutex.Unlock()
	}
	return w
}

func (w *Window) Height() int32 {
	return w.height.Load()
}

func (w *Window) SetHeight(height int) *Window {
	w.height.Store(int32(height))
	if w.glwin != nil {
		w.stateMutex.Lock()
		w.glwin.SetSize(int(w.width.Load()), int(w.height.Load()))
		w.stateMutex.Unlock()
	}
	return w
}

func (w *Window) AspectRatio() float32 {
	return float32(w.Width()) / float32(w.Height())
}

func (w *Window) AspectRatioInv() float32 {
	return float32(w.Height()) / float32(w.Width())
}

func (w *Window) AspectRatio2D() mgl32.Vec2 {
	width := w.Width()
	height := w.Height()

	switch {
	case width > height:
		return mgl32.Vec2{1, float32(width) / float32(height)}
	case height > width:
		return mgl32.Vec2{float32(height) / float32(width), 1}
	default:
		return mgl32.Vec2{1, 1}
	}
}

func (w *Window) AspectRatio2DInv() mgl32.Vec2 {
	width := w.Width()
	height := w.Height()

	switch {
	case width > height:
		return mgl32.Vec2{float32(height) / float32(width), 1}
	case height > width:
		return mgl32.Vec2{1, float32(width) / float32(height)}
	default:
		return mgl32.Vec2{1, 1}
	}
}

func (w *Window) ScaleX(x float32) float32 {
	width := w.Width()
	height := w.Height()

	if width > height {
		x *= float32(height) / float32(width)
	}

	return x
}

func (w *Window) ScaleY(y float32) float32 {
	width := w.Width()
	height := w.Height()

	if height > width {
		y *= float32(width) / float32(height)
	}

	return y
}

func (w *Window) ScaleVec(vec mgl32.Vec3) mgl32.Vec3 {
	width := w.Width()
	height := w.Height()

	aspectRatio := float32(height) / float32(width)
	if height > width {
		aspectRatio = float32(width) / float32(height)
	}

	if width > height {
		vec[0] *= aspectRatio
	} else if height > width {
		vec[1] *= aspectRatio
	}

	return vec
}

func (w *Window) IsFullscreen() bool {
	return w.fullscreen.Load()
}

func (w *Window) TargetFramerate() uint32 {
	return w.targetFramerate.Load()
}

func (w *Window) SetTargetFramerate(framerate int) *Window {
	w.targetFramerate.Store(uint32(framerate))
	return w
}

func (w *Window) SwapMouseButtons(swapped bool) {
	w.mouseStateMutex.Lock()
	w.mouseState.ButtonsSwapped = swapped
	w.mouseStateMutex.Unlock()
}

func (w *Window) Mouse() *MouseState {
	w.mouseStateMutex.Lock()
	ms := w.mouseState
	w.mouseStateMutex.Unlock()
	return &ms
}

func (w *Window) EnableMouseTracking() *Window {
	if w.mouseTrackingEnabled.Load() {
		return w
	}
	w.mouseTrackingEnabled.Store(true)

	w.glwin.SetCursorPosCallback(func(window *glfw.Window, x float64, y float64) {
		width, height := window.GetSize()
		w.mouseStateMutex.Lock()
		w.mouseState.X = float32(((x / float64(width)) * 2.0) - 1.0)
		w.mouseState.Y = float32(((y / float64(height)) * -2.0) + 1.0)
		w.mouseStateMutex.Unlock()
	})

	w.glwin.SetMouseButtonCallback(func(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		w.mouseStateMutex.Lock()
		w.mouseState.Update(button, action)
		w.mouseStateMutex.Unlock()
	})

	return w
}

/******************************************************************************
 New Window Function
******************************************************************************/

func NewWindow() *Window {
	w := &Window{
		objects:          make([]WindowObject, 0),
		keyEventHandlers: make([]*KeyEventHandler, 0),
	}

	w.SetWidth(defaultWinWidth)
	w.SetHeight(defaultWinHeight)
	w.SetTargetFramerate(defaultTargetFramerate)
	w.inputEnabled.Store(true)
	return w
}
