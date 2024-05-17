package gfx

import (
	"github.com/go-gl/mathgl/mgl32"
	"sync"
)

/******************************************************************************
 Transform
******************************************************************************/

// Transform Drawable objects are expected to have a transform that can be
// used to position the object in world space and within the context of a
// hierarchical parent/child structure.  This type defines that contract.
type Transform interface {
	ParentTransform() Transform
	SetParentTransform(Transform) Transform

	Origin() mgl32.Vec3
	WorldOrigin() mgl32.Vec3
	SetOrigin(mgl32.Vec3) Transform

	Position() mgl32.Vec3
	WorldPosition() mgl32.Vec3
	SetPosition(mgl32.Vec3) Transform
	SetPositionX(float32) Transform
	SetPositionY(float32) Transform
	SetPositionZ(float32) Transform

	Rotation() mgl32.Vec3
	WorldRotation() mgl32.Vec3
	SetRotation(mgl32.Vec3) Transform
	SetRotationX(float32) Transform
	SetRotationY(float32) Transform
	SetRotationZ(float32) Transform

	Scale() mgl32.Vec3
	WorldScale() mgl32.Vec3
	SetScale(mgl32.Vec3) Transform
	SetScaleX(float32) Transform
	SetScaleY(float32) Transform
	SetScaleZ(float32) Transform

	RotationQuat() mgl32.Quat
	SetRotationQuat(mgl32.Quat) Transform
	WorldMatrix() mgl32.Mat4
}

/******************************************************************************
 ObjectTransform
******************************************************************************/

type ObjectTransform struct {
	parentTransform Transform

	origin   mgl32.Vec3
	position mgl32.Vec3
	rotation mgl32.Vec3
	scale    mgl32.Vec3

	rotationQuat    mgl32.Quat
	rotationQuatSet bool

	stateMutex sync.Mutex
}

func NewObjectTransform() *ObjectTransform {
	return &ObjectTransform{
		origin:   [3]float32{0, 0, 0},
		position: [3]float32{0, 0, 0},
		rotation: [3]float32{0, 0, 0},
		scale:    [3]float32{1, 1, 1},
	}
}

/******************************************************************************
 Transform Implementation
******************************************************************************/

func (t *ObjectTransform) ParentTransform() Transform {
	t.stateMutex.Lock()
	parent := t.parentTransform
	t.stateMutex.Unlock()
	return parent
}

func (t *ObjectTransform) SetParentTransform(parent Transform) Transform {
	t.stateMutex.Lock()
	t.parentTransform = parent
	t.stateMutex.Unlock()
	return t
}

func (t *ObjectTransform) Origin() mgl32.Vec3 {
	t.stateMutex.Lock()
	origin := t.origin
	t.stateMutex.Unlock()
	return origin
}

func (t *ObjectTransform) WorldOrigin() mgl32.Vec3 {
	t.stateMutex.Lock()
	origin := t.origin
	t.stateMutex.Unlock()

	if p := t.ParentTransform(); p != nil {
		parentOrigin := p.WorldOrigin()
		origin = origin.Add(parentOrigin)
	}

	return origin
}

func (t *ObjectTransform) SetOrigin(origin mgl32.Vec3) Transform {
	t.stateMutex.Lock()
	t.origin = origin
	t.stateMutex.Unlock()
	return t
}

func (t *ObjectTransform) Position() mgl32.Vec3 {
	t.stateMutex.Lock()
	position := t.position
	t.stateMutex.Unlock()
	return position
}

func (t *ObjectTransform) WorldPosition() mgl32.Vec3 {
	t.stateMutex.Lock()
	position := mgl32.Vec3{t.position[0], t.position[1], t.position[2]}
	t.stateMutex.Unlock()

	if p := t.ParentTransform(); p != nil {
		parentPos := p.WorldPosition()
		position = position.Add(parentPos)
	}

	return position
}

func (t *ObjectTransform) SetPosition(position mgl32.Vec3) Transform {
	t.stateMutex.Lock()
	t.position = position
	t.stateMutex.Unlock()
	return t
}

func (t *ObjectTransform) SetPositionX(x float32) Transform {
	t.stateMutex.Lock()
	t.position[0] = x
	t.stateMutex.Unlock()
	return t
}

func (t *ObjectTransform) SetPositionY(y float32) Transform {
	t.stateMutex.Lock()
	t.position[1] = y
	t.stateMutex.Unlock()
	return t
}

func (t *ObjectTransform) SetPositionZ(z float32) Transform {
	t.stateMutex.Lock()
	t.position[2] = z
	t.stateMutex.Unlock()
	return t
}

func (t *ObjectTransform) Rotation() mgl32.Vec3 {
	t.stateMutex.Lock()
	rotation := t.rotation
	t.stateMutex.Unlock()
	return rotation
}

