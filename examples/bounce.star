load("render.star", "render")
load("math.star", "math")
load("encoding/base64.star", "base64")

def linear(x):
    return x

def easeInElastic(x):
    c4 = (2.0 * math.pi) / 3.0

    if x == 0:
        return 0
    elif x == 1:
        return 1

    return -math.pow(2, 10 * x - 10) * math.sin((x * 10 - 10.75) * c4)

def easeOutBounce(x):
    n1 = 7.5625
    d1 = 2.75

    if x < 1 / d1:
        return n1 * x * x
    elif x < 2 / d1:
        x = x - (1.5 / d1)
        return n1 * (x) * x + 0.75
    elif x < 2.5 / d1:
        x = x - (2.25 / d1)
        return n1 * (x) * x + 0.9375
    else:
        x = x - (2.625 / d1)
        return n1 * (x) * x + 0.984375

EARTH = render.Image(
    base64.decode("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"),
    width = 8,
    height = 8,
)

def main():
    #return render.Root(
    #    child = render.Box(Loader(), color = "#222"),
    #    delay = 20,
    #)
    #return render.Root(
    #    child = Orbit(child, 12),
    #    delay = 20,
    #)
    #return render.Root(
    #    child = Bounce("Hello", 64, 16, 20),
    #    delay = 20,
    #)
    duration_ms = 2000
    delay_ms = 50
    duration_frames = int(math.round(duration_ms / delay_ms))

    return render.Root(
        child = render.Column([
            render.Marquee(render.Text(content = "Hello World!    ", height = 8), width = 64),
            render.Marquee(render.Text(content = "Hello Tidbyt! Whats up!", height = 8), width = 64),
            Marquee("Hello World!    ", 64, 8),
            Marquee("Hello Tidbyt! Whats up!", 64, 8),
        ]),
        delay = delay_ms,
    )

def Abs(x):
    return x if x > 0 else -x

def Sign(x):
    return 1 if x > 0 else -1

def Loader():
    diameter = 5
    distance = 5

    return render.Box(
        render.Stack(children = [
            render.Animate(
                child = render.Box(render.Circle(diameter = diameter, color = "#f00")),
                duration = 100,
                delay = 0,
                curve = "ease_in_out",
                origin = render.Origin("50%", "50%"),
                keyframes = [
                    render.Keyframe("from", [render.Rotate(0), render.Translate(-distance, 0), render.Rotate(0)]),
                    render.Keyframe("to", [render.Rotate(360), render.Translate(-distance, 0), render.Rotate(-360)]),
                ],
            ),
            render.Animate(
                child = render.Box(render.Circle(diameter = diameter, color = "#0f0")),
                duration = 100,
                delay = 10,
                curve = "ease_in_out",
                origin = render.Origin("50%", "50%"),
                keyframes = [
                    render.Keyframe("from", [render.Rotate(0), render.Translate(-distance, 0), render.Rotate(0)]),
                    render.Keyframe("to", [render.Rotate(360), render.Translate(-distance, 0), render.Rotate(-360)]),
                ],
            ),
            render.Animate(
                child = render.Box(render.Circle(diameter = diameter, color = "#00f")),
                duration = 100,
                delay = 20,
                curve = "ease_in_out",
                origin = render.Origin("50%", "50%"),
                keyframes = [
                    render.Keyframe("from", [render.Rotate(0), render.Translate(-distance, 0), render.Rotate(0)]),
                    render.Keyframe("to", [render.Rotate(360), render.Translate(-distance, 0), render.Rotate(-360)]),
                ],
            ),
            render.Animate(
                child = render.Box(render.Circle(diameter = diameter, color = "#f0f")),
                duration = 100,
                delay = 30,
                curve = "ease_in_out",
                origin = render.Origin("50%", "50%"),
                keyframes = [
                    render.Keyframe("from", [render.Rotate(0), render.Translate(-distance, 0), render.Rotate(0)]),
                    render.Keyframe("to", [render.Rotate(360), render.Translate(-distance, 0), render.Rotate(-360)]),
                ],
            ),
            render.Animate(
                child = render.Box(render.Circle(diameter = diameter, color = "#ff0")),
                duration = 100,
                delay = 40,
                curve = "ease_in_out",
                origin = render.Origin("50%", "50%"),
                keyframes = [
                    render.Keyframe("from", [render.Rotate(0), render.Translate(-distance, 0), render.Rotate(0)]),
                    render.Keyframe("to", [render.Rotate(360), render.Translate(-distance, 0), render.Rotate(-360)]),
                ],
            ),
            render.Animate(
                child = render.Box(render.Circle(diameter = diameter, color = "#0ff")),
                duration = 100,
                delay = 50,
                curve = "ease_in_out",
                origin = render.Origin("50%", "50%"),
                keyframes = [
                    render.Keyframe("from", [render.Rotate(0), render.Translate(-distance, 0), render.Rotate(0)]),
                    render.Keyframe("to", [render.Rotate(360), render.Translate(-distance, 0), render.Rotate(-360)]),
                ],
            ),
        ]),
        width = 16,
        height = 16,
    )

def Orbit(child, distance):
    return render.Animate(
        child = render.Box(child),
        duration = 100,
        delay = 0,
        curve = "linear",
        origin = render.Origin("50%", "50%"),
        keyframes = [
            render.Keyframe("from", [render.Rotate(0), render.Translate(-distance, 0), render.Rotate(0)]),
            render.Keyframe("to", [render.Rotate(360), render.Translate(-distance, 0), render.Rotate(-360)]),
        ],
    )

def Bounce(str, width, height, hold = 0, always = True):
    text = render.Text(height = height, content = str)

    diff = Abs(width - text.width)
    sign = Sign(width - text.width)
    duration = 2 * diff + hold
    offset = (hold / 2) / duration

    if not always and text.width < width:
        return text

    return render.Box(width = width, height = height, child = render.Animate(
        child = text,
        duration = duration,
        delay = hold,
        curve = "linear",
        keyframes = [
            render.Keyframe(0.0, [render.Translate(0, 0)]),
            render.Keyframe(0.5 - offset, [render.Translate(sign * diff, 0)]),
            render.Keyframe(0.5 + offset, [render.Translate(sign * diff, 0)]),
            render.Keyframe(1.0, [render.Translate(1, 0)]),
        ],
    ))

def Marquee(str, width, height, delay = 0, always = True):
    text = render.Text(height = height, content = str)
    duration = (text.width + 1 + width)
    mid = (text.width + 1) / duration

    if not always and text.width < width:
        return text

    return render.Box(width = width, height = height, child = render.Animate(
        child = text,
        duration = duration,
        delay = delay,
        curve = "linear",
        rounding = "floor",
        keyframes = [
            render.Keyframe(0.0, [render.Translate(0, 0)]),
            render.Keyframe(mid, [render.Translate(-text.width, 0)]),
            render.Keyframe(mid, [render.Translate(width, 0)]),
            render.Keyframe(1.0, [render.Translate(1, 0)]),
        ],
    ))

def main_bounce():
    return render.Root(
        child = render.Box(
            render.Bounce(
                render.Row(children = [
                    render.Box(width = 1, height = 3, color = "#f00"),
                    render.Box(width = 1, height = 3, color = "#0f0"),
                    render.Box(width = 1, height = 3, color = "#00f"),
                ]),
                width = 64,
                bounce_direction = "horizontal",
                bounce_always = True,
                #curve = "ease_in",
                curve = easeOutBounce,
            ),
            width = 64,
            height = 3,
            color = "#fff3",
        ),
        delay = 100,
    )
