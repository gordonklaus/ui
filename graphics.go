package ui

import (
	"fmt"
	"log"

	gl "github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Graphics struct {
	program uint32
	mvp     int32
	pos     uint32
	color   uint32

	proj, view mgl32.Mat4
}

func NewGraphics() *Graphics {
	const vertexShader = `#version 100
		uniform mat4 mvp;
		attribute vec2 pos;
		attribute vec4 color;
		varying vec4 vColor;

		void main() {
			gl_Position = mvp * vec4(pos, 0, 1);
			vColor = color;
		}` + "\x00"

	const fragmentShader = `#version 100
		precision mediump float;
		varying vec4 vColor;

		void main() {
			gl_FragColor = vColor;
		}` + "\x00"

	program, err := createProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Fatalf("error creating GL program: %v", err)
	}

	mvp := gl.GetUniformLocation(program, gl.Str("mvp\x00"))
	pos := uint32(gl.GetAttribLocation(program, gl.Str("pos\x00")))
	color := uint32(gl.GetAttribLocation(program, gl.Str("color\x00")))

	g := &Graphics{
		program: program,
		mvp:     mvp,
		pos:     pos,
		color:   color,
	}

	// Using attribute arrays in OpenGL 3.3 requires the use of a vertex array.
	var va uint32
	gl.GenVertexArrays(1, &va)
	gl.BindVertexArray(va)

	return g
}

func createProgram(vertexSrc, fragmentSrc string) (uint32, error) {
	program := gl.CreateProgram()
	if program == 0 {
		return 0, fmt.Errorf("no programs available")
	}

	vertexShader, err := loadShader(gl.VERTEX_SHADER, vertexSrc)
	if err != nil {
		return 0, err
	}
	fragmentShader, err := loadShader(gl.FRAGMENT_SHADER, fragmentSrc)
	if err != nil {
		gl.DeleteShader(vertexShader)
		return 0, err
	}

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	// Flag shaders for deletion when program is unlinked.
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		defer gl.DeleteProgram(program)
		// TODO: gl.GetProgramInfoLog
		return 0, fmt.Errorf("error linking program")
	}
	return program, nil
}

func loadShader(shaderType uint32, src string) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	if shader == 0 {
		return 0, fmt.Errorf("could not create shader (type %v)", shaderType)
	}
	srcStr, free := gl.Strs(src)
	defer free()
	gl.ShaderSource(shader, 1, srcStr, nil)
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		defer gl.DeleteShader(shader)
		var buf [256]byte
		var bufLen int32
		gl.GetShaderInfoLog(shader, int32(len(buf)), &bufLen, &buf[0])
		return 0, fmt.Errorf("error compiling shader: %s", string(buf[:bufLen]))
	}
	return shader, nil
}

func (g *Graphics) Release() {
	gl.DeleteProgram(g.program)
}

func (g *Graphics) Size(s Size) {
	w := float32(s.Width)
	h := float32(s.Height)
	g.proj = mgl32.Ortho2D(-w/2, w/2, -h/2, h/2)
	g.view = mgl32.LookAt(w/2, h/2, 1, w/2, h/2, 0, 0, 1, 0)
}

func (g *Graphics) Clip2World(x, y float32) (float32, float32) {
	v := g.proj.Mul4(g.view).Inv().Mul4x1(mgl32.Vec4{x, y, 0, 1})
	return v.X(), v.Y()
}

func (g *Graphics) Draw(buffer *TriangleBuffer, model mgl32.Mat4) {
	gl.UseProgram(g.program)

	mvp := g.proj.Mul4(g.view).Mul4(model)
	gl.UniformMatrix4fv(g.mvp, 1, false, &mvp[0])

	buffer.Draw(g.pos, g.color)
}

type TriangleBuffer struct {
	buffer uint32
	length int32
}

type Triangle [3]Vertex

type Vertex struct {
	Position Position
	Color    Color
}

type Color struct {
	R, G, B, A float64
}

const coordsPerVertex = 6

func NewTriangleBuffer(ts []Triangle) *TriangleBuffer {
	var buffer uint32
	gl.GenBuffers(1, &buffer)

	data := []float32{}
	for _, t := range ts {
		for _, v := range t {
			data = append(data,
				float32(v.Position.X),
				float32(v.Position.Y),
				float32(v.Color.R),
				float32(v.Color.G),
				float32(v.Color.B),
				float32(v.Color.A),
			)
		}
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(data), gl.Ptr(data), gl.STATIC_DRAW)

	return &TriangleBuffer{buffer, int32(3 * len(ts))}
}

func (b *TriangleBuffer) Release() {
	gl.DeleteBuffers(1, &b.buffer)
}

func (b *TriangleBuffer) Draw(pos, color uint32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, b.buffer)
	gl.VertexAttribPointer(pos, 2, gl.FLOAT, false, 4*coordsPerVertex, nil)
	gl.EnableVertexAttribArray(pos)
	gl.VertexAttribPointer(color, 4, gl.FLOAT, false, 4*coordsPerVertex, gl.PtrOffset(4*2))
	gl.EnableVertexAttribArray(color)

	gl.DrawArrays(gl.TRIANGLES, 0, b.length)
}
