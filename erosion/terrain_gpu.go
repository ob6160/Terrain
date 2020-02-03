package erosion

import (
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/ob6160/Terrain/core"
	"github.com/ob6160/Terrain/generators"
	"github.com/ob6160/Terrain/utils"
	_ "github.com/ob6160/Terrain/utils"
	"math/rand"
)

type PackedData struct {
	heightData   []float32
	velocityData []float32
	outflowData  []float32
}

type GPUEroder struct {
	heightmap                                                                                              generators.TerrainGenerator
	simulationState                                                                                        *PackedData
	copyFrameBufferHeight, copyFrameBufferOutflow, copyFrameBufferVelocity                                 uint32
	displayFrameBuffer, displayTexture                                                                     uint32
	currentOutflowColorBuffer, currentVelocityColorBuffer, currentHeightColorBuffer                        uint32
	nextOutflowColorBuffer                                                                                 uint32 // o1, o2, o3, o4
	nextVelocityColorBuffer                                                                                uint32 // vX, vY
	nextHeightColorBuffer                                                                                  uint32 // landHeight, waterHeight, sediment
	waterPassProgram, outflowProgram, waterHeightProgram, velocityProgram, erosionProgram, sedimentProgram uint32
}

func NewGPUEroder(heightmap generators.TerrainGenerator) *GPUEroder {
	var e = new(GPUEroder)
	e.heightmap = heightmap
	width, height := heightmap.Dimensions()

	gl.GenTextures(1, &e.displayTexture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, e.displayTexture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(width),
		int32(height),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		nil)

	gl.GenFramebuffers(1, &e.displayFrameBuffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, e.displayFrameBuffer)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, e.displayTexture, 0)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	e.packData()
	e.setupShaders()
	e.setupTextures()
	return e
}

func (e *GPUEroder) BindDrawFramebuffer() {
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, e.displayFrameBuffer)
}

func (e *GPUEroder) BindHeightReadFramebuffer() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.copyFrameBufferHeight)
}

func (e *GPUEroder) BindOutflowReadFramebuffer() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.copyFrameBufferOutflow)
}

func (e *GPUEroder) BindVelocityReadFramebuffer() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.copyFrameBufferVelocity)
}

func (e *GPUEroder) DisplayTexture() uint32 {
	return e.displayTexture
}

func (e *GPUEroder) HeightColorbuffer() uint32 {
	return e.nextHeightColorBuffer
}

func (e *GPUEroder) packData() {
	var width, height = e.heightmap.Dimensions()
	heightmap := e.heightmap.Heightmap()
	packedData := PackedData{
		heightData:   make([]float32, (width)*(height)*4),
		velocityData: make([]float32, (width)*(height)*4),
		outflowData:  make([]float32, (width)*(height)*4),
	}
	// Place heightmap data into a packed array (for sending to GPU)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			index := utils.ToIndex(x, y, width)
			height := heightmap[index]
			location := (x + (y * width)) * 4
			packedData.heightData[location+0] = height         // height val
			packedData.heightData[location+1] = 0.0            // water height val
			packedData.heightData[location+2] = 0.0            // sediment val
			packedData.heightData[location+3] = rand.Float32() // rain rate
		}
	}

	e.simulationState = &packedData
}

