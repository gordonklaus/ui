package ui

type View interface {
	SetParent(View)

	Do(func())
	Size() Size
	Width() float64
	Height() float64
	Resize(Size)

	Draw(*Graphics)

	PointerDown(Pointer)
	PointerMove(Pointer)
	PointerUp(Pointer)

	Redraw()
}

type view struct {
	self   View
	parent View
	size   Size
}

func NewView(v View) View {
	return &view{
		self: v,
	}
}

func (v *view) Do(f func()) {
	if v.parent != nil {
		v.parent.Do(f)
	}
}

func (v *view) Size() Size      { return v.size }
func (v *view) Width() float64  { return v.size.Width }
func (v *view) Height() float64 { return v.size.Height }
func (v *view) Resize(s Size)   { v.size = s }

func (v *view) SetParent(parent View) {
	v.parent = parent
}

func (v *view) Draw(*Graphics) {}

func (v *view) PointerDown(p Pointer) {}
func (v *view) PointerMove(p Pointer) {}
func (v *view) PointerUp(p Pointer)   {}

func (v *view) Redraw() {
	if v.parent != nil {
		v.parent.Redraw()
	}
}
