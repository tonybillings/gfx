package _test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"image/color"
	"testing"
	"tonysoft.com/gfx"
	"tonysoft.com/gfx/_test"
)

func TestAspectRatio(t *testing.T) {
	w := gfx.NewWindow()
	w.SetWidth(800)
	w.SetHeight(600)
	expected := float32(800) / float32(600)
	assert.Equal(t, expected, w.AspectRatio(), "expected AspectRatio %f, got %f", expected, w.AspectRatio())
}

func TestSetSize(t *testing.T) {
	w := gfx.NewWindow()
	w.SetSize(1024, 768)
	assert.Equal(t, int32(1024), w.Width(), "expected Width 1024, got %d", w.Width())
	assert.Equal(t, int32(768), w.Height(), "expected height 768, got %d", w.Height())
}

func TestScaleX1(t *testing.T) {
	w := gfx.NewWindow()
	w.SetSize(2000, 1000)
	assert.InEpsilon(t, .25, w.ScaleX(.5), 0.00001, "expected ScaleX .25, got %f", w.ScaleX(.5))
}

func TestScaleY1(t *testing.T) {
	w := gfx.NewWindow()
	w.SetSize(2000, 1000)
	assert.InEpsilon(t, .5, w.ScaleY(.5), 0.00001, "expected ScaleY .5, got %f", w.ScaleY(.5))
}

func TestScaleX2(t *testing.T) {
	w := gfx.NewWindow()
	w.SetSize(1000, 2000)
	assert.InEpsilon(t, .5, w.ScaleX(.5), 0.00001, "expected ScaleX .5, got %f", w.ScaleX(.5))
}

func TestScaleY2(t *testing.T) {
	w := gfx.NewWindow()
	w.SetSize(1000, 2000)
	assert.InEpsilon(t, .25, w.ScaleY(.5), 0.00001, "expected ScaleY .25, got %f", w.ScaleY(.5))
}

func TestScaleX3(t *testing.T) {
	w := gfx.NewWindow()
	w.SetSize(2000, 2000)
	assert.InEpsilon(t, .5, w.ScaleX(.5), 0.00001, "expected ScaleX .5, got %f", w.ScaleX(.5))
}

func TestScaleY3(t *testing.T) {
	w := gfx.NewWindow()
	w.SetSize(2000, 2000)
	assert.InEpsilon(t, .5, w.ScaleY(.5), 0.00001, "expected ScaleY .5, got %f", w.ScaleY(.5))
}

func TestWidthAndSetWidth(t *testing.T) {
	w := gfx.NewWindow()
	w.SetWidth(800)
	assert.Equal(t, int32(800), w.Width(), fmt.Sprintf("expected Width 800, got %d", w.Width()))
}

func TestHeightAndSetHeight(t *testing.T) {
	w := gfx.NewWindow()
	w.SetHeight(600)
	assert.Equal(t, int32(600), w.Height(), fmt.Sprintf("expected Height 600, got %d", w.Height()))
}

func TestClearColorAndSetClearColor(t *testing.T) {
	w := gfx.NewWindow()
	w.SetClearColor(gfx.Green)
	assert.Equal(t, color.RGBA{G: 255, A: 255}, w.ClearColor(), fmt.Sprintf("expected ClearColor %v, got %v", color.RGBA{G: 255, A: 255}, w.ClearColor()))
}

func TestSetTitle(t *testing.T) {
	w := gfx.NewWindow()
	w.SetTitle("Test Title")
	assert.Equal(t, "Test Title", w.Title(), fmt.Sprintf("expected Title 'Test Title', got '%s'", w.Title()))
}

func TestAddAndGetObject(t *testing.T) {
	w := gfx.NewWindow()
	obj := gfx.NewWindowObject()
	obj.SetName("Test Object")
	w.AddObject(obj)
	assert.Equal(t, obj, w.GetObject("Test Object"), "expected to get 'Test Object'")
}

func TestInitAndDisposeObject(t *testing.T) {
	_test.PanicOnErr(gfx.Init())

	w := gfx.NewWindow()
	obj := gfx.NewWindowObject().SetName("Test Object")

	ctx, cancelFunc := context.WithCancel(context.Background())
	w.Init(ctx, cancelFunc)
	<-w.ReadyChan()

	_test.SleepAFewFrames()

	assert.NotPanics(t, func() { w.InitObject(obj) }, "expected InitObject to not panic")
	assert.NotPanics(t, func() { w.DisposeObject("Test Object") }, "expected DisposeObject to not panic")

	w.Close()
	gfx.Close()
}

