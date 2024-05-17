// Package gfx is a 2D/3D graphics library powered by OpenGL.
// Easily render shapes, text, and UI controls, including a
// performant line graph designed for real-time signal analysis.
package gfx

import (
	"context"
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

/******************************************************************************
 Default Configuration
******************************************************************************/

const (
	openGlVersionMajor     = 4
	openGlVersionMinor     = 1
	defaultWinWidth        = 1000
	defaultWinHeight       = 1000
	defaultTargetFramerate = 60
	defaultVSyncEnabled    = true
)

/******************************************************************************
 Configuration
******************************************************************************/

var (
	targetFramerate atomic.Uint32
	vSyncEnabled    atomic.Bool
)

// TargetFramerate returns the target framerate in frames per second.
// The actual, rendered framerate will not exceed the monitor's refresh
// rate when V-Sync is enabled and will always be limited by system
// performance.
func TargetFramerate() (framesPerSec uint32) {
	framesPerSec = targetFramerate.Load()
	return
}

// SetTargetFramerate changes the current target framerate, which
// must be done before a Window has been initialized.
func SetTargetFramerate(framesPerSec uint32) {
	targetFramerate.Store(framesPerSec)
}

// VSyncEnabled returns true if V-Sync has been enabled.
// V-Sync ensures the active framebuffer will not be
// written to while being read by the monitor, which
// prevents "screen-tearing," but limits the maximum
// framerate to that of the monitor's refresh rate.
func VSyncEnabled() (enabled bool) {
	enabled = vSyncEnabled.Load()
	return
}

// SetVSyncEnabled is used to enable/disable V-Sync, which
// must be done before a Window has been initialized.
func SetVSyncEnabled(enabled bool) {
	vSyncEnabled.Store(enabled)
}

/******************************************************************************
 init Function
******************************************************************************/

// init is called by the Go runtime, before main() is called, and here
// it is used for ensuring the engine runs on the application's main
// thread and for setting the initial global configuration.
func init() {
	// Ensure that the initial Go routine used to execute your
	// application is locked to the thread to which it was assigned
	// (by the Go runtime).  By locking at the point init() is called,
	// hopefully that means locking the routine to the "main" thread
	// (see GLFW docs for more info).
	runtime.LockOSThread()

	// Set default configuration
	SetTargetFramerate(defaultTargetFramerate)
	SetVSyncEnabled(defaultVSyncEnabled)
}

/******************************************************************************
 State
******************************************************************************/

var (
	gfxContext    context.Context
	gfxCancelFunc context.CancelFunc

	gfxInitialized bool
	gfxRunning     bool
	gfxClosedChan  chan struct{}

	gfxWindow  GlfwWindow
	gfxWindows []GlfwWindow

	gfxWindowInitQueue  []GlfwWindow
	gfxWindowCloseQueue []GlfwWindow

	gfxStateMutex sync.Mutex
)

/******************************************************************************
 GlfwWindow
******************************************************************************/

// GlfwWindow is a wrapper for a GLFW window that also contains
// the services, objects, and assets needed to render and manage
// a scene or user interface, etc.  Life-cycle functions Init(),
// Update(), and Close() should NOT be directly invoked by importers
// of this package and are only exported so consumers may provide
// custom GlfwWindow implementations.
type GlfwWindow interface {
	GLFW() *glfw.Window

	// Title shall return the window's title, which is rendered in the
	// title bar, if the window is decorated (not borderless).
	Title() string

	// Width shall return the window's width, in pixels.
	Width() int

	// Height shall return the window's height, in pixels.
	Height() int

	// Borderless shall return true if the window will not be decorated
	// with a border/frame or a title bar.
	Borderless() bool

	// Resizable shall return true if the window can be resized either
	// by the maximize button or by dragging the edges/corners; both
	// of which are only possible with a decorated window.  Note that
	// you can always change the size of a window programmatically.
	Resizable() bool

	// MultiSamplingEnabled shall return true if antialiasing has been
	// enabled, which is provided by OpenGL's native MSAA capability,
	// with the sample size set to 4.
	MultiSamplingEnabled() bool

	// IsSecondary shall return true if the window has been designated
	// as a secondary/tool/dialog window, which means closing it will
	// not result in the application closing.
	IsSecondary() bool

	// Init should only be called by the engine and shall cause the window
	// to initialize itself, its services/objects, and begin the process of
	// updating ("ticking") them at a rate based on the target framerate.
	// Consumers of this package should initialize windows by passing them
	// to gfx.InitWindowAsync().
	Init(*glfw.Window, context.Context)

	// Update should only be called by the engine and shall cause the window
	// to update its services/objects, passing in the amount of time (in
	// microseconds) that has passed since the last update.
	Update(deltaTime int64)

	// Close should only be called by the engine and shall cause the window
	// to release any resources it consumed and call Close() on its
	// services/objects, causing them to do the same. Consumers of this
	// package should close windows by passing them to gfx.CloseWindowAsync().
	Close()
}

/******************************************************************************
 gfx Functions
******************************************************************************/

func gfxNewWindow(title string, width, height int, borderless, resizable, multisampling bool) (*glfw.Window, error) {
	if !gfxInitialized {
		return nil, fmt.Errorf("GLFW not initialized: must call gfx.Init() from the main thread first")
	}

	w := defaultWinWidth
	h := defaultWinHeight
	if width != 0 {
		w = width
	}
	if height != 0 {
		h = height
	}

	if borderless {
		glfw.WindowHint(glfw.Decorated, glfw.False)
	} else {
		glfw.WindowHint(glfw.Decorated, glfw.True)
	}

	if resizable {
		glfw.WindowHint(glfw.Resizable, glfw.True)
	} else {
		glfw.WindowHint(glfw.Resizable, glfw.False)
	}

	if multisampling {
		glfw.WindowHint(glfw.Samples, 4)
	} else {
		glfw.WindowHint(glfw.Samples, 0)
	}

	win, err := glfw.CreateWindow(w, h, title, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GLFW window: %w", err)
	}
	win.MakeContextCurrent()

	if vSyncEnabled.Load() {
		glfw.SwapInterval(1)
	} else {
		glfw.SwapInterval(0)
	}

	monitor := glfw.GetPrimaryMonitor()
	monMode := monitor.GetVideoMode()
	xPos := (monMode.Width - w) / 2
	yPos := (monMode.Height - h) / 2
	win.SetPos(xPos, yPos)

	if err = gl.Init(); err != nil {
		win.Destroy()
		return nil, fmt.Errorf("error initializing OpenGL: %w", err)
	}

	return win, nil
}

func gfxProcessInitQueue() {
	for i := len(gfxWindowInitQueue) - 1; i >= 0; i-- {
		win := gfxWindowInitQueue[i]
		if glwin, err := gfxNewWindow(
			win.Title(),
			win.Width(), win.Height(),
			win.Borderless(), win.Resizable(),
			win.MultiSamplingEnabled()); err != nil {
			panic(err)
		} else {
			win.Init(glwin, gfxContext)
		}
		gfxWindowInitQueue = gfxWindowInitQueue[:i]
	}
}

func gfxProcessCloseQueue() {
	for i := len(gfxWindowCloseQueue) - 1; i >= 0; i-- {
		win := gfxWindowCloseQueue[i]
		glwin := win.GLFW()
		if glwin != nil {
			glwin.MakeContextCurrent()
		}
		win.Close()
		if glwin != nil {
			glwin.Destroy()
		}
		gfxWindowCloseQueue = gfxWindowCloseQueue[:i]
	}
}

func gfxResetState() {
	gfxInitialized = false
	gfxRunning = false
	gfxWindows = make([]GlfwWindow, 0)
	gfxContext = nil
	gfxCancelFunc = nil
	gfxWindowInitQueue = make([]GlfwWindow, 0)
	gfxWindowCloseQueue = make([]GlfwWindow, 0)
	if gfxClosedChan != nil {
		close(gfxClosedChan)
	}
}

func gfxClose() {
	gfxProcessCloseQueue()
	glfw.Terminate()
	gfxResetState()
}

// Init Initialize the gfx package (and GLFW), which must be
// done before calling Run().  Must call from the main routine.
func Init() error {
	gfxStateMutex.Lock()

	if gfxInitialized {
		gfxStateMutex.Unlock()
		return nil
	}

	if err := glfw.Init(); err != nil {
		gfxStateMutex.Unlock()
		return fmt.Errorf("error initializing GLFW: %w", err)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, openGlVersionMajor)
	glfw.WindowHint(glfw.ContextVersionMinor, openGlVersionMinor)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.DoubleBuffer, glfw.True)

	gfxInitialized = true
	gfxClosedChan = make(chan struct{})
	gfxStateMutex.Unlock()

	return nil
}

