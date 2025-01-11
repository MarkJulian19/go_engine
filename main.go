package main

import (
	"engine/src/camera"
	"engine/src/config"
	"engine/src/mainloop"
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

	// Создаём FBO и текстуру для карты теней
	render.CreateDepthMap()

	// Создаём FBO и текстуру для отражений
	render.CreateReflectionFBO()

	// Настраиваем мир и камеру
	worldObj := world.NewWorld(16, 128, 16)
	cameraObj := camera.NewCamera(mgl32.Vec3{0, 60, 0})

	chunkGenCh := make(chan [2]int, 100)
	chunkDelCh := make(chan [2]int, 1000000)
	vramGCCh := make(chan [3]uint32, 1000000)
	workers.UpdateWorld(worldObj, cameraObj, chunkGenCh, chunkDelCh, Config)
	for i := 0; i < Config.NumWorkers; i++ {
		workers.ChunkСreatorWorker(worldObj, chunkGenCh, chunkDelCh, vramGCCh)
		workers.ChunkDeleterWorker(worldObj, chunkGenCh, chunkDelCh, vramGCCh)
	}
	workers.InitMouseHandler(window, cameraObj)

	mainloop.RunMainLoop(window, renderProgram, depthProgram, textProgram, Config, worldObj, cameraObj, vramGCCh)
}
