package world

import (
	"math/rand"
	"sync"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/ojrac/opensimplex-go"
)

// Тип блока
type Block struct {
	Id    uint8
	Color [3]float32
}

// Структура чанка
type Chunk struct {
	Blocks              []Block // Одномерный массив блоков
	VAO                 uint32
	VBO                 uint32
	EBO                 uint32
	IndicesCount        int
	SizeX, SizeY, SizeZ int
	Vertices            []float32
	Indices             []uint32
	UpdateBuf           bool
	CreateBuf           bool
}

// Структура мира
type World struct {
	Mu                  sync.RWMutex
	Chunks              map[[2]int]*Chunk
	SizeX, SizeY, SizeZ int
}

// Создает новый пустой мир
func NewWorld(sizeX, sizeY, sizeZ int) *World {
	return &World{
		Chunks: make(map[[2]int]*Chunk),
		SizeX:  sizeX,
		SizeY:  sizeY,
		SizeZ:  sizeZ,
	}
}

// Функция для вычисления индекса в одномерном массиве по (x, y, z)
func blockIndex(x, y, z, sizeX, sizeY, sizeZ int) int {
	return x + y*sizeX + z*sizeX*sizeY
}

// Создает новый чанк
// func NewChunk(sizeX, sizeY, sizeZ int) *Chunk {
// 	offsetX := 0
// 	offsetZ := 0
// 	seed := int64(1000)
// 	noiseScale := 500.0
// 	blocks := make([]Block, sizeX*sizeY*sizeZ)
// 	var wg sync.WaitGroup

// 	// Инициализируем генератор шума
// 	noise := opensimplex.New(seed)

// 	for x := 0; x < sizeX; x++ {
// 		wg.Add(1) // Увеличиваем счётчик для новой горутины
// 		go func(x int) {
// 			defer wg.Done() // Уменьшаем счётчик после завершения работы горутины
// 			for z := 0; z < sizeZ; z++ {
// 				// Генерируем высоту блока с использованием многослойного шума
// 				absoluteX := x + offsetX*sizeX
// 				absoluteZ := z + offsetZ*sizeZ

// 				baseHeight := noise.Eval2(float64(absoluteX)/noiseScale, float64(absoluteZ)/noiseScale)
// 				detailHeight := noise.Eval2(float64(absoluteX)/(noiseScale/2), float64(absoluteZ)/(noiseScale/2)) * 0.25
// 				height := int((baseHeight + detailHeight + 1) * 0.5 * float64(sizeY-1)) // Высота от 0 до sizeY-1

// 				for y := 0; y < sizeY; y++ {
// 					idx := blockIndex(x, y, z, sizeX, sizeY, sizeZ)

// 					if y <= height {
// 						switch {
// 						case y == height: // Верхний слой — трава
// 							blocks[idx] = Block{
// 								Id:    2, // Трава
// 								Color: [3]float32{0.1 + 0.3*rand.Float32(), 0.8 + 0.2*rand.Float32(), 0.1 + 0.3*rand.Float32()},
// 							}
// 						case y > height-4: // Слои земли
// 							blocks[idx] = Block{
// 								Id:    1, // Земля
// 								Color: [3]float32{0.5, 0.3, 0.1},
// 							}
// 						default: // Глубокие слои — камень
// 							blocks[idx] = Block{
// 								Id:    3, // Камень
// 								Color: [3]float32{0.4, 0.4, 0.4},
// 							}
// 						}
// 					} else {
// 						// Воздух
// 						blocks[idx] = Block{
// 							Id:    0,
// 							Color: [3]float32{0.5, 0.8, 1.0}, // Цвет для воздуха (не используется в рендере)
// 						}
// 					}
// 				}
// 			}
// 		}(x) // Передаём `x` как параметр, чтобы избежать замыкания
// 	}

// 	wg.Wait() // Ожидаем завершения всех горутин

// 	return &Chunk{
// 		Blocks: blocks,
// 		SizeX:  sizeX,
// 		SizeY:  sizeY,
// 		SizeZ:  sizeZ,
// 	}
// }

