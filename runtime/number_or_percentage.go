package runtime

import (
	"fmt"

	"go.starlark.net/starlark"

	"tidbyt.dev/pixlet/render/animation"
)

func NumberOrPercentageFromStarlark(value starlark.Value, min, max float64, mapping map[string]float64) (animation.NumberOrPercentage, error) {
	if val, ok := starlark.AsFloat(value); ok {
		if min <= val && val <= max {
			return animation.Number{Value: val}, nil
		} else {
			return nil, fmt.Errorf("invalid range for number: %f (expected number in range [0.0, 1.0])", val)
		}
	} else if str, ok := starlark.AsString(value); ok {
		return animation.ParsePercentage(str, mapping)
	}

	return nil, fmt.Errorf("invalid type for number or percentage: %s (expected number or string)", value.Type())
}
