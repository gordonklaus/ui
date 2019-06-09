// +build updd

package ui

import (
	"github.com/gordonklaus/ui/internal/updd"
)

func init() {
	touches := map[uint8]*Pointer{}

	updd.RegisterTouchCallback(func(id uint8, x, y int32, touchingLeft bool) {
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
		} else if touchingLeft {
			p.Button = PointerButtonNone
		} else {
			up = true
			p.Button = PointerButtonTouchContact
			p.Buttons = PointerButtonNone
			activePointers.delete(*p)
			delete(touches, id)
		}
		p.Position = Position{float64(x), float64(y)}

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