// Генерирует буферы для чанка
func (chunk *Chunk) CreateBuffers(neighbors map[string]*Chunk) {
	Vertices, indices := chunk.GenerateMesh(neighbors)

	chunk.IndicesCount = len(indices)
	chunk.Vertices = Vertices
	chunk.Indices = indices
	chunk.CreateBuf = true

	// Удаляем старые буферы, если они существуют
	// if chunk.VAO != 0 {
	// 	gl.DeleteVertexArrays(1, &chunk.VAO)
	// }
	// if chunk.VBO != 0 {
	// 	gl.DeleteBuffers(1, &chunk.VBO)
	// }
	// if chunk.EBO != 0 {
	// 	gl.DeleteBuffers(1, &chunk.EBO)
	// }

	// var vao, vbo, ebo uint32
	// gl.GenVertexArrays(1, &vao)
	// gl.BindVertexArray(vao)

	// gl.GenBuffers(1, &vbo)
	// gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	// gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// gl.GenBuffers(1, &ebo)
	// gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	// gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	// gl.EnableVertexAttribArray(0)
	// gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	// gl.EnableVertexAttribArray(1)

	// chunk.VAO, chunk.VBO, chunk.EBO = vao, vbo, ebo
}

// Обновляет буферы и меш чанка
func (chunk *Chunk) UpdateBuffers(neighbors map[string]*Chunk) {
	Vertices, indices := chunk.GenerateMesh(neighbors)

	chunk.IndicesCount = len(indices)
	chunk.Vertices = Vertices
	chunk.Indices = indices
	chunk.UpdateBuf = true
	// Удаляем старые буферы, если они существуют
	// if chunk.VAO != 0 {
	// 	gl.DeleteVertexArrays(1, &chunk.VAO)
	// }
	// if chunk.VBO != 0 {
	// 	gl.DeleteBuffers(1, &chunk.VBO)
	// }
	// if chunk.EBO != 0 {
	// 	gl.DeleteBuffers(1, &chunk.EBO)
	// }

	// var vao, vbo, ebo uint32
	// gl.GenVertexArrays(1, &vao)
	// gl.BindVertexArray(vao)

	// gl.GenBuffers(1, &vbo)
	// gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	// gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// gl.GenBuffers(1, &ebo)
	// gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	// gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	// gl.EnableVertexAttribArray(0)
	// gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	// gl.EnableVertexAttribArray(1)

	// chunk.VAO, chunk.VBO, chunk.EBO = vao, vbo, ebo
}

// Генерирует меш чанка
func (chunk *Chunk) GenerateMesh(neighbors map[string]*Chunk) ([]float32, []uint32) {
	var vertices []float32
	var indices []uint32

	// var mutex sync.Mutex  // Для защиты общей памяти
	// var wg sync.WaitGroup // Для ожидания завершения всех горутин

	// Горутинно обрабатываем данные для каждой оси X
	for x := 0; x < chunk.SizeX; x++ {
		// wg.Add(1) // Увеличиваем счётчик горутин
		// go func(x int) {
		// defer wg.Done() // Уменьшаем счётчик после завершения работы

		var localVertices []float32
		var localIndices []uint32

		for y := 0; y < chunk.SizeY; y++ {
			for z := 0; z < chunk.SizeZ; z++ {
				// Получаем текущий блок
				block := chunk.Blocks[blockIndex(x, y, z, chunk.SizeX, chunk.SizeY, chunk.SizeZ)]
				if block.Id == 0 {
					continue // Пропускаем блоки воздуха
				}

				// Генерация граней для блока
				for _, face := range cubeFaces {
					nx, ny, nz := x+face.OffsetX, y+face.OffsetY, z+face.OffsetZ
					if IsAirWithNeighbors(chunk, nx, ny, nz, neighbors) {
						startIdx := len(localVertices) / 6
						for _, vertex := range face.Vertices {
							localVertices = append(localVertices,
								float32(x)+vertex[0],
								float32(y)+vertex[1],
								float32(z)+vertex[2],
								block.Color[0], block.Color[1], block.Color[2])
						}
						localIndices = append(localIndices,
							uint32(startIdx), uint32(startIdx+1), uint32(startIdx+2),
							uint32(startIdx+2), uint32(startIdx+3), uint32(startIdx))
					}
				}
			}
		}

		// Добавляем локальные данные в общие массивы с блокировкой
		// mutex.Lock()
		offset := len(vertices) / 6
		vertices = append(vertices, localVertices...)
		for _, idx := range localIndices {
			indices = append(indices, idx+uint32(offset))
		}
		// mutex.Unlock()
		// }(x)
	}

	// Ожидаем завершения всех горутин
	// wg.Wait()

	return vertices, indices
}

