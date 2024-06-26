package gfx

import (
	"strconv"
)

const (
	defaultFpsUpdateIntervalMilli = 250
)

/******************************************************************************
 FpsCounter
******************************************************************************/

// FpsCounter When added to a Window, will display the current
// frames per second, calculating the metric at the update
// interval specified (defaults to 250 ms).  Note that this
// object cannot be disabled or hidden, only added/removed from
// a Window.
type FpsCounter struct {
	WindowObjectBase

	text *Label

	interval int64
	sum      int64
	count    int
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (c *FpsCounter) Init() (ok bool) {
	if c.Initialized() {
		return true
	}

	c.sum = 0
	c.count = 0

	if ok = c.text.Init(); !ok {
		return
	}

	c.initialized.Store(true)

	return true
}

func (c *FpsCounter) Update(deltaTime int64) (ok bool) {
	if !c.initialized.Load() {
		return false
	}

	c.sum += deltaTime
	c.count++

	if c.sum >= c.interval {
		fps := int((1_000_000 / (float64(c.sum) / float64(c.count))) + 0.5)
		c.text.SetText(strconv.Itoa(fps))
		c.sum = 0
		c.count = 0
		return c.text.Update(deltaTime)
	}

	return true
}

func (c *FpsCounter) Draw(deltaTime int64) bool {
	if !c.initialized.Load() {
		return false
	}

	return c.text.Draw(deltaTime)
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (c *FpsCounter) SetMaintainAspectRatio(maintainAspectRatio bool) WindowObject {
	c.maintainAspectRatio = maintainAspectRatio
	c.text.SetMaintainAspectRatio(maintainAspectRatio)
	return c
}

func (c *FpsCounter) SetWindow(window *Window) WindowObject {
	c.WindowObjectBase.SetWindow(window)
	c.text.SetWindow(window)
	return c
}

/******************************************************************************
 Resizer Implementation
******************************************************************************/

func (c *FpsCounter) Resize(newWidth, newHeight int) {
	c.WindowObjectBase.Resize(newWidth, newHeight)
	c.text.Resize(newWidth, newHeight)
}

/******************************************************************************
 FpsCounter Functions
******************************************************************************/

func (c *FpsCounter) defaultLayout() {
	c.text.SetColor(Magenta)
	c.text.SetScaleX(.1)
	c.text.SetFontSize(0.05)
	c.text.SetAnchor(TopCenter)
	c.text.SetCacheEnabled(true)
	c.SetMaintainAspectRatio(false)
}

func (c *FpsCounter) SetUpdateInterval(milliseconds int) *FpsCounter {
	c.interval = int64(milliseconds) * 1000
	return c
}

func (c *FpsCounter) Label() *Label {
	return c.text
}

/******************************************************************************
 New FpsCounter Function
******************************************************************************/

func NewFpsCounter(updateIntervalMilli ...int) *FpsCounter {
	interval := defaultFpsUpdateIntervalMilli
	if len(updateIntervalMilli) > 0 {
		interval = updateIntervalMilli[0]
	}

	f := &FpsCounter{
		WindowObjectBase: *NewWindowObject(),
		text:             NewLabel(),
	}

	f.defaultLayout()
	f.SetUpdateInterval(interval)

	return f
}
