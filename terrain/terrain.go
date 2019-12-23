package terrain

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/ob6160/Terrain/generators"
	"github.com/ob6160/Terrain/utils"
	"math"
	"math/rand"
	_ "math/rand"
)

type LayerData struct {
	heightmap []float32
	outflowFlux []mgl64.Vec4
	velocity []mgl64.Vec2
	waterHeight []float64
	suspendedSediment []float64
	rainRate []float64
	tiltMap []float64
}


type ErosionState struct {
	IsRaining bool
	WaterIncrementRate, GravitationalConstant, PipeCrossSectionalArea, EvaporationRate, TimeStep float64
	SedimentCarryCapacity, SoilSuspensionRate, MaximalErodeDepth float64
}

type Terrain struct {
	initial *LayerData
	swap *LayerData
	state *ErosionState
	width, height int
	heightmap []float32
	persistCopy []float32 // TODO: Can we keep a persistent copy somewhere better?
}

func NewTerrain(heightmap generators.TerrainGenerator, state* ErosionState) *Terrain {
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
		state: state,
		width: width,
		height: height,
	}
}

func (t *Terrain) Initialise(heightmap []float32) {
	t.heightmap = make([]float32, (t.width + 1) * (t.height + 1))
	t.persistCopy =  make([]float32, (t.width + 1) * (t.height + 1))
	
	copy(t.heightmap, heightmap)
	copy(t.persistCopy, heightmap)
	// Set a constant rain rate for each cell
	// TODO: Customise the area that is being rained on?
	// TODO: Single point sources, multiple point sources of custom radius.
	for i := range t.initial.rainRate {
		var val = rand.Float64()
		t.initial.rainRate[i] = val
		t.swap.rainRate[i] = val
	}
}

func (t *Terrain) Heightmap() []float32 {
	for i, _ := range t.initial.waterHeight {
		//t.heightmap[i] = float32(t.initial.velocity[i].Len()*5) + t.persistCopy[i] // Shows the velocity for each cell.
		//t.heightmap[i] = t.persistCopy[i] + float32(t.initial.waterHeight[i]) // Shows the height of the water overlayed on terrain
		//t.heightmap[i] = float32(t.initial.outflowFlux[i].Len() * 1)// Shows the outflow flux for each cell.
		t.heightmap[i] = t.initial.heightmap[i]
		//t.heightmap[i] = float32(t.initial.suspendedSediment[i]) * 100
	}
	return t.heightmap
}

func WithinBounds(index, dimensions int) bool {
	if index >= 0 && index < dimensions {
		return true
	}
	return false
}

