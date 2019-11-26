package generators

import (
	"github.com/ob6160/Terrain/utils"
	"math/rand"
)

type MidpointDisplacement struct {
	width, height int
	heightmap []float32
}

func NewMidPointDisplacement(width, height int) *MidpointDisplacement {
	return &MidpointDisplacement{width+1, height+1, make([]float32, width * height)}
}

func (m *MidpointDisplacement) Get(p utils.Point, value float32) float32 {
	return m.heightmap[p.ToIndex(m.width)]
}

func (m *MidpointDisplacement) set(p utils.Point, value float32) {
	m.heightmap[p.ToIndex(m.width)] = value
}

func (m *MidpointDisplacement) Generate() {
	m.heightmap = make([]float32, m.width * m.height)

	// Set all four corners to random values
	topLeft := utils.Point{0,0}
	topRight := utils.Point{m.width-1, 0}
	bottomLeft := utils.Point{0, m.height-1}
	bottomRight := utils.Point{m.width-1, m.height-1}
	m.set(topLeft, rand.Float32())
	m.set(topRight, rand.Float32())
	m.set(bottomLeft, rand.Float32())
	m.set(bottomRight, rand.Float32())

	var boundary = utils.Rectangle{
		TopLeft:     topLeft.ToIndex(m.width),
		TopRight:    topRight.ToIndex(m.width),
		BottomLeft:  bottomLeft.ToIndex(m.width),
		BottomRight: bottomRight.ToIndex(m.width),
	}
	m.displace(boundary, 0.5, 0.6)
}

func (m *MidpointDisplacement) displace(boundary utils.Rectangle, spread, reduce float32) {
	if boundary.TopRight - boundary.TopLeft <= m.width - 1 {
		return
	}
	var topMid, leftMid, rightMid, bottomMid, centre int
	topMid = utils.Midpoint(boundary.TopLeft, boundary.TopRight)
	leftMid = utils.Midpoint(boundary.TopLeft, boundary.BottomLeft)
	rightMid = utils.Midpoint(boundary.TopRight, boundary.BottomRight)
	bottomMid = utils.Midpoint(boundary.BottomLeft, boundary.BottomRight)
	centre = utils.Midpoint(leftMid, rightMid)

	if m.heightmap[topMid] == 0 {
		avg := utils.Average(m.heightmap[boundary.TopLeft], m.heightmap[boundary.TopRight])
		m.heightmap[topMid] = utils.Jitter(avg, spread)
	}
	if m.heightmap[leftMid] == 0 {
		avg := utils.Average(m.heightmap[boundary.TopLeft], m.heightmap[boundary.BottomLeft])
		m.heightmap[topMid] = utils.Jitter(avg, spread)
	}
	if m.heightmap[rightMid] == 0 {
		avg := utils.Average(m.heightmap[boundary.TopRight], m.heightmap[boundary.BottomRight])
		m.heightmap[topMid] = utils.Jitter(avg, spread)
	}
	if m.heightmap[bottomMid] == 0 {
		avg := utils.Average(m.heightmap[boundary.BottomLeft], m.heightmap[boundary.BottomRight])
		m.heightmap[topMid] = utils.Jitter(avg, spread)
	}
	avg := utils.Average(m.heightmap[topMid], m.heightmap[leftMid], m.heightmap[rightMid], m.heightmap[bottomMid])
	m.heightmap[centre] = utils.Jitter(avg, spread)

	next := spread * reduce
	m.displace(utils.Rectangle{TopLeft: boundary.TopLeft, TopRight: topMid, BottomLeft: leftMid, BottomRight: centre}, next, reduce)
	m.displace(utils.Rectangle{TopLeft: topMid, TopRight: boundary.TopRight, BottomLeft: centre, BottomRight: rightMid}, next, reduce)
	m.displace(utils.Rectangle{TopLeft: leftMid, TopRight: centre, BottomLeft: boundary.BottomLeft, BottomRight: bottomMid}, next, reduce)
	m.displace(utils.Rectangle{TopLeft: centre, TopRight: rightMid, BottomLeft: bottomMid, BottomRight: boundary.BottomRight}, next, reduce)
}