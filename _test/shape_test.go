package _test

import (
	"context"
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"testing"
	"time"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/_test"
)

const (
	winWidth1  = 1000
	winHeight1 = 1000
	winWidth2  = 1000
	winHeight2 = 3000
	winWidth3  = 3000
	winHeight3 = 1000
	winWidth4  = 2000
	winHeight4 = 2000
)

func TestShape2DAnchoring(t *testing.T) {
	_test.PanicOnErr(gfx.Init())

	win := gfx.NewWindow().SetTitle(_test.WindowTitle).
		SetWidth(_test.WindowWidth).
		SetHeight(_test.WindowHeight).
		SetTargetFramerate(_test.TargetFramerate).
		SetClearColor(_test.BackgroundColor)
	ctx, cancelFunc := context.WithCancel(context.Background())

	win.EnableQuitKey(cancelFunc)
	win.EnableFullscreenKey()

	circ1 := gfx.NewCircle(.10)
	circ1.SetScale(mgl32.Vec3{.5, .5})
	circ1.SetColor(gfx.Yellow)
	circ1.SetMaintainAspectRatio(false)

	quad1 := gfx.NewQuad()
	quad1.SetScale(mgl32.Vec3{.25, .25})
	quad1.SetAnchor(gfx.TopLeft)
	quad1.SetColor(gfx.Magenta)
	quad1.SetMaintainAspectRatio(true)

	quad2 := gfx.NewQuad()
	quad2.SetScale(mgl32.Vec3{.25, .25})
	quad2.SetAnchor(gfx.TopRight)
	quad2.SetColor(gfx.Red)
	quad2.SetMaintainAspectRatio(false)

	dot1 := gfx.NewDot()
	dot1.SetScale(mgl32.Vec3{.25, .25})
	dot1.SetAnchor(gfx.BottomLeft)
	dot1.SetColor(gfx.Green)
	dot1.SetMaintainAspectRatio(true)

	dot2 := gfx.NewDot()
	dot2.SetScale(mgl32.Vec3{.25, .25})
	dot2.SetAnchor(gfx.BottomRight)
	dot2.SetColor(gfx.Green)
	dot2.SetMaintainAspectRatio(true)

	dot2Label := gfx.NewLabel()
	dot2Label.SetText("X")
	dot2Label.SetColor(gfx.Blue)
	dot2.AddChildren(dot2Label)

	validator := _test.NewSceneValidator(win)
	validator.AddPixelSampler(func() (x, y float32) { return 0, 0 }, _test.BackgroundColor, "center screen")
	validator.AddPixelSampler(func() (x, y float32) { return -.49, 0 }, gfx.Yellow, "circle in the center, left side") // no scaling because MAR=false
	validator.AddPixelSampler(func() (x, y float32) { return .49, 0 }, gfx.Yellow, "circle in the center, right side")
	validator.AddPixelSampler(func() (x, y float32) { return 0, .49 }, gfx.Yellow, "circle in the center, top side")
	validator.AddPixelSampler(func() (x, y float32) { return 0, -.49 }, gfx.Yellow, "circle in the center, bottom side")
	validator.AddPixelSampler(func() (x, y float32) { return -1.0 + win.ScaleX(.24), 1.0 - win.ScaleY(.24) }, gfx.Magenta, "quad in top-left corner")
	validator.AddPixelSampler(func() (x, y float32) { return 1.0 - .24, 1.0 - .24 }, gfx.Red, "quad in top-right corner") // no scaling because MAR=false
	validator.AddPixelSampler(func() (x, y float32) { return -.96, -1.0 + win.ScaleY(.12) }, gfx.Green, "dot in bottom-left corner, left side")
	validator.AddPixelSampler(func() (x, y float32) { return -1.0 + win.ScaleX(.24), -1.0 + win.ScaleY(.12) }, gfx.Green, "dot in bottom-left corner, right side")
	validator.AddPixelSampler(func() (x, y float32) { return -1.0 + win.ScaleX(.12), -1.0 + win.ScaleY(.24) }, gfx.Green, "dot in bottom-left corner, top side")
	validator.AddPixelSampler(func() (x, y float32) { return -1.0 + win.ScaleX(.12), -.96 }, gfx.Green, "dot in bottom-left corner, bottom side")
	validator.AddPixelSampler(func() (x, y float32) { return 1.0 - win.ScaleX(.12), -1.0 + win.ScaleY(.08) }, gfx.Green, "dot in bottom-right corner, fill color")
	validator.AddPixelSampler(func() (x, y float32) { return 1.0 - win.ScaleX(.12), -1.0 + win.ScaleY(.12) }, gfx.Blue, "dot in bottom-right corner, text color")

	win.AddObjects(quad1, quad2, dot1, dot2, circ1, validator)
	win.Init(ctx, cancelFunc)
	<-win.ReadyChan()
	_test.SleepAFewFrames()

	time.Sleep(200 * time.Millisecond) // *optional; give us some time to see the size change

	win.SetSize(winWidth1, winHeight1)
	_test.SleepAFewFrames()
	_test.ValidateScene(t, validator)

	time.Sleep(200 * time.Millisecond) // *optional; give us some time to see the size change

	win.SetSize(winWidth2, winHeight2)
	_test.SleepAFewFrames()
	_test.ValidateScene(t, validator)

	time.Sleep(200 * time.Millisecond) // *optional; give us some time to see the size change

	win.SetSize(winWidth3, winHeight3)
	_test.SleepAFewFrames()
	_test.ValidateScene(t, validator)

	time.Sleep(200 * time.Millisecond) // *optional; give us some time to see the size change

	win.SetSize(winWidth4, winHeight4)
	_test.SleepAFewFrames()
	_test.ValidateScene(t, validator)

	time.Sleep(200 * time.Millisecond) // *optional; give us some time to see the size change

	win.Close()
	gfx.Close()
}

