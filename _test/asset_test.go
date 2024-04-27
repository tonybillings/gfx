package _test

import (
	"testing"
	"tonysoft.com/gfx"
)

func TestAssetLibrarySetName(t *testing.T) {
	lib := gfx.NewAssetLibrary()
	lib.SetName("NewName")
	if lib.Name() != "NewName" {
		t.Errorf("expected Name %s, got %s", "NewName", lib.Name())
	}
}

func TestAssetLibraryAddAndGet(t *testing.T) {
	lib := gfx.NewAssetLibrary()
	asset := gfx.NewAssetBase("test_asset", nil)
	lib.Add(asset)

	retrieved := lib.Get("test_asset")
	if retrieved == nil || retrieved.Name() != "test_asset" {
		t.Errorf("expected to retrieve 'test_asset', got %v", retrieved)
	}
}

func TestAssetLibraryRemove(t *testing.T) {
	lib := gfx.NewAssetLibrary()
	asset := gfx.NewAssetBase("test_asset", nil)
	lib.Add(asset)
	lib.Remove("test_asset")

	if lib.Get("test_asset") != nil {
		t.Error("expected asset 'test_asset' to not be found")
	}
}

func TestAssetLibraryDispose(t *testing.T) {
	lib := gfx.NewAssetLibrary()
	asset := gfx.NewAssetBase("test_asset", nil)
	lib.Add(asset)
	lib.Dispose("test_asset")

	if lib.Get("test_asset") != nil {
		t.Error("expected asset 'test_asset' to not be found")
	}
}

func TestAssetLibraryDisposeProtectedAsset(t *testing.T) {
	lib := gfx.NewAssetLibrary()
	asset := gfx.NewAssetBase("test_asset", nil)
	asset.SetProtected(true)
	lib.Add(asset)
	lib.Dispose("test_asset")

	if lib.Get("test_asset") == nil {
		t.Error("expected protected asset 'test_asset' to not have been disposed")
	}
}

func TestAssetLibraryDisposeAll(t *testing.T) {
	lib := gfx.NewAssetLibrary()
	asset1 := gfx.NewAssetBase("asset1", nil)
	asset2 := gfx.NewAssetBase("asset2", nil)
	asset2.SetProtected(true)

	lib.Add(asset1).Add(asset2)
	lib.DisposeAll()

	if lib.Get("asset1") != nil {
		t.Error("expected unprotected asset 'asset1' to have been disposed")
	}
	if lib.Get("asset2") == nil {
		t.Error("expected protected asset 'asset2' to not have been disposed")
	}
}

func TestAssetLibraryGetReader(t *testing.T) {
	lib := gfx.NewAssetLibrary()
	data := []byte("hello world")
	asset := gfx.NewAssetBase("test_asset", data)
	lib.Add(asset)

	reader, closeFunc := lib.GetReader("test_asset")
	if reader == nil {
		t.Error("expected a reader for 'test_asset', got nil")
	}
	if closeFunc == nil {
		t.Error("expected a non-nil close function")
	}

	readData := make([]byte, len(data))
	if _, err := reader.Read(readData); err != nil {
		t.Error(err)
	} else {
		if string(readData) != string(data) {
			t.Errorf("expected to read %s, got %s", string(data), string(readData))
		}
	}

	closeFunc()
}

func TestAssetLibraryGetReaderNonExistent(t *testing.T) {
	lib := gfx.NewAssetLibrary()
	reader, closeFunc := lib.GetReader("non_existent")

	if reader != nil {
		t.Error("expected nil reader for non-existent asset")
	}
	if closeFunc == nil {
		t.Error("expected non-nil close function for safety")
	}

	closeFunc()
}
