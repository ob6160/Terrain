package erosion

import (
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/ob6160/Terrain/core"
	"github.com/ob6160/Terrain/generators"
	"github.com/ob6160/Terrain/utils"
	_"github.com/ob6160/Terrain/utils"
	"math/rand"
)

type PackedData struct {
	heightData []float32
	velocityData []float32
	outflowData []float32
}

type GPUEroder struct {
	heightmap generators.TerrainGenerator
	simulationState *PackedData
	frameBuffer uint32
	outflowColorBuffer uint32 // o1, o2, o3, o4
	velocityColorBuffer uint32 // vX, vY
	heightColorBuffer uint32 // landHeight, waterHeight, sediment
	waterPassProgram, outflowProgram, waterHeightProgram, velocityProgram, erosionProgram, sedimentProgram uint32
}

func NewGPUEroder(heightmap generators.TerrainGenerator) *GPUEroder {
	var e = new(GPUEroder)
	e.heightmap = heightmap
	e.packData()
	e.setupShaders()
	e.setupTextures()
	return e
}


func (e *GPUEroder) Bind() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.frameBuffer)
}

func (e *GPUEroder) packData() {
	var width, height = e.heightmap.Dimensions()
	heightmap := e.heightmap.Heightmap()
	packedData := PackedData{
		heightData:   make([]float32, (width) * (height) * 4),
		velocityData: make([]float32, (width) * (height) * 4),
		outflowData:  make([]float32, (width) * (height) * 4),
	}
	// Place heightmap data into a packed array (for sending to GPU)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			index := utils.ToIndex(x, y, width)
			height := heightmap[index]
			location := (x + (y * width)) * 4
			packedData.heightData[location + 0] = height // height val
			packedData.heightData[location + 1] = 0.0 // water height val
			packedData.heightData[location + 2] = 0.0 // sediment val
			packedData.heightData[location + 3] = rand.Float32() // rain rate
		}
	}

	e.simulationState = &packedData
}

func (e *GPUEroder) setupTextures() {
	var width, height = e.heightmap.Dimensions()
	
	// Gen textures
	gl.GenTextures(1, &e.outflowColorBuffer)
	gl.GenTextures(1, &e.velocityColorBuffer)
	gl.GenTextures(1, &e.heightColorBuffer)

	// Bind textures as colour attachments to the FBO
	// Create texture for height, waterHeight, sediment
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, e.heightColorBuffer)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, int32(width), int32(height), 0, gl.RGBA, gl.FLOAT, gl.Ptr(e.simulationState.heightData))
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.BindImageTexture(0, e.heightColorBuffer, 0, false, 0, gl.READ_WRITE, gl.RGBA32F)

	// Create texture for Water Outflow
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, e.outflowColorBuffer)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, int32(width), int32(height), 0, gl.RGBA, gl.FLOAT, nil)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.BindImageTexture(1, e.outflowColorBuffer, 0, false, 0, gl.READ_WRITE, gl.RGBA32F)

	// Create texture for velocity
	gl.ActiveTexture(gl.TEXTURE2)
	gl.BindTexture(gl.TEXTURE_2D, e.velocityColorBuffer)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, int32(width), int32(height), 0, gl.RGBA, gl.FLOAT, nil)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.BindImageTexture(2, e.velocityColorBuffer, 0, false, 0, gl.READ_WRITE, gl.RGBA32F)

	// Send the textures to a framebuffer so that we can reference them successfully.
	gl.GenFramebuffers(1, &e.frameBuffer)
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, e.frameBuffer)
	gl.FramebufferTexture2D(gl.READ_FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, e.heightColorBuffer, 0)
	//gl.FramebufferTexture2D(gl.READ_FRAMEBUFFER, gl.COLOR_ATTACHMENT1, gl.TEXTURE_2D, e.outflowColorBuffer, 1)
	//gl.FramebufferTexture2D(gl.READ_FRAMEBUFFER, gl.COLOR_ATTACHMENT2, gl.TEXTURE_2D, e.velocityColorBuffer, 2)

}

func (e *GPUEroder) Pass() {
	// Render a plane to the FBO
	width, height := e.heightmap.Dimensions()
	gl.UseProgram(e.waterPassProgram)
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