// InitWindowAsync Asynchronously initialize a window to ensure
// it gets updated/rendered during the main processing loop.
// Can call at any time and from any routine.
func InitWindowAsync(window GlfwWindow) {
	gfxStateMutex.Lock()
	gfxWindowInitQueue = append(gfxWindowInitQueue, window)
	gfxWindows = append(gfxWindows, window)
	gfxStateMutex.Unlock()
}

// CloseWindowAsync Asynchronously stop updating/rendering and
// close the specified window, releasing any resources it
// consumed while in use (such as services, objects, and assets).
// Removing the primary window will also stop the main processing
// loop by canceling the context that was passed to the Run()
// function. Can call at any time and from any routine.
func CloseWindowAsync(window GlfwWindow) {
	gfxStateMutex.Lock()

	found := false
	for i := len(gfxWindows) - 1; i >= 0; i-- {
		if gfxWindows[i] == window {
			gfxWindows = append(gfxWindows[:i], gfxWindows[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		gfxStateMutex.Unlock()
		return
	}

	gfxWindowCloseQueue = append(gfxWindowCloseQueue, window)

	if !window.IsSecondary() {
		gfxCancelFunc()
	}

	gfxStateMutex.Unlock()
}

// Run Start the main processing loop, which will call Update() on
// all windows at an interval based on the target framerate but
// limited by system performance. Must call from the main routine.
func Run(ctx context.Context, cancelFunc context.CancelFunc) {
	gfxStateMutex.Lock()

	gfxContext = ctx
	gfxCancelFunc = cancelFunc

	gfxProcessInitQueue()

	if glfw.GetCurrentContext() != nil {
		if vSyncEnabled.Load() {
			glfw.SwapInterval(1)
		} else {
			glfw.SwapInterval(0)
		}
	}

	now := time.Now().UnixMicro()
	lastTick := now
	deltaTime := int64(0)
	drawInterval := int64(1000000 / targetFramerate.Load())
	gfxRunning = true
	doneChan := gfxContext.Done()
	gfxStateMutex.Unlock()

	for {
		select {
		case <-doneChan:
			gfxStateMutex.Lock()
			for _, w := range gfxWindows {
				gfxWindowCloseQueue = append(gfxWindowCloseQueue, w)
			}
			gfxClose()
			gfxStateMutex.Unlock()
			return
		default:
		}

		deltaTime = time.Now().UnixMicro() - lastTick
		if deltaTime < drawInterval {
			// This is done to give more CPU time to other, higher
			// priority tasks/processes.  Setting a lower target framerate
			// will give them more time to work, while a higher value
			// increases the potential rendering performance of the windows.
			time.Sleep(time.Microsecond * time.Duration(drawInterval-deltaTime))
			deltaTime = time.Now().UnixMicro() - lastTick
		}
		lastTick = time.Now().UnixMicro()

		gfxStateMutex.Lock()

		gfxProcessInitQueue()
		gfxProcessCloseQueue()

		glfw.PollEvents()

		for _, win := range gfxWindows {
			if win != gfxWindow {
				gfxWindow = win
				gfxWindow.GLFW().MakeContextCurrent()
				gfxWindow.Update(deltaTime)
			} else {
				gfxWindow.Update(deltaTime)
			}
		}

		gfxStateMutex.Unlock()
	}
}

// Close Destroy any remaining windows and release any
// remaining resources allocated on the graphics card
// (VRAM).  After closing, you must again call Init()
// if you need to reuse the package.  Can call from any
// routine but should only be called after initializing
// via Init().
func Close() {
	gfxStateMutex.Lock()

	if !gfxInitialized {
		gfxStateMutex.Unlock()
		return
	} else if !gfxRunning {
		gfxClose()
		gfxStateMutex.Unlock()
		return
	}

	gfxCancelFunc()
	gfxStateMutex.Unlock()
	<-gfxClosedChan
}
