package obj

import (
	"github.com/tonybillings/gfx"
)

/******************************************************************************
 Mesh
******************************************************************************/

type Mesh struct {
	gfx.MeshBase

	name  string
	faces []*Face
}

/******************************************************************************
 gfx.Mesh Implementation
******************************************************************************/

func (m *Mesh) Name() string {
	return m.name
}

func (m *Mesh) Faces() []gfx.Face {
	faces := make([]gfx.Face, len(m.faces))
	for i, f := range m.faces {
		faces[i] = f
	}
	return faces
}

/******************************************************************************
 gfx.Initer Implementation
******************************************************************************/

func (m *Mesh) Init() bool {
	ok := true
	for _, f := range m.faces {
		ok = ok && f.Init()
	}
	return ok
}

/******************************************************************************
 gfx.Closer Implementation
******************************************************************************/

func (m *Mesh) Close() {
	for _, f := range m.faces {
		f.Close()
	}
}

/******************************************************************************
 New Mesh Function
******************************************************************************/

func NewMesh() *Mesh {
	return &Mesh{
		MeshBase: gfx.MeshBase{
			ObjectTransform: *gfx.NewObjectTransform(),
		},
	}
}
