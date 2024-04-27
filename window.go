package gfx

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"image"
	"image/color"
	"image/png"
	"reflect"
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

	objects         []Object
	drawableObjects []DrawableObject
	windowObjects   []WindowObject

	services []Service

	objectInitQueue  []*asyncBoolInvocation
	objectCloseQueue []*asyncVoidInvocation

	serviceInitQueue  []*asyncBoolInvocation
	serviceCloseQueue []*asyncVoidInvocation

	pngExportRequest *asyncByteSliceInvocation

	keyEventHandlers      []*KeyEventHandler
	keyEventHandlersMutex sync.Mutex
	inputEnabled          atomic.Bool

	mouseTrackingEnabled atomic.Bool
	mouseState           MouseState
	mouseStateMutex      sync.Mutex

	initialized atomic.Bool
	ctx         context.Context
	cancelFunc  context.CancelFunc
	doneChan    <-chan struct{}
	closedChan  chan struct{}

	stateMutex    sync.Mutex
	readyChannels []chan bool
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

func (w *Window) initAssets() {
	Assets.SetName("_assets")
	Assets.SetWindow(w)
	w.AddService(Assets)
}

func (w *Window) initObjects() {
	for i := len(w.objectInitQueue) - 1; i >= 0; i-- {
		initInv := w.objectInitQueue[i]
		ok := initInv.Func()
		if initInv.ReturnChan != nil {
			select {
			case initInv.ReturnChan <- ok:
			default:
			}
		}
		close(initInv.ReturnChan)
		w.objectInitQueue = w.objectInitQueue[:i]
	}
}

func (w *Window) initServices() {
	for i := len(w.serviceInitQueue) - 1; i >= 0; i-- {
		initInv := w.serviceInitQueue[i]
		ok := initInv.Func()
		if initInv.ReturnChan != nil {
			select {
			case initInv.ReturnChan <- ok:
			default:
			}
		}
		close(initInv.ReturnChan)
		w.serviceInitQueue = w.serviceInitQueue[:i]
	}
}

func (w *Window) updateObjects(deltaTime int64) {
	for _, o := range w.objects {
		o.Update(deltaTime)
	}
}

func (w *Window) updateServices(deltaTime int64) {
	for _, s := range w.services {
		s.Update(deltaTime)
	}
}

func (w *Window) drawObjects(deltaTime int64) {
	for _, o := range w.drawableObjects {
		o.Draw(deltaTime)
	}
}

func (w *Window) closeObjects() {
	for i := len(w.objectCloseQueue) - 1; i >= 0; i-- {
		closeInv := w.objectCloseQueue[i]
		closeInv.Func()
		if closeInv.DoneChan != nil {
			select {
			case closeInv.DoneChan <- true:
			default:
			}
		}
		close(closeInv.DoneChan)
		w.objectCloseQueue = w.objectCloseQueue[:i]
	}
}

func (w *Window) closeServices() {
	for i := len(w.serviceCloseQueue) - 1; i >= 0; i-- {
		closeInv := w.serviceCloseQueue[i]
		closeInv.Func()
		if closeInv.DoneChan != nil {
			select {
			case closeInv.DoneChan <- true:
			default:
			}
		}
		close(closeInv.DoneChan)
		w.serviceCloseQueue = w.serviceCloseQueue[:i]
	}
}

func (w *Window) resizeObjects(oldWidth, oldHeight, newWidth, newHeight int32) {
	for _, o := range w.windowObjects {
		o.Resize(oldWidth, oldHeight, newWidth, newHeight)
	}

	for _, o := range w.objects {
		if resizer, ok := o.(Resizer); ok {
			resizer.Resize(oldWidth, oldHeight, newWidth, newHeight)
		}
	}
}

func (w *Window) handlePngRequests() {
	if w.pngExportRequest == nil {
		return
	}
	pngBytes := w.pngExportRequest.Func()
	select {
	case w.pngExportRequest.ReturnChan <- &pngBytes:
	default:
	}
	close(w.pngExportRequest.ReturnChan)
	w.pngExportRequest = nil
}

