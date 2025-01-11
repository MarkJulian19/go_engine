package mainloop

import (
	"engine/src/camera"
	"engine/src/config"
	"engine/src/garbageCollector"
	"engine/src/render"
	"engine/src/world"
	"math"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	timeOfDay = float64(0)
)

func RunMainLoop(
	window *glfw.Window,
	renderProgram, depthProgram, textProgram uint32,
	config *config.Config,
	worldObj *world.World,
	cameraObj *camera.Camera,
	vramGCCh chan [3]uint32,
) {
	lastFrame := time.Now()

	// Каналы для асинхронной генерации/удаления чанков

	lightUp := mgl32.Vec3{0, 1, 0}

	for !window.ShouldClose() {
		currentFrame := time.Now()
		deltaTime := currentFrame.Sub(lastFrame).Seconds()
		lastFrame = currentFrame
		timeOfDay += deltaTime * 0.1
		if timeOfDay > 2*math.Pi {
			timeOfDay -= 2 * math.Pi
		}

		// Обработка клавиш и физики
		cameraObj.ProcessKeyboard(window, deltaTime, worldObj)
		cameraObj.UpdatePhysics(deltaTime, worldObj)

		// Обновляем мир (генерация / удаление чанков)

		// Освобождаем буферы из VRAM
		garbageCollector.VramGC(vramGCCh, &render.Cunt_ch)

		// Рендер теней, отражений и сцены
		dynamicLightPos := render.GetDynamicLightPos(cameraObj.Position, timeOfDay)
		lightProjection := render.GetLightProjection()
		lightView := mgl32.LookAtV(dynamicLightPos, cameraObj.Position, lightUp)
		lightSpaceMatrix := lightProjection.Mul4(lightView)

		render.RenderDepthMap(depthProgram, worldObj, lightSpaceMatrix)
		render.RenderReflection(renderProgram, config, worldObj, cameraObj, lightSpaceMatrix, dynamicLightPos)
		render.RenderScene(window, renderProgram, config, worldObj, cameraObj, lightSpaceMatrix, dynamicLightPos, deltaTime, textProgram)
	}
}
