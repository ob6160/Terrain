package core

import (
	"github.com/go-gl/gl/v3.2-core/gl"
	"unsafe"
)

type Mesh struct {
	Vertices      []float32
	Texture       uint32
	Indices       []uint32
	stride int32
	vao, vbo, ebo uint32
	RenderMode uint32
}

func (m Mesh) Draw() {
	gl.BindVertexArray(m.vao)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, m.Texture)
	gl.DrawElements(m.RenderMode, int32(len(m.Indices)), gl.UNSIGNED_INT, unsafe.Pointer(nil))
	gl.BindVertexArray(0)
}

func (m *Mesh) Construct() {
	// Free up memory used for last buffers
	gl.DeleteVertexArrays(1, &m.vao)
	gl.DeleteBuffers(1, &m.vbo)
	gl.DeleteBuffers(1, &m.ebo)
	
	// Vertex Array Object Setup
	gl.GenVertexArrays(1, &m.vao)

	// Element Buffer Object Setup
	gl.GenBuffers(1, &m.ebo)

	// Vertex Buffer Object Setup
	gl.GenBuffers(1, &m.vbo)

	gl.BindVertexArray(m.vao)

	// Send vertex data to a VBO
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(m.Vertices)*4, gl.Ptr(m.Vertices), gl.STATIC_DRAW)

	// Send index data to a EBO
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(m.Indices)*4, gl.Ptr(m.Indices), gl.STATIC_DRAW)

	// Setup Vertex Attrib Pointers so VBO is read correctly (positions, normals, texcoords)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8 * 4, gl.PtrOffset(0))

	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 8 * 4, gl.PtrOffset(3*4))

	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 8 * 4, gl.PtrOffset(6*4))

	gl.BindVertexArray(0)
}