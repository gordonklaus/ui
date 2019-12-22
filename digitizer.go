// +build !android,!ios

package ui

import (
	"github.com/gordonklaus/ui/internal/digitizer"
)

func init() {
	touches := map[uint8]*Pointer{}

	go digitizer.Run(func(id uint8, pressed bool, x, y uint16) {
		down, up := false, false
		p := touches[id]
		if p == nil {
			down = true
			p = activePointers.new(Pointer{
				externalID: uint32(id << 2),
				Type:       PointerTypeTouch,
				Button:     PointerButtonTouchContact,
				Buttons:    PointerButtonTouchContact,
			})
			touches[id] = p
		} else if pressed {
			p.Button = PointerButtonNone
		} else {
			up = true
			p.Button = PointerButtonTouchContact
			p.Buttons = PointerButtonNone
			activePointers.delete(*p)
			delete(touches, id)
		}
		p.Position = Position{5120 * float64(x) / float64(1<<16-1), 2880 * float64(y) / float64(1<<16-1)}

		windowsMu.Lock()
		for _, w := range windows {
			p := *p
			// TODO: only send to the active window, otherwise activate the window, if in rect
			p.Position = w.MapFromParent(p.Position)
			w.pointerEvents <- pointerEvent{
				down: down,
				up:   up,
				p:    p,
			}
		}
		windowsMu.Unlock()
	})
}