func (t *ObjectTransform) WorldRotation() mgl32.Vec3 {
	t.stateMutex.Lock()
	rotation := t.rotation
	t.stateMutex.Unlock()

	if p := t.ParentTransform(); p != nil {
		parentRot := p.WorldRotation()
		rotation = rotation.Add(parentRot)
	}

	return rotation
}

func (t *ObjectTransform) SetRotation(rotation mgl32.Vec3) Transform {
	t.stateMutex.Lock()
	t.rotation = rotation
	t.stateMutex.Unlock()
	return t
}

func (t *ObjectTransform) SetRotationX(x float32) Transform {
	t.stateMutex.Lock()
	t.rotation[0] = x
	t.stateMutex.Unlock()
	return t
}

func (t *ObjectTransform) SetRotationY(y float32) Transform {
	t.stateMutex.Lock()
	t.rotation[1] = y
	t.stateMutex.Unlock()
	return t
}

func (t *ObjectTransform) SetRotationZ(z float32) Transform {
	t.stateMutex.Lock()
	t.rotation[2] = z
	t.stateMutex.Unlock()
	return t
}

func (t *ObjectTransform) Scale() mgl32.Vec3 {
	t.stateMutex.Lock()
	scale := t.scale
	t.stateMutex.Unlock()
	return scale
}

func (t *ObjectTransform) WorldScale() mgl32.Vec3 {
	t.stateMutex.Lock()
	scale := t.scale
	t.stateMutex.Unlock()

	if p := t.ParentTransform(); p != nil {
		parentScale := p.WorldScale()
		scale = mgl32.Vec3{t.scale[0] * parentScale.X(), t.scale[1] * parentScale.Y(), t.scale[2] * parentScale.Z()}
	}

	return scale
}

func (t *ObjectTransform) SetScale(scale mgl32.Vec3) Transform {
	t.stateMutex.Lock()
	t.scale = scale
	t.stateMutex.Unlock()
	return t
}

func (t *ObjectTransform) SetScaleX(x float32) Transform {
	t.stateMutex.Lock()
	t.scale[0] = x
	t.stateMutex.Unlock()
	return t
}

func (t *ObjectTransform) SetScaleY(y float32) Transform {
	t.stateMutex.Lock()
	t.scale[1] = y
	t.stateMutex.Unlock()
	return t
}

func (t *ObjectTransform) SetScaleZ(z float32) Transform {
	t.stateMutex.Lock()
	t.scale[2] = z
	t.stateMutex.Unlock()
	return t
}

func (t *ObjectTransform) RotationQuat() mgl32.Quat {
	if !t.rotationQuatSet {
		return mgl32.QuatIdent()
	}

	t.stateMutex.Lock()
	rotation := t.rotationQuat
	t.stateMutex.Unlock()
	return rotation
}

func (t *ObjectTransform) SetRotationQuat(rotation mgl32.Quat) Transform {
	t.stateMutex.Lock()
	t.rotationQuat = rotation
	t.rotationQuatSet = true
	t.stateMutex.Unlock()
	return t
}

func (t *ObjectTransform) WorldMatrix() mgl32.Mat4 {
	tran := t.WorldPosition()
	rot := t.WorldRotation()
	scale := t.WorldScale()

	tranMat := mgl32.Translate3D(tran.X(), tran.Y(), tran.Z())

	var rotMat mgl32.Mat4
	t.stateMutex.Lock()
	if t.rotationQuatSet {
		rotMat = t.rotationQuat.Mat4()
		t.stateMutex.Unlock()
	} else {
		t.stateMutex.Unlock()
		rotLen := rot.Len()
		if rotLen > 0.0001 {
			rotMat = mgl32.HomogRotate3D(rotLen, rot.Normalize())
		} else {
			rotMat = mgl32.Ident4()
		}
	}

	scaleMat := mgl32.Scale3D(scale.X(), scale.Y(), scale.Z())

	return tranMat.Mul4(rotMat.Mul4(scaleMat))
}

/******************************************************************************
 ShaderTransform
******************************************************************************/

type ShaderTransform struct {
	Origin   mgl32.Vec4
	Position mgl32.Vec4
	Rotation mgl32.Vec4
	Scale    mgl32.Vec4
}

/******************************************************************************
 shaderTransform
******************************************************************************/

type shaderTransform struct {
	Transform *ShaderTransform
}

func (s *shaderTransform) setOrigin(value mgl32.Vec3) {
	s.Transform.Origin = mgl32.Vec4{value[0], value[1], value[2], 1}
}

func (s *shaderTransform) setPosition(value mgl32.Vec3) {
	s.Transform.Position = mgl32.Vec4{value[0], value[1], value[2], 1}
}

func (s *shaderTransform) setRotation(value mgl32.Vec3) {
	s.Transform.Rotation = mgl32.Vec4{value[0], value[1], value[2], 1}
}

func (s *shaderTransform) setScale(value mgl32.Vec3) {
	s.Transform.Scale = mgl32.Vec4{value[0], value[1], value[2], 1}
}
