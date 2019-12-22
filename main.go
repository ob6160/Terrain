package main

import (
	"fmt"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang-ui/nuklear/nk"
	"github.com/ob6160/Terrain/core"
	"github.com/ob6160/Terrain/generators"
	"github.com/ob6160/Terrain/terrain"
	_"github.com/ob6160/Terrain/utils"
	"github.com/xlab/closer"
	"gopkg.in/oleiade/reflections.v1"
	"log"
	"math"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	windowWidth = 1200
	windowHeight = 800
	maxVertexBuffer  = 512 * 1024
	maxElementBuffer = 128 * 1024
	strBufferSize int32 = 256 * 1024
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
	TerrainEroder *terrain.Terrain
	Spread, Reduce float32
	ErosionState *terrain.ErosionState
	//UI
	TerrainTreeState   nk.CollapseStates
	CameraTreeState nk.CollapseStates
	ErosionTreeState nk.CollapseStates
	DebugField []byte
	DebugFieldLen int32
	InfoValueString string
}


func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func setupOpenGl() *glfw.Window {
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Terrain", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	return window
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

	state.Uniforms["heightUniform"] = heightUniform
	state.Uniforms["projectionUniform"] = projectionUniform
	state.Uniforms["cameraUniform"] = cameraUniform
	state.Uniforms["modelUniform"] = modelUniform
	state.Uniforms["angleUniform"] = angleUniform
	state.Uniforms["terrainUniform"] = terrainHitPos
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize GLFW:", err)
	}
	defer glfw.Terminate()

	window := setupOpenGl()
	ctx := nk.NkPlatformInit(window, nk.PlatformInstallCallbacks)

	var testPlane = core.NewPlane(256,256)
	var midpointDisp = generators.NewMidPointDisplacement(256,256)
	midpointDisp.Generate(0.5, 0.5)
	
	var erosionState = terrain.ErosionState{
		WaterIncrementRate:     0.012,
		GravitationalConstant:  9.81,
		PipeCrossSectionalArea: 20,
		EvaporationRate:        0.2,
		TimeStep:               0.01,
	}
	var terrainEroder = terrain.NewTerrain(midpointDisp, &erosionState)

	// TODO: Move defaults into configurable constants.
	var state = &State{
		WorldPos: mgl32.Vec3{-200, 200, -200},
		Uniforms: make(map[string]int32),
		Plane: testPlane,
		FOV: 30.0,
		Height: 0.0,
		Spread: 0.5,
		Reduce: 0.5,
		MidpointGen: midpointDisp,
		TerrainEroder: terrainEroder,
		ErosionState: &erosionState,
		ErosionTreeState: nk.Maximized,
		TerrainTreeState: nk.Maximized,
		CameraTreeState: nk.Maximized,
		DebugField: make([]byte, 1000),
		DebugFieldLen: 0,
		InfoValueString: "",
	}


	// Setup terrain
	state.TerrainEroder.Initialise(midpointDisp.Heightmap())
	testPlane.Construct(midpointDisp)


	program, err := core.NewProgramFromPath(vertexShaderPath, fragShaderPath)
	if err != nil {
		panic(err)
	}
	state.Program = program
	setupUniforms(state)

	atlas := nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	nk.NkFontStashEnd()

	exitC := make(chan struct{}, 1)
	doneC := make(chan struct{}, 1)
	closer.Bind(func() {
		close(exitC)
		<-doneC
	})

	fpsTicker := time.NewTicker(time.Second / 30)
	for {
		select {
		case <-exitC:
			nk.NkPlatformShutdown()
			glfw.Terminate()
			fpsTicker.Stop()
			close(doneC)
			return
		case t := <-fpsTicker.C:
			if window.ShouldClose() {
				close(exitC)
				continue
			}

			glfw.PollEvents()

			render(window, ctx, state, t)
		}
	}
}

