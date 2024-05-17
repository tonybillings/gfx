package gfx

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	defaultAssetLibraryName = "AssetLibrary"
)

/******************************************************************************
 Asset
******************************************************************************/

// Asset Resources that are needed/shared by objects, have a separate
// lifecycle than objects, and that are initialized before objects, are
// considered to be assets.  These include resources like fonts, shaders,
// textures, materials, and models.  It's recommended to use an AssetLibrary
// to manage a collection of related assets to facilitate life-cycle management.
type Asset interface {
	Initer
	Closer

	// Initialized shall return true if Init() has already been called.
	Initialized() bool

	// Name shall return the unique name/id given to the asset.
	Name() string

	// Source shall return whatever was used to construct the asset or
	// what otherwise represents the asset itself.  For example, with a
	// Texture asset the source could be the raw binary for a PNG file,
	// which would also be the case when loading the PNG file as an
	// asset itself for other purposes, etc.
	Source() any

	// SourceLibrary shall return the library this asset can use
	// to look for a source that is needed for construction.
	SourceLibrary() *AssetLibrary
	SetSourceLibrary(*AssetLibrary) Asset

	// Protected shall return true if this asset has been marked
	// for protection from closure/removal while being managed
	// by an AssetLibrary. It is used, for example, to allow
	// consumers to freely call DisposeAll() on the Window's
	// default AssetLibrary without concern that it will affect
	// the default assets added therein.
	Protected() bool
	SetProtected(bool) Asset
}

/******************************************************************************
 AssetBase
******************************************************************************/

type AssetBase struct {
	initialized atomic.Bool
	srcLibrary  *AssetLibrary
	name        string
	source      any
	protected   atomic.Bool
}

/******************************************************************************
 Asset Implementation
******************************************************************************/

func (a *AssetBase) Init() bool {
	a.initialized.Store(true)
	return true
}

func (a *AssetBase) Close() {
	a.initialized.Store(false)
}

func (a *AssetBase) Initialized() bool {
	return a.initialized.Load()
}

func (a *AssetBase) Name() string {
	return a.name
}

func (a *AssetBase) Source() any {
	return a.source
}

func (a *AssetBase) SourceLibrary() *AssetLibrary {
	return a.srcLibrary
}

func (a *AssetBase) SetSourceLibrary(library *AssetLibrary) Asset {
	a.srcLibrary = library
	return a
}

func (a *AssetBase) Protected() bool {
	return a.protected.Load()
}

func (a *AssetBase) SetProtected(protected bool) Asset {
	a.protected.Store(protected)
	return a
}

/******************************************************************************
 AssetBase Functions
******************************************************************************/

func (a *AssetBase) getSourceReader(name string) (reader *bufio.Reader, closeFunc func()) {
	closeFunc = func() {}
	if srcLib := a.SourceLibrary(); srcLib == nil {
		if len(name) > 200 {
			return
		}
		if _, err := os.Stat(name); err != nil && os.IsNotExist(err) {
			return
		}
		if file, err := os.Open(name); err != nil {
			panic(fmt.Errorf("open file error: %w", err))
		} else {
			reader = bufio.NewReader(file)
			closeFunc = func() {
				_ = file.Close()
			}
			return
		}
	} else {
		return srcLib.GetFileReader(name)
	}
}

func NewAssetBase(name string, source any) *AssetBase {
	return &AssetBase{
		name:   name,
		source: source,
	}
}

/******************************************************************************
 BinaryAsset
******************************************************************************/

type BinaryAsset struct {
	AssetBase
}

func NewBinaryAsset(name string, data []byte) *BinaryAsset {
	return &BinaryAsset{
		AssetBase: AssetBase{
			name:   name,
			source: data,
		},
	}
}

/******************************************************************************
 GlAsset
******************************************************************************/

// GlAsset assets are those with an associated OpenGL name, which is an
// unsigned number that is assigned with creating the object (usually one
// taking up memory in VRAM).
type GlAsset interface {
	Asset

	// GlName shall return the unique name/id assigned to the asset by the
	// OpenGL framework.
	GlName() uint32
}

/******************************************************************************
 Asset Sources
******************************************************************************/

type FontSource interface {
	[]byte | string
}

type ShaderSource interface {
	[]byte | string
}

type ModelSource interface {
	[]byte | string
}

type MaterialLibrarySource interface {
	[]byte | string
}

type TextureSource interface {
	[]byte | string | color.RGBA | *image.RGBA | *image.NRGBA
}

/******************************************************************************
 AssetLibrary
******************************************************************************/