func NewChunk(sizeX, sizeY, sizeZ int, offsetX, offsetZ int, noise opensimplex.Noise) *Chunk {
	blocks := make([]Block, sizeX*sizeY*sizeZ)
	noiseScale := 100.0 // Масштаб шума
	biomeScale := 100.0 // Масштаб для биомов

	// var wg sync.WaitGroup
	// numWorkers := runtime.NumCPU() // Количество воркеров = количество ядер
	// workQueue := make(chan int, sizeX)

	// // Worker function
	// worker := func() {
	// 	defer wg.Done()
	// 	for x := range workQueue {
	for x := 0; x < sizeX; x++ {

		for z := 0; z < sizeZ; z++ {
			// Глобальные координаты
			absoluteX := x + offsetX*sizeX
			absoluteZ := z + offsetZ*sizeZ

			// Генерация высоты с использованием шума
			baseHeight := noise.Eval2(float64(absoluteX)/noiseScale, float64(absoluteZ)/noiseScale)
			detailHeight := noise.Eval2(float64(absoluteX)/(noiseScale/2), float64(absoluteZ)/(noiseScale/2)) * 0.25
			height := int((baseHeight + detailHeight + 1) * 0.5 * float64(sizeY-1))

			// Генерация типа биома (пустыня, лес, горы)
			biomeValue := noise.Eval2(float64(absoluteX)/biomeScale, float64(absoluteZ)/biomeScale)

			for y := 0; y < sizeY; y++ {
				idx := blockIndex(x, y, z, sizeX, sizeY, sizeZ)

				if y <= height {
					switch {
					case absoluteX%5 != 0:
						blocks[idx] = Block{
							Id:    0,
							Color: [3]float32{0.1 + 0.3*rand.Float32(), 0.8 + 0.2*rand.Float32(), 0.1 + 0.3*rand.Float32()},
						}
					case biomeValue > 0.3 && y == height: // Лес, верхний слой — трава
						blocks[idx] = Block{
							Id:    2,
							Color: [3]float32{0.1 + 0.3*rand.Float32(), 0.8 + 0.2*rand.Float32(), 0.1 + 0.3*rand.Float32()},
						}
					case biomeValue <= 0.3 && y == height: // Пустыня, верхний слой — песок
						blocks[idx] = Block{
							Id:    4,
							Color: [3]float32{0.9 * float32(y) / float32(sizeY), 0.8 * float32(y) / float32(sizeY), 0.5 * float32(y) / float32(sizeY)},
						}
					case y > height-4: // Слои земли
						blocks[idx] = Block{
							Id:    1,
							Color: [3]float32{0.5, 0.3, 0.1},
						}
					default: // Глубокие слои — камень
						blocks[idx] = Block{
							Id:    3,
							Color: [3]float32{0.4, 0.4, 0.4},
						}
					}
				} else {
					// Воздух
					blocks[idx] = Block{
						Id:    0,
						Color: [3]float32{0.5, 0.8, 1.0},
					}
				}
			}
		}
	}
	// 	}
	// }

	// Запуск воркеров
	// for i := 0; i < numWorkers; i++ {
	// 	wg.Add(1)
	// 	go worker()
	// }

	// // Добавление задач в очередь
	// for x := 0; x < sizeX; x++ {
	// 	workQueue <- x
	// }
	// close(workQueue)

	// // Ожидание завершения всех горутин
	// wg.Wait()

	return &Chunk{
		Blocks: blocks,
		SizeX:  sizeX,
		SizeY:  sizeY,
		SizeZ:  sizeZ,
	}
}

