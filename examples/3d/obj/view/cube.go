package view

import (
	"github.com/go-gl/mathgl/mgl32"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/examples/3d/obj/models"
	"tonysoft.com/gfx/examples/3d/obj/object"
	"tonysoft.com/gfx/examples/3d/obj/textures"
	"tonysoft.com/gfx/obj"
)

func NewCubeView(window *gfx.Window) gfx.WindowObject {
	// Make the Model/Texture source available to other Assets
	gfx.Assets.AddEmbeddedFiles(models.Assets)
	gfx.Assets.AddEmbeddedFiles(textures.Assets)

	// Use the obj package to import Wavefront .OBJ files.  Since
	// the OBJ source file was just added as an Asset, the Model
	// will load from that source, otherwise it will search for the
	// .obj file on the local file system.
	model := obj.NewModel("cube", "cube.obj")
	// Unlike with the "in_mem" example, this Model asset needs to
	// participate in the Init()/Close() life-cycle.  Like all
	// Assets, it's best to manage them via an AssetLibrary anyway,
	// so we'll add to gfx.Assets.
	gfx.Assets.Add(model)

	// Optionally, if you plan to use a normal map and the model
	// doesn't come with tangent/bitangent vectors, you can have it
	// generate them automatically.
	model.ComputeTangents(true)

	// The default shader that the obj package uses is gfx.Shape3DShader.
	// That shader fully supports obj.BasicMaterial, except for
	// sampling normal and specular maps. For that, we need to
	// provide a capable shader such as this one:
	model.SetDefaultShader(gfx.Assets.Get(gfx.Shape3DBumpedSpecularShader).(gfx.Shader))
	// That will set the default shader assigned to materials,
	// though you can always load the model now and change the
	// shader for any given material and each face can be rendered
	// by a separate material, if desired.

	camera := gfx.NewCamera()
	camera.SetProjection(45, window.AspectRatio(), .1, 1000)
	camera.Lock()                                  // not really necessary at this time, but is recommended once the Window is running
	camera.Properties.Target = mgl32.Vec4{0, 0, 0} // this is the default Target vector
	camera.Properties.Up = mgl32.Vec4{0, 1, 0}     // this is the default Up vector
	camera.Properties.Position = mgl32.Vec4{0, 0, 2}
	camera.Unlock() // the Properties struct is bound to the shader as a UBO, hence the need to lock later

	// A Camera is not an Asset, just an Object, but it still needs
	// to participate in the Update() loop so it can update its
	// view-projection matrix, so we'll add it to the Window.
	window.AddObject(camera)

	lighting := gfx.NewQuadDirectionalLighting() // "lighting" can be any shader-bindable object (see README)
	lighting.Lock()                              // like the camera, not really necessary to lock now
	lighting.LightCount = 2                      // can be 0 through 4 with this lighting object
	lighting.Lights[0].Color = mgl32.Vec3{0, 0, .9}
	lighting.Lights[0].Direction = mgl32.Vec3{.3, .2, -.3}
	lighting.Lights[1].Color = mgl32.Vec3{.9, .5, .13}
	lighting.Lights[1].Direction = mgl32.Vec3{.5, 1, .3}
	lighting.Unlock()

	// If you need to change the material at runtime...
	// model.Load() // since it hasn't been initialized yet (by the Window), we need to pre-load
	// material := model.Meshes()[0].Faces()[0].AttachedMaterial().(*obj.BasicMaterial)
	// material.Lock()                       // again, not needed now
	// material.Properties.Transparency = .5 // make it transparent!
	// material.Unlock()

	// You can also override the texture that's set based on the OBJ/MTL:
	// material.DiffuseMap = gfx.NewTexture2D("MyTexture", gfx.Red)
	// In that example, the texture is set to a solid color, but regardless
	// since you provided the texture (it wasn't loaded by the material library
	// as per the materials defined therein), you're responsible for its life-cycle.
	// Adding it to an AssetLibrary ensures the life-cycle is properly handled:
	// gfx.Assets.Add(material.DiffuseMap)

	// You could also change the transforms for individual meshes, where
	// the transform for the Shape3D object becomes their parent transform:
	// model.Load() // have to do this first, if model is uninitialized
	// model.Meshes()[0].SetRotationZ(1.5)
	// Note that changing the model asset means affecting all future instances
	// that are based off this asset and does not affect existing model instances.
	// After a Shape3D has been initialized, changes can be made to that specific
	// instance using the same syntax:
	// cube.Meshes()[0].SetRotationZ(1.5)

	cube := gfx.NewShape3D()
	cube.
		SetModel(gfx.Assets.Get("cube").(gfx.Model)).
		SetCamera(camera).
		SetLighting(lighting).
		SetName("TestCube")

	go func() {
		<-window.ReadyChan()            // cube must be initialized before Meshes() is ready
		cube.Meshes()[0].SetScaleY(.5)  // changes just this instance
		model.Meshes()[0].SetScaleX(.5) // changes the model asset itself, affecting only new instances
	}()

	return object.NewViewer(window, cube)
}
