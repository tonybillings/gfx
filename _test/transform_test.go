package _test

import (
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"testing"
	"tonysoft.com/gfx"
)

type mockTransform struct {
	gfx.ObjectTransform
}

func (m *mockTransform) WorldOrigin() mgl32.Vec3 {
	return mgl32.Vec3{1, 2, 3}
}

func (m *mockTransform) WorldPosition() mgl32.Vec3 {
	return mgl32.Vec3{4, 5, 6}
}

func (m *mockTransform) WorldRotation() mgl32.Vec3 {
	return mgl32.Vec3{0.1, 0.2, 0.3}
}

func (m *mockTransform) WorldScale() mgl32.Vec3 {
	return mgl32.Vec3{1.5, 1.5, 1.5}
}

func TestSetAndGetPositions(t *testing.T) {
	ot := &gfx.ObjectTransform{}
	newPos := mgl32.Vec3{1, 2, 3}
	ot.SetPosition(newPos)
	if ot.Position() != newPos {
		t.Errorf("unexpected position: expected %v, got: %v", newPos, ot.Position())
	}
}

func TestWorldPositionWithoutParent(t *testing.T) {
	ot := &gfx.ObjectTransform{}
	ot.SetPosition(mgl32.Vec3{10, 20, 30})
	expected := mgl32.Vec3{10, 20, 30}
	worldPos := ot.WorldPosition()
	if worldPos != expected {
		t.Errorf("unexpected world position: expected %v, got: %v", expected, worldPos)
	}
}

func TestWorldPositionWithParent(t *testing.T) {
	parent := &mockTransform{}
	child := &gfx.ObjectTransform{}
	child.SetPosition(mgl32.Vec3{1, 1, 1})
	child.SetParentTransform(parent)
	expected := mgl32.Vec3{5, 6, 7}
	worldPos := child.WorldPosition()
	if worldPos != expected {
		t.Errorf("unexpected world position: expected %v, got: %v", expected, worldPos)
	}
}

func TestWorldMatrix(t *testing.T) {
	ot := &gfx.ObjectTransform{}
	ot.SetPosition(mgl32.Vec3{1, 2, 3})
	ot.SetScale(mgl32.Vec3{1, 1, 1})
	expected := mgl32.Translate3D(1, 2, 3)
	worldMatrix := ot.WorldMatrix()
	if !worldMatrix.ApproxEqualThreshold(expected, 1e-4) {
		t.Errorf("unexpected world matrix: expected %v, got %v", expected, worldMatrix)
	}
}

func TestSetAndGetRotation(t *testing.T) {
	ot := &gfx.ObjectTransform{}
	newRot := mgl32.Vec3{0.1, 0.2, 0.3}
	ot.SetRotation(newRot)
	if ot.Rotation() != newRot {
		t.Errorf("unexpected rotation: expected %v, got: %v", newRot, ot.Rotation())
	}
}

func TestWorldRotationWithParent(t *testing.T) {
	parent := &mockTransform{}
	child := &gfx.ObjectTransform{}
	child.SetRotation(mgl32.Vec3{0.1, 0.1, 0.1})
	child.SetParentTransform(parent)
	expected := mgl32.Vec3{0.2, 0.3, 0.4}
	worldRot := child.WorldRotation()
	if worldRot != expected {
		t.Errorf("unexpected world rotation: expected %v, got: %v", expected, worldRot)
	}
}

func TestSetAndGetScale(t *testing.T) {
	ot := &gfx.ObjectTransform{}
	newScale := mgl32.Vec3{2, 2, 2}
	ot.SetScale(newScale)
	if ot.Scale() != newScale {
		t.Errorf("unexpected scale: expected %v, got: %v", newScale, ot.Scale())
	}
}

func TestWorldScaleWithParent(t *testing.T) {
	parent := &mockTransform{}
	child := &gfx.ObjectTransform{}
	child.SetScale(mgl32.Vec3{2, 2, 2})
	child.SetParentTransform(parent)
	expected := mgl32.Vec3{3, 3, 3}
	worldScale := child.WorldScale()
	if worldScale != expected {
		t.Errorf("unexpected world scale: expected %v, got: %v", expected, worldScale)
	}
}

func TestWorldMatrixWithRotationAndScale(t *testing.T) {
	ot := &gfx.ObjectTransform{}
	ot.SetPosition(mgl32.Vec3{1, 2, 3})
	ot.SetRotation(mgl32.Vec3{0, mgl32.DegToRad(90), 0})
	ot.SetScale(mgl32.Vec3{2, 2, 2})

	tranMat := mgl32.Translate3D(1, 2, 3)
	rotMat := mgl32.HomogRotate3DY(mgl32.DegToRad(90))
	scaleMat := mgl32.Scale3D(2, 2, 2)

	expected := tranMat.Mul4(rotMat).Mul4(scaleMat)
	worldMatrix := ot.WorldMatrix()

	if !worldMatrix.ApproxEqualThreshold(expected, 1e-4) {
		t.Errorf("unexpected world matrix: expected %v, got %v", expected, worldMatrix)
	}
}

func TestQuaternionRotation(t *testing.T) {
	isEqual := func(a, b mgl32.Mat4, threshold float32) bool {
		for i := 0; i < 16; i++ {
			if math.Abs(float64(a[i]-b[i])) > float64(threshold) {
				return false
			}
		}
		return true
	}

	ot := &gfx.ObjectTransform{}

	quat := mgl32.QuatRotate(mgl32.DegToRad(90), mgl32.Vec3{0, 1, 0})
	ot.SetRotationQuat(quat)

	expectedRotMat := mgl32.HomogRotate3DY(mgl32.DegToRad(90))
	actualRotMat := ot.RotationQuat().Mat4()

	if !isEqual(expectedRotMat, actualRotMat, 0.000001) {
		t.Errorf("unexpected rotation matrix: expected %v, got %v", expectedRotMat, actualRotMat)
	}
}
