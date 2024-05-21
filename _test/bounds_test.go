package _test

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/stretchr/testify/assert"
	"github.com/tonybillings/gfx"
	"testing"
)

const (
	winWidth1  = 1000
	winHeight1 = 1000
	winWidth2  = 1000
	winHeight2 = 3000
	winWidth3  = 3000
	winHeight3 = 1000
)

func TestNewBoundingObjectBase(t *testing.T) {
	bo := gfx.NewBoundingObject()
	assert.False(t, bo.MouseOver(), "unexpected MouseOver state: expected false, got %v", bo.MouseOver())
}

func TestBoundingBoxMouseEnterAndLeave(t *testing.T) {
	w := gfx.NewWindow()

	b := gfx.NewBoundingBox()
	b.SetMouseSurface(w)

	entered := false
	left := false
	reset := func() {
		entered = false
		left = false
	}
	b.OnMouseEnter(func(_ gfx.WindowObject, _ *gfx.MouseState) {
		entered = true
	})
	b.OnMouseLeave(func(_ gfx.WindowObject, _ *gfx.MouseState) {
		left = true
	})

	test := func(x, y float32, testEnter, testLeave bool) {
		ms := gfx.MouseState{X: x, Y: y}
		w.OverrideMouseState(&ms)
		b.Update(0)
		if testEnter {
			assert.True(t, entered, "expected MouseEnter event to be triggered")
		} else {
			assert.False(t, entered, "expected 'entered' to be false")
		}
		if testLeave {
			assert.True(t, left, "expected MouseLeave event to be triggered")
		} else {
			assert.False(t, left, "expected 'left' to be false")
		}
		reset()
	}

	w.SetWidth(winWidth1)
	w.SetHeight(winHeight1)

	// Anywhere inside [-1,1] for X/Y means inside window...
	test(0, 0, true, false)      //...so for the 1st time we enter the window, which the bounding box fills
	test(.5, .9, false, false)   // still within the window/box bounds, so no change in mouse-over state
	test(.5, 1.01, false, true)  // here we exit the bounds for the 1st time
	test(.5, .99, true, false)   // dip back inside the window/box
	test(.5, -.5, false, false)  // still technically within bounds, so no change in state
	test(-1.01, .5, false, true) // leave...
	test(-.99, .5, true, false)  // re-enter...
	test(0, 0, false, false)     // still within bounds, just resetting the position back to origin

	// Now the bounding box will be for a non-square shape that doesn't fill the entire window
	container := gfx.NewView() // need to use View for anchoring to work, which is utilized later
	container.SetWindow(w)
	container.SetScale(mgl32.Vec3{.5, .75})
	container.AddChild(b) // bounding boxes should have a (1,1) scale so that they fill their parent

	test(.5, .75, false, false)   // on edge of the shape/box, but still inside so no state change
	test(.5, .75001, false, true) // now we exit, dip back in/out for the following tests:
	test(.5, .75, true, false)
	test(.5001, .75, false, true)
	test(.5, .99, false, false)
	test(.5, .5, true, false)
	test(-1.01, .5, false, true)
	test(-.99, .5, false, false)
	test(.5, .51, true, false)
	test(-.50001, .5, false, true)
	test(-.49999, .7499, true, false)
	test(-.49999, .75001, false, true)
	test(0, 0, true, false)

	// With a non-square window and MaintainAspectRatio (MAR) set to true, the shape
	// will no longer consume screen space precisely relative to its scale.  Instead,
	// its effective scale (screen space consumed) will be adjusted to maintain its
	// original shape. This means we need to adjust the simulated mouse position similarly
	// or adjust our assertions accordingly when it comes to mouse enter/leave events.
	w.SetWidth(winWidth2)
	w.SetHeight(winHeight2)

	test(.5, w.ScaleY(.75), false, false) // when testing, we can adjust for MAR...
	test(.5, .75, false, true)            // ...or not, but then the assertion must change
	test(.5, w.ScaleY(.75), true, false)
	test(.5001, .75, false, true)
	test(.5, .99, false, false)
	test(.5, w.ScaleY(.5), true, false)
	test(-1.01, .5, false, true)
	test(-.99, .5, false, false)
	test(.5, w.ScaleY(.51), true, false)
	test(-.50001, .5, false, true)
	test(-.49999, .7499, false, false)          // without adjusting for MAR, we're still outside the bounds
	test(-.49999, w.ScaleY(.7499), true, false) // now we're back inside
	test(-.49999, .75001, false, true)
	test(0, 0, true, false)

	// Now testing the inverted aspect ratio...
	w.SetWidth(winWidth3)
	w.SetHeight(winHeight3)

	test(w.ScaleX(.5), .75, false, false) // ...so we need to adjust for MAR on the X axis
	test(.5, .75001, false, true)
	test(w.ScaleX(.5), .75, true, false)
	test(.5001, .75, false, true)
	test(.5, .99, false, false)
	test(w.ScaleX(.5), .5, true, false)
	test(-1.01, .5, false, true)
	test(-.99, .5, false, false)
	test(w.ScaleX(.5), .51, true, false)
	test(-.50001, .5, false, true)
	test(w.ScaleX(-.49999), .7499, true, false)
	test(w.ScaleX(-.49999), .75001, false, true)
	test(0, 0, true, false)

	// Now the container/box will consume the lower-right quadrant of the
	// lower-right quadrant of the screen (based on anchor and scale). Child
	// objects will anchor to their parent based on the bounds of the parent.
	container.SetAnchor(gfx.BottomRight)
	container.SetScale(mgl32.Vec3{.25, .25})
	container.RefreshLayout() // only necessary when not added to a running window

	test(-1.1, 1.1, false, true)
	test(1.0-w.ScaleX(.24)*2, -.5, true, false)     // the aspect ratio is still the same, so still scaling on the X axis
	test(1.0-w.ScaleX(.24)*2, -1.0+.6, false, true) // ...but now we're testing the edges of the box *from* the window edges,
	// which also means multiplying by 2 because we're translating across the [-1,1] range
	test(.76, -.99, false, false)                  // this is still outside because MAR=true and the 3:1 ultra-wide aspect ratio
	test(.85, -.99, true, false)                   // this is inside the box, however
	test(.99, -.99, false, false)                  // if the window were wide enough, this assertion would fail as we'd be outside!
	test(1.0-w.ScaleX(.24)*2, -.4999, false, true) // out of bounds here...too high on the Y axis
	test(1.0-w.ScaleX(.24)*2, -.5, true, false)    // now we're back inside the box, on the top edge (top-left corner, actually)
	test(1.0-w.ScaleX(.25)*2, -.5, false, true)    // just outside the top-left corner now, too far left to be inside the box
	test(.99, -.99, true, false)                   // come back inside the box in preparation for the next test

	// Finally, the container will be manually positioned/scaled to consume
	// the same space, but with anchoring disabled and MAR set to false.
	container.SetAnchor(gfx.NoAnchor)
	container.RefreshLayout() // only necessary when not added to a running window
	container.SetPosition(mgl32.Vec3{1.0 - w.ScaleX(.25), -1.0 + w.ScaleY(.25)})
	container.SetScale(mgl32.Vec3{w.ScaleX(.25), w.ScaleY(.25)})
	container.SetMaintainAspectRatio(false)
	container.RefreshLayout() // only necessary when not added to a running window
	b.SetMaintainAspectRatio(false)
	b.RefreshLayout() // only necessary when not added to a running window

	// Run the same tests as before...
	test(-1.1, 1.1, false, true)
	test(1.0-w.ScaleX(.24)*2, -.5, true, false)
	test(1.0-w.ScaleX(.24)*2, -1.0+.6, false, true)
	test(.76, -.99, false, false)
	test(.85, -.99, true, false)
	test(.99, -.99, false, false)
	test(1.0-w.ScaleX(.24)*2, -.4999, false, true)
	test(1.0-w.ScaleX(.24)*2, -.5, true, false)
	test(1.0-w.ScaleX(.25)*2, -.5, false, true)
	test(.99, -.99, true, false)
}

