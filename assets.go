package gfx

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	loadedAssets      = make(map[string][]byte)
	loadedAssetsMutex sync.Mutex
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func GetAsset(name string) []byte {
	loadedAssetsMutex.Lock()
	if a, ok := loadedAssets[name]; ok {
		loadedAssetsMutex.Unlock()
		return a
	}
	loadedAssetsMutex.Unlock()
	return nil
}

func GetAssetReader(name string) io.Reader {
	asset := GetAsset(name)
	if asset == nil {
		return bytes.NewReader([]byte{})
	}
	return bytes.NewReader(asset)
}

func AddAsset(name string, asset []byte) {
	loadedAssetsMutex.Lock()
	loadedAssets[name] = asset
	loadedAssetsMutex.Unlock()
}

func AddEmbeddedAsset(name string, fs embed.FS) {
	loadedAssetsMutex.Lock()
	asset, err := fs.Open(name)
	if err != nil {
		panic(fmt.Errorf("error opening embedded asset: %w", err))
	}
	assetBytes, err := io.ReadAll(asset)
	if err != nil {
		panic(fmt.Errorf("error reading embedded asset: %w", err))
	}
	loadedAssets[name] = assetBytes
	loadedAssetsMutex.Unlock()
}
