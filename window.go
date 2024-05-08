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
	"sync"
	"sync/atomic"
)

/******************************************************************************
 Window
******************************************************************************/

type Window struct {
	glwin *glfw.Window

	title         string
	borderless    bool
	resizable     bool
	multisampling bool
	secondary     bool

	width      int
	height     int
	lastWidth  int
	lastHeight int
	lastPosX   int
	lastPosY   int
	reqPosX    int
	reqPosY    int

	opacity        uint8
	clearColorRgba color.RGBA
	clearColorVec  mgl32.Vec4

	fullscreen              bool
	fullscreenButtonEnabled bool
	fullscreenRequested     bool
	windowRequested         bool

	quitButtonEnabled bool

	titleChanged      bool
	sizeChanged       bool
	positionChanged   bool
	opacityChanged    bool
	clearColorChanged bool

	configChanged bool
	configMutex   sync.Mutex

	objects         []Object
	drawableObjects []DrawableObject
	windowObjects   []WindowObject

	services []Service
	assets   *AssetLibrary

	objectInitQueue  []*asyncBoolInvocation
	objectCloseQueue []*asyncVoidInvocation

	serviceInitQueue  []*asyncBoolInvocation
	serviceCloseQueue []*asyncVoidInvocation

	pngExportRequest *asyncByteSliceInvocation

	labelCache map[string]*Texture2D

	keyEventChan          chan *KeyEvent
	keyEventHandlers      map[uint64][]*KeyEventHandler
	keyEventHandlersMutex sync.Mutex

	mouseTrackingEnabled atomic.Bool
	mouseState           MouseState
	mouseStateMutex      sync.Mutex

	initialized   atomic.Bool
	doneChan      <-chan struct{}
	readyChannels []chan bool

	hasFocus      bool
	disableOnBlur bool

	stateMutex sync.Mutex
}

/******************************************************************************
 WindowHints
******************************************************************************/

type WindowHints struct {
	Borderless    bool
	Resizable     bool
	MultiSampling bool
}

func NewWindowHints(borderless, resizable, multisampling bool) *WindowHints {
	return &WindowHints{
		Borderless:    borderless,
		Resizable:     resizable,
		MultiSampling: multisampling,
	}
}

/******************************************************************************
 GlfwWindow Implementation
******************************************************************************/

func (w *Window) GLFW() (glwin *glfw.Window) {
	w.stateMutex.Lock()
	glwin = w.glwin
	w.stateMutex.Unlock()
	return
}

func (w *Window) Title() (title string) {
	w.configMutex.Lock()
	title = w.title
	w.configMutex.Unlock()
	return
}

func (w *Window) Width() (width int) {
	w.configMutex.Lock()
	width = w.width
	w.configMutex.Unlock()
	return
}

func (w *Window) Height() (height int) {
	w.configMutex.Lock()
	height = w.height
	w.configMutex.Unlock()
	return
}

func (w *Window) Borderless() (borderless bool) {
	w.configMutex.Lock()
	borderless = w.borderless
	w.configMutex.Unlock()
	return
}

func (w *Window) Resizable() (resizable bool) {
	w.configMutex.Lock()
	resizable = w.resizable
	w.configMutex.Unlock()
	return
}

func (w *Window) MultiSamplingEnabled() (enabled bool) {
	w.configMutex.Lock()
	enabled = w.multisampling
	w.configMutex.Unlock()
	return
}

func (w *Window) IsSecondary() (secondary bool) {
	w.configMutex.Lock()
	secondary = w.secondary
	w.configMutex.Unlock()
	return
}

func (w *Window) Init(glwin *glfw.Window, ctx context.Context) {
	w.stateMutex.Lock()

	if w.initialized.Load() {
		w.stateMutex.Unlock()
		return
	}

	w.glwin = glwin
	w.doneChan = ctx.Done()

	w.addAssetLibraryService()
	w.registerFocusCallback()
	w.registerMaximizeCallback()
	w.registerKeyCallback()

	w.initServices()
	w.initObjects()

	w.clearScreen()
	w.refreshScreen()

	go w.handleKeyEvents()

	w.initialized.Store(true)

	for _, c := range w.readyChannels {
		close(c)
	}

	w.stateMutex.Unlock()
}

func (w *Window) Update(deltaTime int64) {
	w.stateMutex.Lock()

	if w.glwin.ShouldClose() {
		w.stateMutex.Unlock()
		go CloseWindowAsync(w)
		return
	}

	if !w.hasFocus && w.disableOnBlur {
		w.stateMutex.Unlock()
		return
	}

	w.clearScreen()
	w.tick(deltaTime)
	w.refreshScreen()
	w.handlePngRequests()

	if w.configChanged {
		w.refreshLayout()
		w.configChanged = false
	}

	w.stateMutex.Unlock()
}

