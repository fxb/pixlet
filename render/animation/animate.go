package animation

import (
	"image"
	"sort"

	"github.com/fogleman/gg"

	"tidbyt.dev/pixlet/render"
)

func adjacentKeyframes(arr []Keyframe, percentage float64) (Keyframe, Keyframe) {
	for i := len(arr) - 1; i >= 0; i-- {
		if arr[i].Percentage.Transform(1.0) > percentage {
			continue
		} else if i+1 >= len(arr) {
			return arr[i], arr[i]
		} else {
			return arr[i], arr[i+1]
		}
	}

	return Keyframe{}, Keyframe{}
}

// Animate is a widget that does keyframe animations by interpolating
// between transforms for a child widget.
//
// It supports animating translation, scale and rotation of its child.
//
// DOC(Child): Widget to animate
// DOC(Keyframes): List of animation keyframes
// DOC(Duration): Duration of animation in frames
// DOC(Delay): Delay of animation in frames
// DOC(Origin): Origin for transforms, default is '50%, 50%'
// DOC(Curve): Easing curve to use, default is 'linear'
// DOC(Rounding): Rounding to use for interpolated translation coordinates (not used for scale and rotate), default is 'round'
//
// EXAMPLE BEGIN
// render.Animate(
//   child = render.Box(render.Circle(diameter = 6, color = "#0f0")),
//   duration = 100,
//   delay = 0,
//   curve = "linear",
//   origin = render.Origin("50%", "50%"),
//   keyframes = [
//     render.Keyframe("from", [render.Rotate(0), render.Translate(-10, 0), render.Rotate(0)]),
//     render.Keyframe("to", [render.Rotate(360), render.Translate(-10, 0), render.Rotate(-360)]),
//   ],
// ),
// EXAMPLE END
type Animate struct {
	render.Widget
	Child     render.Widget `starlark:"child,required"`
	Keyframes []Keyframe    `starlark:"keyframes,required"`
	Duration  int           `starlark:"duration,required"`
	Delay     int           `starlark:"delay"`
	Origin    Origin        `starlark:"origin"`
	Curve     Curve         `starlark:"curve"`
	Rounding  Rounding      `starlark:"rounding"`
}

func (a *Animate) Init() {
	sort.SliceStable(a.Keyframes, func(i, j int) bool {
		return a.Keyframes[i].Percentage.Transform(1.0) <
			a.Keyframes[j].Percentage.Transform(1.0)
	})
}

func (a *Animate) FrameCount() int {
	return a.Delay + a.Duration
}

func (a *Animate) Paint(bounds image.Rectangle, frameIdx int) image.Image {
	img := a.Child.Paint(bounds, frameIdx)
	ctx := gg.NewContext(bounds.Dx(), bounds.Dy())
	origin := a.Origin.Transform(img.Bounds())

	var progress float64
	if frameIdx < a.Delay {
		progress = a.Curve.Transform(0.0)
	} else if frameIdx >= a.FrameCount() {
		progress = a.Curve.Transform(0.0)
	} else {
		progress = a.Curve.Transform(float64(frameIdx-a.Delay) / float64(a.Duration-1))
	}

	from, to := adjacentKeyframes(a.Keyframes, progress)
	progress = Rescale(
		from.Percentage.Transform(1.0),
		to.Percentage.Transform(1.0),
		0.0,
		1.0,
		progress)

	for _, transform := range InterpolateTransforms(from.Transforms, to.Transforms, progress) {
		transform.Apply(ctx, origin, a.Rounding)
	}

	ctx.DrawImage(img, 0, 0)

	return ctx.Image()
}
