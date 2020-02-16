package erosion

import (
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/ob6160/Terrain/core"
	"github.com/ob6160/Terrain/generators"
	"github.com/ob6160/Terrain/utils"
	_ "github.com/ob6160/Terrain/utils"
	"math/rand"
	"unsafe"
)

type PackedData struct {
	heightData   []float32
	velocityData []float32
	outflowData  []float32
}

type GPUEroder struct {
	heightmap                                                                                              generators.TerrainGenerator
	simulationState                                                                                        *PackedData
	nextFrameBufferHeight, nextFrameBufferOutflow, nextFrameBufferVelocity                                 uint32
	displayFrameBufferHeight, displayTextureHeight                                                         uint32
	displayFrameBufferOutflow, displayTextureOutflow                                                       uint32
	currentOutflowColorBuffer, currentVelocityColorBuffer, currentHeightColorBuffer                        uint32
	nextOutflowColorBuffer                                                                                 uint32 // o1, o2, o3, o4
	nextVelocityColorBuffer                                                                                uint32 // vX, vY
	nextHeightColorBuffer                                                                                  uint32 // landHeight, waterHeight, sediment
	waterPassProgram, outflowProgram, waterHeightProgram, velocityProgram, erosionProgram, sedimentProgram uint32
}

func NewGPUEroder(heightmap generators.TerrainGenerator) *GPUEroder {
	var e = new(GPUEroder)
	e.heightmap = heightmap
	e.packData()
	e.loadComputeShaders()
	e.setupTextures()
	e.setupFramebuffers()
	return e
}

func (e *GPUEroder) BindOutflowDrawFramebuffer() {
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, e.displayFrameBufferOutflow)
}

func (e *GPUEroder) BindHeightDrawFramebuffer() {
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, e.displayFrameBufferHeight)
}

func (e *GPUEroder) BindNextHeightReadFramebuffer() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.nextFrameBufferHeight)
}

func (e *GPUEroder) BindNextOutflowReadFramebuffer() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.nextFrameBufferOutflow)
}

func (e *GPUEroder) BindNextVelocityReadFramebuffer() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.nextFrameBufferVelocity)
}

func (e *GPUEroder) HeightDisplayTexture() uint32 {
	return e.displayTextureHeight
}

func (e *GPUEroder) OutflowDisplayTexture() uint32 {
	return e.displayTextureOutflow
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
			if x < 512 && y < 512 && x > 450 && y > 450 {

				packedData.heightData[location+3] = rand.Float32() * 100.0 // rain rate
			}

			packedData.outflowData[location+0] = 0.0           // left outflow
			packedData.outflowData[location+1] = 0.0           // right outflow
			packedData.outflowData[location+2] = 0.0           // top outflow
			packedData.outflowData[location+3] = 0.0           // bottom outflow

			packedData.velocityData[location+0] = 0.0          // x velocity
			packedData.velocityData[location+1] = 0.0		   // y velocity
			packedData.velocityData[location+2] = 0.0          // nil
			packedData.velocityData[location+3] = 0.0          // nil
		}
	}

	e.simulationState = &packedData
}

func (e *GPUEroder) setupTextures() {
	var width, height = e.heightmap.Dimensions()

	// Display Textures
	// These are used to reference each texture for rendering to the screen as debug output.

	// Setup texture for height display.
	e.displayTextureHeight = core.NewTexture(width, height, nil)
	// Setup texture for outflow display.
	e.displayTextureOutflow = core.NewTexture(width, height, nil)

	// ===========================

	// State Textures
	// These are used to write to from the compute shader (they represent the new state of the simulation).
	// We eventually bind each texture as a colour attachment to a FBO

	// Next state textures (written to by the Compute Shader).

	/**
	 * Texture stored state:
	 * 	- Terrain Height
	 *  - Water Height
	 *  - Sediment
	 *  - Rain Rate
	 */
	e.nextHeightColorBuffer = createStateTexture(width, height, gl.Ptr(e.simulationState.heightData))
	gl.BindImageTexture(0, e.nextHeightColorBuffer, 0, false, 0, gl.READ_WRITE, gl.RGBA32F)

	/**
	 * Texture stored state:
	 * 	- left outflow
	 *  - right outflow
	 *  - top outflow
	 *  - bottom outflow
	 */
	e.nextOutflowColorBuffer = createStateTexture(width, height, gl.Ptr(e.simulationState.outflowData))
	gl.BindImageTexture(1, e.nextOutflowColorBuffer, 0, false, 0, gl.READ_WRITE, gl.RGBA32F)

	/**
	 * Texture stored state:
	 * 	- Vel X
	 *  - Vel Y
	 *  - nil
	 *  - nil
	 */
	e.nextVelocityColorBuffer = createStateTexture(width, height, gl.Ptr(e.simulationState.velocityData))
	gl.BindImageTexture(2, e.nextVelocityColorBuffer, 0, false, 0, gl.READ_WRITE, gl.RGBA32F)

	// Current state textures (written to by the Compute Shader).

	// Current state textures, the contents of the next shaders are copied to these at the end of each pass. //
	e.currentHeightColorBuffer = createStateTexture(width, height, gl.Ptr(e.simulationState.heightData))
	gl.BindImageTexture(3, e.currentHeightColorBuffer, 0, false, 0, gl.READ_ONLY, gl.RGBA32F)

	e.currentOutflowColorBuffer = createStateTexture(width, height, gl.Ptr(e.simulationState.velocityData))
	gl.BindImageTexture(4, e.currentOutflowColorBuffer, 0, false, 0, gl.READ_ONLY, gl.RGBA32F)

	e.currentVelocityColorBuffer = createStateTexture(width, height, gl.Ptr(e.simulationState.velocityData))
	gl.BindImageTexture(5, e.currentVelocityColorBuffer, 0, false, 0, gl.READ_ONLY, gl.RGBA32F)
}