func (e *GPUEroder) setupTextures() {
	var width, height = e.heightmap.Dimensions()


	// TODO: Abstract texture creation into its own function
	// TODO: Split up creation of the two stages of buffers into separate functions

	// Gen current textures
	gl.GenTextures(1, &e.currentHeightColorBuffer)
	gl.GenTextures(1, &e.currentOutflowColorBuffer)
	gl.GenTextures(1, &e.currentVelocityColorBuffer)

	gl.BindTexture(gl.TEXTURE_2D, e.currentHeightColorBuffer)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, int32(width), int32(height), 0, gl.RGBA, gl.FLOAT, nil)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.BindImageTexture(3, e.currentHeightColorBuffer, 0, false, 0, gl.READ_ONLY, gl.RGBA32F)
	
	gl.BindTexture(gl.TEXTURE_2D, e.currentOutflowColorBuffer)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, int32(width), int32(height), 0, gl.RGBA, gl.FLOAT, nil)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.BindImageTexture(4, e.currentOutflowColorBuffer, 0, false, 0, gl.READ_ONLY, gl.RGBA32F)

	gl.BindTexture(gl.TEXTURE_2D, e.currentVelocityColorBuffer)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, int32(width), int32(height), 0, gl.RGBA, gl.FLOAT, nil)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.BindImageTexture(5, e.currentVelocityColorBuffer, 0, false, 0, gl.READ_ONLY, gl.RGBA32F)
	
	// Gen next textures
	gl.GenTextures(1, &e.nextHeightColorBuffer)
	gl.GenTextures(1, &e.nextOutflowColorBuffer)
	gl.GenTextures(1, &e.nextVelocityColorBuffer)

	// These are used to write to (they represent the new state of the simulation).
	// BindFramebuffer textures as colour attachments to the FBO
	// Create texture for height, waterHeight, sediment
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, e.nextHeightColorBuffer)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, int32(width), int32(height), 0, gl.RGBA, gl.FLOAT, gl.Ptr(e.simulationState.heightData))
	gl.BindImageTexture(0, e.nextHeightColorBuffer, 0, false, 0, gl.READ_WRITE, gl.RGBA32F)

	// Create texture for Water Outflow
	gl.BindTexture(gl.TEXTURE_2D, e.nextOutflowColorBuffer)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, int32(width), int32(height), 0, gl.RGBA, gl.FLOAT, nil)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.BindImageTexture(1, e.nextOutflowColorBuffer, 0, false, 0, gl.READ_WRITE, gl.RGBA32F)

	// Create texture for velocity
	gl.BindTexture(gl.TEXTURE_2D, e.nextVelocityColorBuffer)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, int32(width), int32(height), 0, gl.RGBA, gl.FLOAT, nil)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.BindImageTexture(2, e.nextVelocityColorBuffer, 0, false, 0, gl.READ_WRITE, gl.RGBA32F)

	// Send the textures to a framebuffer for our bulk copy operation every pass.
	gl.GenFramebuffers(1, &e.copyFrameBufferHeight)
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.copyFrameBufferHeight)
	gl.FramebufferTexture2D(gl.READ_FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, e.nextHeightColorBuffer, 0)

	gl.GenFramebuffers(1, &e.copyFrameBufferOutflow)
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.copyFrameBufferOutflow)
	gl.FramebufferTexture2D(gl.READ_FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, e.nextOutflowColorBuffer, 0)

	gl.GenFramebuffers(1, &e.copyFrameBufferVelocity)
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.copyFrameBufferVelocity)
	gl.FramebufferTexture2D(gl.READ_FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, e.nextVelocityColorBuffer, 0)
}

func (e *GPUEroder) copyToCurrent() {
	width, height := e.heightmap.Dimensions()
	// Copy next to current at the start of each pass.
	// Expose the modified textures from last pass using a framebuffer for each.
	// Copy the bound framebuffer to the current texture for each.
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.copyFrameBufferHeight)
	gl.BindTexture(gl.TEXTURE_2D, e.currentHeightColorBuffer)
	gl.CopyTexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, 0, 0, int32(width), int32(height))

	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.copyFrameBufferOutflow)
	gl.BindTexture(gl.TEXTURE_2D, e.currentOutflowColorBuffer)
	gl.CopyTexSubImage2D(gl.TEXTURE_2D, 1, 0, 0, 0, 0, int32(width), int32(height))

	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.copyFrameBufferVelocity)
	gl.BindTexture(gl.TEXTURE_2D, e.currentVelocityColorBuffer)
	gl.CopyTexSubImage2D(gl.TEXTURE_2D, 2, 0, 0, 0, 0, int32(width), int32(height))

}

func (e *GPUEroder) Pass() {
	// Copy "next" textures into "current"
	width, height := e.heightmap.Dimensions()

	// Transfer the newly computed values from the previous pass into readonly buffers.
	e.copyToCurrent()

	// Render a plane to the FBO
	gl.UseProgram(e.waterPassProgram)
	gl.DispatchCompute(uint32(width), uint32(height), 1)
	gl.MemoryBarrier(gl.SHADER_IMAGE_ACCESS_BARRIER_BIT)

	gl.UseProgram(e.outflowProgram)
	gl.DispatchCompute(uint32(width), uint32(height), 1)
	gl.MemoryBarrier(gl.SHADER_IMAGE_ACCESS_BARRIER_BIT)
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
