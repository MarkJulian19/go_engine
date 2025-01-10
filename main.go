package main

import (
	"engine/src/camera"
	"engine/src/config"
	"engine/src/garbageCollector"
	"engine/src/render"
	"engine/src/windows"
	"engine/src/workers"
	"engine/src/world"
	"math"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	frameCount int
	prevTime       = float64(0)
	numWorkers int = 6
	timeOfDay      = float64(0)
)

func init() {
	// GLFW должен работать в главной (main) горутине
	runtime.LockOSThread()
}

func main() {
	config := config.LoadConfig("config.json")
	window := windows.InitWindow(config)
	defer glfw.Terminate()

	gl.Enable(gl.DEPTH_TEST)
	// (Опционально) Включаем сглаживание, если нужно:
	// gl.Enable(gl.MULTISAMPLE)

	// Инициализируем шейдеры
	renderProgram := render.InitShaders()    // Основной шейдер с тенями и отражениями
	depthProgram := render.InitDepthShader() // Шейдер для рендера карты глубины (теней)
	textProgram := render.InitTextShader()

	// Создаём FBO и текстуру для карты теней
	render.CreateDepthMap()

	// Создаём FBO и текстуру для отражений
	render.CreateReflectionFBO()

	// Настраиваем мир и камеру
	worldObj := world.NewWorld(16, 32, 16) // Пример размера чанка
	cameraObj := camera.NewCamera(mgl32.Vec3{0, 60, 0})

	initMouseHandler(window, cameraObj)

	runMainLoop(window, renderProgram, depthProgram, textProgram, config, worldObj, cameraObj)
}

func initMouseHandler(window *glfw.Window, camera *camera.Camera) {
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	go func() {
		for !window.ShouldClose() {
			xpos, ypos := window.GetCursorPos()
			camera.ProcessMouse(xpos, ypos)
			time.Sleep(10 * time.Millisecond)
		}
	}()
}

func runMainLoop(
	window *glfw.Window,
	renderProgram, depthProgram, textProgram uint32,
	config *config.Config,
	worldObj *world.World,
	cameraObj *camera.Camera,
) {
	lastFrame := time.Now()

	// Каналы для асинхронной генерации/удаления чанков
	chunkGenCh := make(chan [2]int, 100)
	chunkDelCh := make(chan [2]int, 1000000)
	vramGCCh := make(chan [3]uint32, 1000000)
	for i := 0; i < numWorkers; i++ {
		workers.ChunkСreatorWorker(worldObj, chunkGenCh, chunkDelCh, vramGCCh)
		workers.ChunkDeleterWorker(worldObj, chunkGenCh, chunkDelCh, vramGCCh)
	}

	// Позиция "солнца" и направление
	// lightPos := mgl32.Vec3{100, 200, 100}
	// lightTarget := mgl32.Vec3{0, 0, 0}
	lightUp := mgl32.Vec3{0, 1, 0}

	// Ортопроекция для теней (чтобы «охватить» сцену)
	// lightProjection := mgl32.Ortho(-200, 200, -200, 200, 0.1, 500.0)
	for !window.ShouldClose() {

		currentFrame := time.Now()
		deltaTime := currentFrame.Sub(lastFrame).Seconds()
		lastFrame = currentFrame
		timeOfDay += deltaTime * 0.1 // Скорость вращения (0.1 для медленных суток)
		if timeOfDay > 2*math.Pi {   // Цикл повторяется каждые 360° (2π радиан)
			timeOfDay -= 2 * math.Pi
		}
		// Обработка клавиш + физики
		go cameraObj.ProcessKeyboard(window, deltaTime, worldObj)
		cameraObj.UpdatePhysics(deltaTime, worldObj)

		// Обновляем мир (генерация / удаление чанков)
		updateWorld(worldObj, cameraObj, chunkGenCh, chunkDelCh)

		// Освобождаем буферы из VRAM, если надо
		//print(render.Cunt_ch)
		garbageCollector.VramGC(vramGCCh, &render.Cunt_ch)

		if cameraObj.ShowInfoPanel {
			print("HERE")
			render.RenderText("AAAAAAAAAAAAAA", 0, 0, 1920, 1080, textProgram, [4]float32{1, 1, 1, 1})
		}

		dynamicLightPos := render.GetDynamicLightPos(cameraObj.Position, timeOfDay)

		// 2) Ортопроекция вокруг игрока
		lightProjection := render.GetLightProjection()

		// 3) Направляем луч из динамического света к игроку
		lightView := mgl32.LookAtV(dynamicLightPos, cameraObj.Position, lightUp)

		// 4) Итоговая матрица
		lightSpaceMatrix := lightProjection.Mul4(lightView)

		// 5) Дальше рендерим depth map, reflection и т.д.:
		render.RenderDepthMap(depthProgram, worldObj, lightSpaceMatrix)
		render.RenderReflection(renderProgram, config, worldObj, cameraObj, lightSpaceMatrix, dynamicLightPos)
		render.RenderScene(window, renderProgram, config, worldObj, cameraObj, lightSpaceMatrix, dynamicLightPos)

		frameCount++
		windows.ChangeWindowTitle(&frameCount, &prevTime, window, render.Cunt_ch)
	}
}

func updateWorld(
	worldObj *world.World,
	cameraObj *camera.Camera,
	chunkGenCh, chunkDelCh chan [2]int,
) {
	playerPos := cameraObj.Position
	// Радиус 32 для прогрузки
	worldObj.UpdateChunks(int(playerPos.X()), int(playerPos.Z()), 32, chunkGenCh, chunkDelCh)
}
