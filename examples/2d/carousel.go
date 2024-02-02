package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"tonysoft.com/gfx"
)

type Carousel struct {
	gfx.Shape
	rotIndex int
	triangle *gfx.Shape
	quad     *gfx.Shape
	dot      *gfx.Shape
	hello    *gfx.Label
	world    *gfx.Label
}

func (c *Carousel) defaultLayout() {
	c.triangle.SetColor(gfx.Red)
	c.triangle.SetScale(mgl32.Vec3{.2, .2})
	c.triangle.SetBlurEnabled(true)
	c.triangle.SetBlurIntensity(3)

	triangleLabel := gfx.NewLabel()
	triangleLabel.SetText("A")
	triangleLabel.SetScale(mgl32.Vec3{.5, .5})
	c.triangle.AddChild(triangleLabel)

	c.quad.SetColor(gfx.Green)
	c.quad.SetScale(mgl32.Vec3{.15, .15})
	c.quad.SetBlurEnabled(true)
	c.quad.SetBlurIntensity(3)
	quadLabel := gfx.NewLabel()
	quadLabel.SetText("B")
	quadLabel.SetScale(mgl32.Vec3{.7, .7})
	c.quad.AddChild(quadLabel)

	c.dot.SetColor(gfx.Blue)
	c.dot.SetScale(mgl32.Vec3{.1, .1})
	c.dot.SetBlurEnabled(true)
	c.dot.SetBlurIntensity(3)
	dotLabel := gfx.NewLabel()
	dotLabel.SetText("C")
	dotLabel.SetScale(mgl32.Vec3{.9, .9})
	c.dot.AddChild(dotLabel)

	c.hello.SetText("hello")
	c.hello.SetScale(mgl32.Vec3{.2, .05})
	c.hello.SetColor(gfx.Magenta)
	c.hello.SetFillColor(gfx.Purple)

	c.world.SetText("world")
	c.world.SetScale(mgl32.Vec3{.2, .05})
	c.world.SetColor(gfx.Lighten(gfx.Purple, .4))
	c.world.SetBorderColor(gfx.Magenta)
	c.world.SetBorderThickness(.05)

	c.SetSides(5)
	c.SetColor(gfx.Orange)
	c.SetScale(mgl32.Vec3{.8, .8})
}

func (c *Carousel) Update(deltaTime int64) bool {
	if !c.WindowObjectBase.Update(deltaTime) {
		return false
	}

	rotation := float32(c.rotIndex) * 0.01
	c.SetRotationZ(rotation)
	c.rotIndex++

	window := c.Window()

	seatAngle := float32(math.Pi / 5)
	scaledX := float64(window.ScaleX(c.Scale().X()))
	scaledY := float64(window.ScaleY(c.Scale().Y()))

	triangleX := math.Cos(float64(rotation+seatAngle)) * scaledX
	triangleY := math.Sin(float64(rotation+seatAngle)) * scaledY
	c.triangle.SetPosition(mgl32.Vec3{-float32(triangleX), float32(triangleY)})
	c.triangle.SetRotationZ(rotation * 2)

	quadX := math.Cos(float64(rotation+seatAngle*3)) * scaledX
	quadY := math.Sin(float64(rotation+seatAngle*3)) * scaledY
	c.quad.SetPosition(mgl32.Vec3{-float32(quadX), float32(quadY)})
	c.quad.SetRotationZ(rotation)

	dotX := math.Cos(float64(rotation+seatAngle*5)) * scaledX
	dotY := math.Sin(float64(rotation+seatAngle*5)) * scaledY
	c.dot.SetPosition(mgl32.Vec3{-float32(dotX), float32(dotY)})
	c.dot.SetOrigin(mgl32.Vec3{-float32(dotX) * 1.2, float32(dotY) * 1.2})
	c.dot.SetRotationZ(rotation * 8)

	helloX := math.Cos(float64(rotation+seatAngle*7)) * scaledX
	helloY := math.Sin(float64(rotation+seatAngle*7)) * scaledY
	c.hello.SetPosition(mgl32.Vec3{-float32(helloX), float32(helloY)})
	c.hello.SetRotationZ(-rotation)

	worldX := math.Cos(float64(rotation+seatAngle*9)) * scaledX
	worldY := math.Sin(float64(rotation+seatAngle*9)) * scaledY
	c.world.SetPosition(mgl32.Vec3{-float32(worldX), float32(worldY)})
	c.world.SetRotationZ(-rotation)

	return true
}

func NewCarousel() *Carousel {
	c := &Carousel{
		Shape:    *gfx.NewShape(),
		triangle: gfx.NewShape(),
		quad:     gfx.NewQuad(),
		dot:      gfx.NewDot(),
		hello:    gfx.NewLabel(),
		world:    gfx.NewLabel(),
	}

	c.AddChildren(c.triangle, c.quad, c.dot, c.hello, c.world)
	c.defaultLayout()
	return c
}
