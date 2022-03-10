package animation

type Keyframe struct {
	Percentage NumberOrPercentage `starlark:"percentage,required"`
	Transforms []Transform        `starlark:"transforms,required"`
}
