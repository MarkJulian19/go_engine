package main

import (
	"engine/src/config"
	"engine/src/mainloop"
	"engine/src/player"
	"engine/src/render"
	"engine/src/windows"
	"engine/src/workers"
	"engine/src/world"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	// frameCount int
	// prevTime       = float64(0)

	Config = config.LoadConfig("config.json")
)

func init() {
	runtime.LockOSThread()
}

func main() {

	window := windows.InitWindow(Config)
	defer glfw.Terminate()

	gl.Enable(gl.DEPTH_TEST)

	// Инициализируем шейдеры
	renderProgram := render.InitShaders()
	depthProgram := render.InitDepthShader()
	textProgram := render.InitTextShader()
	crosshairProgram := render.InitCrosshairShader()

	// Создаём FBO и текстуру для карты теней
	render.CreateDepthMap(Config)

	// Создаём FBO и текстуру для отражений
	render.CreateReflectionFBO(Config)

	// Настраиваем мир и камеру
	worldObj := world.NewWorld(Config.ChunkX, Config.ChunkY, Config.ChunkZ)
	cameraObj := player.NewCamera(mgl32.Vec3{0, 120, 0})

	chunkGenCh := make(chan [2]int, 100)
	chunkDelCh := make(chan [2]int, 1000000)
	vramGCCh := make(chan [3]uint32, 1000000)
	workers.UpdateWorld(worldObj, cameraObj, chunkGenCh, chunkDelCh, Config)
	for i := 0; i < Config.NumWorkers; i++ {
		workers.ChunkCreatorWorker(worldObj, chunkGenCh, chunkDelCh, vramGCCh)
		workers.ChunkDeleterWorker(worldObj, chunkGenCh, chunkDelCh, vramGCCh)
	}
	workers.InitMouseHandler(window, cameraObj)
	workers.InitPhysicsHandler(cameraObj, worldObj)

	mainloop.RunMainLoop(window, renderProgram, depthProgram, textProgram, crosshairProgram, Config, worldObj, cameraObj, vramGCCh)
}
