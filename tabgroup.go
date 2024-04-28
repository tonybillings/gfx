package gfx

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"math"
)

const (
	defaultTabGroupName = "TabGroup"
)

/******************************************************************************
 TabGroup
******************************************************************************/

type TabGroup struct {
	WindowObjectBase

	tabIdx      int
	tabAction   TabAction
	navKeys     NavKeys
	keyHandlers []*KeyEventHandler

	tranTarget      WindowObject
	tranQuad        *Shape2D
	tranQuadShowing bool
	tranQuadHiding  bool
	tranQuadOpacity float64
	tranSpeed       float64
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (g *TabGroup) Init() (ok bool) {
	if g.Initialized() {
		return true
	}

	for _, c := range g.children {
		c.SetWindow(g.window)
	}

	if !g.WindowObjectBase.Init() {
		return false
	}

	g.initTransitionQuad()
	g.updateNav()
	g.Activate(0)

	return true
}

func (g *TabGroup) Close() {
	if !g.Initialized() {
		return
	}

	for _, h := range g.keyHandlers {
		g.window.RemoveKeyEventHandler(h)
	}

	g.WindowObjectBase.Close()
}

/******************************************************************************
 DrawableObject Implementation
******************************************************************************/

func (g *TabGroup) Draw(deltaTime int64) (ok bool) {
	if !g.visible.Load() {
		return false
	}

	g.stateMutex.Lock()

	const tranSpeedMultiplier = 0.000001 // allow speed to be specified using larger numbers

	if g.tranQuadShowing {
		g.tranQuadOpacity += g.tranSpeed * float64(deltaTime) * tranSpeedMultiplier
		g.tranQuad.SetOpacity(uint8(math.Min(255.0, g.tranQuadOpacity*255.0)))
		if g.tranQuad.Opacity() == 255 {
			g.tranQuadShowing = false
			g.tranQuadHiding = true
			g.tranQuadOpacity = 1
			for _, c := range g.Children() {
				g.deactivate(c)
			}
			g.activate(g.tranTarget)
		}
		g.stateMutex.Unlock()
		g.drawChildren(deltaTime)
		g.drawTransitionQuad(deltaTime)
	} else if g.tranQuadHiding {
		g.tranQuadOpacity -= g.tranSpeed * float64(deltaTime) * tranSpeedMultiplier
		g.tranQuad.SetOpacity(uint8(math.Max(0, g.tranQuadOpacity*255.0)))
		if g.tranQuad.Opacity() == 0 {
			g.tranQuadHiding = false
			g.tranQuadOpacity = 0
		}
		g.stateMutex.Unlock()
		g.drawChildren(deltaTime)
		g.drawTransitionQuad(deltaTime)
	} else {
		g.stateMutex.Unlock()
		g.drawChildren(deltaTime)
	}

	return true
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (g *TabGroup) Resize(newWidth, newHeight int) {
	g.tranQuad.Resize(newWidth, newHeight)
	g.WindowObjectBase.Resize(newWidth, newHeight)
}

/******************************************************************************
 TabGroup Functions
******************************************************************************/

func (g *TabGroup) initTransitionQuad() {
	g.tranQuad = NewQuad()
	g.tranQuad.
		SetColor(g.window.ClearColor()).
		SetOpacity(0).
		SetMaintainAspectRatio(false)
	g.tranQuad.SetParent(g)
	g.tranQuad.SetWindow(g.window)
	g.tranQuad.Init()
}

func (g *TabGroup) drawTransitionQuad(deltaTime int64) {
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	g.tranQuad.Draw(deltaTime)
	gl.Disable(gl.BLEND)
}

func (g *TabGroup) updateNav() {
	g.stateMutex.Lock()
	switch g.navKeys {
	case TabAndArrowKeys:
		keyHandler := g.window.AddKeyEventHandler(glfw.KeyTab, glfw.Press, func(w *Window, key glfw.Key, action glfw.Action) {
			g.Next()
		})
		g.keyHandlers = append(g.keyHandlers, keyHandler)

		keyHandler = g.window.AddKeyEventHandler(glfw.KeyRight, glfw.Press, func(w *Window, key glfw.Key, action glfw.Action) {
			g.Next()
		})
		g.keyHandlers = append(g.keyHandlers, keyHandler)

		keyHandler = g.window.AddKeyEventHandler(glfw.KeyLeft, glfw.Press, func(w *Window, key glfw.Key, action glfw.Action) {
			g.Previous()
		})
		g.keyHandlers = append(g.keyHandlers, keyHandler)
	case TabKeyOnly:
		keyHandler := g.window.AddKeyEventHandler(glfw.KeyTab, glfw.Press, func(w *Window, key glfw.Key, action glfw.Action) {
			g.Next()
		})
		g.keyHandlers = append(g.keyHandlers, keyHandler)
	case ArrowKeysOnly:
		keyHandler := g.window.AddKeyEventHandler(glfw.KeyRight, glfw.Press, func(w *Window, key glfw.Key, action glfw.Action) {
			g.Next()
		})
		g.keyHandlers = append(g.keyHandlers, keyHandler)

		keyHandler = g.window.AddKeyEventHandler(glfw.KeyLeft, glfw.Press, func(w *Window, key glfw.Key, action glfw.Action) {
			g.Previous()
		})
		g.keyHandlers = append(g.keyHandlers, keyHandler)
	}
	g.stateMutex.Unlock()
}

func (g *TabGroup) deactivate(child WindowObject) {
	switch g.tabAction {
	case HideOnly:
		child.SetVisibility(false)
	case DisableOnly:
		child.SetEnabled(false)
	case HideAndDisable:
		child.SetVisibility(false)
		child.SetEnabled(false)
	}

	for _, c := range child.Children() {
		g.deactivate(c)
	}
}

func (g *TabGroup) activate(child WindowObject) {
	switch g.tabAction {
	case HideOnly:
		child.SetVisibility(true)
	case DisableOnly:
		child.SetEnabled(true)
	case HideAndDisable:
		child.SetEnabled(true)
		child.SetVisibility(true)
	}

	for _, c := range child.Children() {
		g.activate(c)
	}
}

func (g *TabGroup) IndexOf(name string) int {
	g.stateMutex.Lock()
	idx := -1
	for i, c := range g.Children() {
		if c.Name() == name {
			idx = i
			break
		}
	}
	g.stateMutex.Unlock()
	return idx
}

func (g *TabGroup) ActiveIndex() int {
	g.stateMutex.Lock()
	idx := g.tabIdx
	g.stateMutex.Unlock()
	return idx
}

func (g *TabGroup) Activate(index int) {
	g.stateMutex.Lock()

	if g.tranQuadShowing || g.tranQuadHiding {
		g.stateMutex.Unlock()
		return
	}

	if index < 0 || index >= len(g.children) {
		g.stateMutex.Unlock()
		return
	}

	for _, c := range g.Children() {
		g.deactivate(c)
	}

	g.activate(g.children[index])
	g.tabIdx = index
	g.Window().glwin.SetCursor(glfw.CreateStandardCursor(glfw.ArrowCursor))

	g.stateMutex.Unlock()
}

func (g *TabGroup) Next() {
	g.stateMutex.Lock()

	if g.tranQuadShowing || g.tranQuadHiding {
		g.stateMutex.Unlock()
		return
	}

	if len(g.children) < 2 {
		g.stateMutex.Unlock()
		return
	}

	g.tabIdx = (g.tabIdx + 1) % len(g.children)
	g.stateMutex.Unlock()
	g.Activate(g.tabIdx)
}

func (g *TabGroup) Previous() {
	g.stateMutex.Lock()

	if g.tranQuadShowing || g.tranQuadHiding {
		g.stateMutex.Unlock()
		return
	}

	if len(g.children) < 2 {
		g.stateMutex.Unlock()
		return
	}

	if g.tabIdx == 0 {
		g.tabIdx = len(g.children) - 1
	} else {
		g.tabIdx = g.tabIdx - 1
	}

	g.stateMutex.Unlock()
	g.Activate(g.tabIdx)
}

func (g *TabGroup) Transition(index int, speed ...float64) {
	g.stateMutex.Lock()

	if g.tranQuadShowing || g.tranQuadHiding {
		g.stateMutex.Unlock()
		return
	}

	if index < 0 || index >= len(g.children) {
		g.stateMutex.Unlock()
		return
	}

	if len(speed) > 0 {
		g.tranSpeed = speed[0]
	} else {
		g.tranSpeed = 5.0
	}

	g.tranTarget = g.Children()[index]
	g.tabIdx = index
	g.tranQuadShowing = true
	g.stateMutex.Unlock()
}

func (g *TabGroup) TransitionNext(speed ...float64) {
	idx := (g.ActiveIndex() + 1) % len(g.Children())
	g.Transition(idx, speed...)
}

func (g *TabGroup) TransitionPrevious(speed ...float64) {
	idx := g.ActiveIndex() - 1
	if idx == -1 {
		idx = len(g.Children()) - 1
	}
	g.Transition(idx, speed...)
}

func (g *TabGroup) TabAction() TabAction {
	g.stateMutex.Lock()
	action := g.tabAction
	g.stateMutex.Unlock()
	return action
}

func (g *TabGroup) SetTabAction(action TabAction) *TabGroup {
	g.stateMutex.Lock()
	g.tabAction = action
	g.stateMutex.Unlock()
	return g
}

func (g *TabGroup) NavKeys() NavKeys {
	g.stateMutex.Lock()
	keys := g.navKeys
	g.stateMutex.Unlock()
	return keys
}

func (g *TabGroup) SetNavKeys(keys NavKeys) *TabGroup {
	g.stateMutex.Lock()
	for _, h := range g.keyHandlers {
		g.window.RemoveKeyEventHandler(h)
	}

	g.navKeys = keys
	g.stateMutex.Unlock()
	g.updateNav()

	return g
}

/******************************************************************************
 New TabGroup Function
******************************************************************************/

func NewTabGroup(objects ...WindowObject) *TabGroup {
	tg := &TabGroup{
		WindowObjectBase: *NewWindowObject(),
		tabAction:        HideAndDisable,
		navKeys:          TabAndArrowKeys,
	}

	tg.SetName(defaultTabGroupName)

	for _, o := range objects {
		tg.AddChild(o)
	}

	return tg
}

/******************************************************************************
 TabGroup Enums
******************************************************************************/

type TabAction int

const (
	HideOnly TabAction = iota
	DisableOnly
	HideAndDisable
)

type NavKeys int

const (
	TabAndArrowKeys NavKeys = iota
	TabKeyOnly
	ArrowKeysOnly
)
