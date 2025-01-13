package render

import (
	"engine/src/config"
	"engine/src/player"
	"engine/src/world"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Текстура и FBO для «зеркального» отражения
func CreateReflectionFBO(Config *config.Config) {
	gl.GenFramebuffers(1, &reflectionFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, reflectionFBO)

	// Текстура цвета
	gl.GenTextures(1, &reflectionTex)
	gl.BindTexture(gl.TEXTURE_2D, reflectionTex)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
		reflectWidth, reflectHeight, 0,
		gl.RGBA, gl.UNSIGNED_BYTE, nil)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// Прикрепляем цветовой буфер
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, reflectionTex, 0)

	// Текстура глубины или рендербуфер глубины — на выбор
	var rbo uint32
	gl.GenRenderbuffers(1, &rbo)
	gl.BindRenderbuffer(gl.RENDERBUFFER, rbo)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT24, reflectWidth, reflectHeight)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, rbo)

	// Проверка статуса
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("Reflection Framebuffer is not complete!")
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func RenderReflection(
	program uint32,
	config *config.Config,
	worldObj *world.World,
	cameraObj *player.Camera,
	lightSpaceMatrix mgl32.Mat4,
	lightPos mgl32.Vec3,
) {
	// 1) Сохраняем старое положение камеры
	originalPos := cameraObj.Position
	originalPitch := cameraObj.Pitch

	// 2) «Отражаем» камеру по оси Y относительно waterLevel
	distance := (cameraObj.Position.Y() - waterLevel) * 2.0
	cameraObj.Position = mgl32.Vec3{
		cameraObj.Position.X(),
		cameraObj.Position.Y() - distance,
		cameraObj.Position.Z(),
	}
	cameraObj.Pitch = -cameraObj.Pitch

	// 3) Рендер в reflectionFBO
	gl.Viewport(0, 0, reflectWidth, reflectHeight)
	gl.BindFramebuffer(gl.FRAMEBUFFER, reflectionFBO)
	gl.ClearColor(0.0, 0.5, 0.8, 1.0) // фоновый цвет для отражения (небо)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.UseProgram(program)

	// Матрицы вида/проекции с «отражённой» камерой
	view := cameraObj.GetViewMatrix()
	projection := mgl32.Perspective(mgl32.DegToRad(60),
		float32(reflectWidth)/float32(reflectHeight),
		0.1, 3000.0)

	setUniformMatrix4fv(program, "view", view)
	setUniformMatrix4fv(program, "projection", projection)
	setUniformMatrix4fv(program, "lightSpaceMatrix", lightSpaceMatrix)

	// Направление света
	lightDir := lightPos.Sub(cameraObj.Position).Normalize()
	lightDirLoc := gl.GetUniformLocation(program, gl.Str("lightDir\x00"))
	gl.Uniform3f(lightDirLoc, lightDir.X(), lightDir.Y(), lightDir.Z())

	// Остальные uniform’ы (тени, туман и т.д.)
	setupCommonUniforms(program, cameraObj.Position, config)

	// Привязываем shadowMap
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, depthMap)
	shadowMapLoc := gl.GetUniformLocation(program, gl.Str("shadowMap\x00"))
	gl.Uniform1i(shadowMapLoc, 1)

	// Отрисовываем чанки (как обычно)
	// frustumPlanes := calculateFrustumPlanes(view, projection)

	// worldObj.Mu.Lock()
	// for coord, chunk := range worldObj.Chunks {
	// 	// Если буферы не созданы / обновлены, создадим:
	// 	if chunk.VAO == 0 && len(chunk.Vertices) > 0 && len(chunk.Indices) > 0 {
	// 		safeDeleteBuffers(&chunk.VAO, &chunk.VBO, &chunk.EBO)
	// 		createChunkBuffers(chunk)
	// 		Cunt_ch++
	// 	} else if chunk.UpdateBuf {
	// 		updateChunkBuffers(chunk)
	// 	}

	// 	// Фрустум-отсечение
	// 	chunkBounds := chunk.GetBoundingBox(coord)
	// 	if !isChunkVisible(frustumPlanes, chunkBounds) {
	// 		continue
	// 	}

	// 	if chunk.VAO == 0 {
	// 		continue
	// 	}

	// 	model := mgl32.Translate3D(
	// 		float32(coord[0]*chunk.SizeX),
	// 		0,
	// 		float32(coord[1]*chunk.SizeZ),
	// 	)
	// 	setUniformMatrix4fv(program, "model", model)

	// 	gl.BindVertexArray(chunk.VAO)
	// 	gl.DrawElements(gl.TRIANGLES, int32(chunk.IndicesCount), gl.UNSIGNED_INT, gl.PtrOffset(0))
	// }
	// worldObj.Mu.Unlock()

	// 4) Возвращаем камеру на место
	cameraObj.Position = originalPos
	cameraObj.Pitch = originalPitch

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}