func TestInitAndCloseObjectAsync(t *testing.T) {
	_test.PanicOnErr(gfx.Init())

	w := gfx.NewWindow()
	obj := gfx.NewWindowObject()
	w.AddObject(obj)

	ctx, cancelFunc := context.WithCancel(context.Background())
	w.Init(ctx, cancelFunc)
	<-w.ReadyChan()
	_test.SleepAFewFrames()

	assert.True(t, obj.Initialized(), "expected object to be initialized")

	w.CloseObjectAsync(obj)
	_test.SleepNFrames(10)
	assert.False(t, obj.Initialized(), "expected object to not be initialized")

	w.Close()
	gfx.Close()
}

func TestRemoveObject(t *testing.T) {
	_test.PanicOnErr(gfx.Init())

	w := gfx.NewWindow()
	obj := gfx.NewWindowObject().SetName("Test Object")
	w.AddObject(obj)

	ctx, cancelFunc := context.WithCancel(context.Background())
	w.Init(ctx, cancelFunc)
	<-w.ReadyChan()
	_test.SleepAFewFrames()

	w.RemoveObject(obj.Name())
	_test.SleepAFewFrames()

	assert.Nil(t, w.GetObject(obj.Name()), "expected GetObject to return nil")
}

func TestDisposeAllObjects(t *testing.T) {
	_test.PanicOnErr(gfx.Init())

	w := gfx.NewWindow()
	obj1 := gfx.NewWindowObject().SetName("obj1")
	obj2 := gfx.NewWindowObject().SetName("obj2")
	obj3 := gfx.NewWindowObject().SetName("obj3")
	w.AddObjects(obj1, obj2, obj3)

	ctx, cancelFunc := context.WithCancel(context.Background())
	w.Init(ctx, cancelFunc)
	<-w.ReadyChan()

	_test.SleepAFewFrames()
	w.DisposeAllObjects()
	_test.SleepAFewFrames()

	assert.Nil(t, w.GetObject("obj1"), "expected nil, got an Object")
	assert.Nil(t, w.GetObject("obj2"), "expected nil, got an Object")
	assert.Nil(t, w.GetObject("obj3"), "expected nil, got an Object")

	w.Close()
	gfx.Close()
}

func TestAddAndRemoveService(t *testing.T) {
	w := gfx.NewWindow()

	service := gfx.NewAssetLibrary()
	service.SetName("Test Service")

	w.AddService(service)
	assert.Equal(t, service, w.GetService(service.Name()), "expected to find 'Test Service'")
	w.RemoveService(service)
	assert.Nil(t, w.GetService(service.Name()), "expected not to find 'Test Service'")
}

func TestInitAndDisposeProtectedServiceAsync(t *testing.T) {
	_test.PanicOnErr(gfx.Init())

	w := gfx.NewWindow()

	ctx, cancelFunc := context.WithCancel(context.Background())
	w.Init(ctx, cancelFunc)
	<-w.ReadyChan()

	_test.SleepAFewFrames()

	service := gfx.NewAssetLibrary()
	service.SetProtected(true)

	w.InitServiceAsync(service)
	_test.SleepNFrames(10)
	assert.True(t, service.Initialized(), "expected service to be initialized")

	w.DisposeServiceAsync(service)
	_test.SleepNFrames(10)
	assert.True(t, service.Initialized(), "expected service to still be initialized (protected=true)")

	service.SetProtected(false)
	w.DisposeServiceAsync(service)
	_test.SleepNFrames(10)
	assert.False(t, service.Initialized(), "expected service to no longer be initialized (protected=false)")

	w.Close()
	gfx.Close()
}

func TestDisposeAllServicesAsync(t *testing.T) {
	_test.PanicOnErr(gfx.Init())

	w := gfx.NewWindow()

	services := []gfx.Service{gfx.NewAssetLibrary(), gfx.NewAssetLibrary()}
	for _, svc := range services {
		w.AddService(svc)
	}
	services[0].SetProtected(true)

	ctx, cancelFunc := context.WithCancel(context.Background())
	w.Init(ctx, cancelFunc)
	<-w.ReadyChan()

	_test.SleepAFewFrames()
	w.DisposeAllServicesAsync(true)
	_test.SleepNFrames(10)

	for _, svc := range services {
		assert.False(t, svc.Initialized(), "expected all services to not be initialized")
	}

	w.Close()
	gfx.Close()
}
