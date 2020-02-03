package erosion

import (
	"fmt"
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/ob6160/Terrain/generators"
	"github.com/ob6160/Terrain/utils"
	"math"
	_ "math/rand"
)

type LayerData struct {
	heightmap         []float32
	outflowFlux       []mgl32.Vec4
	velocity          []mgl32.Vec2
	waterHeight       []float32
	suspendedSediment []float32
	rainRate          []float32
	tiltMap           []float32
}

type State struct {
	IsRaining                                                                                    bool
	WaterIncrementRate, GravitationalConstant, PipeCrossSectionalArea, EvaporationRate, TimeStep float32
	SedimentCarryCapacity, SoilSuspensionRate, SoilDepositionRate, MaximalErodeDepth             float32
}

type CPUEroder struct {
	initial                                     *LayerData
	swap                                        *LayerData
	state                                       *State
	running                                     bool
	width, height                               int
	WaterHeightBuffer, WaterHeightBufferTexture uint32
	HeightmapBuffer, HeightmapBufferTexture     uint32
	heightmap                                   *generators.TerrainGenerator
	iterations                                  int
}

func NewCPUEroder(heightmap generators.TerrainGenerator, state *State) *CPUEroder {
	width, height := heightmap.Dimensions()
	var eroder = CPUEroder{
		state:                    state,
		running:                  false,
		width:                    width,
		height:                   height,
		WaterHeightBuffer:        0,
		WaterHeightBufferTexture: 0,
		HeightmapBuffer:          0,
		HeightmapBufferTexture:   0,
		heightmap:                &heightmap,
		iterations:               0,
	}
	// Initialise layerdata
	eroder.Reset()

	return &eroder
}

func (t *CPUEroder) Initialise() {
	t.iterations = 0
	t.running = false

	// Water Height Update (from rainRate array or constant water sources).
	// Modify based on the constant rain volume array.
	gl.GenBuffers(1, &t.HeightmapBuffer)
	gl.BindBuffer(gl.TEXTURE_BUFFER, t.HeightmapBuffer)
	gl.BufferData(gl.TEXTURE_BUFFER, len(t.initial.heightmap)*4, gl.Ptr(t.initial.heightmap), gl.STATIC_DRAW)
	gl.GenTextures(1, &t.HeightmapBufferTexture)
	gl.BindBuffer(gl.TEXTURE_BUFFER, 0)

	// Setup water height buffer and associated storage.
	gl.GenBuffers(1, &t.WaterHeightBuffer)
	gl.BindBuffer(gl.TEXTURE_BUFFER, t.WaterHeightBuffer)
	gl.BufferData(gl.TEXTURE_BUFFER, len(t.swap.waterHeight)*4, gl.Ptr(t.swap.waterHeight), gl.STATIC_DRAW)
	gl.GenTextures(1, &t.WaterHeightBufferTexture)
	gl.BindBuffer(gl.TEXTURE_BUFFER, 0)
}

func WithinBounds(index, dimensions int) bool {
	if index >= 0 && index < dimensions {
		return true
	}
	return false
}

func (t *CPUEroder) Toggle() {
	t.running = !t.running
}

func (t *CPUEroder) newLayerData() *LayerData {
	var width, height = (*t.heightmap).Dimensions()
	initialCopy := make([]float32, (width+1)*(height+1))
	copy(initialCopy, (*t.heightmap).Heightmap())
	return &LayerData{
		rainRate: make([]float32, (width+1)*(height+1)),
		velocity: make([]mgl32.Vec2, (width+1)*(height+1)),
		// L=0, R=1, T=2, B=3
		outflowFlux:       make([]mgl32.Vec4, (width+1)*(height+1)),
		suspendedSediment: make([]float32, (width+1)*(height+1)),
		waterHeight:       make([]float32, (width+1)*(height+1)),
		heightmap:         initialCopy,
	}
}

func (t *CPUEroder) Reset() {
	t.initial = t.newLayerData()
	t.swap = t.newLayerData()
}

func (t *CPUEroder) IsRunning() bool {
	return t.running
}