// Модифицируем GenerateChunk для передачи глобальных координат
func (w *World) GenerateChunk(cx, cz int) {

	coord := [2]int{cx, cz}
	w.Mu.Lock()
	if _, exists := w.Chunks[coord]; exists {
		w.Mu.Unlock()
		return
	}
	w.Mu.Unlock()
	noise := opensimplex.New(2000)
	newChunk := NewChunk(w.SizeX, w.SizeY, w.SizeZ, cx, cz, noise)

	// defer
	w.Mu.Lock()
	w.Chunks[coord] = newChunk

	neighbors := w.collectNeighbors(cx, cz)
	w.Mu.Unlock()
	newChunk.CreateBuffers(neighbors)

	// Обновляем соседей
	for direction, neighbor := range neighbors {
		if neighbor != nil {
			w.Mu.Lock()
			updatedNeighbors := w.collectNeighbors(cx+offsets[direction][0], cz+offsets[direction][1])
			w.Mu.Unlock()
			neighbor.UpdateBuffers(updatedNeighbors)
		}
	}

}

// func NewChunk(sizeX, sizeY, sizeZ int) *Chunk {
// 	blocks := make([]Block, sizeX*sizeY*sizeZ)

// 	for x := 0; x < sizeX; x++ {
// 		for y := 0; y < sizeY; y++ {
// 			for z := 0; z < sizeZ; z++ {
// 				idx := blockIndex(x, y, z, sizeX, sizeY, sizeZ)
// 				if y < 8 || z > 5 && y < 10 {
// 					blocks[idx].Id = 1 // Земля
// 				} else {
// 					blocks[idx].Id = 0 // Воздух
// 				}
// 				blocks[idx].Color = [3]float32{
// 					rand.Float32(),
// 					rand.Float32(),
// 					rand.Float32(),
// 				}
// 			}
// 		}
// 	}

// 	return &Chunk{
// 		Blocks: blocks,
// 		SizeX:  sizeX,
// 		SizeY:  sizeY,
// 		SizeZ:  sizeZ,
// 	}
// }

// func (w *World) GenerateChunk(cx, cz int) {
// 	coord := [2]int{cx, cz}
// 	if _, exists := w.Chunks[coord]; exists {
// 		return // Чанк уже существует
// 	}

// 	chunk := w.NewChunk(w.SizeX, w.SizeY, w.SizeZ)

// 	// Собираем соседние чанки
// 	neighbors := map[string]*Chunk{
// 		"left":  w.Chunks[[2]int{cx - 1, cz}],
// 		"right": w.Chunks[[2]int{cx + 1, cz}],
// 		"back":  w.Chunks[[2]int{cx, cz - 1}],
// 		"front": w.Chunks[[2]int{cx, cz + 1}],
// 	}

//		chunk.GenerateBuffers(neighbors)
//		w.Chunks[coord] = chunk
//		for direction, neighbor := range neighbors {
//			if neighbor != nil {
//				updatedNeighbors := w.collectNeighbors(cx+offsets[direction][0], cz+offsets[direction][1])
//				neighbor.UpdateBuffers(updatedNeighbors)
//			}
//		}
//	}
func (chunk *Chunk) GenerateBuffers(neighbors map[string]*Chunk) {
	vertices, indices := chunk.GenerateMesh(neighbors)
	chunk.IndicesCount = len(indices)

	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// Настройка атрибутов вершин
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0)) // Координаты
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4)) // Цвет
	gl.EnableVertexAttribArray(1)

	chunk.VAO = vao
	chunk.VBO = vbo
	chunk.EBO = ebo
}

// Собирает соседние чанки
func (w *World) collectNeighbors(cx, cz int) map[string]*Chunk {
	return map[string]*Chunk{
		"left":  w.Chunks[[2]int{cx - 1, cz}],
		"right": w.Chunks[[2]int{cx + 1, cz}],
		"back":  w.Chunks[[2]int{cx, cz - 1}],
		"front": w.Chunks[[2]int{cx, cz + 1}],
	}
}

// Смещения для соседей
var offsets = map[string][2]int{
	"left":  {-1, 0},
	"right": {1, 0},
	"back":  {0, -1},
	"front": {0, 1},
}

// Описание граней куба
var cubeFaces = []struct {
	OffsetX, OffsetY, OffsetZ int
	Vertices                  [4][3]float32
}{
	{0, 0, 1, [4][3]float32{{0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1}}},
	{0, 0, -1, [4][3]float32{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0}}},
	{-1, 0, 0, [4][3]float32{{0, 0, 0}, {0, 0, 1}, {0, 1, 1}, {0, 1, 0}}},
	{1, 0, 0, [4][3]float32{{1, 0, 0}, {1, 0, 1}, {1, 1, 1}, {1, 1, 0}}},
	{0, 1, 0, [4][3]float32{{0, 1, 0}, {1, 1, 0}, {1, 1, 1}, {0, 1, 1}}},
	{0, -1, 0, [4][3]float32{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}}},
}

