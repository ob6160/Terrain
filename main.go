package main

import "C"
import (
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/inkyblackness/imgui-go"
	"github.com/ob6160/Terrain/core"
	"github.com/ob6160/Terrain/erosion"
	"github.com/ob6160/Terrain/generators"
	"github.com/ob6160/Terrain/gui"
	_ "github.com/ob6160/Terrain/utils"
	"github.com/xlab/closer"
	"math"
	"time"
)

const (
	windowWidth = 1200
	windowHeight = 800
	vertexShaderPath = "./shaders/main.vert"
	fragShaderPath = "./shaders/main.frag"
)


type State struct {
	Program            uint32
	Uniforms           map[string]int32 //name -> handle
	Projection         mgl32.Mat4
	Camera             mgl32.Mat4
	CameraPos          mgl32.Vec3
	WorldPos           mgl32.Vec3
	TerrainHitPos      mgl32.Vec3
	Model              mgl32.Mat4
	MousePos           mgl32.Vec4
	Angle, Height, FOV float32
	Plane              *core.Plane
	MidpointGen *generators.MidpointDisplacement
	TerrainEroder *erosion.CPUEroder
	GPUEroder *erosion.GPUEroder
	Spread, Reduce float32
	ErosionState *erosion.State
	//UI
	DebugField []byte
	DebugFieldLen int32
	InfoValueString string
}

func setupUniforms(state *State) {
	var program = state.Program

	// Uniforms
	gl.UseProgram(program)

	state.Projection = mgl32.Perspective(mgl32.DegToRad(state.FOV), float32(windowWidth)/windowHeight, 0.01, 10000.0)
	//state.Projection = mgl32.Ortho(-state.Scale, state.Scale, -state.Scale, state.Scale, 0.01, 10000.0)
	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &state.Projection[0])

	state.Camera = mgl32.LookAtV(state.CameraPos, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &state.Camera[0])

	state.Model = mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &state.Model[0])

	state.TerrainHitPos = mgl32.Vec3{0,0,0}
	terrainHitPos := gl.GetUniformLocation(program, gl.Str("hitpos\x00"))
	gl.Uniform3fv(terrainHitPos, 1, &state.TerrainHitPos[0])

	textureUniform := gl.GetUniformLocation(program, gl.Str("tex\x00"))
	gl.Uniform1i(textureUniform, 0)

	angleUniform := gl.GetUniformLocation(program, gl.Str("angle\x00"))
	gl.Uniform1fv(angleUniform, 1, &state.Angle)
	
	heightUniform := gl.GetUniformLocation(program, gl.Str("height\x00"))
	gl.Uniform1fv(heightUniform, 1, &state.Height)

	waterHeightUniform := gl.GetUniformLocation(program, gl.Str("tboWaterHeight\x00"))
	gl.Uniform1i(waterHeightUniform, 0)

	heightmapUniform := gl.GetUniformLocation(program, gl.Str("tboHeightmap\x00"))
	gl.Uniform1i(heightmapUniform, 1)

	state.Uniforms["heightUniform"] = heightUniform
	state.Uniforms["projectionUniform"] = projectionUniform
	state.Uniforms["cameraUniform"] = cameraUniform
	state.Uniforms["modelUniform"] = modelUniform
	state.Uniforms["angleUniform"] = angleUniform
	state.Uniforms["terrainUniform"] = terrainHitPos
	state.Uniforms["waterHeightUniform"] = waterHeightUniform
	state.Uniforms["heightmapUniform"] = heightmapUniform
}

