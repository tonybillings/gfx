package _test

import (
	"testing"
	"tonysoft.com/gfx"
)

func TestNewViewport(t *testing.T) {
	windowWidth, windowHeight := int32(1920), int32(1080)
	v := gfx.NewViewport(windowWidth, windowHeight)

	expectedX, expectedY := int32(0), int32(0)
	expectedW, expectedH := windowWidth, windowHeight

	x, y, w, h := v.Get()
	if x != expectedX || y != expectedY || w != expectedW || h != expectedH {
		t.Errorf("unexpected viewport: expected (%d, %d, %d, %d), got (%d, %d, %d, %d)", expectedX, expectedY, expectedW, expectedH, x, y, w, h)
	}
}

func TestViewportSet(t *testing.T) {
	v := gfx.NewViewport(800, 600)

	v.Set(0.1, 0.1, 0.5, 0.5)

	expectedX, expectedY := int32(80), int32(60)
	expectedW, expectedH := int32(400), int32(300)

	x, y, w, h := v.Get()
	if x != expectedX || y != expectedY || w != expectedW || h != expectedH {
		t.Errorf("unexpected viewport: expected (%d, %d, %d, %d), got (%d, %d, %d, %d)", expectedX, expectedY, expectedW, expectedH, x, y, w, h)
	}
}

func TestViewportSetWindowSize(t *testing.T) {
	v := gfx.NewViewport(800, 600)

	v.Set(0.1, 0.1, 0.5, 0.5)
	v.SetWindowSize(1600, 1200)

	expectedX, expectedY := int32(160), int32(120)
	expectedW, expectedH := int32(800), int32(600)

	x, y, w, h := v.Get()
	if x != expectedX || y != expectedY || w != expectedW || h != expectedH {
		t.Errorf("unexpected viewport: expected (%d, %d, %d, %d), got (%d, %d, %d, %d)", expectedX, expectedY, expectedW, expectedH, x, y, w, h)
	}
}

func TestViewportGet(t *testing.T) {
	v := gfx.NewViewport(1024, 768)

	x, y, w, h := v.Get()

	if x != 0 || y != 0 || w != 1024 || h != 768 {
		t.Errorf("unexpected viewport: expected (0, 0, 1024, 768), got (%d, %d, %d, %d)", x, y, w, h)
	}
}
