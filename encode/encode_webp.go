//go:build !wasm

package encode

import (
	"time"

	"github.com/pkg/errors"
	"github.com/tidbyt/go-libwebp/webp"
)

const (
	WebPKMin = 0
	WebPKMax = 0
)

// Renders a screen to WebP. Optionally pass filters for
// postprocessing each individual frame.
func (s *Screens) EncodeWebP(filters ...ImageFilter) ([]byte, error) {
	images, err := s.render(filters...)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return []byte{}, nil
	}

	bounds := images[0].Bounds()
	anim, err := webp.NewAnimationEncoder(
		bounds.Dx(),
		bounds.Dy(),
		WebPKMin,
		WebPKMax,
	)
	if err != nil {
		return nil, errors.Wrap(err, "initializing encoder")
	}
	defer anim.Close()

	frameDuration := time.Duration(s.delay) * time.Millisecond
	for _, im := range images {
		if err := anim.AddFrame(im, frameDuration); err != nil {
			return nil, errors.Wrap(err, "adding frame")
		}
	}

	buf, err := anim.Assemble()
	if err != nil {
		return nil, errors.Wrap(err, "encoding animation")
	}

	return buf, nil
}
