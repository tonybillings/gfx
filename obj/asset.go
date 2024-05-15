package obj

import (
	"bufio"
	"fmt"
	"github.com/tonybillings/gfx"
	"os"
)

func getSourceReader(asset gfx.Asset, sourceName string) (reader *bufio.Reader, closeFunc func()) {
	closeFunc = func() {}
	if srcLib := asset.SourceLibrary(); srcLib == nil {
		if len(sourceName) > 200 {
			return
		}
		if _, err := os.Stat(sourceName); err != nil && os.IsNotExist(err) {
			return
		}
		if file, err := os.Open(sourceName); err != nil {
			panic(fmt.Errorf("open file error: %w", err))
		} else {
			reader = bufio.NewReader(file)
			closeFunc = func() {
				_ = file.Close()
			}
			return
		}
	} else {
		return srcLib.GetFileReader(sourceName)
	}
}