func (w *Window) tick(deltaTime int64) {
	w.initServices()
	w.closeServices()
	w.updateServices(deltaTime)

	w.initObjects()
	w.closeObjects()

	w.updateObjects(deltaTime)
	w.drawObjects(deltaTime)

	w.handlePngRequests()
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

func (w *Window) disposeAllObjects() {
	for _, obj := range w.objects {
		closeInv := newAsyncVoidInvocation(obj.Close)
		w.objectCloseQueue = append(w.objectCloseQueue, closeInv)
	}
	for _, obj := range w.drawableObjects {
		closeInv := newAsyncVoidInvocation(obj.Close)
		w.objectCloseQueue = append(w.objectCloseQueue, closeInv)
	}
	for _, obj := range w.windowObjects {
		closeInv := newAsyncVoidInvocation(obj.Close)
		w.objectCloseQueue = append(w.objectCloseQueue, closeInv)
	}
	w.closeObjects()
	w.objects = make([]Object, 0)
	w.drawableObjects = make([]DrawableObject, 0)
	w.windowObjects = make([]WindowObject, 0)
}

func (w *Window) disposeAllServices(ignoreProtection bool) {
	for _, svc := range w.services {
		if ignoreProtection {
			svc.SetProtected(false)
		}
		if svc.Protected() {
			continue
		}
		closeInv := newAsyncVoidInvocation(svc.Close)
		w.serviceCloseQueue = append(w.serviceCloseQueue, closeInv)
	}
	w.closeServices()
	w.services = make([]Service, 0)
}

func (w *Window) close() {
	w.stateMutex.Unlock()
	w.disposeAllObjects()
	w.disposeAllServices(true)
	w.glwin.SetKeyCallback(nil)
	glfw.DetachCurrentContext()
	w.glwin.Destroy()
	close(w.closedChan)
	runtime.UnlockOSThread()
}

func (w *Window) Init(ctx context.Context, cancelFunc context.CancelFunc) {
	if w.initialized.Load() {
		return
	}

	w.ctx = ctx
	w.cancelFunc = cancelFunc
	w.doneChan = w.ctx.Done()
	w.closedChan = make(chan struct{})

	go func() {
		runtime.LockOSThread()

		window, err := newGlfwWindow(w.Title(), w.Width(), w.Height())
		if err != nil {
			panic(err)
		}

		w.initAssets()

		w.stateMutex.Lock()
		w.glwin = window
		w.glwin.SetKeyCallback(w.keyEventCallback)
		w.initServices()
		w.initObjects()
		w.clearScreen()
		w.glwin.SwapBuffers()
		w.stateMutex.Unlock()

		for _, c := range w.readyChannels {
			close(c)
		}

		now := time.Now().UnixMicro()
		lastTick := now
		deltaTime := now
		drawInterval := int64(1000000 / w.targetFramerate.Load())

		w.initialized.Store(true)

		for {
			w.stateMutex.Lock()

			if w.glwin.ShouldClose() {
				w.cancelFunc()
			}

			select {
			case <-w.doneChan:
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

			w.clearScreen()
			w.tick(deltaTime)
			w.glwin.SwapBuffers()
			w.stateMutex.Unlock()

			if w.fullscreenRequested.Load() {
				w.fullscreenRequested.Store(false)
				w.enableFullscreenMode()
				w.resizeObjects(w.lastWidth.Load(), w.lastHeight.Load(), w.width.Load(), w.height.Load())
			} else if w.windowRequested.Load() {
				w.windowRequested.Store(false)
				w.enableWindowedMode()
				w.resizeObjects(w.lastWidth.Load(), w.lastHeight.Load(), w.width.Load(), w.height.Load())
			}

			glfw.PollEvents()
		}
	}()
}

func (w *Window) Close() {
	if w.initialized.Load() {
		w.stateMutex.Lock()
		w.glwin.SetShouldClose(true)
		w.initialized.Store(false)
		w.stateMutex.Unlock()
		<-w.closedChan
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
			go w.Close()
		})
	}
	return w
}

func (w *Window) ClearColor() color.RGBA {
	return FloatArrayToRgba(w.clearColor)
}

