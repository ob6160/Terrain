package terrain

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/ob6160/Terrain/generators"
	"github.com/ob6160/Terrain/utils"
	"math"
	"math/rand"
)

type LayerData struct {
	heightmap []float32
	outflowFlux []mgl64.Vec4
	velocity []mgl64.Vec3
	waterHeight []float64
	suspendedSediment []float64
	rainRate []float64
}

type Terrain struct {
	initial *LayerData
	swap *LayerData
	width, height int
}

func NewTerrain(heightmap generators.TerrainGenerator) *Terrain {
	var width, height = heightmap.Dimensions()

	initial := LayerData{
		rainRate:          make([]float64, (width+1)*(height+1)),
		velocity:          make([]mgl64.Vec3, (width+1)*(height+1)),
		// L=0, R=1, T=2, B=3
		outflowFlux:       make([]mgl64.Vec4, (width+1)*(height+1)),
		suspendedSediment: make([]float64, (width+1)*(height+1)),
		waterHeight:       make([]float64, (width+1)*(height+1)),
		heightmap:         heightmap.Heightmap(),
	}

	swap := LayerData{
		rainRate:          make([]float64, (width+1)*(height+1)),
		velocity:          make([]mgl64.Vec3, (width+1)*(height+1)),
		outflowFlux:       make([]mgl64.Vec4, (width+1)*(height+1)),
		suspendedSediment: make([]float64, (width+1)*(height+1)),
		waterHeight:       make([]float64, (width+1)*(height+1)),
		heightmap:         heightmap.Heightmap(),
	}

	return &Terrain{
		initial: &initial,
		swap: &swap,
		width: width,
		height: height,
	}
}

func (t *Terrain) Initialise() {
	// Set a constant rain rate for each cell
	for i := range t.initial.rainRate {
		var newVal = rand.Float64()
		t.initial.rainRate[i] = newVal
		t.swap.rainRate[i] = newVal
	}
	print(t.initial.rainRate)
}

const (
	WaterIncrementRate = 0.012
	GravitationalConstant = 9.81
	PipeCrossSectionalArea = 20
)

func (t *Terrain) SimulationStep() {
	// == Shallow water flow simulation ==
	var initial = t.initial
	var swap = t.swap
	var delta = 1.0

	// Water Height Update (from rainRate array or constant water sources).
	// Modify based on the constant rain volume array.
	for i := range initial.heightmap {
		swap.waterHeight[i] += initial.rainRate[i] * delta * WaterIncrementRate
	}

	// Water cell flow calculation
	for x := 0; x < t.width; x++ {
		for y := 0; y < t.height; y++ {
			var i = utils.ToIndex(x, y, t.width)
			var initialOutflow = initial.outflowFlux[i]
			var landHeight = float64(initial.heightmap[i])
			var waterHeight = swap.waterHeight[i]

			var leftIndex = utils.ToIndex(x - 1, y, t.width)
			var leftOutflow float64 = 0
			if leftIndex >= 0 && leftIndex < t.width {
				var leftHeight = float64(initial.heightmap[leftIndex])
				// Change in height between current cell and cell to the immediate left.
				leftHeightDiff := landHeight + waterHeight - leftHeight - swap.waterHeight[leftIndex]
				leftOutflow = math.Max(0, delta * PipeCrossSectionalArea * (GravitationalConstant * leftHeightDiff))
			}

			var rightIndex = utils.ToIndex(x + 1, y, t.width)
			var rightOutflow float64 = 0
			if rightIndex >= 0 && rightIndex < t.width {
				var rightHeight = float64(initial.heightmap[rightIndex])
				// Change in height between current cell and cell to the immediate left.
				rightHeightDiff := landHeight + waterHeight - rightHeight - swap.waterHeight[rightIndex]
				rightOutflow = math.Max(0, delta * PipeCrossSectionalArea * (GravitationalConstant * rightHeightDiff))
			}


			var topIndex = utils.ToIndex(x, y - 1, t.width)
			var topOutflow float64 = 0
			if topIndex >= 0 && topIndex < t.height {
				var topHeight = float64(initial.heightmap[topIndex])
				// Change in height between current cell and cell to the immediate left.
				topHeightDiff := landHeight + waterHeight - topHeight - swap.waterHeight[topIndex]
				topOutflow = math.Max(0, delta * PipeCrossSectionalArea * (GravitationalConstant * topHeightDiff))

			}

			var bottomIndex = utils.ToIndex(x, y + 1, t.width)
			var bottomOutflow float64 = 0
			if bottomIndex >= 0 && bottomIndex < t.height {
				var bottomHeight = float64(initial.heightmap[bottomIndex])
				// Change in height between current cell and cell to the immediate left.
				bottomHeightDiff := landHeight + waterHeight - bottomHeight - swap.waterHeight[bottomIndex]
				bottomOutflow = math.Max(0, delta*PipeCrossSectionalArea*(GravitationalConstant*bottomHeightDiff))
			}

			var scaleFactor = math.Max(1, waterHeight / (leftOutflow + rightOutflow + topOutflow + bottomOutflow) * delta)
			// Calculate outflow for all four outgoing pipes at f(x,y)
			swap.outflowFlux[i] = mgl64.Vec4{
				leftOutflow * scaleFactor,
				rightOutflow * scaleFactor,
				topOutflow * scaleFactor,
				bottomOutflow * scaleFactor,
			}.Add(initialOutflow)
		}
	}

	// Water height change calculation

	// Velocity Field calculation

	// Cell sediment carry capacity calculation

	// Erode / Deposit material based on the carry capacity

	// Move dissolved sediment along the water based on the velocity

	// Simulate water evaporation
}