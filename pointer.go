package ui

import "sync"

type Pointer struct {
	externalID uint32
	ID         PointerID
	Type       PointerType
	Position
	Button  PointerButtons
	Buttons PointerButtons
}

type PointerID uint16

type PointerType uint8

const (
	PointerTypeMouse PointerType = iota
	PointerTypeTouch
	PointerTypePen
)

func (t PointerType) Mouse() bool { return t == PointerTypeMouse }
func (t PointerType) Touch() bool { return t == PointerTypeTouch }
func (t PointerType) Pen() bool   { return t == PointerTypePen }

type PointerButtons uint8

const (
	PointerButtonNone                                                          PointerButtons = 0
	PointerButtonLeftMouse, PointerButtonTouchContact, PointerButtonPenContact PointerButtons = 1, 1, 1
	PointerButtonMiddleMouse                                                   PointerButtons = 2
	PointerButtonRightMouse, PointerButtonPenBarrel                            PointerButtons = 4, 4
	PointerButtonX1BackMouse                                                   PointerButtons = 8
	PointerButtonX2ForwardMouse                                                PointerButtons = 16
	PointerButtonPenEraser                                                     PointerButtons = 32
)

func (b PointerButtons) None() bool           { return b == PointerButtonNone }
func (b PointerButtons) LeftMouse() bool      { return b&PointerButtonLeftMouse != 0 }
func (b PointerButtons) TouchContact() bool   { return b&PointerButtonTouchContact != 0 }
func (b PointerButtons) PenContact() bool     { return b&PointerButtonPenContact != 0 }
func (b PointerButtons) MiddleMouse() bool    { return b&PointerButtonMiddleMouse != 0 }
func (b PointerButtons) RightMouse() bool     { return b&PointerButtonRightMouse != 0 }
func (b PointerButtons) PenBarrel() bool      { return b&PointerButtonPenBarrel != 0 }
func (b PointerButtons) X1BackMouse() bool    { return b&PointerButtonX1BackMouse != 0 }
func (b PointerButtons) X2ForwardMouse() bool { return b&PointerButtonX2ForwardMouse != 0 }
func (b PointerButtons) PenEraser() bool      { return b&PointerButtonPenEraser != 0 }

var activePointers = pointers{
	ids:  map[PointerID]bool{},
	ptrs: map[uint32]*Pointer{},
}

type pointers struct {
	sync.Mutex
	ids  map[PointerID]bool
	ptrs map[uint32]*Pointer
}

func (ptrs *pointers) new(p Pointer) *Pointer {
	ptrs.Lock()
	for ; ptrs.ids[p.ID]; p.ID++ {
	}
	ptrs.ids[p.ID] = true
	ptrs.ptrs[p.externalID] = &p
	ptrs.Unlock()

	return &p
}

func (ptrs *pointers) delete(p Pointer) {
	ptrs.Lock()
	delete(ptrs.ids, p.ID)
	delete(ptrs.ptrs, p.externalID)
	ptrs.Unlock()
}
