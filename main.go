package main

import "C"
import (
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/inkyblackness/imgui-go/v2"
	"github.com/ob6160/Terrain/core"
	"github.com/ob6160/Terrain/erosion"
	"github.com/ob6160/Terrain/generators"
	"github.com/ob6160/Terrain/gui"
	"github.com/ob6160/Terrain/utils"
	_ "github.com/ob6160/Terrain/utils"
	"github.com/xlab/closer"
	"math"
	"time"
)

const (
	windowWidth      = 1200
	windowHeight     = 800
	vertexShaderPath = "./shaders/main.vert"
	fragShaderPath   = "./shaders/main.frag"
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
	MidpointGen        *generators.MidpointDisplacement
	TerrainEroder      *erosion.CPUEroder
	GPUEroder          *erosion.GPUEroder
	Spread, Reduce     float32
	ErosionState       *erosion.State
	//UI
	DebugField      []byte
	DebugFieldLen   int32
	InfoValueString string
	testTexture     uint32
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

	state.TerrainHitPos = mgl32.Vec3{0, 0, 0}
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

	var testPlane = core.NewPlane(512, 512)
	var midpointDisp = generators.NewMidPointDisplacement(512, 512)
	midpointDisp.Generate(0.5, 0.5)

	var erosionState = erosion.State{
		WaterIncrementRate:     0.012,
		GravitationalConstant:  9.8,
		PipeCrossSectionalArea: 20,
		EvaporationRate:        0.15,
		TimeStep:               0.002,
		IsRaining:              true,
		SedimentCarryCapacity:  2.0,
		SoilDepositionRate:     0.05,
		SoilSuspensionRate:     0.04,
		MaximalErodeDepth:      0.001,
	}
	var terrainEroder = erosion.NewCPUEroder(midpointDisp, &erosionState)
	var gpuEroder = erosion.NewGPUEroder(midpointDisp)

	testTexture, err := core.NewTexture("./demo_1.png")
	if err != nil {
		print("error loading image")
		print(err)
	}

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
		FOV:             50.0,
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
		testTexture:     testTexture,
	}

	program, err := core.NewProgramFromPath(vertexShaderPath, fragShaderPath)
	if err != nil {
		panic(err)
	}
	state.Program = program
	setupUniforms(state)

	// Setup terrain
	//state.MidpointGen.Generate(state.Spread, state.Reduce)
	//state.TerrainEroder = erosion.NewCPUEroder(midpointDisp, &erosionState)
	//state.TerrainEroder.Initialise()
	//state.Plane.Construct(256, 256)

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
			newGUI.Update()
			//state.TerrainEroder.Update()
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



