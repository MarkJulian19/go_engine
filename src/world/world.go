package world

import (
	"fmt"
	"math"
	"math/rand"
	"sync"

	"github.com/go-gl/mathgl/mgl32"
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
	var vertices []float32 // здесь будем класть по 9 float на вершину
	var indices []uint32

	for x := 0; x < chunk.SizeX; x++ {
		for y := 0; y < chunk.SizeY; y++ {
			for z := 0; z < chunk.SizeZ; z++ {

				block := chunk.Blocks[blockIndex(x, y, z, chunk.SizeX, chunk.SizeY, chunk.SizeZ)]
				if block.Id == 0 {
					continue // Воздух не рисуем
				}
				if block.Id == 5 {
					fmt.Println(5) // Воздух не рисуем
				}

				// Для каждой из 6 граней куба
				for _, face := range cubeFaces {
					if y == 0 && face.OffsetY == -1 {
						continue
					}
					nx, ny, nz := x+face.OffsetX, y+face.OffsetY, z+face.OffsetZ
					if IsAirWithNeighbors(chunk, nx, ny, nz, neighbors) {

						// Нормаль этой грани:
						// face.OffsetX, face.OffsetY, face.OffsetZ
						// (для куба это 0,0,1 или -1,0,0 и т.д.)
						normX := float32(face.OffsetX)
						normY := float32(face.OffsetY)
						normZ := float32(face.OffsetZ)

						// Для удобства
						r := block.Color[0]
						g := block.Color[1]
						b := block.Color[2]

						startIdx := uint32(len(vertices) / 9)
						// т.к. теперь 9 float на вершину

						// Добавляем 4 вершины (квадрат)
						for _, vtx := range face.Vertices {
							px := float32(x) + vtx[0]
							py := float32(y) + vtx[1]
							pz := float32(z) + vtx[2]

							vertices = append(vertices,
								px, py, pz, // позиция
								normX, normY, normZ, // нормаль
								r, g, b) // цвет
						}

						// Индексы
						indices = append(indices,
							startIdx+0, startIdx+1, startIdx+2,
							startIdx+2, startIdx+3, startIdx+0)
					}
				}
			}
		}
	}

	return vertices, indices
}

