package terrain

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/ob6160/Terrain/generators"
	"github.com/ob6160/Terrain/utils"
	"math"
	_ "math/rand"
)

type LayerData struct {
	heightmap []float32
	outflowFlux []mgl64.Vec4
	velocity []mgl64.Vec2
	waterHeight []float64
	suspendedSediment []float64
	rainRate []float64
}

type Terrain struct {
	initial *LayerData
	swap *LayerData
	width, height int
	heightmap []float32
	persistCopy []float32
}

func NewTerrain(heightmap generators.TerrainGenerator) *Terrain {
	var width, height = heightmap.Dimensions()
	initialCopy := make([]float32, (width + 1) * (height + 1))
	copy(initialCopy, heightmap.Heightmap())
	initial := LayerData{
		rainRate:          make([]float64, (width+1)*(height+1)),
		velocity:          make([]mgl64.Vec2, (width+1)*(height+1)),
		// L=0, R=1, T=2, B=3
		outflowFlux:       make([]mgl64.Vec4, (width+1)*(height+1)),
		suspendedSediment: make([]float64, (width+1)*(height+1)),
		waterHeight:       make([]float64, (width+1)*(height+1)),
		heightmap:        initialCopy,
	}
	
	swapCopy := make([]float32, (width + 1) * (height + 1))
	copy(swapCopy, heightmap.Heightmap())
	swap := LayerData{
		rainRate:          make([]float64, (width+1)*(height+1)),
		velocity:          make([]mgl64.Vec2, (width+1)*(height+1)),
		outflowFlux:       make([]mgl64.Vec4, (width+1)*(height+1)),
		suspendedSediment: make([]float64, (width+1)*(height+1)),
		waterHeight:       make([]float64, (width+1)*(height+1)),
		heightmap:        swapCopy,
	}

	return &Terrain{
		initial: &initial,
		swap: &swap,
		width: width,
		height: height,
	}
}

const (
	WaterIncrementRate = 0.012
	GravitationalConstant = 9.81
	PipeCrossSectionalArea = 20
	//EvaporationRate = 0.015
	EvaporationRate = 0.6
	TimeStep = 0.02
)

func (t *Terrain) Initialise(heightmap []float32) {
	t.heightmap = make([]float32, (t.width + 1) * (t.height + 1))
	t.persistCopy =  make([]float32, (t.width + 1) * (t.height + 1))
	
	copy(t.heightmap, heightmap)
	copy(t.persistCopy, heightmap)
	// Set a constant rain rate for each cell
	for i := range t.initial.rainRate {
		var val = 0.5
		t.initial.rainRate[i] = val
		t.swap.rainRate[i] = val
	}

	/*for i := range initial.heightmap {
		swap.waterHeight[i] += initial.rainRate[i] * delta * WaterIncrementRate
	}*/
}

func (t *Terrain) Heightmap() []float32 {

	for i, _ := range t.initial.waterHeight {
		//t.heightmap[i] = float32(t.initial.velocity[i].Len())*100.0+ 0.3
		//t.heightmap[i] = float32(t.initial.velocity[i].Len()*5) + t.persistCopy[i]
				t.heightmap[i] = t.persistCopy[i] + float32(t.initial.waterHeight[i])
	}
	return t.heightmap
}

