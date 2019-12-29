package core

import (
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/ob6160/Terrain/utils"
	"math"
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
		Vertices:   make([]float32, rows * cols * 9),
		Texture:    0,
		Indices:    make([]uint32, (rows-1) * (cols-1) * 3 * 2),
		RenderMode: gl.TRIANGLES,
	}}
	return &newPlane
}

func (p *Plane) Construct(sourceWidth, sourceHeight int) {

	var vertices = &p.m.Vertices

	var dW, dH float64
	dW = float64(sourceWidth) / float64(p.rows)
	dH = float64(sourceHeight) / float64(p.cols)
	
	vertIndex := 0

	for x := 0; x < p.rows; x++ {
		for y := 0; y < p.cols; y++ {
			// Billinear Interpolation on height
			lowSampleX := int(math.Floor(float64(x) * (dW)))
			lowSampleY := int(math.Floor(float64(y) * (dH)))
			//hiSampleX := int(math.Ceil(float64(x) * (dW)))
			//hiSampleY := int(math.Ceil(float64(y) * (dH)))
			//
			//q00, _ := generator.Get(utils.Point{X: lowSampleX, Y: lowSampleY})
			//q10, _ := generator.Get(utils.Point{X: hiSampleX, Y: lowSampleY})
			//q01, _ := generator.Get(utils.Point{X: lowSampleX, Y: hiSampleY})
			//q11, _ := generator.Get(utils.Point{X: hiSampleX, Y: hiSampleY})
			//
			//dX := float32(float64(x)*dW - float64(lowSampleX))
			//dY := float32(float64(y)*dH - float64(lowSampleY))
			//// TODO: Refactor into function for reuse.
			//lerped := q00 * (1 - dX) * (1 - dY) +
			//	q10 * dX * (1 - dY) +
			//	q01 * (1 - dX) * dY +
			//	q11 * dX * dY

			(*vertices)[vertIndex+0] = float32(y - (p.rows - 1) / 2)
			(*vertices)[vertIndex+1] = 1.0
			(*vertices)[vertIndex+2] = float32(x - (p.cols - 1) / 2)
			(*vertices)[vertIndex+3] = float32(utils.ToIndex(lowSampleX, lowSampleY, sourceWidth))
			vertIndex += 9
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