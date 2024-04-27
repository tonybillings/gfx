package gfx

import (
	_ "image/png"
	"sync/atomic"
)

/******************************************************************************
 Initer
******************************************************************************/

type Initer interface {
	Init() bool
}

/******************************************************************************
 Closer
******************************************************************************/

type Closer interface {
	Close()
}

/******************************************************************************
 Object
******************************************************************************/

type Object interface {
	Initer
	Update(int64) bool
	Closer

	Initialized() bool
	Name() string
	SetName(string) Object

	Tag() any
	SetTag(any) Object

	Enabled() bool
	SetEnabled(bool) Object
}

/******************************************************************************
 ObjectBase
******************************************************************************/

type ObjectBase struct {
	initialized atomic.Bool
	name        atomic.Pointer[string]
	tag         atomic.Value
	enabled     atomic.Bool
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (o *ObjectBase) Init() (ok bool) {
	o.initialized.Store(true)
	return true
}

func (o *ObjectBase) Update(_ int64) (ok bool) {
	return o.enabled.Load() && o.initialized.Load()
}

func (o *ObjectBase) Close() {
	o.initialized.Store(false)
}

func (o *ObjectBase) Initialized() bool {
	return o.initialized.Load()
}

func (o *ObjectBase) Name() string {
	if n := o.name.Load(); n != nil {
		return *n
	}
	return ""
}

func (o *ObjectBase) SetName(name string) Object {
	o.name.Store(&name)
	return o
}

func (o *ObjectBase) Tag() any {
	return o.tag.Load()
}

func (o *ObjectBase) SetTag(value any) Object {
	o.tag.Store(value)
	return o
}

func (o *ObjectBase) Enabled() bool {
	return o.enabled.Load()
}

func (o *ObjectBase) SetEnabled(enabled bool) Object {
	o.enabled.Store(enabled)
	return o
}

/******************************************************************************
 DrawableObject
******************************************************************************/

type DrawableObject interface {
	Object

	Draw(int64) bool

	Visible() bool
	SetVisibility(bool) DrawableObject
}

/******************************************************************************
 DrawableObjectBase
******************************************************************************/

type DrawableObjectBase struct {
	ObjectBase
	visible atomic.Bool
}

/******************************************************************************
 DrawableObject Implementation
******************************************************************************/

func (o *DrawableObjectBase) Draw(_ int64) (ok bool) {
	return o.visible.Load() && o.initialized.Load()
}

func (o *DrawableObjectBase) Visible() bool {
	return o.visible.Load()
}

func (o *DrawableObjectBase) SetVisibility(visible bool) DrawableObject {
	o.visible.Store(visible)
	return o
}
