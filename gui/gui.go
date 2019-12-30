package gui

import (
	"fmt"
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/inkyblackness/imgui-go"
	"github.com/ob6160/Terrain/gui/renderers"
	"log"
	"math"
	"os"
	"runtime"
)

type GUI struct {
	window *glfw.Window
	context *imgui.Context
	renderer *renderers.OpenGL3
	io imgui.IO
	buttonsPressed [3]bool
	time float64
}

func NewGUI(windowWidth, windowHeight int) (*GUI, error) {
	runtime.LockOSThread()
	var g = new(GUI)

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

func (g *GUI) Render() {
	w, h := g.window.GetSize()
	displaySize := [2]float32{float32(w), float32(h)}
	fw, fh := g.window.GetFramebufferSize()
	fbSize := [2]float32{float32(fw), float32(fh)}
	g.renderer.Render(displaySize, fbSize, imgui.RenderedDrawData())
}

func (g *GUI) Update() {
	w, h := g.window.GetSize()
	g.io.SetDisplaySize(imgui.Vec2{X: float32(w), Y: float32(h)})

	// Setup time step
	currentTime := glfw.GetTime()
	if g.time > 0 {
		g.io.SetDeltaTime(float32(currentTime - g.time))
	}
	g.time = currentTime

	// Setup inputs
	if g.window.GetAttrib(glfw.Focused) != 0 {
		x, y := g.window.GetCursorPos()
		g.io.SetMousePosition(imgui.Vec2{X: float32(x), Y: float32(y)})
	} else {
		g.io.SetMousePosition(imgui.Vec2{X: -math.MaxFloat32, Y: -math.MaxFloat32})
	}

	for i := 0; i < len(g.buttonsPressed); i++ {
		down := g.buttonsPressed[i] || (g.window.GetMouseButton(glfwButtonIDByIndex[i]) == glfw.Press)
		g.io.SetMouseButtonDown(i, down)
		g.buttonsPressed[i] = false
	}
}

func (g *GUI) Dispose() {
	g.renderer.Dispose()
	g.context.Destroy()
	g.window.Destroy()
	glfw.Terminate()
}

func (g* GUI) InitialiseGLFW(windowWidth, windowHeight int) *glfw.Window {
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
	buttonIndex, known := glfwButtonIndexByID[button]
	if known && (action == glfw.Press) {
		g.buttonsPressed[buttonIndex] = true
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