func (w *Window) Close() {
	w.stateMutex.Lock()
	if !w.initialized.Load() {
		w.stateMutex.Unlock()
		return
	}
	w.disposeAllObjects()
	w.disposeAllServices()
	close(w.keyEventChan)
	w.initialized.Store(false)
	w.stateMutex.Unlock()
}

/******************************************************************************
 Window Functions
******************************************************************************/

func (w *Window) enableFullscreenMode() {
	primaryMonitor := glfw.GetPrimaryMonitor()
	vidMode := primaryMonitor.GetVideoMode()

	x, y := w.glwin.GetPos()
	w.lastPosX = x
	w.lastPosY = y
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

	w.glwin.SetMonitor(nil, w.lastPosX, w.lastPosY, w.lastWidth, w.lastHeight, vidMode.RefreshRate)
	w.width = w.lastWidth
	w.height = w.lastHeight
	w.lastWidth = vidMode.Width
	w.lastHeight = vidMode.Height
	w.fullscreen = false
}

func (w *Window) getFontOrDefault(fontName string) Font {
	asset := w.assets.Get(fontName)
	if asset != nil {
		if fontAsset, ok := asset.(Font); ok {
			return fontAsset
		}
	}

	asset = w.assets.Get(DefaultFont)
	if asset != nil {
		if fontAsset, ok := asset.(Font); ok {
			return fontAsset
		}
	}

	return nil
}

func (w *Window) addAssetLibraryService() (lib *AssetLibrary) {
	for _, svc := range w.services {
		if svc.Name() == "_assets" {
			lib = svc.(*AssetLibrary)
			break
		}
	}
	if lib == nil {
		lib = DefaultAssetLibrary()
		w.services = append(w.services, lib)
		w.serviceInitQueue = append(w.serviceInitQueue, newAsyncBoolInvocation(lib.Init))
	}
	w.assets = lib
	return
}