func TestShape2DBlending(t *testing.T) {
	_test.PanicOnErr(gfx.Init())

	win := gfx.NewWindow().SetTitle(_test.WindowTitle).
		SetWidth(winWidth1).
		SetHeight(winHeight1).
		SetTargetFramerate(_test.TargetFramerate).
		SetClearColor(_test.BackgroundColor)
	ctx, cancelFunc := context.WithCancel(context.Background())

	win.EnableQuitKey(cancelFunc)
	win.EnableFullscreenKey()

	quad1 := gfx.NewQuad()
	quad1.SetScale(mgl32.Vec3{.25, .25})
	quad1.SetPositionX(-.15)
	quad1.SetColor(gfx.Red)

	quad2 := gfx.NewQuad()
	quad2.SetScale(mgl32.Vec3{.25, .25})
	quad2.SetPositionX(.15)
	quad2.SetColor(gfx.Blue)
	quad2.SetOpacity(128)

	validator := _test.NewSceneValidator(win)
	validator.AddPixelSampler(func() (x, y float32) { return -.99, .99 }, _test.BackgroundColor, "upper-left corner")
	validator.AddPixelSampler(func() (x, y float32) { return -.2, 0 }, gfx.Red, "left quad")
	validator.AddPixelSampler(func() (x, y float32) { return .2, 0 }, color.RGBA{R: 0, G: 0, B: 128, A: 191}, "right quad")
	validator.AddPixelSampler(func() (x, y float32) { return 0, 0 }, color.RGBA{R: 127, G: 0, B: 128, A: 191}, "cross-section")

	win.AddObjects(quad1, quad2, validator)
	win.Init(ctx, cancelFunc)
	<-win.ReadyChan()
	_test.SleepAFewFrames()

	_test.ValidateScene(t, validator)

	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the rendering

	win.Close()
	gfx.Close()
}