func (w *Window) SetClearColor(rgba color.RGBA) *Window {
	w.clearColor[0] = float32(rgba.R) / 255.0
	w.clearColor[1] = float32(rgba.G) / 255.0
	w.clearColor[2] = float32(rgba.B) / 255.0
	w.clearColor[3] = float32(rgba.A) / 255.0
	return w
}

func (w *Window) InitObject(object Initer) (ok bool) {
	initInv := newAsyncBoolInvocation(object.Init)
	w.stateMutex.Lock()
	w.objectInitQueue = append(w.objectInitQueue, initInv)
	w.stateMutex.Unlock()
	return <-initInv.ReturnChan
}

func (w *Window) InitObjectAsync(object Initer) {
	initInv := newAsyncBoolInvocation(object.Init)
	w.stateMutex.Lock()
	w.objectInitQueue = append(w.objectInitQueue, initInv)
	w.stateMutex.Unlock()
}

func (w *Window) CloseObject(object Closer) {
	closeInv := newAsyncVoidInvocation(object.Close)
	w.stateMutex.Lock()
	w.objectCloseQueue = append(w.objectCloseQueue, closeInv)
	w.stateMutex.Unlock()
	<-closeInv.DoneChan
}

func (w *Window) CloseObjectAsync(object Closer) {
	closeInv := newAsyncVoidInvocation(object.Close)
	w.stateMutex.Lock()
	w.objectCloseQueue = append(w.objectCloseQueue, closeInv)
	w.stateMutex.Unlock()
}

func (w *Window) GetObject(name string) Object {
	w.stateMutex.Lock()
	for _, obj := range w.windowObjects {
		if obj.Name() == name {
			w.stateMutex.Unlock()
			return obj
		}
		o := obj.Child(name)
		if o != nil {
			w.stateMutex.Unlock()
			return o
		}
	}
	for _, obj := range w.drawableObjects {
		if obj.Name() == name {
			w.stateMutex.Unlock()
			return obj
		}
	}
	for _, obj := range w.objects {
		if obj.Name() == name {
			w.stateMutex.Unlock()
			return obj
		}
	}
	w.stateMutex.Unlock()
	return nil
}

func (w *Window) AddObject(object any, waitForInit ...bool) {
	w.stateMutex.Lock()

	wait := false
	if len(waitForInit) > 0 && waitForInit[0] {
		wait = true
	}

	var initInv *asyncBoolInvocation

	if o, ok := object.(WindowObject); ok {
		o.SetWindow(w)
		w.objects = append(w.objects, o)
		w.drawableObjects = append(w.drawableObjects, o)
		w.windowObjects = append(w.windowObjects, o)
		initInv = newAsyncBoolInvocation(o.Init)
	}

	if initInv == nil {
		if o, ok := object.(DrawableObject); ok {
			w.objects = append(w.objects, o)
			w.drawableObjects = append(w.drawableObjects, o)
			initInv = newAsyncBoolInvocation(o.Init)
		}
	}

	if initInv == nil {
		if o, ok := object.(Object); ok {
			w.objects = append(w.objects, o)
			initInv = newAsyncBoolInvocation(o.Init)
		} else {
			panic(fmt.Errorf("cannot add struct of type %s, does not implement one of [Object|DrawableObject|WindowObject]", reflect.TypeOf(object).Name()))
		}
	}

	w.objectInitQueue = append(w.objectInitQueue, initInv)
	w.stateMutex.Unlock()

	if wait {
		<-initInv.ReturnChan
	}
}

func (w *Window) AddObjects(objects ...any) {
	w.stateMutex.Lock()

	for _, obj := range objects {
		if o, ok := obj.(WindowObject); ok {
			o.SetWindow(w)
			w.objects = append(w.objects, o)
			w.drawableObjects = append(w.drawableObjects, o)
			w.windowObjects = append(w.windowObjects, o)
			w.objectInitQueue = append(w.objectInitQueue, newAsyncBoolInvocation(o.Init))
			continue
		}

		if o, ok := obj.(DrawableObject); ok {
			w.objects = append(w.objects, o)
			w.drawableObjects = append(w.drawableObjects, o)
			w.objectInitQueue = append(w.objectInitQueue, newAsyncBoolInvocation(o.Init))
			continue
		}

		if o, ok := obj.(Object); ok {
			w.objects = append(w.objects, o)
			w.objectInitQueue = append(w.objectInitQueue, newAsyncBoolInvocation(o.Init))
			continue
		} else {
			typeName := reflect.Indirect(reflect.ValueOf(obj)).Type().Name()
			panic(fmt.Errorf("cannot add struct of type %s, does not implement one of [Object|DrawableObject|WindowObject]", typeName))
		}
	}

	w.stateMutex.Unlock()
}

