package core

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/ob6160/Terrain/generators"
	"github.com/ob6160/Terrain/utils"
)

type Plane struct {
	rows, cols int
	m Mesh
}

func NewPlane(rows int, cols int) *Plane {
	var newPlane = Plane{rows: rows, cols: cols}
	return &newPlane
}

func (p *Plane) Construct() *Mesh {
	vertices := make([]float32, p.rows * p.cols * 8)
	vertIndex := 0


	midpointGen := generators.NewMidPointDisplacement(1024,1024)
	midpointGen.Generate()

	for x := 0; x < p.rows; x++ {
		for y := 0; y < p.cols; y++ {
			vertices[vertIndex+0] = float32(y - (p.rows - 1) / 2) * 0.5
			vertices[vertIndex+1] = midpointGen.Get(utils.Point{X:x, Y:y})
			vertices[vertIndex+2] = float32(x - (p.cols - 1) / 2 ) * 0.5
			vertIndex += 8
		}
	}

	indices := make([]uint32, (p.rows-1) * (p.cols-1) * 3 * 2)
	var i int = 0
	for r := 0; r < p.rows-1; r++ {
		for c := 0; c < p.cols-1; c++ {
			index := r * p.rows + c
			indices[i] = uint32(index + p.cols + 1)
			indices[i+1] = uint32(index + 1)
			indices[i+2] = uint32(index)

			indices[i+3] = uint32(index+p.rows)
			indices[i+4] = uint32(index+p.rows+1)
			indices[i+5] = uint32(index)
			i += 6
			
		}
	}
	
	p.m = Mesh{
		Vertices:   vertices,
		Texture:    0,
		Indices:    indices,
		RenderMode: gl.TRIANGLES,
	}
	p.m.Construct()
	return &p.m
}