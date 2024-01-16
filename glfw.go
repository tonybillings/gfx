package gfx

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"sync"
	"sync/atomic"
)

const (
	openGlVersionMajor     = 4
	openGlVersionMinor     = 1
	defaultWinWidth        = 1024
	defaultWinHeight       = 768
	defaultTargetFramerate = 30
)

var (
	glfwInitialized atomic.Bool
	newWindowMutex  sync.Mutex
	windowCount     atomic.Int32
)

func enableVsync() {
	glfw.SwapInterval(1)
}

func newGlfwWindow(title string, width, height int32) (*glfw.Window, error) {
	newWindowMutex.Lock()
	defer newWindowMutex.Unlock()

	if !glfwInitialized.Load() {
		return nil, fmt.Errorf("glfw not initialized: must call gfx.Init() first")
	}

	w := defaultWinWidth
	h := defaultWinHeight
	if width != 0 {
		w = int(width)
	}
	if height != 0 {
		h = int(height)
	}

	win, err := glfw.CreateWindow(w, h, title, nil, nil)
	if err != nil {
		return nil, err
	}

	monitor := glfw.GetPrimaryMonitor()
	monMode := monitor.GetVideoMode()
	xPos := (monMode.Width - w) / 2
	yPos := (monMode.Height - h) / 2
	win.SetPos(xPos, yPos)

	win.MakeContextCurrent()

	if err = gl.Init(); err != nil {
		win.SetShouldClose(true)
		return nil, fmt.Errorf("error initializing gl: %w", err)
	}

	enableVsync()
	windowCount.Add(1)

	return win, nil
}

func Init() error {
	if glfwInitialized.Load() {
		return nil
	}

	if err := glfw.Init(); err != nil {
		return fmt.Errorf("error initializing glfw: %w", err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, openGlVersionMajor)
	glfw.WindowHint(glfw.ContextVersionMinor, openGlVersionMinor)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.DoubleBuffer, glfw.True)

	glfwInitialized.Store(true)
	return nil
}