func (coreState *State) renderUI(guiState *gui.State) {
	imgui.NewFrame()

	treeNodeFlags := imgui.TreeNodeFlagsDefaultOpen
	windowFlags := imgui.WindowFlagsMenuBar
	if imgui.BeginV("GPU Debug View", &guiState.GPUDebugWindowOpen, windowFlags) {
		imgui.Image(utils.FullColourTextureId(coreState.GPUEroder.DisplayTexture(), utils.RED), imgui.Vec2{256, 256})
		imgui.SameLine()
		imgui.Image(utils.FullColourTextureId(coreState.GPUEroder.DisplayTexture(), utils.GREEN), imgui.Vec2{256, 256})

		imgui.Image(utils.FullColourTextureId(coreState.GPUEroder.DisplayTexture(), utils.BLUE), imgui.Vec2{256, 256})
		imgui.SameLine()
		imgui.Image(utils.FullColourTextureId(coreState.GPUEroder.DisplayTexture(), utils.ALPHA), imgui.Vec2{256, 256})
	}
	imgui.End()

	if imgui.BeginV("Terrain", &guiState.TerrainWindowOpen, windowFlags) {
		if imgui.TreeNodeV("Camera", treeNodeFlags) {
			imgui.PushItemWidth(80)
			{
				imgui.SliderFloat("FOV", &coreState.FOV, 0.0, 100.0)
				imgui.SameLine()
				imgui.SliderFloat("Angle", &coreState.Angle, 0.0, math.Pi*2.0)
				imgui.PopItemWidth()
			}
			imgui.TreePop()
		}
		imgui.Separator()
		if imgui.TreeNodeV("Terrain", treeNodeFlags) {
			imgui.PushItemWidth(80)
			{
				imgui.SliderFloat("Height", &coreState.Height, 0.0, 100.0)
				imgui.SliderFloat("Spread", &coreState.Spread, 0.0, 10.0)
				imgui.SliderFloat("Reduce", &coreState.Reduce, 0.0, 10.0)
				imgui.PopItemWidth()
			}
			if imgui.Button("Regenerate Terrain") {
				coreState.MidpointGen.Generate(coreState.Spread, coreState.Reduce)
				coreState.TerrainEroder.Reset()
				coreState.TerrainEroder.Initialise()
			}
			imgui.TreePop()
		}
		imgui.Separator()
		if imgui.TreeNodeV("Simulation", treeNodeFlags) {
			runningLabel := "Start Simulation"
			if imgui.TreeNodeV("Control", treeNodeFlags) {
				if coreState.TerrainEroder.IsRunning() {
					runningLabel = "Stop Simulation"
				}
				if imgui.Button(runningLabel) {
					coreState.TerrainEroder.Toggle()
				}
				imgui.SameLine()
				if imgui.Button("Step Simulation") {
					coreState.TerrainEroder.SimulationStep()
					coreState.TerrainEroder.SimulationStep()
				}
				if imgui.Button("Reset Simulation") {
					coreState.TerrainEroder.Reset()
					coreState.TerrainEroder.Initialise()
				}
				imgui.TreePop()
			}
			if imgui.TreeNodeV("Settings", treeNodeFlags) {
				imgui.PushItemWidth(80)
				{
					imgui.SliderFloat("Delta Time", &coreState.ErosionState.TimeStep, 0.001, 0.05)
					imgui.SliderFloat("Evaporation Rate", &coreState.ErosionState.EvaporationRate, 0.001, 1.0)
					imgui.SliderFloat("Water Increment Rate", &coreState.ErosionState.WaterIncrementRate, 0.001, 0.01)
					imgui.PopItemWidth()
				}
				imgui.TreePop()
			}
			imgui.TreePop()
		}
	}
	imgui.End()
	imgui.EndFrame()
	imgui.Render()
}

func render(g *gui.GUI, coreState *State, timer time.Time) {
	gl.Enable(gl.DEPTH_TEST)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// Render Terrain CPU
	//{
	//	gl.UseProgram(coreState.Program)
	//	updateUniforms(coreState)
	//	coreState.TerrainEroder.UpdateBuffers()
	//	{
	//		gl.ActiveTexture(gl.TEXTURE0)
	//		gl.Uniform1i(coreState.Uniforms["waterHeightUniform"], 0)
	//		gl.ActiveTexture(gl.TEXTURE1)
	//		gl.Uniform1i(coreState.Uniforms["heightmapUniform"], 1)
	//	}
	//	coreState.Plane.M().Draw()
	//}

	// Render Terrain GPU
	//{
	//	gl.UseProgram(coreState.Program)
	//	updateUniforms(coreState)
	//	coreState.TerrainEroder.UpdateBuffers()
	//	{
	//		gl.ActiveTexture(gl.TEXTURE0)
	//		gl.Uniform1i(coreState.Uniforms["waterHeightUniform"], 0)
	//		gl.ActiveTexture(gl.TEXTURE1)
	//		gl.Uniform1i(coreState.Uniforms["heightmapUniform"], 1)
	//	}
	//	coreState.Plane.M().Draw()
	//}
	width, height := g.GetSize()
	coreState.GPUEroder.Pass()

	coreState.GPUEroder.BindDrawFramebuffer()
	coreState.GPUEroder.BindHeightReadFramebuffer()
	gl.BlitFramebuffer(0, 0, int32(width), int32(height),
		0, 0, int32(width), int32(height), gl.COLOR_BUFFER_BIT, gl.NEAREST)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, 0)
	// Render UI
	{
		g.Render(coreState.renderUI)
	}

	gl.Viewport(0, 0, int32(width), int32(height))
	g.SwapBuffers()
}
