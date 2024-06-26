package view

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tonybillings/gfx"
	"github.com/tonybillings/gfx/examples/3d/obj/models"
	"github.com/tonybillings/gfx/examples/3d/obj/object"
	"github.com/tonybillings/gfx/examples/3d/obj/textures"
	"github.com/tonybillings/gfx/obj"
)

func NewCubeView(window *gfx.Window) gfx.WindowObject {
	// Make the Model/Texture source available to other Assets
	window.Assets().AddEmbeddedFiles(models.Assets)
	window.Assets().AddEmbeddedFiles(textures.Assets)

	// Use the obj package to import Wavefront .OBJ files.  Since
	// the OBJ source file was just added as an Asset, the Model
	// will load from that source, otherwise it will search for the
	// .obj file on the local file system.
	model := obj.NewModel("cube", "cube.obj")
	// Unlike with the "in_mem" example, this Model asset needs to
	// participate in the Init()/Close() life cycle.  Like all
	// Assets, it's best to manage them via an AssetLibrary anyway,
	// so we'll add it to the Window's asset library.  If we didn't
	// use an asset library then the file references in the OBJ/MTL
	// files would either need to be full, absolute paths or as
	// relative paths they must work by assuming the correct working
	// directory of the application; with an asset library, the
	// references can simply be the names given to them as Assets
	// when added to the library.
	window.Assets().Add(model)

	// Optionally, if you plan to use a normal map and the model
	// doesn't come with tangent/bitangent vectors, you can have it
	// generate them automatically.
	model.ComputeTangents(true)

	// The default shader that the obj package uses is gfx.Shape3DShader.
	// That shader fully supports obj.BasicMaterial. This is how you could
	// change it to the "no lights" Shape3D shader:
	// model.SetDefaultShader(window.Assets().Get(gfx.Shape3DNoLightsShader).(gfx.Shader))
	// That will set the default shader assigned to materials when loading,
	// though you can always load the model now and change the shader
	// for any given material and each face can be rendered by a separate
	// material, if desired.  Note that *contiguous* groups of faces belonging
	// to the same mesh and having the same material will be rendered together
	// in the same draw call, so having a mesh with 100 faces, all using the same
	// material except for face #50, will result in 3 draw calls for the rendering
	// of that mesh (for faces 1-49, 50, and 51-100).

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
	// model.Load() // since it hasn't been initialized yet (by the Window), we need to preload
	// material := model.Meshes()[0].Faces()[0].AttachedMaterial().(*obj.BasicMaterial)
	// material.Lock()                       // again, not needed now
	// material.Properties.Transparency = .5 // make it transparent!
	// material.Unlock()
	// Remember that you are changing the material referenced to by many faces!

	// You can also override the texture that's set based on the OBJ/MTL:
	// material.DiffuseMap = gfx.NewTexture2D("MyTexture", gfx.Red)
	// In that example, the texture is set to a solid color, but regardless
	// since you provided the texture (it wasn't loaded by the material library
	// as per the materials defined therein), you're responsible for its life cycle.
	// Adding it to an AssetLibrary ensures the life cycle is properly handled:
	// window.Assets().Add(material.DiffuseMap)

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
		SetModel(window.Assets().Get("cube").(gfx.Model)).
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