func main() {
	var newGUI, _ = gui.NewGUI(windowWidth, windowHeight)
	defer newGUI.Dispose()
	
	var testPlane = core.NewPlane(128,128)
	var midpointDisp = generators.NewMidPointDisplacement(64,64)
	midpointDisp.Generate(0.5, 0.5)

	var erosionState = erosion.State{
		WaterIncrementRate:     0.012,
		GravitationalConstant:  9.81,
		PipeCrossSectionalArea: 20,
		EvaporationRate:        0.015,
		TimeStep:               0.02,
		IsRaining: true,
		SedimentCarryCapacity: 1.0,
		SoilDepositionRate: 1.0,
		SoilSuspensionRate: 0.5,
		MaximalErodeDepth: 0.001,
	}
	var terrainEroder = erosion.NewCPUEroder(midpointDisp, &erosionState)
	var gpuEroder = erosion.NewGPUEroder(midpointDisp)
	
	// TODO: Move defaults into configurable constants.
	var state = &State{
		Program:         0,
		Uniforms:        make(map[string]int32),
		Projection:      mgl32.Mat4{},
		Camera:          mgl32.Mat4{},
		CameraPos:       mgl32.Vec3{},
		WorldPos:        mgl32.Vec3{-200, 200, -200},
		TerrainHitPos:   mgl32.Vec3{},
		Model:           mgl32.Mat4{},
		MousePos:        mgl32.Vec4{},
		Angle:           0,
		Height:          0.0,
		FOV:             30.0,
		Plane:           testPlane,
		MidpointGen:     midpointDisp,
		TerrainEroder:   terrainEroder,
		GPUEroder:       gpuEroder,
		Spread:          0.5,
		Reduce:          0.5,
		ErosionState:    &erosionState,
		DebugField:      make([]byte, 1000),
		DebugFieldLen:   0,
		InfoValueString: "",
	}


	program, err := core.NewProgramFromPath(vertexShaderPath, fragShaderPath)
	if err != nil {
		panic(err)
	}
	state.Program = program
	setupUniforms(state)

	// Setup terrain
	state.MidpointGen.Generate(state.Spread, state.Reduce)
	state.TerrainEroder = erosion.NewCPUEroder(midpointDisp, &erosionState)
	state.TerrainEroder.Initialise(midpointDisp.Heightmap())

	state.Plane.Construct(64, 64)
	
	state.TerrainEroder.SimulationStep()
	state.TerrainEroder.SimulationStep()

	exitC := make(chan struct{}, 1)
	doneC := make(chan struct{}, 1)
	closer.Bind(func() {
		close(exitC)
		<-doneC
	})

	fpsTicker := time.NewTicker(time.Second / 60)
	for {
		select {
		case <-exitC:
			fpsTicker.Stop()
			close(doneC)
			return
		case t := <-fpsTicker.C:
			if newGUI.ShouldClose() {
				close(exitC)
				continue
			}

			glfw.PollEvents()
			render(newGUI, state, t)
		}
	}
}

func updateUniforms(state *State) {
	state.CameraPos = mgl32.Rotate3DY(state.Angle).Mul3x1(state.WorldPos)
	state.Camera = mgl32.LookAtV(state.CameraPos, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	state.Projection = mgl32.Perspective(mgl32.DegToRad(state.FOV), float32(windowWidth)/windowHeight, 0.01, 10000.0)

	gl.UniformMatrix4fv(state.Uniforms["projectionUniform"], 1, false, &state.Projection[0])
	gl.UniformMatrix4fv(state.Uniforms["cameraUniform"], 1, false, &state.Camera[0])
	gl.Uniform1fv(state.Uniforms["heightUniform"], 1, &state.Height)
	gl.Uniform1fv(state.Uniforms["angleUniform"], 1, &state.Angle)
}

func render(gui *gui.GUI, state *State, timer time.Time) {
	gl.Enable(gl.DEPTH_TEST)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gui.Update()
	imgui.NewFrame()
	{
		imgui.SetNextWindowSize(imgui.Vec2{X: 200, Y: 100})
		imgui.Begin("Camera Settings")
		imgui.SliderFloat("FOV", &state.FOV, 0.0, 100.0)
		imgui.SliderFloat("Angle", &state.Angle, 0.0, math.Pi*2.0)
		imgui.End()
	}
	imgui.Render()

	// Start Terrain render

	gl.UseProgram(state.Program)
	updateUniforms(state)
	state.TerrainEroder.UpdateBuffers()
	gl.ActiveTexture(gl.TEXTURE0)
	gl.Uniform1i(state.Uniforms["waterHeightUniform"], 0)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.Uniform1i(state.Uniforms["heightmapUniform"], 1)

	state.Plane.M().Draw()

	// Start IMGUI render
	gui.Render()

	width, height := gui.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	gui.SwapBuffers()
}