func (t *CPUEroder) Update() {
	if t.running {
		t.SimulationStep()
	}
}

func (t *CPUEroder) UpdateBuffers() {
	// Update heightmap buffer data.
	gl.BindBuffer(gl.TEXTURE_BUFFER, t.HeightmapBuffer)
	gl.BufferSubData(gl.TEXTURE_BUFFER, 0, len(t.initial.heightmap)*4, gl.Ptr(t.initial.heightmap))
	gl.BindBuffer(gl.TEXTURE_BUFFER, 0)
	// Update water height buffer data.
	gl.BindBuffer(gl.TEXTURE_BUFFER, t.WaterHeightBuffer)
	gl.BufferSubData(gl.TEXTURE_BUFFER, 0, len(t.initial.waterHeight)*4, gl.Ptr(t.initial.waterHeight))
	gl.BindBuffer(gl.TEXTURE_BUFFER, 0)

	// Update the associated textures.
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_BUFFER, t.WaterHeightBufferTexture)
	gl.TexBuffer(gl.TEXTURE_BUFFER, gl.R32F, t.WaterHeightBuffer)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_BUFFER, t.HeightmapBufferTexture)
	gl.TexBuffer(gl.TEXTURE_BUFFER, gl.R32F, t.HeightmapBuffer)

}

