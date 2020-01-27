package erosion

import (
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/ob6160/Terrain/core"
	"github.com/ob6160/Terrain/generators"
)
type GPUEroder struct {
	heightmap generators.TerrainGenerator
	frameBuffer uint32
	outflowColorBuffer uint32 // o1, o2, o3, o4
	velocityColorBuffer uint32 // vX, vY
	heightColorBuffer uint32 // landHeight, waterHeight, sediment
	waterPassProgram uint32
}

func NewGPUEroder(heightmap generators.TerrainGenerator) *GPUEroder {
	var e = new(GPUEroder)
	e.heightmap = heightmap
	e.setupShaders()
	e.setupTextures()
	e.setupFBO()
	return e
}

func (e *GPUEroder) setupShaders() {
	var waterProgram, err1 = core.NewComputeProgramFromPath("./shaders/WaterPass.comp")
	if err1 != nil {
		panic(err1)
	}

	var outflowProgram, err2 = core.NewComputeProgramFromPath("./shaders/OutFlow.comp")
	if err2 != nil {
		panic(err2)
	}

	var waterHeightProgram, err3 = core.NewComputeProgramFromPath("./shaders/WaterHeight.comp")
	if err3 != nil {
		panic(err3)
	}

	var velocityProgram, err4 = core.NewComputeProgramFromPath("./shaders/Velocity.comp")
	if err4 != nil {
		panic(err4)
	}

	var erosionProgram, err5 = core.NewComputeProgramFromPath("./shaders/Erosion.comp")
	if err5 != nil {
		panic(err5)
	}

	var sedimentProgram, err6 = core.NewComputeProgramFromPath("./shaders/Sediment.comp")
	if err6 != nil {
		panic(err6)
	}
	gl.UseProgram(waterProgram)
	gl.UseProgram(outflowProgram)
	gl.UseProgram(waterHeightProgram)
	gl.UseProgram(velocityProgram)
	gl.UseProgram(erosionProgram)
	gl.UseProgram(sedimentProgram)
}

func (e *GPUEroder) setupFBO() {
	gl.GenFramebuffers(1, &e.frameBuffer)
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.frameBuffer)
	// Bind textures as colour attachments to the FBO
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, e.heightColorBuffer, 0)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT1, gl.TEXTURE_2D, e.outflowColorBuffer, 0)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT2, gl.TEXTURE_2D, e.velocityColorBuffer, 0)
}

func (e *GPUEroder) setupTextures() {
	var width, height = e.heightmap.Dimensions()
	// Gen textures
	gl.GenTextures(1, &e.outflowColorBuffer)
	gl.GenTextures(1, &e.velocityColorBuffer)
	gl.GenTextures(1, &e.heightColorBuffer)

	// Create texture for height, waterHeight, sediment
	gl.BindTexture(gl.TEXTURE_2D, e.heightColorBuffer)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, int32(width), int32(height), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	// Create texture for Water Outflow
	gl.BindTexture(gl.TEXTURE_2D, e.outflowColorBuffer)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	// Create texture for velocity
	gl.BindTexture(gl.TEXTURE_2D, e.velocityColorBuffer)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RG, int32(width), int32(height), 0, gl.RG, gl.UNSIGNED_BYTE, nil)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
}

func (e *GPUEroder) Pass() {
	// Render a plane to the FBO
	width, height := e.heightmap.Dimensions()
	gl.DispatchCompute(uint32(width), uint32(height), 1)
}