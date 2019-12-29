package erosion

import (
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/ob6160/Terrain/generators"
)
type GPUEroder struct {
	frameBuffer uint32
	outflowColorBuffer uint32 // o1, o2, o3, o4
	velocityColorBuffer uint32 // vX, vY
	heightColorBuffer uint32 // landHeight, waterHeight, sediment
	intermediateBuffer uint32 // Framebuffer
}

func NewGPUEroder(heightmap generators.TerrainGenerator) *GPUEroder {
	var width, height = heightmap.Dimensions()
	var e = new(GPUEroder)
	
	gl.GenFramebuffers(1, &e.frameBuffer)

	// Gen textures
	gl.GenTextures(1, &e.outflowColorBuffer)
	gl.GenTextures(1, &e.velocityColorBuffer)
	gl.GenTextures(1, &e.heightColorBuffer)
	gl.GenTextures(1, &e.intermediateBuffer)

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

	gl.BindTexture(gl.TEXTURE_2D, e.intermediateBuffer)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, int32(width), int32(height), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	
	// Bind textures as colour attachments to the FBO
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, e.heightColorBuffer, 0)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT1, gl.TEXTURE_2D, e.outflowColorBuffer, 0)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT2, gl.TEXTURE_2D, e.velocityColorBuffer, 0)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT3, gl.TEXTURE_2D, e.intermediateBuffer, 0)
	return e
}

func Pass() {
	// Render a plane to the FBO
}