func (t *CPUEroder) SimulationStep() {

	// == Shallow water flow simulation ==
	var initial = *t.initial
	var swap = *t.swap
	var dimensions = len(initial.heightmap)

	t.iterations++
	println(fmt.Sprintf("%d Iterations", t.iterations))

	for i := range t.initial.heightmap {
		if t.state.IsRaining {
			t.initial.waterHeight[i] += t.state.TimeStep * t.state.WaterIncrementRate
		}
	}

	// Water cell outflow flux calculation
	for x := 0; x < t.width; x++ {
		for y := 0; y < t.height; y++ {
			var i = utils.ToIndex(x, y, t.width)
			var iL, iR, iT, iB = initial.outflowFlux[i].Elem()
			var landHeight = initial.heightmap[i]
			var waterHeight = initial.waterHeight[i]

			var currentHeight = landHeight + waterHeight

			var pressure = t.state.TimeStep * t.state.PipeCrossSectionalArea * t.state.GravitationalConstant

			var leftIndex = utils.ToIndex(x-1, y, t.width)
			var leftOutflow = 0.0
			//TODO: Use the next water height as we've just incremented it with rain?
			if WithinBounds(leftIndex, dimensions) {
				var leftHeight = initial.heightmap[leftIndex]
				leftHeightDiff := currentHeight - (leftHeight + initial.waterHeight[leftIndex])
				leftOutflow = math.Max(0.0, float64(iL+pressure*leftHeightDiff))
			}

			var rightIndex = utils.ToIndex(x+1, y, t.width)
			var rightOutflow = 0.0
			if WithinBounds(rightIndex, dimensions) {
				var rightHeight = initial.heightmap[rightIndex]
				rightHeightDiff := currentHeight - (rightHeight + initial.waterHeight[rightIndex])
				rightOutflow = math.Max(0.0, float64(iR+pressure*rightHeightDiff))
			}

			var topIndex = utils.ToIndex(x, y-1, t.width)
			var topOutflow = 0.0
			if WithinBounds(topIndex, dimensions) {
				var topHeight = initial.heightmap[topIndex]
				topHeightDiff := currentHeight - (topHeight + initial.waterHeight[topIndex])
				topOutflow = math.Max(0, float64(iT+pressure*topHeightDiff))
			}

			var bottomIndex = utils.ToIndex(x, y+1, t.width)
			var bottomOutflow = 0.0
			if WithinBounds(bottomIndex, dimensions) {
				var bottomHeight = initial.heightmap[bottomIndex]
				bottomHeightDiff := currentHeight - (bottomHeight + initial.waterHeight[bottomIndex])
				bottomOutflow = math.Max(0, float64(iB+pressure*bottomHeightDiff))
			}

			if x == 0 {
				leftOutflow = 0.0
			}

			if y == 0 {
				topOutflow = 0.0
			}

			if x == t.width-1 {
				rightOutflow = 0.0
			}

			if y == t.height-1 {
				bottomOutflow = 0.0
			}

			// Find k
			var sumFluxOut = leftOutflow + rightOutflow + topOutflow + bottomOutflow
			var scaleFactor = math.Min(1.0,
				float64(waterHeight)/(sumFluxOut*float64(t.state.TimeStep)))

			// Calculate outflow for all four outgoing pipes at f(x,y)
			swap.outflowFlux[i] = mgl32.Vec4{
				float32(math.Max(0, leftOutflow*scaleFactor)),
				float32(math.Max(0, rightOutflow*scaleFactor)),
				float32(math.Max(0, topOutflow*scaleFactor)),
				float32(math.Max(0, bottomOutflow*scaleFactor)),
			}
		}
	}

	// Water height change calculation
	for x := 0; x < t.width; x++ {
		for y := 0; y < t.height; y++ {
			var i = utils.ToIndex(x, y, t.width)

			// Calculate inflow..
			// Right Pipe of the Left Neighbour + Left Pipe of the Right Neighbour + ...
			// TODO: Can we make a safe function that wraps out of bounds accesses on the grid?
			var leftIndex = utils.ToIndex(x-1, y, t.width)
			var leftCellInflow float32 = 0
			if WithinBounds(leftIndex, dimensions) {
				_, leftCellInflow, _, _ = swap.outflowFlux[leftIndex].Elem()
			}

			var rightIndex = utils.ToIndex(x+1, y, t.width)
			var rightCellInflow float32 = 0
			if WithinBounds(rightIndex, dimensions) {
				rightCellInflow, _, _, _ = swap.outflowFlux[rightIndex].Elem()
			}

			var topIndex = utils.ToIndex(x, y-1, t.width)
			var topCellInflow float32 = 0
			if WithinBounds(topIndex, dimensions) {
				_, _, _, topCellInflow = swap.outflowFlux[topIndex].Elem()
			}

			var bottomIndex = utils.ToIndex(x, y+1, t.width)
			var bottomCellInflow float32 = 0
			if WithinBounds(bottomIndex, dimensions) {
				_, _, bottomCellInflow, _ = swap.outflowFlux[bottomIndex].Elem()
			}

			// Calculate the outflow.
			var o1, o2, o3, o4 = swap.outflowFlux[i].Elem()

			var outFlow = o1 + o2 + o3 + o4
			var inFlow = leftCellInflow + rightCellInflow + topCellInflow + bottomCellInflow

			var TimeStepWaterHeight = t.state.TimeStep * (inFlow - outFlow)
			t.swap.waterHeight[i] += TimeStepWaterHeight
			t.swap.waterHeight[i] *= 1 - t.state.EvaporationRate*t.state.TimeStep
			t.swap.waterHeight[i] = float32(math.Max(0.0, float64(t.swap.waterHeight[i])))
		}
	}

	// Velocity Field calculation
	for x := 0; x < t.width; x++ {
		for y := 0; y < t.height; y++ {
			var i = utils.ToIndex(x, y, t.width)
			var li = utils.ToIndex(x-1, y, t.width)
			var ri = utils.ToIndex(x+1, y, t.width)
			var ti = utils.ToIndex(x, y-1, t.width)
			var bi = utils.ToIndex(x, y+1, t.width)

			var centreLeft, centreRight, centreTop, centreBottom = swap.outflowFlux[i].Elem()

			var leftInFlow float32 = 0
			if WithinBounds(li, dimensions) {
				_, leftInFlow, _, _ = swap.outflowFlux[li].Elem()
			}

			var rightInFlow float32 = 0
			if WithinBounds(ri, dimensions) {
				rightInFlow, _, _, _ = swap.outflowFlux[ri].Elem()
			}

			var topInFlow float32 = 0
			if WithinBounds(ti, dimensions) {
				_, _, _, topInFlow = swap.outflowFlux[ti].Elem()
			}

			var bottomInFlow float32 = 0
			if WithinBounds(bi, dimensions) {
				_, _, bottomInFlow, _ = swap.outflowFlux[bi].Elem()
			}

			var velX = (leftInFlow - centreLeft + centreRight - rightInFlow) * 0.5
			var velY = (topInFlow - centreTop + centreBottom - bottomInFlow) * 0.5
			t.swap.velocity[i] = mgl32.Vec2{velX, velY}
		}
	}

	for x := 0; x < t.width; x++ {
		for y := 0; y < t.height; y++ {
			var i = utils.ToIndex(x, y, t.width)
			var li = utils.ToIndex(x-1, y, t.width)
			var ri = utils.ToIndex(x+1, y, t.width)
			var ti = utils.ToIndex(x, y-1, t.width)
			var bi = utils.ToIndex(x, y+1, t.width)

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

			var dx = rh - lh
			var dy = th - bh

			var dxv = mgl32.Vec3{1, dx, 0}
			var dyv = mgl32.Vec3{0, dy, 1}
			var normal = dxv.Cross(dyv)
			var tiltAngle = (math.Abs(float64(normal.Y()))) / float64(normal.Len())

			var waterHeight = t.initial.waterHeight[i]
			var velocity = t.swap.velocity[i].Len()
			var sediment = t.initial.suspendedSediment[i]

			var maximum float32 = 0
			if waterHeight <= 0 {
				maximum = 0
			} else if waterHeight >= t.state.MaximalErodeDepth {
				maximum = 1
			} else {
				maximum = 1 - (t.state.MaximalErodeDepth-waterHeight)/t.state.MaximalErodeDepth
			}

			var carryCapacity = t.state.SedimentCarryCapacity * velocity * float32(math.Min(tiltAngle, 0.05)) * maximum

			if sediment < carryCapacity {
				var delta = t.state.TimeStep * t.state.SoilSuspensionRate * (carryCapacity - sediment)
				t.swap.heightmap[i] -= delta
				t.swap.suspendedSediment[i] += delta
				t.swap.waterHeight[i] += delta
			} else {
				var delta = t.state.TimeStep * t.state.SoilDepositionRate * (sediment - carryCapacity)
				t.swap.heightmap[i] += delta
				t.swap.suspendedSediment[i] -= delta
				t.swap.waterHeight[i] -= delta
			}

			t.swap.waterHeight[i] *= 1 - t.state.EvaporationRate*t.state.TimeStep
			t.swap.waterHeight[i] = float32(math.Max(0, float64(t.swap.waterHeight[i])))
			t.swap.heightmap[i] = float32(math.Max(0, float64(t.swap.heightmap[i])))
		}

		for x := 0; x < t.width; x++ {
			for y := 0; y < t.height; y++ {
				var i = utils.ToIndex(x, y, t.width)
				var vel = t.swap.velocity[i]
				var aVal float32 = 0.0
				var aCoord = utils.ToIndex(int(float32(x)-vel.X()*t.state.TimeStep), int(float32(y)-vel.Y()*t.state.TimeStep), t.width)
				aVal = t.initial.suspendedSediment[i]

				if WithinBounds(aCoord, dimensions) {
					t.swap.suspendedSediment[i] = t.initial.suspendedSediment[i] - aVal
					t.swap.suspendedSediment[aCoord] = t.initial.suspendedSediment[aCoord] + aVal
				}

			}
		}
	}

	// Copy swap into initial
	copy(t.initial.waterHeight, t.swap.waterHeight)
	copy(t.initial.rainRate, t.swap.rainRate)
	copy(t.initial.velocity, t.swap.velocity)
	copy(t.initial.outflowFlux, t.swap.outflowFlux)
	copy(t.initial.suspendedSediment, t.swap.suspendedSediment)
	copy(t.initial.heightmap, t.swap.heightmap)

	*t.initial, *t.swap = *t.swap, *t.initial

	//

	//}

	//*t.initial, *t.swap = *t.swap, *t.initial
	// Cell sediment carry capacity calculation

	// Erode / Deposit material based on the carry capacity

	// Move dissolved sediment along the water based on the velocity

	// Simulate water evaporation
}
