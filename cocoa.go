// +build darwin
// +build 386 amd64
// +build !ios

package ui

/*
#cgo CFLAGS: -x objective-c -DGL_SILENCE_DEPRECATION
#cgo LDFLAGS: -framework Cocoa -framework OpenGL
#include <OpenGL/gl3.h>
#import <Carbon/Carbon.h> // for HIToolbox/Events.h
#import <Cocoa/Cocoa.h>
#include <pthread.h>
#include <stdint.h>
#include <stdlib.h>

uint64_t threadID();
void runApp();
uintptr_t newWindow(double width, double height);
void makeCurrentContext(uintptr_t ctx);
void flushContext(uintptr_t ctx);
NSPoint mapFromScreen(uintptr_t window, NSPoint pt);
*/
import "C"

import (
	"log"
	"runtime"
)

var initThreadID C.uint64_t

func init() {
	runtime.LockOSThread()
	initThreadID = C.threadID()
}

var appCallback func()

func run(cb func()) {
	if tid := C.threadID(); tid != initThreadID {
		log.Fatalf("ui.Run called on thread %d, but ui.init ran on %d", tid, initThreadID)
	}

	appCallback = cb
	C.runApp()
}

//export applicationDidFinishLaunching
func applicationDidFinishLaunching() {
	go appCallback()
}

func newWindow(size Size) uintptr {
	return uintptr(C.newWindow(C.double(size.Width), C.double(size.Height)))
}

//export preparedOpenGL
func preparedOpenGL(window uintptr, ctx uintptr) {
	go windowLoop(window, ctx)
}

func makeCurrentContext(ctx uintptr) {
	C.makeCurrentContext(C.uintptr_t(ctx))
}

func flushContext(ctx uintptr) {
	C.flushContext(C.uintptr_t(ctx))
}

// //export drawgl
// func drawgl(id uintptr) {
// 	theScreen.mu.Lock()
// 	w := theScreen.windows[id]
// 	theScreen.mu.Unlock()

// 	if w == nil {
// 		return // closing window
// 	}

// 	// TODO: is this necessary?
// 	w.lifecycler.SetVisible(true)
// 	w.lifecycler.SendEvent(w, w.glctx)

// 	w.Send(paint.Event{External: true})
// 	<-w.drawDone
// }

//export resize
func resize(window uintptr, width, height, pxWidth, pxHeight float64) {
	windowsMu.Lock()
	w := windows[window]
	windowsMu.Unlock()

	w.sizeEvents <- sizeEvent{
		size: Size{width, height},
		px:   Size{pxWidth, pxHeight},
	}
}

// //export windowClosing
// func windowClosing(id uintptr) {
// 	sendLifecycle(id, (*lifecycler.State).SetDead, true)
// }

var mousePointer = activePointers.new(Pointer{
	Type: PointerTypeMouse,
})

//export mouseEvent
func mouseEvent(window uintptr, x, y float64, typ, button int32, flags uint32) {
	windowsMu.Lock()
	defer windowsMu.Unlock()
	w := windows[window]

	var down, up bool

	mousePointer.X = x
	mousePointer.Y = y

	switch typ {
	case C.NSMouseMoved, C.NSLeftMouseDragged, C.NSRightMouseDragged, C.NSOtherMouseDragged:
		mousePointer.Button = PointerButtonNone
	case C.NSLeftMouseDown, C.NSRightMouseDown, C.NSOtherMouseDown:
		down = true
		mousePointer.Button = cocoaMouseButton(button)
		mousePointer.Buttons |= mousePointer.Button
	case C.NSLeftMouseUp, C.NSRightMouseUp, C.NSOtherMouseUp:
		up = true
		mousePointer.Button = cocoaMouseButton(button)
		mousePointer.Buttons &^= mousePointer.Button
	}

	w.pointerEvents <- pointerEvent{
		down: down,
		up:   up,
		p:    *mousePointer,
	}
}

func cocoaMouseButton(button int32) PointerButtons {
	switch button {
	default:
		// TODO: PointerButtonOther/Unknown?
		return PointerButtonNone
	case 0:
		return PointerButtonLeftMouse
	case 1:
		return PointerButtonRightMouse
	case 2:
		return PointerButtonMiddleMouse
	case 3:
		return PointerButtonX1BackMouse
	case 4:
		return PointerButtonX2ForwardMouse
	}
}

