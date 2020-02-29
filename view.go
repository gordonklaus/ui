package ui

type View interface {
	view() *view

	Parent() View
	SetParent(View)

	Do(func())

	// SizePolicy() SizePolicy
	// SetSizePolicy(SizePolicy)

	// Position in parent's coordinate system.
	Position() Position
	Move(Position)

	// Size in parent's coordinate system.
	Size() Size
	Resize(Size)

	// Rect defines the internal coordinate system.
	// If unset, defaults to a rectangle with Min == (0, 0) and Size == v.Size().
	Rect() Rectangle
	SetRect(Rectangle)

	ViewAt(Position) View

	MapToParent(Position) Position
	MapFromParent(Position) Position

	Draw(*Graphics)

	PointerDown(Pointer)
	PointerMove(Pointer)
	PointerUp(Pointer)

	Redraw()
}

type SizePolicy struct {
	Min, Max, Preferred Size
}

type view struct {
	self     View
	parent   *view
	children []View
	position Position
	size     Size
	rect     Rectangle

	transformToWindow      transform
	transformToWindowValid bool
}

func NewView(self, parent View) View {
	v := &view{
		self: self,
	}

	if parent != nil {
		v.parent = parent.view()
		v.parent.children = append(v.parent.children, self)
	}

	return v
}

func (v *view) view() *view { return v }

func (v *view) Parent() View { return v.parent.self }
func (v *view) SetParent(p View) {
	if v.parent != nil {
		for i, c := range v.parent.children {
			if c == v.self {
				v.parent.children = append(v.parent.children[:i], v.parent.children[i+1:]...)
				break
			}
		}
	}
	if p != nil {
		v.parent = p.view()
		v.parent.children = append(v.parent.children, v.self)
	} else {
		v.parent = nil
	}
	v.invalidateTransformToWindow()
}

func (v *view) Do(f func()) {
	if v.parent != nil {
		v.parent.Do(f)
	}
}

func (v *view) Position() Position { return v.position }
func (v *view) Move(p Position) {
	v.position = p
	v.invalidateTransformToWindow()
}

func (v *view) Size() Size { return v.size }
func (v *view) Resize(s Size) {
	v.size = s
	v.invalidateTransformToWindow()
}

func (v *view) Rect() Rectangle {
	if v.rect == (Rectangle{}) {
		return Rectangle{Max: Position{v.size.Width, v.size.Height}}
	}
	return v.rect
}
func (v *view) SetRect(r Rectangle) {
	v.rect = r
	v.invalidateTransformToWindow()
}

func (v *view) draw(gfx *Graphics) {
	gfx.setViewTransform(v.getTransformToWindow())
	v.self.Draw(gfx)
	for _, v := range v.children {
		v.view().draw(gfx)
	}
}

func (v *view) Draw(*Graphics) {}

func (v *view) PointerDown(p Pointer) {}
func (v *view) PointerMove(p Pointer) {}
func (v *view) PointerUp(p Pointer)   {}

func (v *view) Redraw() {
	if v.parent != nil {
		v.parent.self.Redraw()
	}
}

func (v *view) ViewAt(p Position) View {
	if !p.In(v.Rect()) {
		return nil
	}
	for _, child := range v.children {
		if v := child.ViewAt(child.MapFromParent(p)); v != nil {
			return v
		}
	}
	return v.self
}

func (v *view) mapFromWindow(p Position) Position {
	return v.getTransformToWindow().invert().transform(p)
}

func (v *view) MapToParent(p Position) Position {
	return v.getTransformToParent().transform(p)
}

func (v *view) MapFromParent(p Position) Position {
	return v.getTransformToParent().invert().transform(p)
}

func (v *view) invalidateTransformToWindow() {
	v.transformToWindowValid = false
	for _, v := range v.children {
		v.view().invalidateTransformToWindow()
	}
	v.Redraw()
}

func (v *view) getTransformToWindow() transform {
	if v == nil {
		return identityTransform()
	}

	if !v.transformToWindowValid {
		v.transformToWindow = v.getTransformToParent().compose(v.parent.getTransformToWindow())
		v.transformToWindowValid = true
	}

	return v.transformToWindow
}

func (v *view) getTransformToParent() transform {
	t := identityTransform()
	if v.rect != (Rectangle{}) {
		t = t.translate(-v.rect.Min.X, -v.rect.Min.Y)
		t = t.scale(v.size.Width/v.rect.Width(), v.size.Height/v.rect.Height())
	}
	return t.translate(v.position.X, v.position.Y)
}

type transform struct {
	scaleX, scaleY, translateX, translateY float64
}

func identityTransform() transform {
	return transform{
		scaleX: 1,
		scaleY: 1,
	}
}

func (t transform) scale(x, y float64) transform {
	t.scaleX *= x
	t.scaleY *= y
	t.translateX *= x
	t.translateY *= y
	return t
}

func (t transform) translate(x, y float64) transform {
	t.translateX += x
	t.translateY += y
	return t
}

func (t transform) invert() transform {
	t.scaleX = 1 / t.scaleX
	t.scaleY = 1 / t.scaleY
	t.translateX *= -t.scaleX
	t.translateY *= -t.scaleY
	return t
}

func (t transform) compose(u transform) transform {
	return t.scale(u.scaleX, u.scaleY).translate(u.translateX, u.translateY)
}

func (t transform) transform(p Position) Position {
	p.X = t.scaleX*p.X + t.translateX
	p.Y = t.scaleY*p.Y + t.translateY
	return p
}
