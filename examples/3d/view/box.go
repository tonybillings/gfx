package view

import (
	"github.com/go-gl/mathgl/mgl32"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/examples/3d/models"
	"tonysoft.com/gfx/examples/3d/object"
	"tonysoft.com/gfx/examples/3d/textures"
)

func NewBoxView(window *gfx.Window) gfx.WindowObject {
	gfx.AddEmbeddedAsset("box.obj", models.Assets)
	gfx.AddEmbeddedAsset("box.mtl", models.Assets)
	gfx.AddEmbeddedAsset("box.png", textures.Assets)

	camera1 := gfx.NewCamera()
	camera1.
		SetProjection(45, window.AspectRatio(), .5, 100).
		SetUp(mgl32.Vec3{0, 1, 0}).
		SetPosition(mgl32.Vec3{-2, 0, 2})

	light1 := gfx.NewLight()
	light1.SetDirection(mgl32.Vec3{0, 0, -1})

	model := gfx.NewModel()
	model.
		SetOBJ("box.obj").
		SetMTL("box.mtl").
		SetTexture("box.png").
		SetCamera(camera1).
		AddLight(light1). // omit this for non-directional/specular lighting
		SetName("Box")

	rotObj := object.NewRotatingObject(model)
	return rotObj
}
