package main

import (
	"C"
	"fmt"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang-ui/nuklear/nk"
	"github.com/ob6160/Terrain/core"
	"github.com/xlab/closer"
	"log"
	"runtime"
	"time"
)

const (
	windowWidth = 1200
	windowHeight = 800
	maxVertexBuffer  = 512 * 1024
	maxElementBuffer = 128 * 1024
	vertexShaderPath = "./shaders/main.vert"
	fragShaderPath = "./shaders/main.frag"
)

type NKState struct {
	bgColor nk.Color
	prop    int32
	text    nk.TextEdit
}

type State struct {
	Nk *NKState
	Program uint32
	Uniforms map[string]int32 //name -> handle
	Projection mgl32.Mat4
	Camera mgl32.Mat4
	Model mgl32.Mat4
	Angle float32
	Plane *core.Plane
}

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func setupOpenGl() *glfw.Window {
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
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

	gl.Enable(gl.DEPTH_TEST)

	return window
}

func setupUniforms(state *State) {
	var program = state.Program

	// Uniforms
	gl.UseProgram(program)

	state.Projection = mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 0.01, 10000.0)
	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &state.Projection[0])

	state.Camera = mgl32.LookAtV(mgl32.Vec3{-400, 200, -400}, mgl32.Vec3{100, 0, 100}, mgl32.Vec3{0, 1, 0})
	cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &state.Camera[0])

	state.Model = mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &state.Model[0])

	textureUniform := gl.GetUniformLocation(program, gl.Str("tex\x00"))
	gl.Uniform1i(textureUniform, 0)

	angleUniform := gl.GetUniformLocation(program, gl.Str("angle\x00"))
	gl.Uniform1fv(angleUniform, 1, &state.Angle)

	state.Uniforms["projectionUniform"] = projectionUniform
	state.Uniforms["cameraUniform"] = cameraUniform
	state.Uniforms["modelUniform"] = modelUniform
	state.Uniforms["angleUniform"] = angleUniform
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize GLFW:", err)
	}
	defer glfw.Terminate()

	window := setupOpenGl()

	
	var testPlane = core.NewPlane(1000,1000)
	testPlane.Construct()
	var state = &State{
		Nk: &NKState{
			bgColor: nk.NkRgba(28, 48, 62, 255),
		},
		Uniforms: make(map[string]int32),
		Plane: testPlane,
	}

	program, err := core.NewProgramFromPath(vertexShaderPath, fragShaderPath)
	if err != nil {
		panic(err)
	}
	state.Program = program

	setupUniforms(state)

	ctx := nk.NkPlatformInit(window, nk.PlatformInstallCallbacks)
	
	atlas := nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	// sansFont := nk.NkFontAtlasAddFromBytes(atlas, MustAsset("assets/FreeSans.ttf"), 16, nil)
	// config := nk.NkFontConfig(14)
	// config.SetOversample(1, 1)
	// config.SetRange(nk.NkFontChineseGlyphRanges())
	// simsunFont := nk.NkFontAtlasAddFromFile(atlas, "/Library/Fonts/Microsoft/SimHei.ttf", 14, &config)
	nk.NkFontStashEnd()
	// if simsunFont != nil {
	// 	nk.NkStyleSetFont(ctx, simsunFont.Handle())
	// }

	exitC := make(chan struct{}, 1)
	doneC := make(chan struct{}, 1)
	closer.Bind(func() {
		close(exitC)
		<-doneC
	})

	nk.NkTexteditInitDefault(&state.Nk.text)

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

func render(win *glfw.Window, ctx *nk.Context, state *State, time *time.Time) {

	nk.NkPlatformNewFrame()

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Update
		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time

		angle += float32(elapsed)
		model = mgl32.HomogRotate3D(angle, mgl32.Vec3{0, 1, 0})
		// Render
		gl.UseProgram(program)
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])
		gl.Uniform1fv(angleUniform, 1, &angle)

		mesh.Draw()

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()

	// Layout
	bounds := nk.NkRect(50, 50, 230, 250)
	update := nk.NkBegin(ctx, "Demo", bounds,
		nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle)
	if update > 0 {
		nk.NkLayoutRowStatic(ctx, 30, 80, 1)
		{
			if nk.NkButtonLabel(ctx, "button") > 0 {
				log.Println("[INFO] button pressed!")
			}
		}
	}

	nk.NkEnd(ctx)

	// Render
	bg := make([]float32, 4)
	nk.NkColorFv(bg, state.Nk.bgColor)
	width, height := win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.ClearColor(bg[0], bg[1], bg[2], bg[3])
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
	win.SwapBuffers()
}

func onError(code int32, msg string) {
	log.Printf("[glfw ERR]: error %d: %s", code, msg)
}
