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

	borderless bool

	title         string
	width         int
	height        int
	lastWidth     int
	lastHeight    int
	lastPositionX int
	lastPositionY int
	opacity       uint8

	clearColorRgba color.RGBA
	clearColorVec  mgl32.Vec4

	fullscreen              bool
	fullscreenButtonEnabled bool
	fullscreenRequested     bool
	windowRequested         bool

	targetFramerate atomic.Uint32
	vSyncEnabled    bool

	quitButtonEnabled bool

	titleChanged        bool
	sizeChanged         bool
	opacityChanged      bool
	vSyncEnabledChanged bool
	clearColorChanged   bool

	configChanged bool
	configMutex   sync.Mutex

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

	initialized   atomic.Bool
	ctx           context.Context
	cancelFunc    context.CancelFunc
	doneChan      <-chan struct{}
	closedChan    chan struct{}
	readyChannels []chan bool

	stateMutex sync.Mutex
}

/******************************************************************************
 Window Functions
******************************************************************************/

func (w *Window) clearScreen() {
	gl.ClearColor(w.clearColorVec[0], w.clearColorVec[1], w.clearColorVec[2], w.clearColorVec[3])
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (w *Window) enableFullscreenMode() {
	primaryMonitor := glfw.GetPrimaryMonitor()
	vidMode := primaryMonitor.GetVideoMode()

	x, y := w.glwin.GetPos()
	w.lastPositionX = x
	w.lastPositionY = y
	w.lastWidth = w.width
	w.lastHeight = w.height
	w.width = vidMode.Width
	w.height = vidMode.Height
	w.glwin.SetMonitor(primaryMonitor, 0, 0, vidMode.Width, vidMode.Height, vidMode.RefreshRate)
	w.fullscreen = true
}

func (w *Window) enableWindowedMode() {
	primaryMonitor := glfw.GetPrimaryMonitor()
	vidMode := primaryMonitor.GetVideoMode()

	w.glwin.SetMonitor(nil, w.lastPositionX, w.lastPositionY, w.lastWidth, w.lastHeight, vidMode.RefreshRate)
	w.width = w.lastWidth
	w.height = w.lastHeight
	w.lastWidth = vidMode.Width
	w.lastHeight = vidMode.Height
	w.fullscreen = false
}

func (w *Window) addAssetLibraryService() {
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

func (w *Window) resizeObjects(newWidth, newHeight int) {
	for _, o := range w.objects {
		if resizer, ok := o.(Resizer); ok {
			resizer.Resize(newWidth, newHeight)
		}
	}
}

func (w *Window) refreshLayout() {
	w.configMutex.Lock()

	if w.titleChanged {
		w.glwin.SetTitle(w.title)
		w.titleChanged = false
	}

	if w.opacityChanged {
		w.glwin.SetOpacity(float32(w.opacity) / 255.0)
		w.opacityChanged = false
	}

	if w.vSyncEnabledChanged {
		if w.vSyncEnabled {
			glfw.SwapInterval(1)
		} else {
			glfw.SwapInterval(0)
		}
		w.vSyncEnabledChanged = false
	}

	if w.clearColorChanged {
		w.clearColorVec[0] = float32(w.clearColorRgba.R) / 255.0
		w.clearColorVec[1] = float32(w.clearColorRgba.G) / 255.0
		w.clearColorVec[2] = float32(w.clearColorRgba.B) / 255.0
		w.clearColorVec[3] = float32(w.clearColorRgba.A) / 255.0
		w.clearColorChanged = false
	}

	if w.fullscreenRequested {
		w.enableFullscreenMode()
		width := w.width
		height := w.height
		w.fullscreenRequested = false
		w.configMutex.Unlock()
		w.resizeObjects(width, height)
		return
	} else if w.windowRequested {
		w.enableWindowedMode()
		width := w.width
		height := w.height
		w.windowRequested = false
		w.configMutex.Unlock()
		w.resizeObjects(width, height)
		return
	} else if w.sizeChanged {
		w.glwin.SetSize(w.width, w.height)
		width := w.width
		height := w.height
		w.sizeChanged = false
		w.configMutex.Unlock()
		w.resizeObjects(width, height)
		return
	}

	w.configMutex.Unlock()
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

func (w *Window) keyEventCallback(_ *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
	if !w.inputEnabled.Load() {
		return
	}

	w.keyEventHandlersMutex.Lock()
	for _, h := range w.keyEventHandlers {
		if h.Key == key && h.Action == action {
			h.Callback(w, key, action)
		}
	}
	w.keyEventHandlersMutex.Unlock()
}

func (w *Window) disposeAllObjects() {
	for _, obj := range w.objects {
		obj.Close()
	}
	for _, obj := range w.drawableObjects {
		obj.Close()
	}
	for _, obj := range w.windowObjects {
		obj.Close()
	}
	w.objects = make([]Object, 0)
	w.drawableObjects = make([]DrawableObject, 0)
	w.windowObjects = make([]WindowObject, 0)
}

func (w *Window) disposeAllServices() {
	for _, svc := range w.services {
		svc.SetProtected(false)
		svc.Close()
	}
	w.services = make([]Service, 0)
	w.serviceCloseQueue = make([]*asyncVoidInvocation, 0)
	w.serviceInitQueue = make([]*asyncBoolInvocation, 0)
}

func (w *Window) run() {
	runtime.LockOSThread()

	window, err := newGlfwWindow(w.Title(), w.Width(), w.Height(), w.borderless)
	if err != nil {
		panic(err)
	}

	w.addAssetLibraryService()

	w.stateMutex.Lock()
	w.glwin = window
	w.glwin.SetKeyCallback(w.keyEventCallback)
	w.initServices()
	w.initObjects()
	w.clearScreen()
	w.glwin.SwapBuffers()
	w.stateMutex.Unlock()

	w.initialized.Store(true)

	for _, c := range w.readyChannels {
		close(c)
	}

	now := time.Now().UnixMicro()
	lastTick := now
	deltaTime := now
	drawInterval := int64(1000000 / w.targetFramerate.Load())

	for {
		w.stateMutex.Lock()

		if w.glwin.ShouldClose() || !w.initialized.Load() {
			w.close()
			return
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
		w.handlePngRequests()
		w.stateMutex.Unlock()

		if w.configChanged {
			w.refreshLayout()
			w.configChanged = false
		}

		glfw.PollEvents()
	}
}

func (w *Window) tick(deltaTime int64) {
	w.initServices()
	w.closeServices()
	w.updateServices(deltaTime)

	w.initObjects()
	w.closeObjects()

	w.updateObjects(deltaTime)
	w.drawObjects(deltaTime)
}

func (w *Window) close() {
	w.disposeAllObjects()
	w.disposeAllServices()
	ClearLabelTextureCache()
	w.glwin.SetKeyCallback(nil)
	glfw.PollEvents()
	w.glwin.Destroy()
	runtime.UnlockOSThread()
	w.cancelFunc()
	w.stateMutex.Unlock()
	close(w.closedChan)
}

func (w *Window) Init(ctx context.Context, cancelFunc context.CancelFunc) {
	if w.initialized.Load() {
		return
	}

	w.ctx = ctx
	w.cancelFunc = cancelFunc
	w.doneChan = w.ctx.Done()
	w.closedChan = make(chan struct{})

	go w.run()
}

func (w *Window) Close() {
	if w.initialized.Load() {
		w.stateMutex.Lock()
		w.glwin.SetShouldClose(true)
		w.cancelFunc()
		w.initialized.Store(false)
		w.stateMutex.Unlock()
		<-w.closedChan
	}
}

func (w *Window) AddKeyEventHandler(key glfw.Key, action glfw.Action, callback func(window *Window, key glfw.Key, action glfw.Action)) *KeyEventHandler {
	w.keyEventHandlersMutex.Lock()
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
	if len(w.keyEventHandlers) == 0 {
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
	w.configMutex.Lock()
	if enabled {
		w.fullscreenRequested = true
	} else {
		w.windowRequested = true
	}
	w.configChanged = true
	w.configMutex.Unlock()
	return w
}

func (w *Window) EnableFullscreenKey() *Window {
	w.configMutex.Lock()
	if !w.fullscreenButtonEnabled {
		w.fullscreenButtonEnabled = true
		w.AddKeyEventHandler(glfw.KeyF11, glfw.Press, func(win *Window, _ glfw.Key, _ glfw.Action) {
			win.SetFullscreenEnabled(!win.IsFullscreen())
		})
	}
	w.configMutex.Unlock()
	return w
}

func (w *Window) EnableQuitKey(cancelFuncs ...context.CancelFunc) *Window {
	w.configMutex.Lock()
	if !w.quitButtonEnabled {
		w.quitButtonEnabled = true

		w.AddKeyEventHandler(glfw.KeyEscape, glfw.Press, func(_ *Window, _ glfw.Key, _ glfw.Action) {
			for _, cancelFunc := range cancelFuncs {
				cancelFunc()
			}
			w.Close()
		})
	}
	w.configMutex.Unlock()
	return w
}

func (w *Window) ClearColor() (rgba color.RGBA) {
	w.configMutex.Lock()
	rgba = w.clearColorRgba
	w.configMutex.Unlock()
	return
}

func (w *Window) SetClearColor(rgba color.RGBA) *Window {
	w.configMutex.Lock()
	w.clearColorRgba = rgba
	w.clearColorChanged = true
	w.configChanged = true
	w.configMutex.Unlock()
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
	service.SetWindow(w)
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

func (w *Window) Title() (title string) {
	w.configMutex.Lock()
	title = w.title
	w.configMutex.Unlock()
	return
}

func (w *Window) SetTitle(title string) *Window {
	w.configMutex.Lock()
	w.title = title
	w.titleChanged = true
	w.configChanged = true
	w.configMutex.Unlock()
	return w
}

func (w *Window) Width() (width int) {
	w.configMutex.Lock()
	width = w.width
	w.configMutex.Unlock()
	return
}

func (w *Window) SetWidth(width int) *Window {
	w.configMutex.Lock()
	w.width = width
	w.sizeChanged = true
	w.configChanged = true
	w.configMutex.Unlock()
	return w
}

func (w *Window) Height() (height int) {
	w.configMutex.Lock()
	height = w.height
	w.configMutex.Unlock()
	return
}

func (w *Window) SetHeight(height int) *Window {
	w.configMutex.Lock()
	w.height = height
	w.sizeChanged = true
	w.configChanged = true
	w.configMutex.Unlock()
	return w
}

func (w *Window) Size() (width, height int) {
	w.configMutex.Lock()
	width = w.width
	height = w.height
	w.configMutex.Unlock()
	return
}

func (w *Window) SetSize(width, height int) *Window {
	w.configMutex.Lock()
	w.width = width
	w.height = height
	w.sizeChanged = true
	w.configChanged = true
	w.configMutex.Unlock()
	return w
}

func (w *Window) Opacity() (alpha uint8) {
	w.configMutex.Lock()
	alpha = w.opacity
	w.configMutex.Unlock()
	return
}

func (w *Window) SetOpacity(alpha uint8) *Window {
	w.configMutex.Lock()
	w.opacity = alpha
	w.opacityChanged = true
	w.configChanged = true
	w.configMutex.Unlock()
	return w
}

func (w *Window) AspectRatio() (ratio float32) {
	w.configMutex.Lock()
	ratio = float32(w.width) / float32(w.height)
	w.configMutex.Unlock()
	return
}

func (w *Window) AspectRatioInv() (invRatio float32) {
	w.configMutex.Lock()
	invRatio = float32(w.height) / float32(w.width)
	w.configMutex.Unlock()
	return
}

func (w *Window) AspectRatio2D() (ratio2d mgl32.Vec2) {
	w.configMutex.Lock()

	switch {
	case w.width > w.height:
		ratio2d = mgl32.Vec2{1, float32(w.width) / float32(w.height)}
	case w.height > w.width:
		ratio2d = mgl32.Vec2{float32(w.height) / float32(w.width), 1}
	default:
		ratio2d = mgl32.Vec2{1, 1}
	}

	w.configMutex.Unlock()
	return
}

func (w *Window) AspectRatio2DInv() (invRatio2d mgl32.Vec2) {
	w.configMutex.Lock()

	switch {
	case w.width > w.height:
		invRatio2d = mgl32.Vec2{float32(w.height) / float32(w.width), 1}
	case w.height > w.width:
		invRatio2d = mgl32.Vec2{1, float32(w.width) / float32(w.height)}
	default:
		invRatio2d = mgl32.Vec2{1, 1}
	}

	w.configMutex.Unlock()
	return
}

func (w *Window) ScaleX(x float32) (scaledX float32) {
	w.configMutex.Lock()

	if w.width > w.height {
		scaledX = x * (float32(w.height) / float32(w.width))
	} else {
		scaledX = x
	}

	w.configMutex.Unlock()
	return
}

func (w *Window) ScaleY(y float32) (scaledY float32) {
	w.configMutex.Lock()

	if w.height > w.width {
		scaledY = y * (float32(w.width) / float32(w.height))
	} else {
		scaledY = y
	}

	w.configMutex.Unlock()
	return
}

func (w *Window) IsFullscreen() bool {
	return w.fullscreen
}

func (w *Window) TargetFramerate() uint32 {
	return w.targetFramerate.Load()
}

func (w *Window) SetTargetFramerate(framerate int) *Window {
	w.targetFramerate.Store(uint32(framerate))
	return w
}

func (w *Window) VSyncEnabled() bool {
	w.configMutex.Lock()
	enabled := w.vSyncEnabled
	w.configMutex.Unlock()
	return enabled
}

func (w *Window) SetVSyncEnabled(enabled bool) *Window {
	w.configMutex.Lock()
	w.vSyncEnabled = enabled
	w.vSyncEnabledChanged = true
	w.configChanged = true
	w.configMutex.Unlock()
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
	w.configMutex.Lock()
	width := int32(w.width)
	height := int32(w.height)
	w.configMutex.Unlock()

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

func NewWindow(borderless ...bool) *Window {
	winBorderless := false
	if len(borderless) > 0 {
		winBorderless = borderless[0]
	}

	w := &Window{
		borderless:     winBorderless,
		clearColorVec:  mgl32.Vec4{0, 0, 0, 1},
		clearColorRgba: color.RGBA{A: 255},
		opacity:        255,
	}

	w.SetWidth(defaultWinWidth)
	w.SetHeight(defaultWinHeight)
	w.SetTargetFramerate(defaultTargetFramerate)
	w.SetVSyncEnabled(true)
	w.SetInputEnabled(true)

	return w
}
