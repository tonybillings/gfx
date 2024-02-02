package gfx

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	defaultTabGroupName = "TabGroup"
)

/******************************************************************************
 Enums
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

/******************************************************************************
 TabGroup
******************************************************************************/

type TabGroup struct {
	WindowObjectBase
	tabIdx      int
	tabAction   TabAction
	navKeys     NavKeys
	keyHandlers []*KeyEventHandler
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (g *TabGroup) Init(window *Window) (ok bool) {
	if !g.WindowObjectBase.Init(window) {
		return false
	}

	g.updateNav()
	g.Activate(0)

	g.initialized.Store(true)
	return true
}

func (g *TabGroup) Close() {
	if g.initialized.Load() {
		g.WindowObjectBase.Close()

		for _, h := range g.keyHandlers {
			g.window.RemoveKeyEventHandler(h)
		}
	}
}

/******************************************************************************
 TabGroup Functions
******************************************************************************/

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

func (g *TabGroup) Activate(index int) {
	if len(g.children) == 0 {
		return
	}

	for _, c := range g.Children() {
		g.deactivate(c)
	}

	g.activate(g.children[index])
	g.Window().glwin.SetCursor(glfw.CreateStandardCursor(glfw.ArrowCursor))
}

func (g *TabGroup) Next() {
	if len(g.children) < 2 {
		return
	}

	g.tabIdx = (g.tabIdx + 1) % len(g.children)
	g.Activate(g.tabIdx)
}

func (g *TabGroup) Previous() {
	if len(g.children) < 2 {
		return
	}

	if g.tabIdx == 0 {
		g.tabIdx = len(g.children) - 1
	} else {
		g.tabIdx = g.tabIdx - 1
	}

	g.Activate(g.tabIdx)
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
		WindowObjectBase: *NewObject(nil),
		tabAction:        HideAndDisable,
		navKeys:          TabAndArrowKeys,
	}

	tg.SetName(defaultTabGroupName)

	if len(objects) > 0 {
		for _, o := range objects {
			tg.AddChild(o)
		}
	}

	return tg
}