func (w *Window) RemoveObject(name string) {
	w.stateMutex.Lock()

	for i := len(w.objects) - 1; i >= 0; i-- {
		if w.objects[i].Name() == name {
			w.objects = append(w.objects[:i], w.objects[i+1:]...)
			break
		}
	}
	for i := len(w.drawableObjects) - 1; i >= 0; i-- {
		if w.drawableObjects[i].Name() == name {
			w.drawableObjects = append(w.drawableObjects[:i], w.drawableObjects[i+1:]...)
			break
		}
	}
	for i := len(w.windowObjects) - 1; i >= 0; i-- {
		if w.windowObjects[i].Name() == name {
			w.windowObjects = append(w.windowObjects[:i], w.windowObjects[i+1:]...)
			break
		}
	}

	w.stateMutex.Unlock()
}

func (w *Window) RemoveAllObjects() {
	w.stateMutex.Lock()
	w.objects = make([]Object, 0)
	w.drawableObjects = make([]DrawableObject, 0)
	w.windowObjects = make([]WindowObject, 0)
	w.stateMutex.Unlock()
}

func (w *Window) DisposeObject(name string) {
	obj := w.GetObject(name)

	if obj != nil {
		w.stateMutex.Lock()
		for i := len(w.objects) - 1; i >= 0; i-- {
			if w.objects[i].Name() == name {
				w.objects = append(w.objects[:i], w.objects[i+1:]...)
				break
			}
		}
		for i := len(w.drawableObjects) - 1; i >= 0; i-- {
			if w.drawableObjects[i].Name() == name {
				w.drawableObjects = append(w.drawableObjects[:i], w.drawableObjects[i+1:]...)
				break
			}
		}
		for i := len(w.windowObjects) - 1; i >= 0; i-- {
			if w.windowObjects[i].Name() == name {
				w.windowObjects = append(w.windowObjects[:i], w.windowObjects[i+1:]...)
				break
			}
		}
		closeInv := newAsyncVoidInvocation(obj.Close)
		w.objectCloseQueue = append(w.objectCloseQueue, closeInv)
		w.stateMutex.Unlock()
		<-closeInv.DoneChan
	}
}

func (w *Window) DisposeObjectAsync(name string) {
	obj := w.GetObject(name)

	if obj != nil {
		w.stateMutex.Lock()
		for i := len(w.objects) - 1; i >= 0; i-- {
			if w.objects[i].Name() == name {
				w.objects = append(w.objects[:i], w.objects[i+1:]...)
				break
			}
		}
		for i := len(w.drawableObjects) - 1; i >= 0; i-- {
			if w.drawableObjects[i].Name() == name {
				w.drawableObjects = append(w.drawableObjects[:i], w.drawableObjects[i+1:]...)
				break
			}
		}
		for i := len(w.windowObjects) - 1; i >= 0; i-- {
			if w.windowObjects[i].Name() == name {
				w.windowObjects = append(w.windowObjects[:i], w.windowObjects[i+1:]...)
				break
			}
		}
		w.objectCloseQueue = append(w.objectCloseQueue, newAsyncVoidInvocation(obj.Close))
		w.stateMutex.Unlock()
	}
}