func TestShape3DManualAssetMgmt(t *testing.T) {
	_test.PanicOnErr(gfx.Init())

	win := gfx.NewWindow().SetTitle(_test.WindowTitle).
		SetWidth(_test.WindowWidth).
		SetHeight(_test.WindowHeight).
		SetTargetFramerate(_test.TargetFramerate).
		SetClearColor(_test.BackgroundColor)
	ctx, cancelFunc := context.WithCancel(context.Background())

	win.EnableQuitKey(cancelFunc)
	win.EnableFullscreenKey()

	shader := gfx.NewBasicShader("test_shader", _test.VertShader, _test.FragShader)
	texture := gfx.NewTexture2D("purple_texture", gfx.Purple)

	model := _test.NewTestModel()
	model.Meshes()[0].Faces()[0].AttachedMaterial().AttachShader(shader)
	model.Meshes()[0].Faces()[0].AttachedMaterial().(*_test.TestMaterial).DiffuseMap = texture

	quad := gfx.NewShape3D()
	quad.SetModel(model)

	vp := gfx.NewViewport(win.Width(), win.Height())
	vp.Set(-.5, 0, 1, 1) // move left by half the window width
	quad.SetViewport(vp)

	validator := _test.NewSceneValidator(win)
	validator.AddPixelSampler(func() (x, y float32) { return .5, -.5 }, _test.BackgroundColor, "lower-right quadrant")

	// Because objects are initialized in the reverse order in which they're
	// added to the window or to the init queue (which is used like a "stack"),
	// we add the shader last because we need it to be initialized before the
	// quad is initialized.  This is why it's recommended to use an AssetLibrary
	// to manage assets, as services (like AssetLibrary) get initialized before
	// objects (like Shape3D), ensuring assets like shaders get initialized first.
	win.AddObjects(quad, validator)
	win.InitObjectAsync(shader)
	win.InitObjectAsync(texture) // here we use the Async version because the window isn't running yet
	win.Init(ctx, cancelFunc)
	<-win.ReadyChan()
	_test.SleepAFewFrames()

	//win.InitObject(texture) // if the window was running, then you could use the synchronous version

	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the color change

	_test.ValidateScene(t, validator)

	vp.Set(0, -.5, 1, 1) // move down by half the window height

	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the color change

	validator.Samplers[0].ExpectedColor = gfx.Purple
	_test.ValidateScene(t, validator)

	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the color change

	// Since in this test we aren't using an AssetLibrary, we close the assets
	// this way (CloseObjectAsync() also exists, but we don't want that in this
	// case since the window will soon be closed):
	win.CloseObject(texture)
	win.CloseObject(shader)

	win.Close()
	gfx.Close()
}

func TestShape3DWithAssetLibrary(t *testing.T) {
	_test.PanicOnErr(gfx.Init())

	win := gfx.NewWindow().SetTitle(_test.WindowTitle).
		SetWidth(_test.WindowWidth).
		SetHeight(_test.WindowHeight).
		SetTargetFramerate(_test.TargetFramerate).
		SetClearColor(_test.BackgroundColor)
	ctx, cancelFunc := context.WithCancel(context.Background())

	win.EnableQuitKey(cancelFunc)
	win.EnableFullscreenKey()

	shader := gfx.NewBasicShader("test_shader", _test.VertShader, _test.FragShader)
	gfx.Assets.Add(shader)

	texture := gfx.NewTexture2D("orange_texture", gfx.Orange)
	gfx.Assets.Add(texture)

	model := _test.NewTestModel()
	model.Meshes()[0].Faces()[0].AttachedMaterial().AttachShader(shader)
	model.Meshes()[0].Faces()[0].AttachedMaterial().(*_test.TestMaterial).DiffuseMap = texture

	quad := gfx.NewShape3D()
	quad.SetModel(model)

	validator := _test.NewSceneValidator(win)
	validator.AddPixelSampler(func() (x, y float32) { return -.5, -.5 }, _test.BackgroundColor, "lower-left quadrant")

	win.AddObjects(quad, validator)
	win.Init(ctx, cancelFunc)
	<-win.ReadyChan()
	_test.SleepAFewFrames()

	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the color change

	_test.ValidateScene(t, validator)

	vp := gfx.NewViewport(win.Width(), win.Height())
	vp.Set(-.5, -.5, 1, 1) // move left/down by half the window width/height (respectively)
	quad.SetViewport(vp)
	_test.SleepAFewFrames()

	time.Sleep(400 * time.Millisecond) // *optional; give us some time to see the color change

	validator.Samplers[0].ExpectedColor = gfx.Orange
	_test.ValidateScene(t, validator)

	win.Close()
	gfx.Close()
}