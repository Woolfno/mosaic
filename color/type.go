package color

type Color struct {
	R uint32
	G uint32
	B uint32
}

func New(r uint32, g uint32, b uint32) *Color {
	return &Color{R: r, G: g, B: b}
}
