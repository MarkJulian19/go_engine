package src

import "engine/world"

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