func (w *Window) MapFromParent(p Position) Position {
	pt := C.NSPoint{
		x: C.double(p.X),
		y: C.double(p.Y),
	}
	pt = C.mapFromScreen(C.uintptr_t(w.w), pt)
	return Position{
		X: float64(pt.x),
		Y: float64(pt.y),
	}
}

// //export keyEvent
// func keyEvent(id uintptr, runeVal rune, dir uint8, code uint16, flags uint32) {
// 	sendWindowEvent(id, key.Event{
// 		Rune:      cocoaRune(runeVal),
// 		Direction: key.Direction(dir),
// 		Code:      cocoaKeyCode(code),
// 		Modifiers: cocoaMods(flags),
// 	})
// }

// //export flagEvent
// func flagEvent(id uintptr, flags uint32) {
// 	for _, mod := range mods {
// 		if flags&mod.flags == mod.flags && lastFlags&mod.flags != mod.flags {
// 			keyEvent(id, -1, C.NSKeyDown, mod.code, flags)
// 		}
// 		if lastFlags&mod.flags == mod.flags && flags&mod.flags != mod.flags {
// 			keyEvent(id, -1, C.NSKeyUp, mod.code, flags)
// 		}
// 	}
// 	lastFlags = flags
// }

// var lastFlags uint32

// func sendLifecycle(id uintptr, setter func(*lifecycler.State, bool), val bool) {
// 	theScreen.mu.Lock()
// 	w := theScreen.windows[id]
// 	theScreen.mu.Unlock()

// 	if w == nil {
// 		return
// 	}
// 	setter(&w.lifecycler, val)
// 	w.lifecycler.SendEvent(w, w.glctx)
// }

// func sendLifecycleAll(dead bool) {
// 	windows := []*windowImpl{}

// 	theScreen.mu.Lock()
// 	for _, w := range theScreen.windows {
// 		windows = append(windows, w)
// 	}
// 	theScreen.mu.Unlock()

// 	for _, w := range windows {
// 		w.lifecycler.SetFocused(false)
// 		w.lifecycler.SetVisible(false)
// 		if dead {
// 			w.lifecycler.SetDead(true)
// 		}
// 		w.lifecycler.SendEvent(w, w.glctx)
// 	}
// }

// //export lifecycleDeadAll
// func lifecycleDeadAll() { sendLifecycleAll(true) }

// //export lifecycleHideAll
// func lifecycleHideAll() { sendLifecycleAll(false) }

// //export lifecycleVisible
// func lifecycleVisible(id uintptr, val bool) {
// 	sendLifecycle(id, (*lifecycler.State).SetVisible, val)
// }

// //export lifecycleFocused
// func lifecycleFocused(id uintptr, val bool) {
// 	sendLifecycle(id, (*lifecycler.State).SetFocused, val)
// }

// // cocoaRune marks the Carbon/Cocoa private-range unicode rune representing
// // a non-unicode key event to -1, used for Rune in the key package.
// //
// // http://www.unicode.org/Public/MAPPINGS/VENDORS/APPLE/CORPCHAR.TXT
// func cocoaRune(r rune) rune {
// 	if '\uE000' <= r && r <= '\uF8FF' {
// 		return -1
// 	}
// 	return r
// }

