package ui

type Position struct {
	X, Y float64
}

func (p Position) Add(s Size) Position {
	return Position{
		X: p.X + s.Width,
		Y: p.Y + s.Height,
	}
}

func (p Position) Sub(q Position) Size {
	return Size{
		Width:  p.X - q.X,
		Height: p.Y - q.Y,
	}
}

func (p Position) In(r Rectangle) bool {
	return p.X >= r.Min.X && p.X < r.Max.X && p.Y >= r.Min.Y && p.Y < r.Max.Y
}

type Size struct {
	Width, Height float64
}

func (s Size) Mul(x float64) Size {
	return Size{
		Width:  s.Width * x,
		Height: s.Height * x,
	}
}

type Rectangle struct {
	Min, Max Position
}

func (r Rectangle) Size() Size      { return r.Max.Sub(r.Min) }
func (r Rectangle) Width() float64  { return r.Max.X - r.Min.X }
func (r Rectangle) Height() float64 { return r.Max.Y - r.Min.Y }

func (r Rectangle) Bounds() (xMin, xMax, yMin, yMax float64) {
	return r.Min.X, r.Max.X, r.Min.Y, r.Max.Y
}
