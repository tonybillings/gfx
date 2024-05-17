package _test

import (
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, 2, len(model.Meshes()), "unexpected mesh count")

	mesh1 := model.Meshes()[0]
	assert.Equal(t, "FubarMesh001", mesh1.Name(), "unexpected mesh name for mesh1")

	mesh2 := model.Meshes()[1]
	assert.Equal(t, "FubarMesh002", mesh2.Name(), "unexpected mesh name for mesh2")

	assert.Equal(t, 2, len(mesh1.Faces()), "unexpected face count for mesh1")
	assert.Equal(t, 3, len(mesh2.Faces()), "unexpected face count for mesh2")

	face1 := mesh1.Faces()[0]
	face2 := mesh1.Faces()[1]

	vIdx1 := face1.VertexIndices()[0]
	vIdx2 := face1.VertexIndices()[1]
	vIdx3 := face1.VertexIndices()[2]
	v1 := int(model.Vertices()[vIdx1*3])
	v2 := int(model.Vertices()[vIdx2*3])
	v3 := int(model.Vertices()[vIdx3*3])
	assert.Equal(t, 1, v1, "unexpected vertex data for face1, vertex1")
	assert.Equal(t, 4, v2, "unexpected vertex data for face1, vertex2")
	assert.Equal(t, 7, v3, "unexpected vertex data for face1, vertex3")

	vIdx1 = face2.VertexIndices()[0]
	vIdx2 = face2.VertexIndices()[1]
	vIdx3 = face2.VertexIndices()[2]
	v1 = int(model.Vertices()[vIdx1*3])
	v2 = int(model.Vertices()[vIdx2*3])
	v3 = int(model.Vertices()[vIdx3*3])
	assert.Equal(t, -1, v1, "unexpected vertex data for face2, vertex1")
	assert.Equal(t, -4, v2, "unexpected vertex data for face2, vertex2")
	assert.Equal(t, -7, v3, "unexpected vertex data for face2, vertex3")

	vnIdx1 := face1.NormalIndices()[0]
	vnIdx2 := face1.NormalIndices()[1]
	vnIdx3 := face1.NormalIndices()[2]
	vn1 := int(model.Normals()[vnIdx1*3])
	vn2 := int(model.Normals()[vnIdx2*3])
	vn3 := int(model.Normals()[vnIdx3*3])
	assert.Equal(t, 10, vn1, "unexpected normal data for face1, normal1")
	assert.Equal(t, 40, vn2, "unexpected normal data for face1, normal2")
	assert.Equal(t, 70, vn3, "unexpected normal data for face1, normal3")

	vnIdx1 = face2.NormalIndices()[0]
	vnIdx2 = face2.NormalIndices()[1]
	vnIdx3 = face2.NormalIndices()[2]
	vn1 = int(model.Normals()[vnIdx1*3])
	vn2 = int(model.Normals()[vnIdx2*3])
	vn3 = int(model.Normals()[vnIdx3*3])
	assert.Equal(t, -10, vn1, "unexpected normal data for face2, normal1")
	assert.Equal(t, -40, vn2, "unexpected normal data for face2, normal2")
	assert.Equal(t, -70, vn3, "unexpected normal data for face2, normal3")

	vtIdx1 := face1.UvIndices()[0]
	vtIdx2 := face1.UvIndices()[1]
	vtIdx3 := face1.UvIndices()[2]
	vt1 := int(model.UVs()[vtIdx1*2])
	vt2 := int(model.UVs()[vtIdx2*2])
	vt3 := int(model.UVs()[vtIdx3*2])
	assert.Equal(t, 111, vt1, "unexpected uv data for face1, uv1")
	assert.Equal(t, 444, vt2, "unexpected uv data for face1, uv2")
	assert.Equal(t, -111, vt3, "unexpected uv data for face1, uv3")
}