func (t *Terrain) SimulationStep() {
	// == Shallow water flow simulation ==
	var initial = *t.initial
	var swap = *t.swap
	var delta = TimeStep
	var dimensions = len(initial.heightmap)

	// Water Height Update (from rainRate array or constant water sources).
	// Modify based on the constant rain volume array.
	for i := range initial.heightmap {
		swap.waterHeight[i] += initial.rainRate[i] * delta * WaterIncrementRate
	}

	// Water cell outflow flux calculation
	for x := 0; x < t.width; x++ {
		for y := 0; y < t.height; y++ {
			var i = utils.ToIndex(x, y, t.width)
			var iL, iR, iT, iB = initial.outflowFlux[i].Elem()
			var landHeight = float64(initial.heightmap[i])
			var waterHeight = swap.waterHeight[i]

			var leftIndex = utils.ToIndex(x - 1, y, t.width)
			var leftOutflow float64 = 0
			if leftIndex >= 0 && leftIndex < dimensions {
				var leftHeight = float64(initial.heightmap[leftIndex])
				// Change in height between current cell and cell to the immediate left.
				leftHeightDiff := landHeight + waterHeight - leftHeight - swap.waterHeight[leftIndex]
				leftOutflow = math.Max(0, iL + delta * PipeCrossSectionalArea * (GravitationalConstant * leftHeightDiff))
			}

			var rightIndex = utils.ToIndex(x + 1, y, t.width)
			var rightOutflow float64 = 0
			if rightIndex >= 0 && rightIndex < dimensions {
				var rightHeight = float64(initial.heightmap[rightIndex])
				// Change in height between current cell and cell to the immediate left.
				rightHeightDiff := landHeight + waterHeight - rightHeight - swap.waterHeight[rightIndex]
				rightOutflow = math.Max(0, iR + delta * PipeCrossSectionalArea * (GravitationalConstant * rightHeightDiff))
			}


			var topIndex = utils.ToIndex(x, y - 1, t.width)
			var topOutflow float64 = 0
			if topIndex >= 0 && topIndex < dimensions {
				var topHeight = float64(initial.heightmap[topIndex])
				// Change in height between current cell and cell to the immediate left.
				topHeightDiff := landHeight + waterHeight - topHeight - swap.waterHeight[topIndex]
				topOutflow = math.Max(0, iT + delta * PipeCrossSectionalArea * (GravitationalConstant * topHeightDiff))

			}

			var bottomIndex = utils.ToIndex(x, y + 1, t.width)
			var bottomOutflow float64 = 0
			if bottomIndex >= 0 && bottomIndex < dimensions {
				var bottomHeight = float64(initial.heightmap[bottomIndex])
				// Change in height between current cell and cell to the immediate left.
				bottomHeightDiff := landHeight + waterHeight - bottomHeight - swap.waterHeight[bottomIndex]
				bottomOutflow = math.Max(0, iB + delta*PipeCrossSectionalArea*(GravitationalConstant*bottomHeightDiff))
			}

			var scaleFactor = math.Min(1, waterHeight / ((leftOutflow + rightOutflow + topOutflow + bottomOutflow) * delta))

			if x == 0  {
				leftOutflow = 0
			}
			
			if y == 0 {
				bottomOutflow = 0
			}
			
			if x == t.width {
				rightOutflow = 0
			}
			
			if y == t.height {
				topOutflow = 0
			}

			// Calculate outflow for all four outgoing pipes at f(x,y)
			swap.outflowFlux[i] = mgl64.Vec4{
				math.Max(0, leftOutflow * scaleFactor),
				math.Max(0, rightOutflow * scaleFactor),
				math.Max(0, topOutflow * scaleFactor),
				math.Max(0, bottomOutflow * scaleFactor),
			}
		}
	}

	// Water height change calculation
	for x := 0; x < t.width; x++ {
		for y := 0; y < t.height; y++ {
			var i = utils.ToIndex(x, y, t.width)
			//  delta * ( flow_in[4] - flow_out[4] )

			// Calculate the outflow.
			var o1, o2, o3, o4 = swap.outflowFlux[i].Elem()
			var outFlow = o1 + o2 + o3 + o4
			// Calculate inflow..
			// Right Pipe of the Left Neighbour + Left Pipe of the Right Neighbour + ...
			var leftIndex = utils.ToIndex(x - 1, y, t.width)
			var leftCellInflow float64 = 0
			if leftIndex >= 0 && leftIndex < dimensions {
				_, leftCellInflow, _, _ = swap.outflowFlux[leftIndex].Elem()
			}

			var rightIndex = utils.ToIndex(x + 1, y, t.width)
			var rightCellInflow float64 = 0
			if rightIndex >= 0 && rightIndex < dimensions {
				rightCellInflow, _, _, _ = swap.outflowFlux[rightIndex].Elem()
			}

			var topIndex = utils.ToIndex(x, y - 1, t.width)
			var topCellInflow float64 = 0
			if topIndex >= 0 && topIndex < dimensions {
				_, _, _, topCellInflow = swap.outflowFlux[topIndex].Elem()
			}
			
			var bottomIndex = utils.ToIndex(x, y + 1, t.width)
			var bottomCellInflow float64 = 0
			if bottomIndex >= 0 && bottomIndex < dimensions {
				_, _, bottomCellInflow, _ = swap.outflowFlux[bottomIndex].Elem()
			}

			var inFlow = leftCellInflow + rightCellInflow + topCellInflow + bottomCellInflow

			var deltaWaterHeight = delta * ( inFlow - outFlow )

			swap.waterHeight[i] += deltaWaterHeight
		}
	}
	// Velocity Field calculation
	for x := 0; x < t.width; x++ {
		for y := 0; y < t.height; y++ {
			var i  = utils.ToIndex(x, y, t.width)
			var li = utils.ToIndex(x - 1, y, t.width)
			var ri = utils.ToIndex(x + 1, y, t.width)
			var ti = utils.ToIndex(x, y - 1, t.width)
			var bi = utils.ToIndex(x, y + 1, t.width)

			var centreLeft, centreRight, centreTop, centreBottom = swap.outflowFlux[i].Elem()
			
			var leftInFlow float64 = 0
			if li >= 0 && li < dimensions {
				_, leftInFlow, _, _ = swap.outflowFlux[li].Elem()
			}

			var rightInFlow float64 = 0
			if ri >= 0 && ri < dimensions {
				rightInFlow, _, _, _ = swap.outflowFlux[ri].Elem()
			}
			
			var topInFlow float64 = 0
			if ti >= 0 && ti < dimensions {
				_, _, _, topInFlow = swap.outflowFlux[ti].Elem()
			}
			
			var bottomInFlow float64 = 0
			if bi >= 0 && bi < dimensions {
				_, _, bottomInFlow, _ = swap.outflowFlux[bi].Elem()
			}

			var velX = (leftInFlow - centreLeft + centreRight - rightInFlow) / 2
			var velY = (topInFlow - centreTop + centreBottom - bottomInFlow) / 2

			t.swap.velocity[i] = mgl64.Vec2{velX, velY}
			
			t.swap.waterHeight[i] *= 1 - EvaporationRate * delta
		}
	}

	*t.initial, *t.swap = *t.swap, *t.initial
	// Cell sediment carry capacity calculation

	// Erode / Deposit material based on the carry capacity

	// Move dissolved sediment along the water based on the velocity

	// Simulate water evaporation
}