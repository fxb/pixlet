//go:build wasm

package render

import (
	"fmt"
)

func (p *Image) InitFromWebP(data []byte) error {
	return fmt.Errorf("WebP decoding is not implemented for WASM target")
}
