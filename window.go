package ui

import (
	"runtime"
	"sync"

	gl "github.com/go-gl/gl/v3.2-core/gl"
)

var (
	windowsMu sync.Mutex
	windows   = map[uintptr]*Window{}
)

type Window struct {
	View
	w             uintptr
	theView       View
	gfx           *Graphics
	do            chan func()
	sizeEvents    chan sizeEvent
	drawEvents    chan drawEvent
	pointerEvents chan pointerEvent
	size          sizeEvent
}

type sizeEvent struct {
	size Size
	px   Size
}

type drawEvent struct{}

type pointerEvent struct {
	down, up bool
	p        Pointer
}

func NewWindow(size Size, v View) (*Window, error) {
	w := &Window{
		w:             newWindow(size),
		theView:       v,
		do:            make(chan func()),
		sizeEvents:    make(chan sizeEvent, 1),
		drawEvents:    make(chan drawEvent, 1),
		pointerEvents: make(chan pointerEvent, 1),
	}
	w.View = NewView(w)
	v.SetParent(w)

	windowsMu.Lock()
	windows[w.w] = w
	windowsMu.Unlock()

	return w, nil
}

func (w *Window) Do(f func()) {
	done := make(chan struct{})
	w.do <- func() {
		f()
		close(done)
	}
	<-done
}

func (w *Window) Resize(s Size) {
	w.View.Resize(s)
	w.gfx.Size(s)
	w.theView.Resize(s)
	w.Redraw()
}

func (w *Window) Redraw() {
	select {
	case w.drawEvents <- drawEvent{}:
	default:
	}
}

func windowLoop(window uintptr, ctx uintptr) {
	windowsMu.Lock()
	w := windows[window]
	windowsMu.Unlock()

	runtime.LockOSThread()
	makeCurrentContext(ctx)

	gl.Init()

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	w.gfx = NewGraphics()
	defer w.gfx.Release()

	for {
		select {
		case f := <-w.do:
			f()
		case s := <-w.sizeEvents:
			w.size = s
			w.Resize(s.size)
		case <-w.drawEvents:
			gl.ClearColor(0, 0, 0, 1)
			gl.Clear(gl.COLOR_BUFFER_BIT)

			w.theView.Draw(w.gfx)
			flushContext(ctx)
		case p := <-w.pointerEvents:
			p.p.X *= w.size.px.Width
			p.p.Y *= w.size.px.Height
			if p.down {
				w.theView.PointerDown(p.p)
			} else if p.up {
				w.theView.PointerUp(p.p)
			} else {
				w.theView.PointerMove(p.p)
			}
		}
	}
}
