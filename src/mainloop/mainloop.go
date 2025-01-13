package mainloop

import (
	"engine/src/config"
	"engine/src/garbageCollector"
	"engine/src/player"
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
	renderProgram, depthProgram, textProgram, crosshairProgram uint32,
	config *config.Config,
	worldObj *world.World,
	playerObj *player.Camera,
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
		playerObj.ProcessKeyboard(window, deltaTime, worldObj)
		playerObj.UpdatePhysics(deltaTime, worldObj)
		playerObj.InteractWithBlock(window, worldObj)
		// Обновляем мир (генерация / удаление чанков)

		// Освобождаем буферы из VRAM
		garbageCollector.VramGC(vramGCCh, &render.Cunt_ch)

		// Рендер теней, отражений и сцены
		dynamicLightPos := render.GetDynamicLightPos(playerObj.Position, timeOfDay)
		lightProjection := render.GetLightProjection(config)
		lightView := mgl32.LookAtV(dynamicLightPos, playerObj.Position, lightUp)
		lightSpaceMatrix := lightProjection.Mul4(lightView)
		render.RenderDepthMap(depthProgram, worldObj, lightSpaceMatrix, config)
		// render.RenderReflection(renderProgram, config, worldObj, playerObj, lightSpaceMatrix, dynamicLightPos)
		render.RenderScene(window, renderProgram, config, worldObj, playerObj, lightSpaceMatrix, dynamicLightPos, deltaTime, textProgram)
		if playerObj.ShowHUD {
			render.RenderCrosshair(window, crosshairProgram)
			if playerObj.ShowInfoPanel {
				render.RenderDebugHUD(window, textProgram, render.Get_hud_info(deltaTime, worldObj, playerObj))
			}
		}
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
