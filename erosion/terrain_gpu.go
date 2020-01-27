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
	waterPassProgram, outflowProgram, waterHeightProgram, velocityProgram, erosionProgram, sedimentProgram uint32
}

func NewGPUEroder(heightmap generators.TerrainGenerator) *GPUEroder {
	var e = new(GPUEroder)
	e.heightmap = heightmap
	e.setupShaders()
	e.setupTextures()
	return e
}

func (e *GPUEroder) setupShaders() {
	var err error
	e.waterPassProgram, err = core.NewComputeProgramFromPath("./shaders/WaterPass.comp")
	if err != nil {
		panic(err)
	}

	e.outflowProgram, err = core.NewComputeProgramFromPath("./shaders/OutFlow.comp")
	if err != nil {
		panic(err)
	}

	e.waterHeightProgram, err = core.NewComputeProgramFromPath("./shaders/WaterHeight.comp")
	if err != nil {
		panic(err)
	}

	e.velocityProgram, err = core.NewComputeProgramFromPath("./shaders/Velocity.comp")
	if err != nil {
		panic(err)
	}

	e.erosionProgram, err = core.NewComputeProgramFromPath("./shaders/Erosion.comp")
	if err != nil {
		panic(err)
	}

	e.sedimentProgram, err = core.NewComputeProgramFromPath("./shaders/Sediment.comp")
	if err != nil {
		panic(err)
	}
}

func (e *GPUEroder) Bind() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.frameBuffer)
}

func (e *GPUEroder) setupTextures() {
	var width, height = e.heightmap.Dimensions()
	// Gen textures
	//gl.GenTextures(1, &e.outflowColorBuffer)
	//gl.GenTextures(1, &e.velocityColorBuffer)
	gl.GenTextures(1, &e.heightColorBuffer)

	// Bind textures as colour attachments to the FBO
	// Create texture for height, waterHeight, sediment
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, e.heightColorBuffer)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexStorage2D(gl.TEXTURE_2D, 1, gl.RGBA8, int32(width), int32(height))
	gl.BindImageTexture(0, e.heightColorBuffer, 0, false, 0, gl.READ_WRITE, gl.RGBA8)


	gl.GenFramebuffers(1, &e.frameBuffer)
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.frameBuffer)
	gl.FramebufferTexture2D(gl.READ_FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, e.heightColorBuffer, 0)
	// Create texture for Water Outflow
	//gl.ActiveTexture(gl.TEXTURE1)
	//gl.BindTexture(gl.TEXTURE_2D, e.outflowColorBuffer)
	//gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	//gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	//gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	//gl.FramebufferTexture2D(gl.READ_FRAMEBUFFER, gl.COLOR_ATTACHMENT1, gl.TEXTURE_2D, e.outflowColorBuffer, 0)
	//
	//// Create texture for velocity
	//gl.ActiveTexture(gl.TEXTURE2)
	//gl.BindTexture(gl.TEXTURE_2D, e.velocityColorBuffer)
	//gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	//gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	//gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RG, int32(width), int32(height), 0, gl.RG, gl.UNSIGNED_BYTE, nil)
	//gl.FramebufferTexture2D(gl.READ_FRAMEBUFFER, gl.COLOR_ATTACHMENT2, gl.TEXTURE_2D, e.velocityColorBuffer, 0)
}

func (e *GPUEroder) Pass() {
	// Render a plane to the FBO
	width, height := e.heightmap.Dimensions()
	gl.UseProgram(e.waterPassProgram)
	gl.DispatchCompute(uint32(width), uint32(height), 1)
	gl.MemoryBarrier(gl.SHADER_IMAGE_ACCESS_BARRIER_BIT)
}
