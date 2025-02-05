package world

import (
	"engine/src/config"
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

// var (
// 	biomeNoise   = opensimplex.New(rand.Int63()) // Для биомов
// 	terrainNoise = opensimplex.New(rand.Int63())
// )

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

}

// Обновляет буферы и меш чанка
func (chunk *Chunk) UpdateBuffers(neighbors map[string]*Chunk) {
	Vertices, indices := chunk.GenerateMesh(neighbors)

	chunk.IndicesCount = len(indices)
	chunk.Vertices = Vertices
	chunk.Indices = indices
	chunk.UpdateBuf = true
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
				// if block.Id == 5 {
				// 	fmt.Println(5) // Воздух не рисуем
				// }

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

var (
	biomeNoise   = opensimplex.New(rand.Int63())
	terrainNoise = opensimplex.New(rand.Int63())
	warpNoise    = opensimplex.New(rand.Int63())
)

// Параметры для каждого биома: минимальный и максимальный «коэффициент высоты», блок поверхности и подпочвы.
type Biome struct {
	Name            string
	MinHeightFactor float64
	MaxHeightFactor float64
	SurfaceBlock    Block
	SoilBlock       Block
}

// Определяем 4 «чистых» биома
var (
	biomeDesert = Biome{
		Name:            "desert",
		MinHeightFactor: 0.4,
		MaxHeightFactor: 0.6,
		SurfaceBlock: Block{
			Id:    8, // Песок
			Color: [3]float32{0.9, 0.8, 0.4},
		},
		SoilBlock: Block{
			Id:    8, // Песок
			Color: [3]float32{0.9, 0.8, 0.4},
		},
	}
	biomePlains = Biome{
		Name:            "plains",
		MinHeightFactor: 0.55,
		MaxHeightFactor: 0.65,
		SurfaceBlock: Block{
			Id:    9, // Луга
			Color: [3]float32{0.4, 0.7, 0.1},
		},
		SoilBlock: Block{
			Id:    1, // Земля
			Color: [3]float32{0.45, 0.36, 0.2},
		},
	}
	biomeForest = Biome{
		Name:            "forest",
		MinHeightFactor: 0.55,
		MaxHeightFactor: 0.75,
		SurfaceBlock: Block{
			Id:    2, // Трава (лесная)
			Color: [3]float32{0.1, 0.8, 0.1},
		},
		SoilBlock: Block{
			Id:    1,
			Color: [3]float32{0.45, 0.36, 0.2},
		},
	}
	biomeMountains = Biome{
		Name:            "mountains",
		MinHeightFactor: 0.7,
		MaxHeightFactor: 2.3, // Высокие горы
		SurfaceBlock: Block{
			Id:    10, // Камень (грубый)
			Color: [3]float32{0.6, 0.6, 0.6},
		},
		SoilBlock: Block{
			Id:    3, // Камень
			Color: [3]float32{0.5, 0.5, 0.5},
		},
	}
	// --- Новые биомы ---
	biomeSwamp = Biome{
		Name:            "swamp",
		MinHeightFactor: 0.25, // Низкие болота
		MaxHeightFactor: 0.5,
		SurfaceBlock: Block{
			Id:    11,                        // Грязь/или "тёмная трава"
			Color: [3]float32{0.2, 0.4, 0.1}, // Темнее, чем у обычной травы
		},
		SoilBlock: Block{
			Id:    1,                          // Земля
			Color: [3]float32{0.3, 0.25, 0.1}, // Более коричневая
		},
	}
	biomeSnow = Biome{
		Name:            "snow",
		MinHeightFactor: 0.6,
		MaxHeightFactor: 1.2, // Будет чуть повышенный рельеф
		SurfaceBlock: Block{
			Id:    12, // Снег
			Color: [3]float32{1.0, 1.0, 1.0},
		},
		SoilBlock: Block{
			Id:    3, // Камень под снегом
			Color: [3]float32{0.6, 0.6, 0.6},
		},
	}
)

// Линейная интерполяция
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// Плавная функция (s-кривая), чуть более «мягкая», чем линейная
func smoothstep(edge0, edge1, x float64) float64 {
	// Классический smoothstep: 3x^2 - 2x^3
	if x <= edge0 {
		return 0.0
	}
	if x >= edge1 {
		return 1.0
	}
	t := (x - edge0) / (edge1 - edge0)
	return t * t * (3.0 - 2.0*t)
}

// Функция смешивает параметры двух биомов (A,B) с помощью t
func blendBiomes(bA, bB Biome, t float64) Biome {
	return Biome{
		Name:            "mixed",
		MinHeightFactor: lerp(bA.MinHeightFactor, bB.MinHeightFactor, t),
		MaxHeightFactor: lerp(bA.MaxHeightFactor, bB.MaxHeightFactor, t),
		// Упрощённо берём surface/soil от «доминантного» биома (если t<0.5 => bA)
		// Можно усложнить и смешать цвета.
		SurfaceBlock: func() Block {
			if t < 0.5 {
				return bA.SurfaceBlock
			}
			return bB.SurfaceBlock
		}(),
		SoilBlock: func() Block {
			if t < 0.5 {
				return bA.SoilBlock
			}
			return bB.SoilBlock
		}(),
	}
}

// Выдаёт "чистый" биом без смешивания, исходя из bVal
// func pickPureBiome(bVal float64) Biome {
// 	switch {
// 	case bVal < -0.5:
// 		return biomeDesert
// 	case bVal < 0.0:
// 		return biomePlains
// 	case bVal < 0.5:
// 		return biomeForest
// 	default:
// 		return biomeMountains
// 	}
// }

// Выдаёт «смешанный» биом, если bVal попал в переходную зону между биомами
func pickBiomeSmooth(bVal float64) Biome {
	switch {
	case bVal < -0.7:
		// Чистый swamp
		return biomeSwamp

	case bVal < -0.5:
		// Переход swamp -> desert
		//   bVal=-0.7 => swamp, bVal=-0.5 => desert
		t := smoothstep(-0.7, -0.5, bVal)
		return blendBiomes(biomeSwamp, biomeDesert, t)

	case bVal < -0.2:
		// desert
		t := smoothstep(-0.5, -0.2, bVal)
		return blendBiomes(biomeDesert, biomePlains, t)

	case bVal < 0.0:
		return biomePlains

	case bVal < 0.3:
		// Переход plains -> forest
		t := smoothstep(0.0, 0.3, bVal)
		return blendBiomes(biomePlains, biomeForest, t)

	case bVal < 0.5:
		return biomeForest

	case bVal < 0.7:
		// forest -> mountains
		t := smoothstep(0.5, 0.7, bVal)
		return blendBiomes(biomeForest, biomeMountains, t)

	case bVal < 0.8:
		return biomeMountains

	case bVal < 0.9:
		// mountains -> snow
		//   bVal=0.8 => mountains, bVal=0.9 => snow
		t := smoothstep(0.8, 0.9, bVal)
		return blendBiomes(biomeMountains, biomeSnow, t)

	default:
		// Чистый snow
		return biomeSnow
	}
}

// ------------------- NewChunk с «warp» и плавными переходами -------------------
func NewChunk(sizeX, sizeY, sizeZ int, offsetX, offsetZ int,
	biomeNoise, terrainNoise, warpNoise opensimplex.Noise, Config *config.Config,
) *Chunk {

	blocks := make([]Block, sizeX*sizeY*sizeZ)

	// Хотим, чтобы ~60% высоты занимало твёрдое
	maxTerrainHeight := int(Config.MaxTerrainHeight * float64(sizeY))
	seaLevel := int(Config.SeaLevel * float64(sizeY))

	const octaves = 6
	scales := []float64{256, 128, 64, 32, 16, 8}
	amplitudes := []float64{1.0, 0.5, 0.25, 0.125, 0.0625, 0.03125}

	warpScale := Config.WarpScale
	warpAmp := Config.WarpAmp

	for x := 0; x < sizeX; x++ {
		for z := 0; z < sizeZ; z++ {
			worldX := float64(x + offsetX*sizeX)
			worldZ := float64(z + offsetZ*sizeZ)

			// Warp
			wVal := warpNoise.Eval2(worldX/warpScale, worldZ/warpScale)
			warp := wVal * warpAmp

			warpedX := worldX + warp
			warpedZ := worldZ - warp
			bVal := biomeNoise.Eval2(warpedX/300.0, warpedZ/300.0)

			currentBiome := pickBiomeSmooth(bVal)

			var totalNoise float64
			var ampSum float64
			for i := 0; i < octaves; i++ {
				val := terrainNoise.Eval2(worldX/scales[i], worldZ/scales[i])
				totalNoise += val * amplitudes[i]
				ampSum += amplitudes[i]
			}
			totalNoise /= ampSum
			normNoise := (totalNoise + 1) / 2
			if normNoise < 0 {
				normNoise = 0
			} else if normNoise > 1 {
				normNoise = 1
			}
			baseHeight := int(normNoise * float64(maxTerrainHeight))

			factor := lerp(currentBiome.MinHeightFactor, currentBiome.MaxHeightFactor, normNoise)
			finalHeight := int(float64(baseHeight) * factor)
			if finalHeight < 0 {
				finalHeight = 0
			}
			if finalHeight >= sizeY {
				finalHeight = sizeY - 1
			}

			for y := 0; y < sizeY; y++ {
				idx := blockIndex(x, y, z, sizeX, sizeY, sizeZ)

				if y < finalHeight {
					if y < finalHeight-4 {
						// Глубина — камень
						blocks[idx] = Block{
							Id:    3,
							Color: [3]float32{0.5, 0.5, 0.5},
						}
					} else {
						// Почва
						blocks[idx] = currentBiome.SoilBlock
					}
				} else if y == finalHeight {
					// Поверхность
					blocks[idx] = currentBiome.SurfaceBlock
				} else if y < seaLevel {
					blocks[idx] = Block{
						Id:    7, // Вода
						Color: [3]float32{0.0, 0.0, 1.0},
					}
				} else {
					blocks[idx] = Block{
						Id:    0, // Воздух
						Color: [3]float32{0.5, 0.8, 1.0},
					}
				}
			}
			if (currentBiome.Name == "plains" || currentBiome.Name == "forest") &&
				finalHeight >= seaLevel && finalHeight < sizeY-1 {
				if rand.Float64() < 0.02 {
					placeTree(blocks, x, finalHeight+1, z, sizeX, sizeY, sizeZ)
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

// placeTree — простое «майнкрафтовское» дерево
func placeTree(blocks []Block, x, y, z, sizeX, sizeY, sizeZ int) {
	trunkHeight := 4 + rand.Intn(3)
	for i := 0; i < trunkHeight; i++ {
		yy := y + i
		if yy >= sizeY {
			break
		}
		idx := blockIndex(x, yy, z, sizeX, sizeY, sizeZ)
		blocks[idx] = Block{
			Id:    5, // Ствол
			Color: [3]float32{0.5, 0.3, 0.1},
		}
	}
	generateLeaves(blocks, x, y+trunkHeight, z, sizeX, sizeY, sizeZ)
}

// generateLeaves — простая «сфера» листьев радиусом 2
func generateLeaves(blocks []Block, x, y, z, sizeX, sizeY, sizeZ int) {
	const radius = 3
	for offX := -radius; offX <= radius; offX++ {
		for offZ := -radius; offZ <= radius; offZ++ {
			for offY := -radius; offY <= radius; offY++ {
				nx := x + offX
				ny := y + offY
				nz := z + offZ
				if nx < 0 || nx >= sizeX ||
					ny < 0 || ny >= sizeY ||
					nz < 0 || nz >= sizeZ {
					continue
				}
				dist := math.Sqrt(float64(offX*offX + offY*offY + offZ*offZ))
				if dist <= float64(radius) {
					idx := blockIndex(nx, ny, nz, sizeX, sizeY, sizeZ)
					if blocks[idx].Id == 0 || blocks[idx].Id == 7 {
						blocks[idx] = Block{
							Id:    6, // Листва
							Color: [3]float32{0.0, 0.8, 0.0},
						}
					}
				}
			}
		}
	}
}

// Модифицируем GenerateChunk для передачи глобальных координат
func (w *World) GenerateChunk(cx, cz int, Config *config.Config) {

	coord := [2]int{cx, cz}
	w.Mu.Lock()
	if _, exists := w.Chunks[coord]; exists {
		w.Mu.Unlock()
		return
	}
	w.Mu.Unlock()
	// noise := opensimplex.New(2000)

	newChunk := NewChunk(w.SizeX, w.SizeY, w.SizeZ, cx, cz, biomeNoise, terrainNoise, warpNoise, Config)

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

	chunk, exists := w.Chunks[chunkCoord]
	if !exists {
		return
	}
	w.Mu.Unlock()
	idx := blockIndex(lx, y, lz, chunk.SizeX, chunk.SizeY, chunk.SizeZ)
	chunk.Blocks[idx] = block

	// Устанавливаем флаг обновления меша
	w.Mu.Lock()
	neighbors := w.collectNeighbors(cx, cz)
	chunk.UpdateBuffers(neighbors)
	w.Mu.Unlock()
	for direction, neighbor := range neighbors {
		if neighbor != nil {
			w.Mu.Lock()
			updatedNeighbors := w.collectNeighbors(cx+offsets[direction][0], cz+offsets[direction][1])
			w.Mu.Unlock()
			neighbor.UpdateBuffers(updatedNeighbors)
		}
	}
}

// RemoveBlock удаляет блок по мировым координатам (ставит воздух)
func (w *World) RemoveBlock(x, y, z int) {
	w.SetBlock(x, y, z, Block{Id: 0})
}