func (w *Window) getAssetLibraryService() (lib *AssetLibrary) {
	lib = w.assets
	return
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

	if w.clearColorChanged {
		w.clearColorVec[0] = float32(w.clearColorRgba.R) / 255.0
		w.clearColorVec[1] = float32(w.clearColorRgba.G) / 255.0
		w.clearColorVec[2] = float32(w.clearColorRgba.B) / 255.0
		w.clearColorVec[3] = float32(w.clearColorRgba.A) / 255.0
		w.clearColorChanged = false
	}

	if w.positionChanged {
		w.glwin.SetPos(w.reqPosX, w.reqPosY)
		w.positionChanged = false
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

func (w *Window) handleKeyEvents() {
	for {
		select {
		case <-w.doneChan:
			return
		default:
		}

		select {
		case keyEvent, ok := <-w.keyEventChan:
			if !ok {
				return
			}

			w.keyEventHandlersMutex.Lock()
			if handlers, exists := w.keyEventHandlers[keyEvent.Event()]; exists {
				for _, h := range handlers {
					h.Callback(w, keyEvent.Key, keyEvent.Action)
				}
			}
			w.keyEventHandlersMutex.Unlock()
		}
	}
}

func (w *Window) keyEventCallback(_ *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
	select {
	case w.keyEventChan <- &KeyEvent{Key: key, Action: action}:
	default:
	}
}

func (w *Window) registerKeyCallback() {
	w.keyEventChan = make(chan *KeyEvent, 1024)
	w.glwin.SetKeyCallback(w.keyEventCallback)
}

func (w *Window) registerFocusCallback() {
	w.hasFocus = true
	w.glwin.SetFocusCallback(func(win *glfw.Window, focused bool) {
		w.stateMutex.Lock()
		w.hasFocus = focused
		w.stateMutex.Unlock()
	})
}

func (w *Window) registerMaximizeCallback() {
	w.glwin.SetMaximizeCallback(func(win *glfw.Window, maximized bool) {
		w.stateMutex.Lock()
		w.width, w.height = win.GetSize()
		w.sizeChanged = true
		w.configChanged = true
		w.stateMutex.Unlock()
	})
}

func (w *Window) clearScreen() {
	gl.ClearColor(w.clearColorVec[0], w.clearColorVec[1], w.clearColorVec[2], w.clearColorVec[3])
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (w *Window) refreshScreen() {
	w.glwin.SwapBuffers()
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

func (w *Window) AddKeyEventHandler(receiver any, key glfw.Key, action glfw.Action,
	callback func(window *Window, key glfw.Key, action glfw.Action)) *KeyEventHandler {
	w.keyEventHandlersMutex.Lock()
	handler := &KeyEventHandler{
		Receiver: receiver,
		Key:      key,
		Action:   action,
		Callback: callback,
	}

	if handlers, ok := w.keyEventHandlers[handler.Event()]; ok {
		w.keyEventHandlers[handler.Event()] = append(handlers, handler)
	} else {
		w.keyEventHandlers[handler.Event()] = []*KeyEventHandler{handler}
	}
	w.keyEventHandlersMutex.Unlock()
	return handler
}

func (w *Window) RemoveKeyEventHandlers(receiver any) {
	if len(w.keyEventHandlers) == 0 {
		return
	}

	w.keyEventHandlersMutex.Lock()
	for event, handlers := range w.keyEventHandlers {
		index := -1
		for i, h := range handlers {
			if h.Receiver == receiver {
				index = i
				break
			}
		}
		if index != -1 {
			handlers = append(handlers[:index], handlers[index+1:]...)
		}
		w.keyEventHandlers[event] = handlers
	}
	w.keyEventHandlersMutex.Unlock()
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
		w.AddKeyEventHandler(w, glfw.KeyF11, glfw.Press, func(win *Window, _ glfw.Key, _ glfw.Action) {
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
		w.AddKeyEventHandler(w, glfw.KeyEscape, glfw.Press, func(_ *Window, _ glfw.Key, _ glfw.Action) {
			for _, cancelFunc := range cancelFuncs {
				cancelFunc()
			}
			CloseWindowAsync(w)
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

func (w *Window) SetSecondary(secondary bool) *Window {
	w.stateMutex.Lock()
	w.secondary = secondary
	w.stateMutex.Unlock()
	return w
}

func (w *Window) SetTitle(title string) *Window {
	w.configMutex.Lock()
	w.title = title
	w.titleChanged = true
	w.configChanged = true
	w.configMutex.Unlock()
	return w
}

func (w *Window) SetWidth(width int) *Window {
	w.configMutex.Lock()
	w.width = width
	w.sizeChanged = true
	w.configChanged = true
	w.configMutex.Unlock()
	return w
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

func (w *Window) Position() (x, y int) {
	w.configMutex.Lock()
	if w.glwin == nil {
		w.configMutex.Unlock()
		return 0, 0
	}
	x, y = w.glwin.GetPos()
	w.configMutex.Unlock()
	return
}

func (w *Window) SetPosition(x, y int) *Window {
	w.configMutex.Lock()
	w.reqPosX = x
	w.reqPosY = y
	w.positionChanged = true
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

func (w *Window) Maximize() {
	w.stateMutex.Lock()
	if w.initialized.Load() {
		w.glwin.Maximize()
	}
	w.stateMutex.Unlock()
}

func (w *Window) Minimize() {
	w.stateMutex.Lock()
	if w.initialized.Load() {
		w.glwin.Iconify()
	}
	w.stateMutex.Unlock()
}

func (w *Window) Hide() {
	w.stateMutex.Lock()
	if w.initialized.Load() {
		w.glwin.Hide()
	}
	w.stateMutex.Unlock()
}

func (w *Window) Show() {
	w.stateMutex.Lock()
	if w.initialized.Load() {
		w.glwin.Show()
	}
	w.stateMutex.Unlock()
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

func (w *Window) HasFocus() (focused bool) {
	w.stateMutex.Lock()
	focused = w.hasFocus
	w.stateMutex.Unlock()
	return
}

func (w *Window) DisableOnBlur(disableOnBlur bool) *Window {
	w.stateMutex.Lock()
	w.disableOnBlur = disableOnBlur
	w.stateMutex.Unlock()
	return w
}

func (w *Window) SwapMouseButtons(swapped bool) *Window {
	w.mouseStateMutex.Lock()
	w.mouseState.ButtonsSwapped = swapped
	w.mouseStateMutex.Unlock()
	return w
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

	w.mouseStateMutex.Lock()
	w.mouseState.X = -999
	w.mouseState.Y = -999
	w.mouseStateMutex.Unlock()

	return w
}

func (w *Window) ReadyChan() <-chan bool {
	c := make(chan bool)
	w.stateMutex.Lock()
	if w.initialized.Load() {
		close(c)
		return c
	}
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

func (w *Window) ClearLabelCache() {
	w.stateMutex.Lock()
	for _, texture := range w.labelCache {
		texture.Close()
	}
	w.labelCache = make(map[string]*Texture2D)
	w.stateMutex.Unlock()
}

func (w *Window) Assets() (lib *AssetLibrary) {
	lib = w.getAssetLibraryService()
	if lib == nil {
		lib = w.addAssetLibraryService()
	}
	return
}

/******************************************************************************
 New Window Function
******************************************************************************/

func NewWindow(hints ...*WindowHints) *Window {
	winHints := WindowHints{}
	if len(hints) > 0 {
		winHints = *hints[0]
	}

	w := &Window{
		borderless:       winHints.Borderless,
		resizable:        winHints.Resizable,
		multisampling:    winHints.MultiSampling,
		clearColorVec:    mgl32.Vec4{0, 0, 0, 1},
		clearColorRgba:   color.RGBA{A: 255},
		opacity:          255,
		hasFocus:         true,
		labelCache:       make(map[string]*Texture2D),
		keyEventHandlers: make(map[uint64][]*KeyEventHandler),
	}

	w.SetWidth(defaultWinWidth)
	w.SetHeight(defaultWinHeight)

	return w
}