func createStateTexture(width, height int, data unsafe.Pointer) uint32 {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, int32(width), int32(height), 0, gl.RGBA, gl.FLOAT, data)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	return texture
}

func (e *GPUEroder) setupFramebuffers() {
	// DISPLAY FRAMEBUFFERS

	gl.GenFramebuffers(1, &e.displayFrameBufferHeight)
	gl.BindFramebuffer(gl.FRAMEBUFFER, e.displayFrameBufferHeight)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, e.displayTextureHeight, 0)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	gl.GenFramebuffers(1, &e.displayFrameBufferOutflow)
	gl.BindFramebuffer(gl.FRAMEBUFFER, e.displayFrameBufferOutflow)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, e.displayTextureOutflow, 0)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// ===========================

	// STATE FRAMEBUFFERS

	// Generate and store references to each framebuffer.
	gl.GenFramebuffers(1, &e.nextFrameBufferHeight)
	gl.GenFramebuffers(1, &e.nextFrameBufferOutflow)
	gl.GenFramebuffers(1, &e.nextFrameBufferVelocity)

	// Attach each state to an associated read only framebuffer for bulk copy operation.

	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.nextFrameBufferHeight)
	gl.FramebufferTexture2D(gl.READ_FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, e.nextHeightColorBuffer, 0)

	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.nextFrameBufferOutflow)
	gl.FramebufferTexture2D(gl.READ_FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, e.nextOutflowColorBuffer, 0)

	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.nextFrameBufferVelocity)
	gl.FramebufferTexture2D(gl.READ_FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, e.nextVelocityColorBuffer, 0)

	// ===========================
}

/**
 * Copies the state of the last pass to the current textures, ready for the next.
 */
func (e *GPUEroder) copyNextToCurrent() {
	width, height := e.heightmap.Dimensions()
	// Copy next to current at the start of each pass.
	// Expose the modified textures from last pass using a framebuffer for each.
	// Copy the bound framebuffer to the current texture for each.
	e.BindNextHeightReadFramebuffer()
	gl.BindTexture(gl.TEXTURE_2D, e.currentHeightColorBuffer)
	gl.CopyTexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, 0, 0, int32(width), int32(height))

	e.BindNextOutflowReadFramebuffer()
	gl.BindTexture(gl.TEXTURE_2D, e.currentOutflowColorBuffer)
	gl.CopyTexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, 0, 0, int32(width), int32(height))

	e.BindNextVelocityReadFramebuffer()
	gl.BindTexture(gl.TEXTURE_2D, e.currentVelocityColorBuffer)
	gl.CopyTexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, 0, 0, int32(width), int32(height))
}

/**
 * Executes a single compute shader pipeline pass on the simulation state textures.
 */
func (e *GPUEroder) Pass() {
	// Copy "next" textures into "current"
	width, height := e.heightmap.Dimensions()

	// Transfer the newly computed values from the previous pass into readonly "current" buffers.
	e.copyNextToCurrent()

	const subdivideSize int = 32
	subW := uint32(width / subdivideSize)
	subH := uint32(height / subdivideSize)
	
	// Distribute new "water" across the terrain
	gl.UseProgram(e.waterPassProgram)
	gl.DispatchCompute(subW, subH, 1)
	gl.MemoryBarrier(gl.SHADER_IMAGE_ACCESS_BARRIER_BIT)

	// Calculate the movement of water across each cell of the terrain.
	gl.UseProgram(e.outflowProgram)
	gl.DispatchCompute(subW, subH, 1)
	gl.MemoryBarrier(gl.SHADER_IMAGE_ACCESS_BARRIER_BIT)

	// Calculate the resultant height of water in each cell based on previous step.
	gl.UseProgram(e.waterHeightProgram)
	gl.DispatchCompute(subW, subH, 1)
	gl.MemoryBarrier(gl.SHADER_IMAGE_ACCESS_BARRIER_BIT)

	// Calculate the velocity of water as it moves across the terrain.
	gl.UseProgram(e.velocityProgram)
	gl.DispatchCompute(subW, subH, 1)
	gl.MemoryBarrier(gl.SHADER_IMAGE_ACCESS_BARRIER_BIT)

	gl.UseProgram(e.erosionProgram)
	gl.DispatchCompute(subW, subH, 1)
	gl.MemoryBarrier(gl.SHADER_IMAGE_ACCESS_BARRIER_BIT)

	gl.UseProgram(e.sedimentProgram)
	gl.DispatchCompute(subW, subH, 1)
	gl.MemoryBarrier(gl.SHADER_IMAGE_ACCESS_BARRIER_BIT)
}

/**
 * Loads each compute shader in the pipeline.
 */
func (e *GPUEroder) loadComputeShaders() {
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