func (w *Window) DisposeAllObjects() {
	doneChans := make([]chan bool, 0)
	w.stateMutex.Lock()
	for _, obj := range w.objects {
		closeInv := newAsyncVoidInvocation(obj.Close)
		w.objectCloseQueue = append(w.objectCloseQueue, closeInv)
		doneChans = append(doneChans, closeInv.DoneChan)
	}
	for _, obj := range w.drawableObjects {
		closeInv := newAsyncVoidInvocation(obj.Close)
		w.objectCloseQueue = append(w.objectCloseQueue, closeInv)
		doneChans = append(doneChans, closeInv.DoneChan)
	}
	for _, obj := range w.windowObjects {
		closeInv := newAsyncVoidInvocation(obj.Close)
		w.objectCloseQueue = append(w.objectCloseQueue, closeInv)
		doneChans = append(doneChans, closeInv.DoneChan)
	}
	w.objects = make([]Object, 0)
	w.drawableObjects = make([]DrawableObject, 0)
	w.windowObjects = make([]WindowObject, 0)
	w.stateMutex.Unlock()

	for _, doneChan := range doneChans {
		<-doneChan
	}
}

func (w *Window) DisposeAllObjectsAsync() {
	w.stateMutex.Lock()
	for _, obj := range w.objects {
		w.objectCloseQueue = append(w.objectCloseQueue, newAsyncVoidInvocation(obj.Close))
	}
	for _, obj := range w.drawableObjects {
		w.objectCloseQueue = append(w.objectCloseQueue, newAsyncVoidInvocation(obj.Close))
	}
	for _, obj := range w.windowObjects {
		w.objectCloseQueue = append(w.objectCloseQueue, newAsyncVoidInvocation(obj.Close))
	}
	w.objects = make([]Object, 0)
	w.drawableObjects = make([]DrawableObject, 0)
	w.windowObjects = make([]WindowObject, 0)
	w.stateMutex.Unlock()
}

func (w *Window) GetService(name string) Service {
	w.stateMutex.Lock()
	for _, svc := range w.services {
		if svc.Name() == name {
			w.stateMutex.Unlock()
			return svc
		}
	}
	w.stateMutex.Unlock()
	return nil
}

func (w *Window) AddService(service Service) {
	w.stateMutex.Lock()
	w.services = append(w.services, service)
	w.serviceInitQueue = append(w.serviceInitQueue, newAsyncBoolInvocation(service.Init))
	w.stateMutex.Unlock()
}

func (w *Window) RemoveService(service Service) {
	if service == nil || service.Protected() {
		return
	}

	w.stateMutex.Lock()
	for i := len(w.services) - 1; i >= 0; i-- {
		if w.services[i].Name() == service.Name() {
			w.services = append(w.services[:i], w.services[i+1:]...)
			break
		}
	}
	w.stateMutex.Unlock()
}

func (w *Window) InitService(service Service) (ok bool) {
	initInv := newAsyncBoolInvocation(service.Init)
	w.stateMutex.Lock()
	w.serviceInitQueue = append(w.serviceInitQueue, initInv)
	w.stateMutex.Unlock()
	return <-initInv.ReturnChan
}

func (w *Window) InitServiceAsync(service Service) {
	w.stateMutex.Lock()
	w.serviceInitQueue = append(w.serviceInitQueue, newAsyncBoolInvocation(service.Init))
	w.stateMutex.Unlock()
}

func (w *Window) DisposeService(service Service) {
	if service == nil || service.Protected() {
		return
	}

	w.stateMutex.Lock()
	for i := len(w.services) - 1; i >= 0; i-- {
		if w.services[i].Name() == service.Name() {
			w.services = append(w.services[:i], w.services[i+1:]...)
			break
		}
	}
	closeInv := newAsyncVoidInvocation(service.Close)
	w.serviceCloseQueue = append(w.serviceCloseQueue, closeInv)
	w.stateMutex.Unlock()
	<-closeInv.DoneChan
}

func (w *Window) DisposeServiceAsync(service Service) {
	if service == nil || service.Protected() {
		return
	}

	w.stateMutex.Lock()
	w.serviceCloseQueue = append(w.serviceCloseQueue, newAsyncVoidInvocation(service.Close))
	for i := len(w.services) - 1; i >= 0; i-- {
		if w.services[i].Name() == service.Name() {
			w.services = append(w.services[:i], w.services[i+1:]...)
			break
		}
	}
	w.stateMutex.Unlock()
}