func render(win *glfw.Window, ctx *nk.Context, state *State, timer time.Time) {
	nk.NkPlatformNewFrame()
	gl.Enable(gl.DEPTH_TEST)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	width, height := win.GetSize()

	state.TerrainEroder.SimulationStep()
	state.MidpointGen.SetHeightmap(state.TerrainEroder.Heightmap())
	state.Plane.Construct(state.MidpointGen)

	state.CameraPos = mgl32.Rotate3DY(state.Angle).Mul3x1(state.WorldPos)
	state.Camera = mgl32.LookAtV(state.CameraPos, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	state.Projection = mgl32.Perspective(mgl32.DegToRad(state.FOV), float32(windowWidth)/windowHeight, 0.01, 10000.0)

	gl.UseProgram(state.Program)
	gl.UniformMatrix4fv(state.Uniforms["projectionUniform"], 1, false, &state.Projection[0])
	gl.UniformMatrix4fv(state.Uniforms["cameraUniform"], 1, false, &state.Camera[0])
	gl.Uniform1fv(state.Uniforms["heightUniform"], 1, &state.Height)
	gl.Uniform1fv(state.Uniforms["angleUniform"], 1, &state.Angle)

	state.Plane.M().Draw()

	// GUI
	simulBounds := nk.NkRect(50, 50, 300, 350)
	simulUpdate := nk.NkBegin(ctx, "Simulation Controls", simulBounds,
		nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle)

	// TODO: Abstract UI into its own namespace/module
	if simulUpdate > 0 {
		// Camera Settings Panel
		if nk.NkTreeStatePush(ctx, nk.TreeTab, "Camera", &state.CameraTreeState) > 0 {
			nk.NkLayoutRowDynamic(ctx, 15, 3)
			{
				nk.NkLabel(ctx, "Angle", nk.TextAlignLeft)
				newAngle := nk.NkSlideFloat(ctx, 0.0, state.Angle, math.Pi*2, 0.01)
				if newAngle != state.Angle {
					state.Angle = newAngle
				}
				state.InfoValueString = fmt.Sprintf("%.1f", state.Angle)
				if len(state.InfoValueString) != 0 {
					nk.NkLabel(ctx, state.InfoValueString, nk.TextAlignRight)
				}
			}
			nk.NkLayoutRowDynamic(ctx, 15, 3)
			{
				nk.NkLabel(ctx,"FOV", nk.TextAlignLeft)
				newFOV := nk.NkSlideFloat(ctx, 0.0, state.FOV, 120.0, 1.0)
				if newFOV != state.FOV {
					state.FOV = newFOV
				}
				state.InfoValueString = fmt.Sprintf("%.1f", newFOV)
				if len(state.InfoValueString) != 0 {
					nk.NkLabel(ctx, state.InfoValueString, nk.TextAlignRight)
				}
			}
			nk.NkTreePop(ctx)
		}
		// Terrain Settings Panel
		if nk.NkTreeStatePush(ctx, nk.TreeTab, "Terrain", &state.TerrainTreeState) > 0 {
			if nk.NkButtonLabel(ctx, "Recalc Terrain") > 0 {

				state.MidpointGen.Generate(state.Spread, state.Reduce)
				state.TerrainEroder = terrain.NewTerrain(state.MidpointGen, state.ErosionState)
				state.TerrainEroder.Initialise(state.MidpointGen.Heightmap())
				state.Plane.Construct(state.MidpointGen)
			}
			// TODO: Abstract out into a function so we reduce code repetition.
			nk.NkLayoutRowDynamic(ctx, 15, 3)
			{
				nk.NkLabel(ctx, "Height", nk.TextAlignLeft)
				newHeight := nk.NkSlideFloat(ctx, -200.0, state.Height, 200.0, 0.3)
				if newHeight != state.Height {
					state.Height = newHeight
				}
				state.InfoValueString = fmt.Sprintf("%.1f",  state.Height)
				if len(state.InfoValueString) != 0 {
					nk.NkLabel(ctx, state.InfoValueString, nk.TextAlignRight)
				}
			}
			nk.NkLayoutRowDynamic(ctx, 15, 3)
			{
				nk.NkLabel(ctx, "Spread", nk.TextAlignLeft)
				newSpread := nk.NkSlideFloat(ctx, 0.0, state.Spread, 2.0, 0.01)
				if newSpread != state.Spread {
					state.Spread = newSpread
				}
				state.InfoValueString = fmt.Sprintf("%.1f",  state.Spread)
				if len(state.InfoValueString) != 0 {
					nk.NkLabel(ctx, state.InfoValueString, nk.TextAlignRight)
				}
			}
			nk.NkLayoutRowDynamic(ctx, 15, 3)
			{
				nk.NkLabel(ctx, "Reduce", nk.TextAlignLeft)
				newReduce := nk.NkSlideFloat(ctx, 0.0, state.Reduce, 2.0, 0.01)
				if newReduce != state.Reduce {
					state.Reduce = newReduce
				}
				state.InfoValueString = fmt.Sprintf("%.1f",  state.Reduce)
				if len(state.InfoValueString) != 0 {
					nk.NkLabel(ctx, state.InfoValueString, nk.TextAlignRight)
				}
			}
			nk.NkTreePop(ctx)
		}
		// Debug Text Input / Output

		if nk.NkTreeStatePush(ctx, nk.TreeTab, "Erosion", &state.ErosionTreeState) > 0 {
			nk.NkLayoutRowDynamic(ctx, 15, 3)
			{
				nk.NkLabel(ctx, "Water Incrmt Rate", nk.TextAlignLeft)
				newRate := nk.NkSlideFloat(ctx, 0.001, float32(state.ErosionState.WaterIncrementRate), 0.2, 0.0001)
				if newRate != float32(state.ErosionState.WaterIncrementRate) {
					state.ErosionState.WaterIncrementRate = float64(newRate)
				}
				state.InfoValueString = fmt.Sprintf("%.3f", state.ErosionState.WaterIncrementRate)
				if len(state.InfoValueString) != 0 {
					nk.NkLabel(ctx, state.InfoValueString, nk.TextAlignRight)
				}
			}

			nk.NkLayoutRowDynamic(ctx, 15, 3)
			{
				nk.NkLabel(ctx, "Gravitational Constant", nk.TextAlignLeft)
				newRate := nk.NkSlideFloat(ctx, 0.0, float32(state.ErosionState.GravitationalConstant), 20.0, 0.1)
				if newRate != float32(state.ErosionState.GravitationalConstant) {
					state.ErosionState.GravitationalConstant = float64(newRate)
				}
				state.InfoValueString = fmt.Sprintf("%.3f", state.ErosionState.GravitationalConstant)
				if len(state.InfoValueString) != 0 {
					nk.NkLabel(ctx, state.InfoValueString, nk.TextAlignRight)
				}
			}

			nk.NkLayoutRowDynamic(ctx, 15, 3)
			{
				nk.NkLabel(ctx, "Pipe Cross Section Area", nk.TextAlignLeft)
				newRate := nk.NkSlideFloat(ctx, 0.0, float32(state.ErosionState.PipeCrossSectionalArea), 100.0, 1.0)
				if newRate != float32(state.ErosionState.PipeCrossSectionalArea) {
					state.ErosionState.PipeCrossSectionalArea = float64(newRate)
				}
				state.InfoValueString = fmt.Sprintf("%.3f", state.ErosionState.PipeCrossSectionalArea)
				if len(state.InfoValueString) != 0 {
					nk.NkLabel(ctx, state.InfoValueString, nk.TextAlignRight)
				}
			}

			nk.NkLayoutRowDynamic(ctx, 15, 3)
			{
				nk.NkLabel(ctx, "Evaporation Rate", nk.TextAlignLeft)
				newRate := nk.NkSlideFloat(ctx, 0.0, float32(state.ErosionState.EvaporationRate), 2.0, 0.01)
				if newRate != float32(state.ErosionState.EvaporationRate) {
					state.ErosionState.EvaporationRate = float64(newRate)
				}
				state.InfoValueString = fmt.Sprintf("%.3f", state.ErosionState.EvaporationRate)
				if len(state.InfoValueString) != 0 {
					nk.NkLabel(ctx, state.InfoValueString, nk.TextAlignRight)
				}
			}

			nk.NkLayoutRowDynamic(ctx, 15, 3)
			{
				nk.NkLabel(ctx, "Timestep", nk.TextAlignLeft)
				newRate := nk.NkSlideFloat(ctx, 0.0, float32(state.ErosionState.TimeStep), 2.0, 0.001)
				if newRate != float32(state.ErosionState.TimeStep) {
					state.ErosionState.TimeStep = float64(newRate)
				}
				state.InfoValueString = fmt.Sprintf("%.3f", state.ErosionState.TimeStep)
				if len(state.InfoValueString) != 0 {
					nk.NkLabel(ctx, state.InfoValueString, nk.TextAlignRight)
				}
			}
			
			nk.NkTreePop(ctx)
		}
	}

	nk.NkEnd(ctx)

	debugBounds := nk.NkRect(windowWidth - 350, windowHeight - 200, 300, 175)
	debugUpdate := nk.NkBegin(ctx, "Debug Console", debugBounds, nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle)
	// Fix for faulty enter key handling
	// TODO: Abstract debug console handling logic out into its own module
	if debugUpdate > 0 {
		nk.NkLayoutRowBegin(ctx, nk.Static, 30, 2)
		{
			nk.NkLayoutRowPush(ctx, 230)
			nk.NkEditString(ctx, nk.EditField, state.DebugField, &state.DebugFieldLen, strBufferSize, nk.NkFilterDefault)
			nk.NkLayoutRowPush(ctx, 30)

			if nk.NkButtonLabel(ctx, ">>") > 0 ||  win.GetKey(glfw.KeyEnter) == glfw.Press {
				// TODO: Store settings in a Map so we don't need to do ugly reflection here.
				if state.DebugFieldLen > 0 {
					input := string(state.DebugField[:state.DebugFieldLen])
					cmdSplitted := strings.Split(input, "/")[1]
					cmd := strings.Split(cmdSplitted, " ")
					print(fmt.Sprintf("Command: %s\n", input))
					switch strings.ToLower(cmd[0]) {
					case "mode":
						var mesh = state.Plane.M()
						val := cmd[1]
						switch strings.ToLower(val) {
						case "triangles":
							mesh.RenderMode = gl.TRIANGLES
						case "lines":
							mesh.RenderMode = gl.LINES
						}
					case "regen":
						state.MidpointGen.Generate(state.Spread, state.Reduce)
						state.Plane.Construct(state.MidpointGen)
					case "step":
						for i:= 0; i < 100; i++ {
							state.TerrainEroder.SimulationStep()
						}
						state.MidpointGen.SetHeightmap(state.TerrainEroder.Heightmap())
						state.Plane.Construct(state.MidpointGen)
					case "set":
						key := cmd[1]
						val, _ := strconv.ParseFloat(cmd[2], 32)
						print(fmt.Sprintf("Key: %s, Val: %f\n", key, val))
						err := reflections.SetField(state, key, float32(val))
						if err != nil {
							fmt.Print(err)
						}
					}
					state.DebugFieldLen = 0
					state.DebugField = make([]byte, 1000)
				}
			}
		}
	}

	nk.NkEnd(ctx)
	gl.Viewport(0, 0, int32(width), int32(height))
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
	win.SwapBuffers()
}
