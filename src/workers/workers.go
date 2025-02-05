package workers

import (
	"engine/src/config"
	"engine/src/player"
	"engine/src/world"
	"fmt"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func ChunkCreatorWorker(w *world.World, genCh, delCh <-chan [2]int, vramCh chan [3]uint32, Config *config.Config) {
	go func() {
		for {

			coords := <-genCh
			// Если приходят координаты для генерации
			x, z := coords[0], coords[1]
			w.GenerateChunk(x, z, Config)

		}
	}()
}
func ChunkDeleterWorker(w *world.World, genCh, delCh <-chan [2]int, vramCh chan [3]uint32) {
	go func() {
		for {

			coords := <-delCh
			// Если приходят координаты для удаления
			x, z := coords[0], coords[1]
			w.RemoveChunk(x, z, vramCh)

		}
	}()
}
func UpdateWorld(
	worldObj *world.World,
	cameraObj *player.Camera,
	chunkGenCh, chunkDelCh chan [2]int,
	Config *config.Config,
) {
	go func() {
		ticker := time.NewTicker(time.Second / 12) // 12 раз в секунду
		defer ticker.Stop()

		for range ticker.C {
			playerPos := cameraObj.Position
			worldObj.UpdateChunks(int(playerPos.X()), int(playerPos.Z()), Config.ChunkDist, chunkGenCh, chunkDelCh)

			// Проверяем завершение работы программы или других условий
			if cameraObj == nil || worldObj == nil { // Условие для выхода из горутины
				return
			}
		}
	}()
}
func InitMouseHandler(window *glfw.Window, camera *player.Camera) {
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	go func() {
		for !window.ShouldClose() {
			xpos, ypos := window.GetCursorPos()
			camera.ProcessMouse(xpos, ypos)
			time.Sleep(10 * time.Millisecond)
		}
	}()
}
func MonitorMemoryStats(
	cameraObj *player.Camera,
	worldObj *world.World,
	deltaTime float64,
	memStatsCh chan<- []string,
) {
	go func() {
		ticker := time.NewTicker(time.Second / 12) // Обновление 12 раз в секунду
		defer ticker.Stop()

		for range ticker.C {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			// Замер видеопамяти
			var totalVRAM, availableVRAM int32
			gl.GetIntegerv(0x9048 /* GL_GPU_MEMORY_INFO_TOTAL_AVAILABLE_MEMORY_NVX */, &totalVRAM)
			gl.GetIntegerv(0x9049 /* GL_GPU_MEMORY_INFO_CURRENT_AVAILABLE_VIDMEM_NVX */, &availableVRAM)

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

			memStatsCh <- debugInfo // Отправляем данные в канал
		}
	}()
}
func InitPhysicsHandler(player *player.Camera, worldObj *world.World) {
	go func() {
		ticker := time.NewTicker(time.Second / 64) // Обновление 12 раз в секунду
		defer ticker.Stop()
		for range ticker.C {
			player.UpdatePhysics(float64(1.0/64.0), worldObj)
		}
	}()
}
