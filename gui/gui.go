package gui

/**
 * Based upon code written by Christian Hans
 * https://github.com/inkyblackness/imgui-go-examples/blob/master/internal/platforms/glfw.go
 */

import (
	"fmt"
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/inkyblackness/imgui-go/v2"
	"github.com/ob6160/Terrain/gui/renderers"
	"log"
	"math"
	"os"
	"runtime"
)

type State struct {
	CameraWindowOpen, SimulationWindowOpen, TerrainWindowOpen, GPUDebugWindowOpen bool
	ButtonsPressed                                                                [3]bool
	Time                                                                          float64
}

type GUI struct {
	window   *glfw.Window
	context  *imgui.Context
	renderer *renderers.OpenGL3
	state    *State
	io       imgui.IO
}

func NewGUI(windowWidth, windowHeight int) (*GUI, error) {
	runtime.LockOSThread()
	var g = new(GUI)
	g.state = &State{
		CameraWindowOpen:     true,
		SimulationWindowOpen: true,
		ButtonsPressed:       [3]bool{},
		Time:                 0,
	}

	// Setup imgui
	g.context = imgui.CreateContext(nil)
	g.io = imgui.CurrentIO()

	// Setup glew, glfw
	g.window = g.InitialiseGLFW(windowWidth, windowHeight)

	g.installCallbacks()

	// Setup imgui renderer
	renderer, err := renderers.NewOpenGL3(g.io)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	g.renderer = renderer
	return g, nil
}

func (g *GUI) ShouldClose() bool {
	return g.window.ShouldClose()
}

func (g *GUI) SwapBuffers() {
	g.window.SwapBuffers()
}

func (g *GUI) GetSize() (int, int) {
	return g.window.GetSize()
}

func (g *GUI) Render(renderUI func(state *State)) {
	w, h := g.window.GetSize()
	displaySize := [2]float32{float32(w), float32(h)}
	fw, fh := g.window.GetFramebufferSize()
	fbSize := [2]float32{float32(fw), float32(fh)}
	renderUI(g.state)
	g.renderer.Render(displaySize, fbSize, imgui.RenderedDrawData())
}

func (g *GUI) Update() {
	state := *g.state
	w, h := g.window.GetSize()
	g.io.SetDisplaySize(imgui.Vec2{X: float32(w), Y: float32(h)})

	// Setup Time step
	currentTime := glfw.GetTime()
	if state.Time > 0 {
		g.io.SetDeltaTime(float32(currentTime - state.Time))
	}
	state.Time = currentTime

	// Setup inputs
	if g.window.GetAttrib(glfw.Focused) != 0 {
		x, y := g.window.GetCursorPos()
		g.io.SetMousePosition(imgui.Vec2{X: float32(x), Y: float32(y)})
	} else {
		g.io.SetMousePosition(imgui.Vec2{X: -math.MaxFloat32, Y: -math.MaxFloat32})
	}

	for i := 0; i < len(state.ButtonsPressed); i++ {
		down := state.ButtonsPressed[i] || (g.window.GetMouseButton(glfwButtonIDByIndex[i]) == glfw.Press)
		g.io.SetMouseButtonDown(i, down)
		state.ButtonsPressed[i] = false
	}
}

func (g *GUI) Dispose() {
	g.renderer.Dispose()
	g.context.Destroy()
	g.window.Destroy()
	glfw.Terminate()
}

func (g *GUI) InitialiseGLFW(windowWidth, windowHeight int) *glfw.Window {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize GLFW:", err)
	}
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
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
	fmt.Println("Maximum Max Uniform Block Size: ", gl.MAX_UNIFORM_BLOCK_SIZE)
	var workGroupInvocationsI int32
	var workGroupInvocationsJ int32
	var workGroupInvocationsK int32
	gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_COUNT, 0, &workGroupInvocationsI)
	gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_COUNT, 1, &workGroupInvocationsJ)
	gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_COUNT, 2, &workGroupInvocationsK)
	fmt.Println("Maximum work group count: x: ", workGroupInvocationsI, "y: ", workGroupInvocationsJ, "z: ", workGroupInvocationsK)
	return window
}

func (g *GUI) installCallbacks() {
	g.window.SetMouseButtonCallback(g.mouseButtonChange)
	g.window.SetScrollCallback(g.mouseScrollChange)
	g.window.SetKeyCallback(g.keyChange)
	g.window.SetCharCallback(g.charChange)
}

func (g *GUI) mouseScrollChange(w *glfw.Window, xoff float64, yoff float64) {
	g.io.AddMouseWheelDelta(float32(xoff), float32(yoff))
}

var glfwButtonIndexByID = map[glfw.MouseButton]int{
	glfw.MouseButton1: 0,
	glfw.MouseButton2: 1,
	glfw.MouseButton3: 2,
}

var glfwButtonIDByIndex = map[int]glfw.MouseButton{
	0: glfw.MouseButton1,
	1: glfw.MouseButton2,
	2: glfw.MouseButton3,
}

func (g *GUI) mouseButtonChange(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	state := *g.state
	buttonIndex, known := glfwButtonIndexByID[button]
	if known && (action == glfw.Press) {
		state.ButtonsPressed[buttonIndex] = true
	}
}

func (g *GUI) keyChange(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		g.io.KeyPress(int(key))
	}
	if action == glfw.Release {
		g.io.KeyRelease(int(key))
	}
	// Modifiers are not reliable across systems
	g.io.KeyCtrl(int(glfw.KeyLeftControl), int(glfw.KeyRightControl))
	g.io.KeyShift(int(glfw.KeyLeftShift), int(glfw.KeyRightShift))
	g.io.KeyAlt(int(glfw.KeyLeftAlt), int(glfw.KeyRightAlt))
	g.io.KeySuper(int(glfw.KeyLeftSuper), int(glfw.KeyRightSuper))
}

func (g *GUI) charChange(w *glfw.Window, char rune) {
	g.io.AddInputCharacters(string(char))
}