// Проверяет, является ли блок воздухом с учетом соседей
func IsAirWithNeighbors(chunk *Chunk, x, y, z int, neighbors map[string]*Chunk) bool {
	if x >= 0 && x < chunk.SizeX && y >= 0 && y < chunk.SizeY && z >= 0 && z < chunk.SizeZ {
		return chunk.Blocks[blockIndex(x, y, z, chunk.SizeX, chunk.SizeY, chunk.SizeZ)].Id == 0
	}

	// Проверяем соседние чанки
	switch {
	case x < 0:
		neighbor := neighbors["left"]
		if neighbor != nil {
			return neighbor.Blocks[blockIndex(neighbor.SizeX-1, y, z, neighbor.SizeX, neighbor.SizeY, neighbor.SizeZ)].Id == 0
		} else {
			return true
		}
	case x >= chunk.SizeX:
		neighbor := neighbors["right"]
		if neighbor != nil {
			return neighbor.Blocks[blockIndex(0, y, z, neighbor.SizeX, neighbor.SizeY, neighbor.SizeZ)].Id == 0
		} else {
			return true
		}
	case z < 0:
		neighbor := neighbors["back"]
		if neighbor != nil {
			return neighbor.Blocks[blockIndex(x, y, neighbor.SizeZ-1, neighbor.SizeX, neighbor.SizeY, neighbor.SizeZ)].Id == 0
		} else {
			return true
		}
	case z >= chunk.SizeZ:
		neighbor := neighbors["front"]
		if neighbor != nil && x >= 0 && x < neighbor.SizeX && y >= 0 && y < neighbor.SizeY {
			return neighbor.Blocks[blockIndex(x, y, 0, neighbor.SizeX, neighbor.SizeY, neighbor.SizeZ)].Id == 0
		} else {
			return true
		}
	}
	return true // Возвращаем `true`, если сосед отсутствует (вместо false)
}
func (w *World) UpdateChunks(playerX, playerZ int, radius int, chunkGenCh chan [2]int, chunkDelCh chan [2]int) {
	// Вычисляем центральные координаты чанка, в котором находится игрок
	centerX, centerZ := playerX/w.SizeX, playerZ/w.SizeZ

	// Множество чанков, которые должны быть загружены
	newChunks := make(map[[2]int]bool)

	// Вычисляем координаты чанков в пределах радиуса
	for x := centerX - radius; x <= centerX+radius; x++ {
		for z := centerZ - radius; z <= centerZ+radius; z++ {
			coord := [2]int{x, z}
			newChunks[coord] = true

			// Генерируем новый чанк, если он еще не существует
			w.Mu.Lock()
			if _, exists := w.Chunks[coord]; !exists {
				// w.GenerateChunk(x, z)
				if len(chunkGenCh) < 10000 {
					chunkGenCh <- [2]int{x, z}
				}
				// fmt.Printf("%d, %d\n", x, z)
			}
			w.Mu.Unlock()
		}
	}

	// Удаляем чанки, которые больше не находятся в радиусе
	w.Mu.Lock()
	for coord := range w.Chunks {
		if !newChunks[coord] {
			// w.RemoveChunk(coord[0], coord[1])
			chunkDelCh <- [2]int{coord[0], coord[1]}
		}
	}
	w.Mu.Unlock()
}

// Удаляет чанк и освобождает связанные ресурсы (VAO, VBO, EBO)
func (w *World) RemoveChunk(cx, cz int) {
	coord := [2]int{cx, cz}
	w.Mu.Lock()
	defer w.Mu.Unlock()
	// chunk, exists := w.Chunks[coord]
	// if !exists {
	// 	return // Чанк уже удален или не существует
	// }

	// // Удаляем буферы OpenGL
	// gl.DeleteVertexArrays(1, &chunk.VAO)
	// gl.DeleteBuffers(1, &chunk.VBO)
	// gl.DeleteBuffers(1, &chunk.EBO)

	// Удаляем чанк из карты
	delete(w.Chunks, coord)
}
