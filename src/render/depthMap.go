package render

import (
	"engine/src/config"
	"engine/src/world"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

func CreateDepthMap(Config *config.Config) {
	gl.GenFramebuffers(1, &depthMapFBO)

	gl.GenTextures(1, &depthMap)
	gl.BindTexture(gl.TEXTURE_2D, depthMap)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT,
		Config.ShadowWidth, Config.ShadowHeight, 0,
		gl.DEPTH_COMPONENT, gl.FLOAT, nil)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR) // Было GL_NEAREST
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR) // Было GL_NEAREST
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)

	borderColor := []float32{1.0, 1.0, 1.0, 1.0}
	gl.TexParameterfv(gl.TEXTURE_2D, gl.TEXTURE_BORDER_COLOR, &borderColor[0])

	gl.BindFramebuffer(gl.FRAMEBUFFER, depthMapFBO)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, depthMap, 0)
	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func RenderDepthMap(depthProgram uint32, worldObj *world.World, lightSpaceMatrix mgl32.Mat4, Config *config.Config) {
	gl.Viewport(0, 0, Config.ShadowWidth, Config.ShadowHeight)
	gl.BindFramebuffer(gl.FRAMEBUFFER, depthMapFBO)
	gl.Clear(gl.DEPTH_BUFFER_BIT)

	gl.UseProgram(depthProgram)
	setUniformMatrix4fv(depthProgram, "lightSpaceMatrix", lightSpaceMatrix)

	worldObj.Mu.Lock()
	for coord, chunk := range worldObj.Chunks {
		// Создаём буферы, если надо
		if chunk.VAO == 0 && len(chunk.Vertices) > 0 && len(chunk.Indices) > 0 {
			safeDeleteBuffers(&chunk.VAO, &chunk.VBO, &chunk.EBO)
			createChunkBuffers(chunk)
			Cunt_ch++
		} else if chunk.UpdateBuf {
			updateChunkBuffers(chunk)
		}

		if chunk.VAO == 0 {
			continue
		}

		// Здесь НЕ делаем isChunkVisible(...) по КАМЕРНОМУ фрустуму!
		// при желании можно сделать culling со стороны света, но НЕ от камеры
		model := mgl32.Translate3D(
			float32(coord[0]*chunk.SizeX),
			0,
			float32(coord[1]*chunk.SizeZ),
		)
		setUniformMatrix4fv(depthProgram, "model", model)

		gl.BindVertexArray(chunk.VAO)
		gl.DrawElements(gl.TRIANGLES, int32(chunk.IndicesCount), gl.UNSIGNED_INT, nil)
	}
	worldObj.Mu.Unlock()

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}