func (w *Window) DisposeAllServices(ignoreProtection bool) {
	doneChans := make([]chan bool, 0)
	w.stateMutex.Lock()
	for _, svc := range w.services {
		if ignoreProtection {
			svc.SetProtected(false)
		}
		if svc.Protected() {
			continue
		}
		closeInv := newAsyncVoidInvocation(svc.Close)
		w.serviceCloseQueue = append(w.serviceCloseQueue, closeInv)
		doneChans = append(doneChans, closeInv.DoneChan)
	}
	w.services = make([]Service, 0)
	w.stateMutex.Unlock()
	for _, doneChan := range doneChans {
		<-doneChan
	}
}

func (w *Window) DisposeAllServicesAsync(ignoreProtection bool) {
	w.stateMutex.Lock()
	for _, svc := range w.services {
		if ignoreProtection {
			svc.SetProtected(false)
		}
		if svc.Protected() {
			continue
		}
		w.serviceCloseQueue = append(w.serviceCloseQueue, newAsyncVoidInvocation(svc.Close))
	}
	w.services = make([]Service, 0)
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
		oldWidth := w.width.Load()
		oldHeight := w.height.Load()
		w.stateMutex.Lock()
		w.glwin.SetSize(width, int(oldHeight))
		w.resizeObjects(oldWidth, oldHeight, int32(width), oldHeight)
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
		oldWidth := w.width.Load()
		oldHeight := w.height.Load()
		w.stateMutex.Lock()
		w.glwin.SetSize(int(oldWidth), height)
		w.resizeObjects(oldWidth, oldHeight, oldWidth, int32(height))
		w.stateMutex.Unlock()
	}
	return w
}

func (w *Window) Size() (width, height int32) {
	return w.width.Load(), w.height.Load()
}

func (w *Window) SetSize(width, height int) *Window {
	w.width.Store(int32(width))
	w.height.Store(int32(height))
	if w.glwin != nil {
		oldWidth := w.width.Load()
		oldHeight := w.height.Load()
		w.stateMutex.Lock()
		w.glwin.SetSize(width, height)
		w.resizeObjects(oldWidth, oldHeight, int32(width), int32(height))
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
	state := w.mouseState
	w.mouseStateMutex.Unlock()
	return &state
}

func (w *Window) OverrideMouseState(state *MouseState) {
	w.mouseStateMutex.Lock()
	w.mouseState = *state
	w.mouseStateMutex.Unlock()
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

func (w *Window) ReadyChan() <-chan bool {
	c := make(chan bool)
	w.stateMutex.Lock()
	w.readyChannels = append(w.readyChannels, c)
	w.stateMutex.Unlock()
	return c
}

func (w *Window) Initialized() bool {
	return w.initialized.Load()
}

func (w *Window) toPNG() []byte {
	width := w.width.Load()
	height := w.height.Load()
	pixels := make([]byte, width*height*4)
	gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))

	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

	for y := int32(0); y < height; y++ {
		for x := int32(0); x < width; x++ {
			i := y*width*4 + x*4
			j := (height-y-1)*width*4 + x*4
			img.Pix[i+0] = pixels[j+0]
			img.Pix[i+1] = pixels[j+1]
			img.Pix[i+2] = pixels[j+2]
			img.Pix[i+3] = pixels[j+3]
		}
	}

	var resultBuffer bytes.Buffer
	if err := png.Encode(&resultBuffer, img); err != nil {
		panic(err)
	}

	return resultBuffer.Bytes()
}

func (w *Window) ToPNG() []byte {
	pngInv := newAsyncByteSliceInvocation(w.toPNG)
	w.stateMutex.Lock()
	w.pngExportRequest = pngInv
	w.stateMutex.Unlock()
	return *<-pngInv.ReturnChan
}

/******************************************************************************
 New Window Function
******************************************************************************/

func NewWindow() *Window {
	w := &Window{
		objects:          make([]Object, 0),
		drawableObjects:  make([]DrawableObject, 0),
		windowObjects:    make([]WindowObject, 0),
		services:         make([]Service, 0),
		keyEventHandlers: make([]*KeyEventHandler, 0),
	}

	w.SetWidth(defaultWinWidth)
	w.SetHeight(defaultWinHeight)
	w.SetTargetFramerate(defaultTargetFramerate)
	w.SetInputEnabled(true)
	return w
}
