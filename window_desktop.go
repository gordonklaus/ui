// +build !android,!ios

package ui

import (
	"runtime"
	"sync"
	"unsafe"

	"github.com/go-gl/gl/v3.2-core/gl"
	glmobile "golang.org/x/mobile/gl"
)

var (
	windowsMu sync.Mutex
	windows   = map[uintptr]*window{}
)

type window struct {
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

func newWindow(size Size, v View) (Window, error) {
	w := &window{
		w:             newWindowImpl(size),
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

func (w *window) Do(f func()) {
	done := make(chan struct{})
	w.do <- func() {
		f()
		close(done)
	}
	<-done
}

func (w *window) Resize(s Size) {
	w.View.Resize(s)
	w.gfx.Size(s)
	w.theView.Resize(s)
	w.Redraw()
}

func (w *window) Redraw() {
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

	// Using attribute arrays in OpenGL 3.3 requires the use of a vertex array.
	var va uint32
	gl.GenVertexArrays(1, &va)
	gl.BindVertexArray(va)

	w.gfx = newGraphics(glContext{})
	defer w.gfx.release()

	for {
		select {
		case f := <-w.do:
			f()
		case s := <-w.sizeEvents:
			w.size = s
			w.Resize(s.size)
		case <-w.drawEvents:
			w.gfx.clear()
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

type glContext struct{ glmobile.Context }

func (glContext) Enable(cap glmobile.Enum) {
	gl.Enable(uint32(cap))
}

func (glContext) BlendFunc(sfactor, dfactor glmobile.Enum) {
	gl.BlendFunc(uint32(sfactor), uint32(dfactor))
}

func (glContext) CreateProgram() glmobile.Program {
	return glmobile.Program{Init: true, Value: gl.CreateProgram()}
}

func (glContext) CreateShader(ty glmobile.Enum) glmobile.Shader {
	return glmobile.Shader{Value: gl.CreateShader(uint32(ty))}
}

func (glContext) ShaderSource(s glmobile.Shader, src string) {
	srcStr, free := gl.Strs(src + "\x00")
	defer free()
	gl.ShaderSource(s.Value, 1, srcStr, nil)
}

func (glContext) CompileShader(s glmobile.Shader) {
	gl.CompileShader(s.Value)
}

func (glContext) GetShaderi(s glmobile.Shader, pname glmobile.Enum) int {
	var status int32
	gl.GetShaderiv(s.Value, uint32(pname), &status)
	return int(status)
}

func (glContext) DeleteShader(s glmobile.Shader) {
	gl.DeleteShader(s.Value)
}

func (glContext) GetShaderInfoLog(s glmobile.Shader) string {
	var buf [256]byte
	var bufLen int32
	gl.GetShaderInfoLog(s.Value, int32(len(buf)), &bufLen, &buf[0])
	return string(buf[:bufLen])
}

func (glContext) AttachShader(p glmobile.Program, s glmobile.Shader) {
	gl.AttachShader(p.Value, s.Value)
}

func (glContext) LinkProgram(p glmobile.Program) {
	gl.LinkProgram(p.Value)
}

func (glContext) GetProgrami(p glmobile.Program, pname glmobile.Enum) int {
	var status int32
	gl.GetProgramiv(p.Value, uint32(pname), &status)
	return int(status)
}

func (glContext) DeleteProgram(p glmobile.Program) {
	gl.DeleteProgram(p.Value)
}

func (glContext) GetProgramInfoLog(p glmobile.Program) string {
	var buf [256]byte
	var bufLen int32
	gl.GetProgramInfoLog(p.Value, int32(len(buf)), &bufLen, &buf[0])
	return string(buf[:bufLen])
}

func (glContext) GetUniformLocation(p glmobile.Program, name string) glmobile.Uniform {
	return glmobile.Uniform{Value: gl.GetUniformLocation(p.Value, gl.Str(name+"\x00"))}
}

func (glContext) GetAttribLocation(p glmobile.Program, name string) glmobile.Attrib {
	return glmobile.Attrib{Value: uint(gl.GetAttribLocation(p.Value, gl.Str(name+"\x00")))}
}

func (glContext) UseProgram(p glmobile.Program) {
	gl.UseProgram(p.Value)
}

func (glContext) ClearColor(red, green, blue, alpha float32) {
	gl.ClearColor(red, green, blue, alpha)
}

func (glContext) Clear(mask glmobile.Enum) {
	gl.Clear(uint32(mask))
}

func (glContext) UniformMatrix4fv(dst glmobile.Uniform, src []float32) {
	gl.UniformMatrix4fv(dst.Value, 1, false, &src[0])
}

func (glContext) CreateBuffer() glmobile.Buffer {
	var buffer uint32
	gl.GenBuffers(1, &buffer)
	return glmobile.Buffer{Value: buffer}
}

func (glContext) BindBuffer(target glmobile.Enum, b glmobile.Buffer) {
	gl.BindBuffer(uint32(target), b.Value)
}

func (glContext) BufferData(target glmobile.Enum, src []byte, usage glmobile.Enum) {
	gl.BufferData(uint32(target), len(src), gl.Ptr(src), uint32(usage))
}

func (glContext) DeleteBuffer(b glmobile.Buffer) {
	gl.DeleteBuffers(1, &b.Value)
}

func (glContext) VertexAttribPointer(dst glmobile.Attrib, size int, ty glmobile.Enum, normalized bool, stride, offset int) {
	gl.VertexAttribPointer(uint32(dst.Value), int32(size), uint32(ty), normalized, int32(stride), unsafe.Pointer(uintptr(offset)))
}

func (glContext) EnableVertexAttribArray(a glmobile.Attrib) {
	gl.EnableVertexAttribArray(uint32(a.Value))
}

func (glContext) DrawArrays(mode glmobile.Enum, first, count int) {
	gl.DrawArrays(uint32(mode), int32(first), int32(count))
}