// // cocoaKeyCode converts a Carbon/Cocoa virtual key code number
// // into the standard keycodes used by the key package.
// //
// // To get a sense of the key map, see the diagram on
// //	http://boredzo.org/blog/archives/2007-05-22/virtual-key-codes
// func cocoaKeyCode(vkcode uint16) key.Code {
// 	switch vkcode {
// 	case C.kVK_ANSI_A:
// 		return key.CodeA
// 	case C.kVK_ANSI_B:
// 		return key.CodeB
// 	case C.kVK_ANSI_C:
// 		return key.CodeC
// 	case C.kVK_ANSI_D:
// 		return key.CodeD
// 	case C.kVK_ANSI_E:
// 		return key.CodeE
// 	case C.kVK_ANSI_F:
// 		return key.CodeF
// 	case C.kVK_ANSI_G:
// 		return key.CodeG
// 	case C.kVK_ANSI_H:
// 		return key.CodeH
// 	case C.kVK_ANSI_I:
// 		return key.CodeI
// 	case C.kVK_ANSI_J:
// 		return key.CodeJ
// 	case C.kVK_ANSI_K:
// 		return key.CodeK
// 	case C.kVK_ANSI_L:
// 		return key.CodeL
// 	case C.kVK_ANSI_M:
// 		return key.CodeM
// 	case C.kVK_ANSI_N:
// 		return key.CodeN
// 	case C.kVK_ANSI_O:
// 		return key.CodeO
// 	case C.kVK_ANSI_P:
// 		return key.CodeP
// 	case C.kVK_ANSI_Q:
// 		return key.CodeQ
// 	case C.kVK_ANSI_R:
// 		return key.CodeR
// 	case C.kVK_ANSI_S:
// 		return key.CodeS
// 	case C.kVK_ANSI_T:
// 		return key.CodeT
// 	case C.kVK_ANSI_U:
// 		return key.CodeU
// 	case C.kVK_ANSI_V:
// 		return key.CodeV
// 	case C.kVK_ANSI_W:
// 		return key.CodeW
// 	case C.kVK_ANSI_X:
// 		return key.CodeX
// 	case C.kVK_ANSI_Y:
// 		return key.CodeY
// 	case C.kVK_ANSI_Z:
// 		return key.CodeZ
// 	case C.kVK_ANSI_1:
// 		return key.Code1
// 	case C.kVK_ANSI_2:
// 		return key.Code2
// 	case C.kVK_ANSI_3:
// 		return key.Code3
// 	case C.kVK_ANSI_4:
// 		return key.Code4
// 	case C.kVK_ANSI_5:
// 		return key.Code5
// 	case C.kVK_ANSI_6:
// 		return key.Code6
// 	case C.kVK_ANSI_7:
// 		return key.Code7
// 	case C.kVK_ANSI_8:
// 		return key.Code8
// 	case C.kVK_ANSI_9:
// 		return key.Code9
// 	case C.kVK_ANSI_0:
// 		return key.Code0
// 	// TODO: move the rest of these codes to constants in key.go
// 	// if we are happy with them.
// 	case C.kVK_Return:
// 		return key.CodeReturnEnter
// 	case C.kVK_Escape:
// 		return key.CodeEscape
// 	case C.kVK_Delete:
// 		return key.CodeDeleteBackspace
// 	case C.kVK_Tab:
// 		return key.CodeTab
// 	case C.kVK_Space:
// 		return key.CodeSpacebar
// 	case C.kVK_ANSI_Minus:
// 		return key.CodeHyphenMinus
// 	case C.kVK_ANSI_Equal:
// 		return key.CodeEqualSign
// 	case C.kVK_ANSI_LeftBracket:
// 		return key.CodeLeftSquareBracket
// 	case C.kVK_ANSI_RightBracket:
// 		return key.CodeRightSquareBracket
// 	case C.kVK_ANSI_Backslash:
// 		return key.CodeBackslash
// 	// 50: Keyboard Non-US "#" and ~
// 	case C.kVK_ANSI_Semicolon:
// 		return key.CodeSemicolon
// 	case C.kVK_ANSI_Quote:
// 		return key.CodeApostrophe
// 	case C.kVK_ANSI_Grave:
// 		return key.CodeGraveAccent
// 	case C.kVK_ANSI_Comma:
// 		return key.CodeComma
// 	case C.kVK_ANSI_Period:
// 		return key.CodeFullStop
// 	case C.kVK_ANSI_Slash:
// 		return key.CodeSlash
// 	case C.kVK_CapsLock:
// 		return key.CodeCapsLock
// 	case C.kVK_F1:
// 		return key.CodeF1
// 	case C.kVK_F2:
// 		return key.CodeF2
// 	case C.kVK_F3:
// 		return key.CodeF3
// 	case C.kVK_F4:
// 		return key.CodeF4
// 	case C.kVK_F5:
// 		return key.CodeF5
// 	case C.kVK_F6:
// 		return key.CodeF6
// 	case C.kVK_F7:
// 		return key.CodeF7
// 	case C.kVK_F8:
// 		return key.CodeF8
// 	case C.kVK_F9:
// 		return key.CodeF9
// 	case C.kVK_F10:
// 		return key.CodeF10
// 	case C.kVK_F11:
// 		return key.CodeF11
// 	case C.kVK_F12:
// 		return key.CodeF12
// 	// 70: PrintScreen
// 	// 71: Scroll Lock
// 	// 72: Pause
// 	// 73: Insert
// 	case C.kVK_Home:
// 		return key.CodeHome
// 	case C.kVK_PageUp:
// 		return key.CodePageUp
// 	case C.kVK_ForwardDelete:
// 		return key.CodeDeleteForward
// 	case C.kVK_End:
// 		return key.CodeEnd
// 	case C.kVK_PageDown:
// 		return key.CodePageDown
// 	case C.kVK_RightArrow:
// 		return key.CodeRightArrow
// 	case C.kVK_LeftArrow:
// 		return key.CodeLeftArrow
// 	case C.kVK_DownArrow:
// 		return key.CodeDownArrow
// 	case C.kVK_UpArrow:
// 		return key.CodeUpArrow
// 	case C.kVK_ANSI_KeypadClear:
// 		return key.CodeKeypadNumLock
// 	case C.kVK_ANSI_KeypadDivide:
// 		return key.CodeKeypadSlash
// 	case C.kVK_ANSI_KeypadMultiply:
// 		return key.CodeKeypadAsterisk
// 	case C.kVK_ANSI_KeypadMinus:
// 		return key.CodeKeypadHyphenMinus
// 	case C.kVK_ANSI_KeypadPlus:
// 		return key.CodeKeypadPlusSign
// 	case C.kVK_ANSI_KeypadEnter:
// 		return key.CodeKeypadEnter
// 	case C.kVK_ANSI_Keypad1:
// 		return key.CodeKeypad1
// 	case C.kVK_ANSI_Keypad2:
// 		return key.CodeKeypad2
// 	case C.kVK_ANSI_Keypad3:
// 		return key.CodeKeypad3
// 	case C.kVK_ANSI_Keypad4:
// 		return key.CodeKeypad4
// 	case C.kVK_ANSI_Keypad5:
// 		return key.CodeKeypad5
// 	case C.kVK_ANSI_Keypad6:
// 		return key.CodeKeypad6
// 	case C.kVK_ANSI_Keypad7:
// 		return key.CodeKeypad7
// 	case C.kVK_ANSI_Keypad8:
// 		return key.CodeKeypad8
// 	case C.kVK_ANSI_Keypad9:
// 		return key.CodeKeypad9
// 	case C.kVK_ANSI_Keypad0:
// 		return key.CodeKeypad0
// 	case C.kVK_ANSI_KeypadDecimal:
// 		return key.CodeKeypadFullStop
// 	case C.kVK_ANSI_KeypadEquals:
// 		return key.CodeKeypadEqualSign
// 	case C.kVK_F13:
// 		return key.CodeF13
// 	case C.kVK_F14:
// 		return key.CodeF14
// 	case C.kVK_F15:
// 		return key.CodeF15
// 	case C.kVK_F16:
// 		return key.CodeF16
// 	case C.kVK_F17:
// 		return key.CodeF17
// 	case C.kVK_F18:
// 		return key.CodeF18
// 	case C.kVK_F19:
// 		return key.CodeF19
// 	case C.kVK_F20:
// 		return key.CodeF20
// 	// 116: Keyboard Execute
// 	case C.kVK_Help:
// 		return key.CodeHelp
// 	// 118: Keyboard Menu
// 	// 119: Keyboard Select
// 	// 120: Keyboard Stop
// 	// 121: Keyboard Again
// 	// 122: Keyboard Undo
// 	// 123: Keyboard Cut
// 	// 124: Keyboard Copy
// 	// 125: Keyboard Paste
// 	// 126: Keyboard Find
// 	case C.kVK_Mute:
// 		return key.CodeMute
// 	case C.kVK_VolumeUp:
// 		return key.CodeVolumeUp
// 	case C.kVK_VolumeDown:
// 		return key.CodeVolumeDown
// 	// 130: Keyboard Locking Caps Lock
// 	// 131: Keyboard Locking Num Lock
// 	// 132: Keyboard Locking Scroll Lock
// 	// 133: Keyboard Comma
// 	// 134: Keyboard Equal Sign
// 	// ...: Bunch of stuff
// 	case C.kVK_Control:
// 		return key.CodeLeftControl
// 	case C.kVK_Shift:
// 		return key.CodeLeftShift
// 	case C.kVK_Option:
// 		return key.CodeLeftAlt
// 	case C.kVK_Command:
// 		return key.CodeLeftGUI
// 	case C.kVK_RightControl:
// 		return key.CodeRightControl
// 	case C.kVK_RightShift:
// 		return key.CodeRightShift
// 	case C.kVK_RightOption:
// 		return key.CodeRightAlt
// 	// TODO key.CodeRightGUI
// 	default:
// 		return key.CodeUnknown
// 	}
// }
