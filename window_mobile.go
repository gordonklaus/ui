// +build android ios

package ui

import (
	"errors"
	"log"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
)

var theWindow *window

type window struct {
	*windowBase
	drawEvents chan drawEvent

	app         app.App
	pixelsPerPt float32
	pointers    map[touch.Sequence]*Pointer
}

type drawEvent struct{}

func newWindow(size Size, v View) (Window, error) {
	if theWindow != nil {
		return nil, errors.New("only a single window is supported on mobile platforms")
	}

	w := &window{
		drawEvents: make(chan drawEvent, 1),
		pointers:   map[touch.Sequence]*Pointer{},
	}
	w.windowBase = newWindowBase(w, v)
	v.SetParent(w)

	theWindow = w

	return w, nil
}

func (w *window) Resize(s Size) {
	w.View.Resize(s)
	w.theView.Resize(s)
	if w.gfx != nil {
		w.gfx.Size(s)
		w.Redraw()
	}
}

func (w *window) Redraw() {
	select {
	case w.drawEvents <- drawEvent{}:
	default:
	}
}

func run(cb func()) {
	app.Main(func(a app.App) {
		cb()

		if theWindow == nil {
			log.Println("no window was created")
			return
		}

		theWindow.app = a
		theWindow.handleEvents()
	})
}

func (w *window) handleEvents() {
	for {
		select {
		case f := <-w.do:
			f()
		case <-w.drawEvents:
			w.app.Send(paint.Event{})
		case e := <-w.app.Events():
			switch e := w.app.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					w.gfx = newGraphics(e.DrawContext.(gl.Context))
					w.Redraw()
				case lifecycle.CrossOff:
					w.gfx.release()
					w.gfx = nil
				}
			case size.Event:
				w.pixelsPerPt = e.PixelsPerPt
				w.Resize(Size{
					Width:  ptToMM(e.WidthPt),
					Height: ptToMM(e.HeightPt),
				})
			case paint.Event:
				if w.gfx != nil {
					w.windowBase.draw()
					w.app.Publish()
				}
			case touch.Event:
				w.handleTouchEvent(e)
			}
		}
	}
}

func (w *window) handleTouchEvent(e touch.Event) {
	pos := Position{
		X: ptToMM(geom.Pt(e.X / w.pixelsPerPt)),
		Y: ptToMM(geom.Pt(e.Y / w.pixelsPerPt)),
	}

	switch e.Type {
	case touch.TypeBegin:
		p := &Pointer{
			ID:       PointerID(e.Sequence),
			Type:     PointerTypeTouch,
			Position: pos,
			Button:   PointerButtonTouchContact,
			Buttons:  PointerButtonTouchContact,
		}
		w.pointers[e.Sequence] = p

		w.windowBase.pointerDown(*p)
	case touch.TypeMove:
		p := w.pointers[e.Sequence]
		p.Position = pos
		p.Button = PointerButtonNone

		w.windowBase.pointerMove(*p)
	case touch.TypeEnd:
		p := w.pointers[e.Sequence]
		p.Position = pos
		p.Button = PointerButtonTouchContact
		p.Buttons = PointerButtonNone
		delete(w.pointers, e.Sequence)

		w.windowBase.pointerUp(*p)
	}
}

func ptToMM(pt geom.Pt) float64 {
	const mmPerPt = 10 * 2.54 / 72
	return mmPerPt * float64(pt)
}