func TestBoundingRadiusMouseEnterAndLeave(t *testing.T) {
	w := gfx.NewWindow()

	b := gfx.NewBoundingRadius()
	b.SetMouseSurface(w)

	entered := false
	left := false
	reset := func() {
		entered = false
		left = false
	}
	b.OnMouseEnter(func(_ gfx.WindowObject, _ *gfx.MouseState) {
		entered = true
	})
	b.OnMouseLeave(func(_ gfx.WindowObject, _ *gfx.MouseState) {
		left = true
	})

	test := func(x, y float32, testEnter, testLeave bool) {
		ms := gfx.MouseState{X: x, Y: y}
		w.OverrideMouseState(&ms)
		b.Update(0)
		if testEnter {
			assert.True(t, entered, "expected MouseEnter event to be triggered")
		} else {
			assert.False(t, entered, "expected 'entered' to be false")
		}
		if testLeave {
			assert.True(t, left, "expected MouseLeave event to be triggered")
		} else {
			assert.False(t, left, "expected 'left' to be false")
		}
		reset()
	}

	w.SetWidth(winWidth1)
	w.SetHeight(winHeight1)

	// Anywhere inside a radius of 1.0 (the default) is inside the bounding radius
	test(0, 0, true, false)        // world/screen origin is also the bounding radius origin
	test(.99, 0, false, false)     // right side, still inside the radius
	test(0, -.99, false, false)    // bottom side
	test(-.99, 0, false, false)    // left side
	test(0, .99, false, false)     // top side
	test(.95, .95, false, true)    // upper-right corner of the window would not be covered by the radius
	test(.95, -.95, false, false)  // ...or any other corner, of course
	test(-.95, -.95, false, false) // lower-left corner
	test(-.95, .95, false, false)  // upper-left corner

	// Change to a 3:1 aspect ratio
	w.SetWidth(winWidth3)
	w.SetHeight(winHeight3)

	// Like the BoundingBox test, will use a container (View) to put
	// the bounding radius in the lower-right quadrant of the lower-right
	// quadrant of the window.
	container := gfx.NewView()
	container.SetWindow(w)
	container.SetScale(mgl32.Vec3{.25, .25})
	container.AddChild(b)
	container.SetAnchor(gfx.BottomRight)
	container.RefreshLayout()

	test(1.0-w.ScaleX(.125), -1.0+.125, true, false) // this is center of the bounding radius / container
	test(0, 0, false, true)                          // world/screen origin is no longer inside the container
	test(1.0-w.ScaleX(.24)*2, -.5, false, false)     // unlike with the box, using a radius means we're still outside
	test(1.0-w.ScaleX(.17)*2, -.52, true, false)     // now we're inside
}
