package _test

import (
	"github.com/tonybillings/gfx/obj"
	"testing"
)

var objFile = `
# a comment that should be ignored

v 1 2 3
v 4.0 5.0 6.0
v 7 8.0 9
v -1 -2 -3
v -4.0 -5.0 -6.0
v -7 -8.0 -9

vn 10 20 30
vn 40.0 50.0 60.0
vn 70 80.0 90
vn -10 -20 -30
vn -40.0 -50.0 -60.0
vn -70 -80.0 -90

vt 111 222 333
vt 444 555 666
vt -111 -222 -333
vt -444 -555 -666

o FubarModel001

g FubarMesh001
usemtl FubarMat001
f 1/1/1 2/2/2 3/3/3
f 4/4/4 5/3/5 6/2/6

g FubarMesh002
usemtl FubarMat001
f 1/1/1 2/2/1 3/3/1
usemtl FubarMat002
f 3/3/1 2/2/1 4/4/1
usemtl FubarMat001
f 1/1/1 2/2/1 3/3/1
`

func TestOBJLoading(t *testing.T) {
	model := obj.NewModel("TestModel", objFile)
	model.Load()

	if len(model.Meshes()) != 2 {
		t.Errorf("unexpected mesh count: expected %d, got %d", 2, len(model.Meshes()))
	}

	mesh1 := model.Meshes()[0]
	if mesh1.Name() != "FubarMesh001" {
		t.Errorf("unexpected mesh name: expected %s, got %s", "FubarMesh001", mesh1.Name())
	}

	mesh2 := model.Meshes()[1]
	if mesh2.Name() != "FubarMesh002" {
		t.Errorf("unexpected mesh name: expected %s, got %s", "FubarMesh002", mesh2.Name())
	}

	if len(mesh1.Faces()) != 2 {
		t.Errorf("unexpected face count: expected %d, got %d", 2, len(mesh1.Faces()))
	}

	if len(mesh2.Faces()) != 3 {
		t.Errorf("unexpected face count: expected %d, got %d", 3, len(mesh2.Faces()))
	}

	face1 := mesh1.Faces()[0]
	face2 := mesh1.Faces()[1]

	vIdx1 := face1.VertexIndices()[0]
	vIdx2 := face1.VertexIndices()[1]
	vIdx3 := face1.VertexIndices()[2]
	v1 := int(model.Vertices()[vIdx1*3])
	v2 := int(model.Vertices()[vIdx2*3])
	v3 := int(model.Vertices()[vIdx3*3])
	if v1 != 1 || v2 != 4 || v3 != 7 {
		t.Errorf("unexpected vertex data: expected 1, 4, 7, got %d, %d, %d",
			int(model.Vertices()[vIdx1*3]), int(model.Vertices()[vIdx2*3]), int(model.Vertices()[vIdx3*3]))
	}

	vIdx1 = face2.VertexIndices()[0]
	vIdx2 = face2.VertexIndices()[1]
	vIdx3 = face2.VertexIndices()[2]
	v1 = int(model.Vertices()[vIdx1*3])
	v2 = int(model.Vertices()[vIdx2*3])
	v3 = int(model.Vertices()[vIdx3*3])
	if v1 != -1 || v2 != -4 || v3 != -7 {
		t.Errorf("unexpected vertex data: expected -1, -4, -7, got %d, %d, %d",
			int(model.Vertices()[vIdx1*3]), int(model.Vertices()[vIdx2*3]), int(model.Vertices()[vIdx3*3]))
	}

	vnIdx1 := face1.NormalIndices()[0]
	vnIdx2 := face1.NormalIndices()[1]
	vnIdx3 := face1.NormalIndices()[2]
	vn1 := int(model.Normals()[vnIdx1*3])
	vn2 := int(model.Normals()[vnIdx2*3])
	vn3 := int(model.Normals()[vnIdx3*3])
	if vn1 != 10 || vn2 != 40 || vn3 != 70 {
		t.Errorf("unexpected normal data: expected 10, 40, 70, got %d, %d, %d",
			int(model.Normals()[vnIdx1*3]), int(model.Normals()[vnIdx2*3]), int(model.Normals()[vnIdx3*3]))
	}

	vnIdx1 = face2.NormalIndices()[0]
	vnIdx2 = face2.NormalIndices()[1]
	vnIdx3 = face2.NormalIndices()[2]
	vn1 = int(model.Normals()[vnIdx1*3])
	vn2 = int(model.Normals()[vnIdx2*3])
	vn3 = int(model.Normals()[vnIdx3*3])
	if vn1 != -10 || vn2 != -40 || vn3 != -70 {
		t.Errorf("unexpected normal data: expected -10, -40, -70, got %d, %d, %d",
			int(model.Normals()[vnIdx1*3]), int(model.Normals()[vnIdx2*3]), int(model.Normals()[vnIdx3*3]))
	}

	vtIdx1 := face1.UvIndices()[0]
	vtIdx2 := face1.UvIndices()[1]
	vtIdx3 := face1.UvIndices()[2]
	vt1 := int(model.UVs()[vtIdx1*2])
	vt2 := int(model.UVs()[vtIdx2*2])
	vt3 := int(model.UVs()[vtIdx3*2])
	if vt1 != 111 || vt2 != 444 || vt3 != -111 {
		t.Errorf("unexpected uv data: expected 111, 444, -111, got %d, %d, %d",
			int(model.UVs()[vtIdx1*2]), int(model.UVs()[vtIdx2*2]), int(model.UVs()[vtIdx3*2]))
	}
}
