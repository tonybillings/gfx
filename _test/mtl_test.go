package _test

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tonybillings/gfx/obj"
	"testing"
)

var mtlFile = `
# a comment that should be ignored
newmtl FubarMat001
    Ka 1 2 3
	Kd 4.0 5.0 6.0
	Ks 7 7 7
	Ns 333.0
	Tr 0.642
	d 0.0000
	Tf 1.0000 1.0000 1.0000
	Pr 50.0
	Pm 0.0000
	Ni 0.0000
	Ke 9.0000 8.0000 7.0000
	illum 2
	map_Kd map_diffuse.png
	map_Ks map_specular.png
	norm map_normal.png

newmtl FubarMat002
    Ka 4 5 6
	Kd 7.0 8.0 9.0
	Ks 8 8 8
	Ns 666.0
	Tr .456
	d 0.0000
	Tf 2.0000 2.0000 2.0000
	Pr 60.0
	Pm 0.0000
	Ni 0.0000
	Ke 0.0000 0.0000 0.0000
	illum 2
	map_Kd map_diffuse2.png
	map_Ks map_specular2.png
	norm map_normal2.png
`

func TestMTLLoading(t *testing.T) {
	mtl := obj.NewMaterialLibrary("TestLibrary", mtlFile)
	mtl.Load()

	if len(mtl.GetNames()) != 2 {
		t.Errorf("unexpected material count: expected %d, got %d", 2, len(mtl.GetNames()))
	}

	mat1 := mtl.Get("FubarMat001")
	if mat1 == nil {
		t.Errorf("material not found: expected to find %s", "FubarMat001")
	}

	mat2 := mtl.Get("FubarMat002")
	if mat2 == nil {
		t.Errorf("material not found: expected to find %s", "FubarMat002")
	}

	Ambient := mgl32.Vec4{1, 2, 3}
	if mat1.Properties.Ambient != Ambient {
		t.Errorf("unexpected ambient value: expected %v, got %v", Ambient, mat1.Properties.Ambient)
	}

	Diffuse := mgl32.Vec4{4, 5, 6}
	if mat1.Properties.Diffuse != Diffuse {
		t.Errorf("unexpected diffuse value: expected %v, got %v", Diffuse, mat1.Properties.Diffuse)
	}

	Specular := mgl32.Vec4{7, 7, 7}
	if mat1.Properties.Specular != Specular {
		t.Errorf("unexpected specular value: expected %v, got %v", Specular, mat1.Properties.Specular)
	}

	Shininess := float32(333.0)
	if mat1.Properties.Shininess != Shininess {
		t.Errorf("unexpected shininess value: expected %v, got %v", Shininess, mat1.Properties.Shininess)
	}

	Emissive := mgl32.Vec4{9, 8, 7}
	if mat1.Properties.Emissive != Emissive {
		t.Errorf("unexpected emissive value: expected %v, got %v", Emissive, mat1.Properties.Emissive)
	}

	Transparency := float32(.642)
	if mat1.Properties.Transparency != Transparency {
		t.Errorf("unexpected transparency value: expected %v, got %v", Transparency, mat1.Properties.Transparency)
	}

	Ambient = mgl32.Vec4{4, 5, 6}
	if mat2.Properties.Ambient != Ambient {
		t.Errorf("unexpected ambient value: expected %v, got %v", Ambient, mat2.Properties.Ambient)
	}

	Transparency = float32(.456)
	if mat2.Properties.Transparency != Transparency {
		t.Errorf("unexpected transparency value: expected %v, got %v", Transparency, mat2.Properties.Transparency)
	}
}
