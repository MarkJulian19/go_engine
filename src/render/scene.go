package render

import (
	"engine/src/camera"
	"engine/src/config"
	"engine/src/world"
	"fmt"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	depthMapFBO  uint32
	depthMap     uint32
	shadowWidth  = int32(16384) // Очень высокое разрешение (16k!) — возможно, слишком большое.
	shadowHeight = int32(16384)
)

// Параметры для отражения (planar reflection)
var (
	reflectionFBO uint32
	reflectionTex uint32
	reflectWidth  = int32(1024)
	reflectHeight = int32(1024)
	waterLevel    = float32(0.0) // Плоскость, относительно которой зеркалим камеру
)

func RenderScene(
	window *glfw.Window,
	program uint32,
	config *config.Config,
	worldObj *world.World,
	cameraObj *camera.Camera,
	lightSpaceMatrix mgl32.Mat4,
	lightPos mgl32.Vec3,
	deltaTime float64,
	textProgram uint32,
) {
	// Настраиваем вьюпорт под размер окна
	gl.Viewport(0, 0, int32(config.Width), int32(config.Height))
	gl.ClearColor(0.7, 0.8, 1.0, 1.0) // небо
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.UseProgram(program)

	// Матрицы вида и проекции (с нормальной, не-зеркальной камерой)
	view := cameraObj.GetViewMatrix()
	projection := mgl32.Perspective(mgl32.DegToRad(60),
		float32(config.Width)/float32(config.Height),
		0.1, 3000.0)

	setUniformMatrix4fv(program, "view", view)
	setUniformMatrix4fv(program, "projection", projection)
	setUniformMatrix4fv(program, "lightSpaceMatrix", lightSpaceMatrix)

	// Направление света — из lightPos в сцену
	lightDir := lightPos.Sub(cameraObj.Position).Normalize()
	lightDirLoc := gl.GetUniformLocation(program, gl.Str("lightDir\x00"))
	gl.Uniform3f(lightDirLoc, lightDir.X(), lightDir.Y(), lightDir.Z())

	// Общие uniform’ы (lightColor, ambient, fog, и т.д.)
	setupCommonUniforms(program, cameraObj.Position)

	// Привязываем shadowMap
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, depthMap)
	shadowMapLoc := gl.GetUniformLocation(program, gl.Str("shadowMap\x00"))
	gl.Uniform1i(shadowMapLoc, 1)

	// Привязываем reflectionTex (на TEXTURE2)
	gl.ActiveTexture(gl.TEXTURE2)
	gl.BindTexture(gl.TEXTURE_2D, reflectionTex)
	reflectionMapLoc := gl.GetUniformLocation(program, gl.Str("reflectionMap\x00"))
	gl.Uniform1i(reflectionMapLoc, 2)

	frustumPlanes := calculateFrustumPlanes(view, projection)

	// Рендерим чанки
	worldObj.Mu.Lock()
	for coord, chunk := range worldObj.Chunks {
		// Точно так же: если буфер не создан, создаём
		if chunk.VAO == 0 && len(chunk.Vertices) > 0 && len(chunk.Indices) > 0 {
			safeDeleteBuffers(&chunk.VAO, &chunk.VBO, &chunk.EBO)
			createChunkBuffers(chunk)
			Cunt_ch++
		} else if chunk.UpdateBuf {
			updateChunkBuffers(chunk)
		}

		if !isChunkVisible(frustumPlanes, chunk.GetBoundingBox(coord)) {
			continue
		}

		if chunk.VAO == 0 {
			continue
		}

		model := mgl32.Translate3D(
			float32(coord[0]*chunk.SizeX),
			0,
			float32(coord[1]*chunk.SizeZ),
		)
		setUniformMatrix4fv(program, "model", model)

		gl.BindVertexArray(chunk.VAO)
		gl.DrawElements(gl.TRIANGLES, int32(chunk.IndicesCount), gl.UNSIGNED_INT, gl.PtrOffset(0))
	}
	worldObj.Mu.Unlock()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Замер видеопамяти
	var totalVRAM, availableVRAM int32
	gl.GetIntegerv(0x9048 /* GL_GPU_MEMORY_INFO_TOTAL_AVAILABLE_MEMORY_NVX */, &totalVRAM)       // Общее количество VRAM
	gl.GetIntegerv(0x9049 /* GL_GPU_MEMORY_INFO_CURRENT_AVAILABLE_VIDMEM_NVX */, &availableVRAM) // Доступная VRAM
	var usedVRAM int32
	if totalVRAM > 0 {
		usedVRAM = totalVRAM - availableVRAM
	}
	debugInfo := []string{
		fmt.Sprintf("FPS: %.2f", 1.0/deltaTime),
		fmt.Sprintf("Camera Position: X=%.2f Y=%.2f Z=%.2f", cameraObj.Position.X(), cameraObj.Position.Y(), cameraObj.Position.Z()),
		fmt.Sprintf("Chunks Loaded: %d", len(worldObj.Chunks)),
		fmt.Sprintf("Allocated RAM: %.2f MB", float64(memStats.Alloc)/1024/1024),
		fmt.Sprintf("Total Allocated RAM: %.2f MB", float64(memStats.TotalAlloc)/1024/1024),
		fmt.Sprintf("System RAM: %.2f MB", float64(memStats.Sys)/1024/1024),
		fmt.Sprintf("Total VRAM: %.2f MB", float64(totalVRAM)/1024),
		fmt.Sprintf("Used VRAM: %.2f MB", float64(usedVRAM)/1024),
	}

	// Вызов HUD — используем textProgram
	RenderDebugHUD(window, textProgram, debugInfo)
	window.SwapBuffers()
	glfw.PollEvents()
}
