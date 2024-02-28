package view

import (
	"github.com/go-gl/mathgl/mgl32"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/examples/3d/models"
	"tonysoft.com/gfx/examples/3d/object"
	"tonysoft.com/gfx/examples/3d/textures"
)

func NewCubeView(window *gfx.Window) gfx.WindowObject {
	gfx.AddEmbeddedAsset("cube.obj", models.Assets)
	gfx.AddEmbeddedAsset("cube.mtl", models.Assets)
	gfx.AddEmbeddedAsset("cube.png", textures.Assets)

	camera := gfx.NewCamera()
	camera.
		SetProjection(45, window.AspectRatio(), .5, 100).
		SetUp(mgl32.Vec3{0, 1, 0}).
		SetPosition(mgl32.Vec3{0, 0, 2})

	light := gfx.NewLight()
	light.SetDirection(mgl32.Vec3{0, 0, -1})

	model := gfx.NewModel()
	model.
		SetOBJ("cube.obj").
		SetMTL("cube.mtl").
		SetTexture(gfx.NewTexture("cube.png")). // set using path to local PNG or loaded asset
		//SetTexture(gfx.NewTexture(gfx.Red)).  // set using a solid color
		//SetTexture(gfx.NewTexture(myImage)).  // set using *image.RGBA or *image.NRGBA
		SetCamera(camera).
		AddLight(light). // omit this for non-directional/non-specular lighting
		SetName("Box")

	return object.NewViewer(window, model)
}
