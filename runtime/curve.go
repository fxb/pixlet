package runtime

import (
	"fmt"

	"go.starlark.net/starlark"

	"tidbyt.dev/pixlet/render/animation"
)

func CurveFromStarlark(value starlark.Value) (animation.Curve, error) {
	if str, ok := starlark.AsString(value); ok {
		return animation.ParseCurve(str)
	} else if fn, ok := value.(*starlark.Function); ok {
		if fn.NumParams() != 1 || fn.NumKwonlyParams() != 0 {
			return nil, fmt.Errorf("invalid number of parameters to curve function: %s", fn.String())
		}

		return animation.NewCustomCurve(fn), nil
	}

	return nil, fmt.Errorf("invalid type for curve: %s (expected string or function)", value.Type())
}