func NewChunk(sizeX, sizeY, sizeZ int, offsetX, offsetZ int, noise opensimplex.Noise) *Chunk {
	blocks := make([]Block, sizeX*sizeY*sizeZ)
	noiseScale := 100.0                   // Масштаб шума для базовой высоты
	mountainScale := 200.0                // Масштаб шума для гор
	riverNoiseScale := 300.0              // Масштаб шума для рек
	biomeNoiseScale := 500.0              // Масштаб шума для биомов
	seaLevel := int(float64(sizeY) * 0.3) // Уровень моря (30% от высоты мира)

	for x := 0; x < sizeX; x++ {
		for z := 0; z < sizeZ; z++ {
			// Координаты в мире
			absoluteX := x + offsetX*sizeX
			absoluteZ := z + offsetZ*sizeZ

			// Генерация биома
			biomeNoise := noise.Eval2(float64(absoluteX)/biomeNoiseScale, float64(absoluteZ)/biomeNoiseScale)
			biome := determineBiome(biomeNoise)

			// Генерация базовой высоты
			baseHeight := noise.Eval2(float64(absoluteX)/noiseScale, float64(absoluteZ)/noiseScale)
			height := int((baseHeight + 1.0) * 0.3 * float64(sizeY-1)) // Высота ограничена 30% от размера чанка

			// Добавление гор
			if biome == "mountains" {
				mountainHeight := noise.Eval2(float64(absoluteX)/mountainScale, float64(absoluteZ)/mountainScale)
				height += int((mountainHeight + 1.0) * 0.4 * float64(sizeY-1)) // Горы добавляют до 40% высоты
			}

			// Генерация рек
			riverNoise := noise.Eval2(float64(absoluteX)/riverNoiseScale, float64(absoluteZ)/riverNoiseScale)
			isRiver := math.Abs(riverNoise) < 0.1 // Порог для реки

			for y := 0; y < sizeY; y++ {
				idx := blockIndex(x, y, z, sizeX, sizeY, sizeZ)

				if isRiver && y < seaLevel {
					// Река заполняет блоки ниже уровня моря
					blocks[idx] = Block{
						Id:    7, // ID воды
						Color: [3]float32{0.0, 0.0, 1.0},
					}
					continue
				}

				if y < height {
					// Заполнение горных биомов плотными блоками
					if biome == "mountains" && y > height-5 {
						blocks[idx] = Block{
							Id:    10, // Каменные вершины
							Color: [3]float32{0.7, 0.7, 0.7},
						}
					} else {
						blocks[idx] = generateUndergroundBlock(y, height, biome)
					}
				} else if y == height {
					// Верхний слой в зависимости от биома
					blocks[idx] = generateSurfaceBlock(biome)

					// Случайное дерево или растение
					if biome == "forest" && rand.Float64() < 0.1 {
						placeMinecraftTree(blocks, x, y+1, z, sizeX, sizeY, sizeZ)
					}
				} else if y <= seaLevel {
					// Вода ниже уровня моря
					blocks[idx] = Block{
						Id:    7, // ID воды
						Color: [3]float32{0.0, 0.0, 1.0},
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

	return &Chunk{
		Blocks: blocks,
		SizeX:  sizeX,
		SizeY:  sizeY,
		SizeZ:  sizeZ,
	}
}

func determineBiome(noiseValue float64) string {
	if noiseValue < -0.3 {
		return "desert"
	} else if noiseValue < 0.0 {
		return "plains"
	} else if noiseValue < 0.3 {
		return "forest"
	}
	return "mountains"
}

func generateUndergroundBlock(y, height int, biome string) Block {
	if biome == "desert" {
		return Block{Id: 8, Color: [3]float32{0.9, 0.8, 0.4}} // Песок
	}
	return Block{Id: 3, Color: [3]float32{0.5, 0.5, 0.5}} // Камень
}

func generateSurfaceBlock(biome string) Block {
	switch biome {
	case "desert":
		return Block{Id: 8, Color: [3]float32{0.9, 0.8, 0.4}} // Песок
	case "forest":
		return Block{Id: 2, Color: [3]float32{0.1, 0.8, 0.1}} // Трава
	case "plains":
		return Block{Id: 9, Color: [3]float32{0.4, 0.7, 0.1}} // Луга
	case "mountains":
		return Block{Id: 10, Color: [3]float32{0.7, 0.7, 0.7}} // Каменные вершины
	}
	return Block{Id: 2, Color: [3]float32{0.1, 0.8, 0.1}}
}

func placeMinecraftTree(blocks []Block, x, y, z, sizeX, sizeY, sizeZ int) {
	trunkHeight := 4 + rand.Intn(4) // Случайная высота ствола

	// Генерация ствола
	for i := 0; i < trunkHeight; i++ {
		yy := y + i
		if yy >= sizeY {
			break
		}
		idx := blockIndex(x, yy, z, sizeX, sizeY, sizeZ)
		if idx >= 0 && idx < len(blocks) {
			blocks[idx] = Block{
				Id:    5, // ID ствола
				Color: [3]float32{0.6, 0.3, 0.1},
			}
		}
	}

	// Генерация кроны
	generateLeaves(blocks, x, y+trunkHeight, z, sizeX, sizeY, sizeZ)
}

func generateLeaves(blocks []Block, x, y, z, sizeX, sizeY, sizeZ int) {
	leafRadius := 2
	leafHeight := 3 + rand.Intn(2)
	for dy := 0; dy < leafHeight; dy++ {
		yy := y + dy
		for dx := -leafRadius; dx <= leafRadius; dx++ {
			for dz := -leafRadius; dz <= leafRadius; dz++ {
				nx, nz := x+dx, z+dz
				if dx*dx+dz*dz <= leafRadius*leafRadius {
					idx := blockIndex(nx, yy, nz, sizeX, sizeY, sizeZ)
					if idx >= 0 && idx < len(blocks) && blocks[idx].Id == 0 {
						blocks[idx] = Block{
							Id:    6, // ID листвы
							Color: [3]float32{0.0 + 0.3*rand.Float32(), 0.8 + 0.2*rand.Float32(), 0.0 + 0.3*rand.Float32()},
						}
					}
				}
			}
		}
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
// func (chunk *Chunk) GenerateBuffers(neighbors map[string]*Chunk) {
// 	vertices, indices := chunk.GenerateMesh(neighbors)
// 	chunk.IndicesCount = len(indices)

// 	var vao, vbo, ebo uint32
// 	gl.GenVertexArrays(1, &vao)
// 	gl.BindVertexArray(vao)

// 	gl.GenBuffers(1, &vbo)
// 	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
// 	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

// 	gl.GenBuffers(1, &ebo)
// 	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
// 	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

// 	// Настройка атрибутов вершин
// 	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0)) // Координаты
// 	gl.EnableVertexAttribArray(0)
// 	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4)) // Цвет
// 	gl.EnableVertexAttribArray(1)

// 	chunk.VAO = vao
// 	chunk.VBO = vbo
// 	chunk.EBO = ebo
// }

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
				if len(chunkGenCh) < 99 {
					chunkGenCh <- [2]int{x, z}
				}
				//fmt.Printf("%d, %d\n", x, z)
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
func (w *World) RemoveChunk(cx, cz int, vramCh chan [3]uint32) {
	coord := [2]int{cx, cz}
	w.Mu.Lock()
	defer w.Mu.Unlock()
	chunk, exists := w.Chunks[coord]
	if !exists {
		return // Чанк уже удален или не существует
	}

	// // Удаляем буферы OpenGL
	// gl.DeleteVertexArrays(1, &chunk.VAO)
	// gl.DeleteBuffers(1, &chunk.VBO)
	// gl.DeleteBuffers(1, &chunk.EBO)

	// Удаляем чанк из карты
	vramCh <- [3]uint32{chunk.VAO, chunk.VBO, chunk.EBO}
	delete(w.Chunks, coord)
}
func (chunk *Chunk) GetBoundingBox(coord [2]int) [2]mgl32.Vec3 {
	// Минимальная точка чанка (нижний левый угол в мировых координатах)
	min := mgl32.Vec3{
		float32(coord[0] * chunk.SizeX),
		0, // Минимальная высота чанка всегда 0
		float32(coord[1] * chunk.SizeZ),
	}

	// Максимальная точка чанка (верхний правый угол в мировых координатах)
	max := mgl32.Vec3{
		float32((coord[0] + 1) * chunk.SizeX),
		float32(chunk.SizeY), // Максимальная высота чанка равна его высоте
		float32((coord[1] + 1) * chunk.SizeZ),
	}

	return [2]mgl32.Vec3{min, max}
}
func (w *World) GetBlock(x, y, z int) Block {
	// Проверяем высоту
	if y < 0 || y >= w.SizeY {
		return Block{Id: 0}
	}
	// Координаты чанка
	cx := x / w.SizeX
	cz := z / w.SizeZ

	// Локальные координаты внутри чанка
	lx := x % w.SizeX
	lz := z % w.SizeZ
	// Исправляем, если x или z отрицательные (Go % может давать отрицательное)
	if lx < 0 {
		lx += w.SizeX
		cx -= 1
	}
	if lz < 0 {
		lz += w.SizeZ
		cz -= 1
	}

	chunkCoord := [2]int{cx, cz}

	w.Mu.RLock()
	chunk, exists := w.Chunks[chunkCoord]
	w.Mu.RUnlock()
	if !exists {
		// Чанка нет — возвращаем воздух
		return Block{Id: 0}
	}

	// Индекс в одномерном массиве
	idx := blockIndex(lx, y, lz, chunk.SizeX, chunk.SizeY, chunk.SizeZ)
	return chunk.Blocks[idx]
}

func (w *World) SetBlock(x, y, z int, block Block) {
	// Проверяем, что координаты в допустимых пределах
	if y < 0 || y >= w.SizeY {
		return
	}
	// Координаты чанка
	cx := x / w.SizeX
	cz := z / w.SizeZ

	// Локальные координаты внутри чанка
	lx := x % w.SizeX
	lz := z % w.SizeZ
	if lx < 0 {
		lx += w.SizeX
		cx -= 1
	}
	if lz < 0 {
		lz += w.SizeZ
		cz -= 1
	}

	chunkCoord := [2]int{cx, cz}

	w.Mu.Lock()
	defer w.Mu.Unlock()

	chunk, exists := w.Chunks[chunkCoord]
	if !exists {
		return
	}

	idx := blockIndex(lx, y, lz, chunk.SizeX, chunk.SizeY, chunk.SizeZ)
	chunk.Blocks[idx] = block

	// Устанавливаем флаг обновления меша
	neighbors := w.collectNeighbors(cx, cz)
	chunk.UpdateBuffers(neighbors)
}

// RemoveBlock удаляет блок по мировым координатам (ставит воздух)
func (w *World) RemoveBlock(x, y, z int) {
	w.SetBlock(x, y, z, Block{Id: 0})
}
