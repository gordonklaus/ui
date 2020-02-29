package ui

type Window interface {
	View
}

func NewWindow(size Size, v View) (Window, error) {
	return newWindow(size, v)
}

type windowBase struct {
	View
	theView      View
	do           chan func()
	gfx          *Graphics
	pointerViews map[PointerID]View
}

func newWindowBase(self Window, v View) *windowBase {
	w := &windowBase{
		theView:      v,
		do:           make(chan func()),
		pointerViews: map[PointerID]View{},
	}
	w.View = NewView(self, nil)
	return w
}

func (w *windowBase) Do(f func()) {
	done := make(chan struct{})
	w.do <- func() {
		f()
		close(done)
	}
	<-done
}

func (w *windowBase) draw() {
	w.gfx.clear()
	w.view().draw(w.gfx)
}

func (w *windowBase) pointerDown(p Pointer) {
	if v := w.ViewAt(p.Position); v != nil {
		p.Position = v.view().mapFromWindow(p.Position)
		w.pointerViews[p.ID] = v
		v.PointerDown(p)
	}
}

func (w *windowBase) pointerMove(p Pointer) {
	if v, ok := w.pointerViews[p.ID]; ok {
		p.Position = v.view().mapFromWindow(p.Position)
		v.PointerMove(p)
	}
}

func (w *windowBase) pointerUp(p Pointer) {
	if v, ok := w.pointerViews[p.ID]; ok {
		p.Position = v.view().mapFromWindow(p.Position)
		v.PointerUp(p)
		delete(w.pointerViews, p.ID)
	}
}
