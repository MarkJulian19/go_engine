package workers

import (
	"engine/src/camera"
	"engine/src/config"
	"engine/src/world"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
)

func ChunkСreatorWorker(w *world.World, genCh, delCh <-chan [2]int, vramCh chan [3]uint32) {
	go func() {
		for {

			coords := <-genCh
			// Если приходят координаты для генерации
			x, z := coords[0], coords[1]
			w.GenerateChunk(x, z)

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
	cameraObj *camera.Camera,
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
func InitMouseHandler(window *glfw.Window, camera *camera.Camera) {
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	go func() {
		for !window.ShouldClose() {
			xpos, ypos := window.GetCursorPos()
			camera.ProcessMouse(xpos, ypos)
			time.Sleep(10 * time.Millisecond)
		}
	}()
}
