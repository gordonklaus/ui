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

type Size struct {
	Width, Height float64
}

func (s Size) Mul(x float64) Size {
	return Size{
		Width:  s.Width * x,
		Height: s.Height * x,
	}
}
