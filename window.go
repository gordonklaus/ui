package ui

type Window interface {
	View
}

func NewWindow(size Size, v View) (Window, error) {
	return newWindow(size, v)
}
