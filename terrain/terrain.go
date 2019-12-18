package terrain

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/ob6160/Terrain/generators"
	"math/rand"
)

type LayerData struct {
	heightmap []float32
	outflowFlux []mgl32.Vec4
	velocity []mgl32.Vec3
	waterHeight []float32
	suspendedSediment []float32
	rainRate []float32
}

type Terrain struct {
	initial *LayerData
	swap *LayerData
}

func NewTerrain(heightmap generators.TerrainGenerator) *Terrain {
	var width, height = heightmap.Dimensions()

	initial := LayerData{
		rainRate:          make([]float32, (width+1)*(height+1)),
		velocity:          make([]mgl32.Vec3, (width+1)*(height+1)),
		// L=0, R=1, T=2, B=3
		outflowFlux:       make([]mgl32.Vec4, (width+1)*(height+1)),
		suspendedSediment: make([]float32, (width+1)*(height+1)),
		waterHeight:       make([]float32, (width+1)*(height+1)),
		heightmap:         heightmap.Heightmap(),
	}

	swap := LayerData{
		rainRate:          make([]float32, (width+1)*(height+1)),
		velocity:          make([]mgl32.Vec3, (width+1)*(height+1)),
		outflowFlux:       make([]mgl32.Vec4, (width+1)*(height+1)),
		suspendedSediment: make([]float32, (width+1)*(height+1)),
		waterHeight:       make([]float32, (width+1)*(height+1)),
		heightmap:         heightmap.Heightmap(),
	}

	return &Terrain{
		initial: &initial,
		swap: &swap,
	}
}

func (t *Terrain) Initialise() {
	// Set a constant rain rate for each cell
	for i, _ := range t.initial.rainRate {
		var newVal = rand.Float32()
		t.initial.rainRate[i] = newVal
		t.swap.rainRate[i] = newVal
	}
	print(t.initial.rainRate)
}

const (
	WaterIncrementRate = 0.3
)

func (t *Terrain) SimulationStep() {
	// == Shallow water flow simulation ==
	var initial = t.initial
	var swap = t.swap
	var delta float32 = 1.0

	// Water Height Update (from rainRate array or constant water sources).
	// Modify based on the constant rain volume array.
	for i, _ := range initial.heightmap {
		swap.waterHeight[i] += initial.rainRate[i] * delta * WaterIncrementRate
	}

	// Water cell flow calculation
	for i, _ := range initial.heightmap {
		// Calculate outflow for all four outgoing pipes at f(x,y)
		for i := 0; i < 4; i++ {

		}

	}

	// Water height change calculation

	// Velocity Field calculation

	// Cell sediment carry capacity calculation

	// Erode / Deposit material based on the carry capacity

	// Move dissolved sediment along the water based on the velocity

	// Simulate water evaporation
}