package core

import (
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/ob6160/Terrain/generators"
	"github.com/ob6160/Terrain/utils"
)

type Plane struct {
	rows, cols int
	m Mesh
}

func (p *Plane) M() *Mesh {
	return &p.m
}

func NewPlane(rows int, cols int) *Plane {
	var newPlane = Plane{rows: rows, cols: cols, m: Mesh{
		Vertices:   make([]float32, rows * cols * 8),
		Texture:    0,
		Indices:    make([]uint32, (rows-1) * (cols-1) * 3 * 2),
		RenderMode: gl.TRIANGLES,
	}}
	return &newPlane
}

func (p *Plane) Construct(generator generators.TerrainGenerator) {
	var vertices = &p.m.Vertices
	vertIndex := 0

	for x := 0; x < p.rows; x++ {
		for y := 0; y < p.cols; y++ {
			(*vertices)[vertIndex+0] = float32(y - (p.rows - 1) / 2)
			height, _ := generator.Get(utils.Point{X:x, Y:y})
			(*vertices)[vertIndex+1] = height
			(*vertices)[vertIndex+2] = float32(x - (p.cols - 1) / 2)
			vertIndex += 8
		}
	}
	
	var indices = &p.m.Indices
	var i = 0
	for r := 0; r < p.rows-1; r++ {
		for c := 0; c < p.cols-1; c++ {
			index := r * p.rows + c
			(*indices)[i] = uint32(index + p.cols + 1)
			(*indices)[i+1] = uint32(index + 1)
			(*indices)[i+2] = uint32(index)

			(*indices)[i+3] = uint32(index+p.rows)
			(*indices)[i+4] = uint32(index+p.rows+1)
			(*indices)[i+5] = uint32(index)
			i += 6
		}
	}
	
	p.m.Construct()
}