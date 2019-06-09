package updd

/*
#cgo LDFLAGS: -lupddapi.1
#include "upddapi.h"

void TBCALL connectCallback(unsigned long context, _PointerEvent* ev);
void TBCALL touchCallback(unsigned long context, _PointerEvent* ev);
*/
import "C"

type TouchCallback func(id uint8, x, y int32, down bool)

var touchCallback TouchCallback

func RegisterTouchCallback(cb TouchCallback) {
	touchCallback = cb

	C.TBApiRegisterEvent(0, 0, C._EventConfiguration, (*[0]byte)(C.connectCallback))

	C.TBApiOpen()
}

func Close() {
	C.TBApiUnregisterEvent((*[0]byte)(C.touchCallback))
	C.TBApiClose()
}

var touches = map[uint8]touch{}

type touch struct {
	x, y int32
	touchingLeft bool
}

//export callback
func callback(id uint8, x, y int32, touchingLeft bool) {
	t := touch{x, y, touchingLeft}
	if touches[id] != t {
		touches[id] = t
		touchCallback(id, x, y, touchingLeft)
	}
}