func (t *Terrain) SimulationStep() {
	// == Shallow water flow simulation ==
	var initial = *t.initial
	var swap = *t.swap
	var dimensions = len(initial.heightmap)

	// Water Height Update (from rainRate array or constant water sources).
	// Modify based on the constant rain volume array.
	for i := range initial.heightmap {
		if t.state.IsRaining {
			swap.waterHeight[i] += initial.rainRate[i] * t.state.TimeStep * t.state.WaterIncrementRate
		} else {
			swap.waterHeight[i] += 0.000001
		}
	}

	// Water cell outflow flux calculation
	for x := 0; x < t.width; x++ {
		for y := 0; y < t.height; y++ {
			var i = utils.ToIndex(x, y, t.width)
			var iL, iR, iT, iB = initial.outflowFlux[i].Elem()
			var landHeight = float64(initial.heightmap[i])
			var waterHeight = swap.waterHeight[i]

			var leftIndex = utils.ToIndex(x - 1, y, t.width)
			var leftOutflow = 0.0
			if WithinBounds(leftIndex, dimensions) {
				var leftHeight = float64(initial.heightmap[leftIndex])
				// Change in height between current cell and cell to the immediate left.
				leftHeightDiff := (landHeight + waterHeight) - (leftHeight + swap.waterHeight[leftIndex])
				leftOutflow = math.Max(0, iL + t.state.TimeStep * t.state.PipeCrossSectionalArea * (t.state.GravitationalConstant * leftHeightDiff))
			}

			var rightIndex = utils.ToIndex(x + 1, y, t.width)
			var rightOutflow float64 = 0
			if WithinBounds(rightIndex, dimensions) {
				var rightHeight = float64(initial.heightmap[rightIndex])
				// Change in height between current cell and cell to the immediate left.
				rightHeightDiff := (landHeight + waterHeight) - (rightHeight + swap.waterHeight[rightIndex])
				rightOutflow = math.Max(0, iR + t.state.TimeStep * t.state.PipeCrossSectionalArea * (t.state.GravitationalConstant * rightHeightDiff))
			}

			var topIndex = utils.ToIndex(x, y - 1, t.width)
			var topOutflow float64 = 0
			if WithinBounds(topIndex, dimensions) {
				var topHeight = float64(initial.heightmap[topIndex])
				// Change in height between current cell and cell to the immediate left.
				topHeightDiff := (landHeight + waterHeight) - (topHeight + swap.waterHeight[topIndex])
				topOutflow = math.Max(0, iT + t.state.TimeStep * t.state.PipeCrossSectionalArea * (t.state.GravitationalConstant * topHeightDiff))

			}

			var bottomIndex = utils.ToIndex(x, y + 1, t.width)
			var bottomOutflow float64 = 0
			if WithinBounds(bottomIndex, dimensions) {
				var bottomHeight = float64(initial.heightmap[bottomIndex])
				// Change in height between current cell and cell to the immediate left.
				bottomHeightDiff := (landHeight + waterHeight) - (bottomHeight + swap.waterHeight[bottomIndex])
				bottomOutflow = math.Max(0, iB + t.state.TimeStep*t.state.PipeCrossSectionalArea*(t.state.GravitationalConstant*bottomHeightDiff))
			}

			var scaleFactor = math.Min(1, waterHeight / ((leftOutflow + rightOutflow + topOutflow + bottomOutflow) * t.state.TimeStep))

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
			// Calculate the outflow.
			var o1, o2, o3, o4 = swap.outflowFlux[i].Elem()
			var outFlow = o1 + o2 + o3 + o4
			// Calculate inflow..
			// Right Pipe of the Left Neighbour + Left Pipe of the Right Neighbour + ...
			// TODO: Can we make a safe function that wraps out of bounds accesses on the grid?
			var leftIndex = utils.ToIndex(x - 1, y, t.width)
			var leftCellInflow float64 = 0
			if WithinBounds(leftIndex, dimensions) {
				_, leftCellInflow, _, _ = swap.outflowFlux[leftIndex].Elem()
			}

			var rightIndex = utils.ToIndex(x + 1, y, t.width)
			var rightCellInflow float64 = 0
			if WithinBounds(rightIndex, dimensions) {
				rightCellInflow, _, _, _ = swap.outflowFlux[rightIndex].Elem()
			}

			var topIndex = utils.ToIndex(x, y - 1, t.width)
			var topCellInflow float64 = 0
			if WithinBounds(topIndex, dimensions) {
				_, _, _, topCellInflow = swap.outflowFlux[topIndex].Elem()
			}
			
			var bottomIndex = utils.ToIndex(x, y + 1, t.width)
			var bottomCellInflow float64 = 0
			if WithinBounds(bottomIndex, dimensions) {
				_, _, bottomCellInflow, _ = swap.outflowFlux[bottomIndex].Elem()
			}

			var inFlow = leftCellInflow + rightCellInflow + topCellInflow + bottomCellInflow

			var TimeStepWaterHeight = t.state.TimeStep * ( inFlow - outFlow )

			swap.waterHeight[i] += TimeStepWaterHeight
			swap.waterHeight[i] = math.Max(0, swap.waterHeight[i])
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
			if WithinBounds(li, dimensions) {
				_, leftInFlow, _, _ = swap.outflowFlux[li].Elem()
			}

			var rightInFlow float64 = 0
			if WithinBounds(ri, dimensions) {
				rightInFlow, _, _, _ = swap.outflowFlux[ri].Elem()
			}

			var topInFlow float64 = 0
			if WithinBounds(ti, dimensions) {
				_, _, _, topInFlow = swap.outflowFlux[ti].Elem()
			}

			var bottomInFlow float64 = 0
			if WithinBounds(bi, dimensions) {
				_, _, bottomInFlow, _ = swap.outflowFlux[bi].Elem()
			}

			var velX = (leftInFlow - centreLeft + centreRight - rightInFlow) / 2
			var velY = (topInFlow - centreTop + centreBottom - bottomInFlow) / 2
			t.swap.velocity[i] = mgl64.Vec2{velX, velY}
		}
	}

	for x := 0; x < t.width; x++ {
		for y := 0; y < t.height; y++ {
			// Plan:

			// Find delta in height between the Left and Right neighbours
			// Find delta in height between the Top and Bottom neighbours
			// Use deltas to find the normal
			// Take y coordinate of normal and divide through by magnitude to find the sin(tilt_angle)
			// If out of bounds, use the central value TODO: make a note of this in the writeup?
			// =========
			var dimensions = len(t.initial.heightmap)
			var i = utils.ToIndex(x, y, t.width)
			var li = utils.ToIndex(x - 1, y, t.width)
			var ri = utils.ToIndex(x + 1, y, t.width)
			var ti = utils.ToIndex(x, y - 1, t.width)
			var bi = utils.ToIndex(x, y + 1, t.width)

			var centralValue = t.initial.heightmap[i]

			var lh = centralValue
			if WithinBounds(li, dimensions) {
				lh = t.initial.heightmap[li]
			}

			var rh = centralValue
			if WithinBounds(ri, dimensions) {
				rh = t.initial.heightmap[ri]
			}

			var th = centralValue
			if WithinBounds(ti, dimensions) {
				th = t.initial.heightmap[ti]
			}

			var bh = centralValue
			if WithinBounds(bi, dimensions) {
				bh = t.initial.heightmap[bi]
			}

			var dx = float64(rh - lh)
			var dy = float64(th - bh)

			var dxv = mgl64.Vec3{2, dx, 0}
			var dyv = mgl64.Vec3{0, dy, 2}
			var normal = dxv.Cross(dyv)
			var tiltAngle = math.Abs(normal.Y()) / normal.Len()
			var sediment = t.initial.suspendedSediment[i]
			//var waterHeight = t.initial.waterHeight[i]
			var velocity = t.swap.velocity[i].Len()

			//var maximum = 1 - math.Max(0, t.state.MaximalErodeDepth - waterHeight) / t.state.MaximalErodeDepth
			var carryCapacity = t.state.SedimentCarryCapacity * velocity * math.Min(0.05, tiltAngle)

			if carryCapacity > sediment {
				var delta = t.state.TimeStep * t.state.SoilSuspensionRate * (carryCapacity - sediment)
				t.swap.heightmap[i] -= float32(delta)
				t.swap.suspendedSediment[i] += delta
				t.swap.waterHeight[i] += delta
			} else {
				var delta = t.state.TimeStep * t.state.SoilSuspensionRate * (sediment - carryCapacity)
				t.swap.heightmap[i] += float32(delta)
				t.swap.suspendedSediment[i] -= delta
				t.swap.waterHeight[i] -= delta
			}
			
			t.swap.waterHeight[i] *= 1 - t.state.EvaporationRate * t.state.TimeStep
		}

		for x := 0; x < t.width; x++ {
			for y := 0; y < t.height; y++ {
				var i = utils.ToIndex(x, y, t.width)
				var pos = mgl64.Vec2{float64(x), float64(y)}
				var vel = t.swap.velocity[i]
				var dVel = pos.Sub(vel.Mul(t.state.TimeStep))
	
				var a = mgl64.Vec2{math.Floor(dVel.X()), math.Floor(dVel.Y())}
				var b = mgl64.Vec2{math.Ceil(dVel.X()), math.Ceil(dVel.Y())}
				
				var i1Val = 0.0
				i1 := utils.ToIndex(int(a.X()), int(a.Y()), t.width)
				if WithinBounds(i1, dimensions) {
					i1Val = t.initial.suspendedSediment[i1]
				}
				var i2Val = 0.0
				i2 := utils.ToIndex(int(b.X()), int(a.Y()), t.width)
				if WithinBounds(i2, dimensions) {
					i2Val = t.initial.suspendedSediment[i2]
				}
				var i3Val = 0.0
				i3 := utils.ToIndex(int(a.X()), int(b.Y()), t.width)
				if WithinBounds(i3, dimensions) {
					i3Val = t.initial.suspendedSediment[i3]
				}
				var i4Val = 0.0
				i4 := utils.ToIndex(int(b.X()), int(b.Y()), t.width)
				if WithinBounds(i4, dimensions) {
					i4Val = t.initial.suspendedSediment[i4]
				}
				
				t.swap.suspendedSediment[i] = i1Val * (1 - dVel.X()) * (1- dVel.Y()) +
					i2Val * dVel.X() * (1 - dVel.Y()) +
					i3Val * (1 - dVel.X()) * dVel.Y() +
					i4Val * dVel.X() * dVel.Y()
				
			}
		}
	}

	*t.initial, *t.swap = *t.swap, *t.initial
	// Cell sediment carry capacity calculation

	// Erode / Deposit material based on the carry capacity

	// Move dissolved sediment along the water based on the velocity

	// Simulate water evaporation
}