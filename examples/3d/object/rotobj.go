package object

import (
	"tonysoft.com/gfx"
)

type RotatingObject struct {
	gfx.WindowObjectBase // embed this for default Init/Update/Draw/Close/... behavior
	x, y                 float32
}

func (o *RotatingObject) rotate(deltaTime int64) {
	// Leverage deltaTime for a framerate-independent rotation speed
	o.x += .000001 * float32(deltaTime)
	o.y += .0000005 * float32(deltaTime)
	o.SetRotationX(o.x)
	o.SetRotationY(o.y)
}

// Update We override the default implementation so we can include our
// rotating logic, which gets called once per game "tick".
func (o *RotatingObject) Update(deltaTime int64) (ok bool) {
	// Ensure the object should be updated and if so, update the children and continue
	if !o.WindowObjectBase.Update(deltaTime) {
		return false
	}

	o.rotate(deltaTime) // will rotate "this" object, which affects all children that call WorldRotation()

	return true // tell the parent, if one exists, that all went well...in case it cares
}

func NewRotatingObject(obj gfx.WindowObject) *RotatingObject {
	o := &RotatingObject{
		WindowObjectBase: *gfx.NewObject(nil), // get good defaults
	}

	o.AddChild(obj) // add the object as a child, assuming it will call WorldRotation()

	return o
}
