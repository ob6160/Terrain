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
	return &MidpointDisplacement{width, height, make([]float32, width * height)}
}

func (m *MidpointDisplacement) Get(p utils.Point) float32 {
	return m.heightmap[p.ToIndex(m.width)]
}

func (m *MidpointDisplacement) set(p utils.Point, value float32) {
	m.heightmap[p.ToIndex(m.width)] = value
}

func (m *MidpointDisplacement) normalize() {
	var maxValue float32 = -1000
	var minValue float32 = 1000
	for i := 0; i < len(m.heightmap); i++ {
		if m.heightmap[i] > maxValue {
			maxValue = m.heightmap[i]
		}
		if m.heightmap[i] < minValue {
			minValue = m.heightmap[i]
		}
	}
	diff := maxValue - minValue

	for i := 0; i < len(m.heightmap); i++ {
		m.heightmap[i] = (m.heightmap[i] - minValue) / diff
	}
}

func (m *MidpointDisplacement) Generate() {
	m.heightmap = make([]float32, (m.width+1) * (m.height+1))
	// Set all four corners to random values
	topLeft := utils.Point{0,0}
	topRight := utils.Point{m.width, 0}
	bottomLeft := utils.Point{0, m.height}
	bottomRight := utils.Point{m.width, m.height}
	m.set(topLeft, rand.Float32())
	m.set(topRight, rand.Float32())
	m.set(bottomLeft, rand.Float32())
	m.set(bottomRight, rand.Float32())


	m.displace(topLeft.ToIndex(m.width),topRight.ToIndex(m.width),bottomLeft.ToIndex(m.width),bottomRight.ToIndex(m.width), 0.5, 0.6)
	m.normalize()
}

func (m *MidpointDisplacement) displace(tl, tr, bl, br int, spread, reduce float32) {
	if tr - tl <= m.width + 1 {
		return
	}
	var topMid, leftMid, rightMid, bottomMid, centre int
	topMid = utils.Midpoint(tl, tr)
	leftMid = utils.Midpoint(tl, bl)
	rightMid = utils.Midpoint(tr, br)
	bottomMid = utils.Midpoint(bl, br)
	centre = utils.Midpoint(leftMid, rightMid)

	if m.heightmap[topMid] == 0 {
		avg := utils.Average(m.heightmap[tl], m.heightmap[tr])
		m.heightmap[topMid] = utils.Jitter(avg, spread)
	}
	if m.heightmap[leftMid] == 0 {
		avg := utils.Average(m.heightmap[tl], m.heightmap[bl])
		m.heightmap[leftMid] = utils.Jitter(avg, spread)
	}
	if m.heightmap[rightMid] == 0 {
		avg := utils.Average(m.heightmap[tr], m.heightmap[br])
		m.heightmap[rightMid] = utils.Jitter(avg, spread)
	}
	if m.heightmap[bottomMid] == 0 {
		avg := utils.Average(m.heightmap[bl], m.heightmap[br])
		m.heightmap[bottomMid] = utils.Jitter(avg, spread)
	}
	if m.heightmap[centre] == 0 {
		avg := utils.Average(m.heightmap[topMid], m.heightmap[leftMid], m.heightmap[rightMid], m.heightmap[bottomMid])
		m.heightmap[centre] = utils.Jitter(avg, spread)
	}

	next := spread * reduce
	m.displace(tl, topMid, leftMid, centre, next, reduce)
	m.displace(topMid, tr, centre, rightMid, next, reduce)
	m.displace(leftMid, centre, bl, bottomMid, next, reduce)
	m.displace(centre, rightMid, bottomMid, br, next, reduce)
}