type AssetLibrary struct {
	ServiceBase

	assets     map[string]Asset
	initQueue  []Asset
	closeQueue []Asset

	stateMutex   sync.Mutex
	stateChanged atomic.Bool
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (l *AssetLibrary) Init() bool {
	l.stateMutex.Lock()

	if l.Initialized() {
		l.stateMutex.Unlock()
		return true
	}

	l.initAssets()
	l.ServiceBase.Init()
	l.stateMutex.Unlock()

	return true
}

func (l *AssetLibrary) Update(_ int64) bool {
	if !l.Initialized() {
		return false
	}

	if l.stateChanged.Load() {
		l.stateMutex.Lock()
		l.initAssets()
		l.closeAssets()
		l.stateMutex.Unlock()
		l.stateChanged.Store(false)
	}

	return true
}

func (l *AssetLibrary) Close() {
	if l.Protected() || !l.Initialized() {
		return
	}

	l.stateMutex.Lock()
	l.closeAllAssets()
	l.ServiceBase.Close()
	l.stateMutex.Unlock()
}

/******************************************************************************
 AssetLibrary Functions
******************************************************************************/

func (l *AssetLibrary) initAssets() {
	for i := len(l.initQueue) - 1; i >= 0; i-- {
		l.initQueue[i].Init()
		l.initQueue = l.initQueue[:i]
	}
}

func (l *AssetLibrary) closeAssets() {
	for i := len(l.closeQueue) - 1; i >= 0; i-- {
		l.closeQueue[i].Close()
		l.closeQueue = l.closeQueue[:i]
	}
}

func (l *AssetLibrary) closeAllAssets() {
	for _, asset := range l.assets {
		asset.SetProtected(false)
		asset.Close()
	}
}

func (l *AssetLibrary) setSourceLibrary(asset Asset) {
	if lib := asset.SourceLibrary(); lib == nil {
		asset.SetSourceLibrary(l)
	}
}

func (l *AssetLibrary) Get(name string) Asset {
	return l.assets[name]
}

func (l *AssetLibrary) Add(asset Asset) *AssetLibrary {
	l.stateMutex.Lock()
	l.setSourceLibrary(asset)
	if existing, ok := l.assets[asset.Name()]; ok {
		if existing.Protected() {
			l.stateMutex.Unlock()
			return l
		}
		l.closeQueue = append(l.closeQueue, existing)
	}
	l.assets[asset.Name()] = asset
	l.initQueue = append(l.initQueue, asset)
	l.stateMutex.Unlock()
	l.stateChanged.Store(true)
	return l
}

func (l *AssetLibrary) AddEmbeddedFile(name string, fs embed.FS) *AssetLibrary {
	if asset, err := fs.ReadFile(name); err != nil {
		panic(fmt.Errorf("error opening embedded asset: %w", err))
	} else {
		l.stateMutex.Lock()
		if existing, ok := l.assets[name]; ok {
			if existing.Protected() {
				l.stateMutex.Unlock()
				return l
			}
			l.closeQueue = append(l.closeQueue, existing)
		}
		fileAsset := NewBinaryAsset(name, asset)
		fileAsset.SetSourceLibrary(l)
		l.assets[name] = fileAsset
		l.initQueue = append(l.initQueue, l.assets[name])
		l.stateMutex.Unlock()
		l.stateChanged.Store(true)
	}
	return l
}

func (l *AssetLibrary) AddEmbeddedFiles(fs embed.FS) *AssetLibrary {
	if assets, err := fs.ReadDir("."); err != nil {
		panic(fmt.Errorf("error opening embedded file system: %w", err))
	} else {
		for _, asset := range assets {
			if !asset.IsDir() && !strings.HasSuffix(asset.Name(), ".go") {
				l.AddEmbeddedFile(asset.Name(), fs)
			}
		}
	}
	return l
}

func (l *AssetLibrary) Remove(name string) *AssetLibrary {
	if name == "" {
		return l
	}

	l.stateMutex.Lock()
	if existing, ok := l.assets[name]; ok {
		if existing.Protected() {
			l.stateMutex.Unlock()
			return l
		}
		delete(l.assets, name)
	}
	l.stateMutex.Unlock()
	return l
}

func (l *AssetLibrary) RemoveAll() *AssetLibrary {
	l.stateMutex.Lock()
	for name, existing := range l.assets {
		if existing.Protected() {
			continue
		}
		delete(l.assets, name)
	}
	l.stateMutex.Unlock()
	return l
}

func (l *AssetLibrary) Dispose(name string) *AssetLibrary {
	if name == "" {
		return l
	}

	l.stateMutex.Lock()
	if existing, ok := l.assets[name]; ok {
		if existing.Protected() {
			l.stateMutex.Unlock()
			return l
		}
		l.closeQueue = append(l.closeQueue, existing)
		delete(l.assets, name)
	}
	l.stateMutex.Unlock()
	l.stateChanged.Store(true)
	return l
}

func (l *AssetLibrary) DisposeAll() *AssetLibrary {
	l.stateMutex.Lock()
	for name, existing := range l.assets {
		if existing.Protected() {
			continue
		}
		l.closeQueue = append(l.closeQueue, existing)
		delete(l.assets, name)
	}
	l.stateMutex.Unlock()
	l.stateChanged.Store(true)
	return l
}

func (l *AssetLibrary) GetFileReader(name string) (reader *bufio.Reader, closeFunc func()) {
	closeFunc = func() {}
	asset := l.Get(name)
	if asset != nil {
		if assetBytes, ok := asset.Source().([]byte); ok {
			reader = bufio.NewReader(bytes.NewReader(assetBytes))
			return
		}
	}

	if len(name) > 200 {
		return
	}
	if _, err := os.Stat(name); err != nil && os.IsNotExist(err) {
		return
	}
	if file, err := os.Open(name); err != nil {
		panic(fmt.Errorf("open file error: %w", err))
	} else {
		reader = bufio.NewReader(file)
		closeFunc = func() {
			_ = file.Close()
		}
		return
	}

	return
}

/******************************************************************************
 New AssetLibrary Function
******************************************************************************/

func NewAssetLibrary() *AssetLibrary {
	l := &AssetLibrary{
		assets: make(map[string]Asset),
	}
	l.SetName(defaultAssetLibraryName)
	return l
}

/******************************************************************************
 Default AssetLibrary
******************************************************************************/

func DefaultAssetLibrary() *AssetLibrary {
	lib := NewAssetLibrary()
	lib.SetName("_assets")
	lib.SetProtected(true)
	addDefaultShaders(lib)
	addDefaultFonts(lib)
	return lib
}
