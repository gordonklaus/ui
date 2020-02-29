package ui

import (
	"encoding/binary"
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
)

type Graphics struct {
	glctx   gl.Context
	program gl.Program
	mvp     gl.Uniform
	pos     gl.Attrib
	color   gl.Attrib

	proj, view mgl32.Mat4
}

func newGraphics(glctx gl.Context) *Graphics {
	glctx.Enable(gl.BLEND)
	glctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	const vertexShader = `#version 100
		uniform mat4 mvp;
		attribute vec2 pos;
		attribute vec4 color;
		varying vec4 vColor;

		void main() {
			gl_Position = mvp * vec4(pos, 0, 1);
			vColor = color;
		}`

	const fragmentShader = `#version 100
		precision mediump float;
		varying vec4 vColor;

		void main() {
			gl_FragColor = vColor;
		}`

	program, err := glutil.CreateProgram(glctx, vertexShader, fragmentShader)
	if err != nil {
		log.Fatalf("error creating GL program: %v", err)
	}

	mvp := glctx.GetUniformLocation(program, "mvp")
	pos := glctx.GetAttribLocation(program, "pos")
	color := glctx.GetAttribLocation(program, "color")

	return &Graphics{
		glctx:   glctx,
		program: program,
		mvp:     mvp,
		pos:     pos,
		color:   color,
	}
}

func (g *Graphics) release() {
	g.glctx.DeleteProgram(g.program)
}

func (g *Graphics) Size(s Size) {
	w := float32(s.Width)
	h := float32(s.Height)
	g.proj = mgl32.Mat4{
		2 / w, 0, 0, 0,
		0, -2 / h, 0, 0,
		0, 0, -1, 0,
		-1, 1, -1, 1,
	}
}

func (g *Graphics) setViewTransform(t transform) {
	g.view = mgl32.Mat4{
		float32(t.scaleX), 0, 0, 0,
		0, float32(t.scaleY), 0, 0,
		0, 0, 1, 0,
		float32(t.translateX), float32(t.translateY), 0, 1,
	}
}

func (g *Graphics) clear() {
	g.glctx.ClearColor(0, 0, 0, 1)
	g.glctx.Clear(gl.COLOR_BUFFER_BIT)
}

func (g *Graphics) Draw(buffer *TriangleBuffer, model mgl32.Mat4) {
	g.glctx.UseProgram(g.program)

	mvp := g.proj.Mul4(g.view).Mul4(model)
	g.glctx.UniformMatrix4fv(g.mvp, mvp[:])

	buffer.draw(g.pos, g.color)
}

type TriangleBuffer struct {
	gfx    *Graphics
	buffer gl.Buffer
	length int
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

func NewTriangleBuffer(gfx *Graphics, ts []Triangle) *TriangleBuffer {
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

	buffer := gfx.glctx.CreateBuffer()
	gfx.glctx.BindBuffer(gl.ARRAY_BUFFER, buffer)
	gfx.glctx.BufferData(gl.ARRAY_BUFFER, f32.Bytes(binary.LittleEndian, data...), gl.STATIC_DRAW)

	return &TriangleBuffer{gfx, buffer, 3 * len(ts)}
}

func (b *TriangleBuffer) Release() {
	b.gfx.glctx.DeleteBuffer(b.buffer)
}

func (b *TriangleBuffer) draw(pos, color gl.Attrib) {
	b.gfx.glctx.BindBuffer(gl.ARRAY_BUFFER, b.buffer)
	b.gfx.glctx.VertexAttribPointer(pos, 2, gl.FLOAT, false, 4*coordsPerVertex, 0)
	b.gfx.glctx.EnableVertexAttribArray(pos)
	b.gfx.glctx.VertexAttribPointer(color, 4, gl.FLOAT, false, 4*coordsPerVertex, 4*2)
	b.gfx.glctx.EnableVertexAttribArray(color)

	b.gfx.glctx.DrawArrays(gl.TRIANGLES, 0, b.length)
}
