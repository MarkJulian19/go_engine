package player

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"engine/src/world"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	wireframeMode   = false
	gravity         = 30.0 // сила гравитации
	jumpSpeed       = 20.0 // сила прыжка
	playerHeight    = 1.8  // высота игрока (примерно 2 блока)
	playerEyeOffset = 1.7  // где «глаза» относительно нижней точки
	playerRadius    = 0.3  // Радиус капсулы
	collisionStep   = 0.1  // Шаг проверки для окружности
)

// Camera описывает положение, ориентацию и «физику» игрока
type Camera struct {
	Position    mgl32.Vec3
	Yaw         float64
	Pitch       float64
	Speed       float32
	Sensitivity float64
	lastX       float64
	lastY       float64
	mu          sync.Mutex

	velocityY       float32
	isOnGround      bool
	creativeMode    bool // если true — режим «креатива» (полёт, нет коллизий)
	ShowInfoPanel   bool
	ShowHUD         bool
	lastPlaceAction time.Time
}

func NewCamera(position mgl32.Vec3) *Camera {
	return &Camera{
		Position:      position,
		Yaw:           -90.0,
		Pitch:         0.0,
		Speed:         55.0,
		Sensitivity:   0.05,
		velocityY:     0,
		isOnGround:    false,
		creativeMode:  false,
		ShowInfoPanel: false,
		ShowHUD:       true,
	}
}

// GetViewMatrix возвращает матрицу вида
func (cam *Camera) GetViewMatrix() mgl32.Mat4 {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	yawRad := float64(mgl32.DegToRad(float32(cam.Yaw)))
	pitchRad := float64(mgl32.DegToRad(float32(cam.Pitch)))

	direction := mgl32.Vec3{
		float32(math.Cos(yawRad) * math.Cos(pitchRad)),
		float32(math.Sin(pitchRad)),
		float32(math.Sin(yawRad) * math.Cos(pitchRad)),
	}.Normalize()

	return mgl32.LookAtV(cam.Position, cam.Position.Add(direction), mgl32.Vec3{0, 1, 0})
}

// ProcessKeyboard обрабатывает перемещения и переключения режимов
func (cam *Camera) ProcessKeyboard(window *glfw.Window, deltaTime float64, w *world.World) {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	if window.GetKey(glfw.KeyF1) == glfw.Press {
		cam.ShowHUD = !cam.ShowHUD
		time.Sleep(200 * time.Millisecond) // Задержка для предотвращения дребезга
	}
	if window.GetKey(glfw.KeyF3) == glfw.Press {
		cam.ShowInfoPanel = !cam.ShowInfoPanel
		time.Sleep(200 * time.Millisecond) // Задержка для предотвращения дребезга
	}

	// Переключение креативного режима по клавише ']'
	if window.GetKey(glfw.KeyRightBracket) == glfw.Press {
		cam.creativeMode = !cam.creativeMode
		if cam.creativeMode {
			fmt.Println("Creative mode ON")
		} else {
			fmt.Println("Creative mode OFF")
		}
		time.Sleep(200 * time.Millisecond)
	}

	speed := cam.Speed * float32(deltaTime)

	// Направление (без учёта pitch по Y — движение по плоскости)
	yawRad := float64(mgl32.DegToRad(float32(cam.Yaw)))
	forward := mgl32.Vec3{
		float32(math.Cos(yawRad)),
		0,
		float32(math.Sin(yawRad)),
	}.Normalize()

	// Правый вектор
	right := forward.Cross(mgl32.Vec3{0, 1, 0}).Normalize()

	// === Управление в обычном режиме (не креатив) ===
	if !cam.creativeMode {
		// WASD по XZ
		if window.GetKey(glfw.KeyW) == glfw.Press {
			cam.Position = cam.tryMove(cam.Position, forward, speed, w)
		}
		if window.GetKey(glfw.KeyS) == glfw.Press {
			cam.Position = cam.tryMove(cam.Position, forward.Mul(-1), speed, w)
		}
		if window.GetKey(glfw.KeyA) == glfw.Press {
			cam.Position = cam.tryMove(cam.Position, right.Mul(-1), speed, w)
		}
		if window.GetKey(glfw.KeyD) == glfw.Press {
			cam.Position = cam.tryMove(cam.Position, right, speed, w)
		}

		// Прыжок (пробел)
		if window.GetKey(glfw.KeySpace) == glfw.Press && cam.isOnGround {
			cam.velocityY = float32(jumpSpeed)
			cam.isOnGround = false
		}

	} else {
		// === Управление в креативном режиме ===
		// Игнорируем коллизии, можем летать свободно

		// Вектор вперёд с учётом Pitch (чтобы можно было летать вверх при наклоне)
		pitchRad := float64(mgl32.DegToRad(float32(cam.Pitch)))
		fullForward := mgl32.Vec3{
			float32(math.Cos(yawRad) * math.Cos(pitchRad)),
			float32(0),
			float32(math.Sin(yawRad) * math.Cos(pitchRad)),
		}.Normalize()

		// Вектор вправо (с учётом pitch=0, обычно)
		fullRight := fullForward.Cross(mgl32.Vec3{0, 1, 0}).Normalize()

		if window.GetKey(glfw.KeyW) == glfw.Press {
			cam.Position = cam.Position.Add(fullForward.Mul(speed))
		}
		if window.GetKey(glfw.KeyS) == glfw.Press {
			cam.Position = cam.Position.Sub(fullForward.Mul(speed))
		}
		if window.GetKey(glfw.KeyA) == glfw.Press {
			cam.Position = cam.Position.Sub(fullRight.Mul(speed))
		}
		if window.GetKey(glfw.KeyD) == glfw.Press {
			cam.Position = cam.Position.Add(fullRight.Mul(speed))
		}
		// Подъём/опускание по Y (можно также сделать shift / space)
		if window.GetKey(glfw.KeySpace) == glfw.Press {
			cam.Position = cam.Position.Add(mgl32.Vec3{0, speed, 0})
		}
		if window.GetKey(glfw.KeyLeftShift) == glfw.Press {
			cam.Position = cam.Position.Sub(mgl32.Vec3{0, speed, 0})
		}
	}

	// Выход из программы
	if window.GetKey(glfw.KeyEscape) == glfw.Press {
		os.Exit(0)
	}

	// Включение wireframe
	if window.GetKey(glfw.KeyM) == glfw.Press {
		wireframeMode = !wireframeMode
		if wireframeMode {
			fmt.Println("Wireframe ON")
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
		} else {
			fmt.Println("Wireframe OFF")
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// UpdatePhysics обновляет гравитацию и проверку коллизий
func (cam *Camera) UpdatePhysics(deltaTime float64, w *world.World) {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	// Если креативный режим — игнорируем гравитацию и коллизии
	if cam.creativeMode {
		return
	}

	// Обычный режим: гравитация + коллизии
	cam.velocityY -= float32(gravity) * float32(deltaTime)

	oldPos := cam.Position
	newPos := oldPos
	newPos[1] += cam.velocityY * float32(deltaTime)

	if cam.checkCollision(newPos, w) {
		if cam.velocityY < 0 {
			cam.isOnGround = true
		}
		cam.velocityY = 0
	} else {
		cam.isOnGround = false
		oldPos = newPos
	}
	cam.Position = oldPos
}

// ProcessMouse — вращение камеры
func (cam *Camera) ProcessMouse(xpos, ypos float64) {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	xOffset := xpos - cam.lastX
	yOffset := cam.lastY - ypos
	cam.lastX = xpos
	cam.lastY = ypos

	xOffset *= cam.Sensitivity
	yOffset *= cam.Sensitivity

	cam.Yaw += xOffset
	cam.Pitch += yOffset

	if cam.Pitch > 89.0 {
		cam.Pitch = 89.0
	}
	if cam.Pitch < -89.0 {
		cam.Pitch = -89.0
	}
}

// tryMove — движение в обычном режиме (учёт коллизий)
func (cam *Camera) tryMove(pos mgl32.Vec3, dir mgl32.Vec3, dist float32, w *world.World) mgl32.Vec3 {
	candidate := pos.Add(dir.Mul(dist))

	if cam.checkCollision(candidate, w) {
		// Попробуем скорректировать движение, исключая ось
		if !cam.checkCollision(pos.Add(mgl32.Vec3{dir.X() * dist, 0, 0}), w) {
			return pos.Add(mgl32.Vec3{dir.X() * dist, 0, 0})
		}
		if !cam.checkCollision(pos.Add(mgl32.Vec3{0, 0, dir.Z() * dist}), w) {
			return pos.Add(mgl32.Vec3{0, 0, dir.Z() * dist})
		}
		return pos // Полная блокировка
	}
	return candidate
}

// checkCollision — проверяем коллизию «капсулы» игрока
func (cam *Camera) checkCollision(pos mgl32.Vec3, w *world.World) bool {
	stepCount := int(playerHeight / collisionStep)
	angleStep := math.Pi * 2 / 8 // Проверяем 8 точек на окружности

	for i := 0; i <= stepCount; i++ {
		checkY := pos.Y() - float32(playerEyeOffset) + float32(i)*float32(collisionStep)

		for angle := 0.0; angle < math.Pi*2; angle += angleStep {
			offsetX := float32(math.Cos(angle)) * float32(playerRadius)
			offsetZ := float32(math.Sin(angle)) * float32(playerRadius)

			checkPos := mgl32.Vec3{
				pos.X() + offsetX,
				checkY,
				pos.Z() + offsetZ,
			}

			if isSolidAt(w, float64(checkPos.X()), float64(checkPos.Y()), float64(checkPos.Z())) {
				return true
			}
		}
	}
	return false
}

// isSolidAt — проверяет, твёрдый ли блок (Id != 0)
func isSolidAt(w *world.World, fx, fy, fz float64) bool {
	x := int(math.Floor(float64(fx)))
	y := int(math.Floor(float64(fy)))
	z := int(math.Floor(float64(fz)))

	if y < 0 || y >= w.SizeY {
		return false
	}
	block := w.GetBlock(x, y, z)
	return block.Id != 0
}
func (cam *Camera) InteractWithBlock(window *glfw.Window, w *world.World) {
	// Проверяем нажатие левой кнопки мыши для удаления блока
	if window.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press {
		currentTime := time.Now()
		if currentTime.Sub(cam.lastPlaceAction) < 100*time.Millisecond {
			// Если прошло меньше 500 мс, блок не ставится
			return
		}
		// Используем raycast для получения пересекаемого блока
		targetBlock, _, _ := cam.raycast(w, float64(7.0))
		if targetBlock != nil {
			// Удаляем блок
			fmt.Printf("RemoveBlock %d %d %d\n", targetBlock[0], targetBlock[1], targetBlock[2])
			w.RemoveBlock(targetBlock[0], targetBlock[1], targetBlock[2])
			cam.lastPlaceAction = currentTime
		}
	}

	// Проверяем нажатие правой кнопки мыши для добавления блока
	if window.GetMouseButton(glfw.MouseButtonRight) == glfw.Press {
		currentTime := time.Now()
		if currentTime.Sub(cam.lastPlaceAction) < 100*time.Millisecond {
			// Если прошло меньше 500 мс, блок не ставится
			return
		}
		// Используем raycast для получения пересекаемого блока и позиции нового блока
		_, normal, newBlockPos := cam.raycast(w, float64(7.0))
		// checkPos := mgl32.Vec3{float32(newBlockPos[0]), float32(newBlockPos[1]), float32(newBlockPos[2])}
		// checkPos.Normalize()
		// println(checkPos.X, checkPos.Y, checkPos.Z)
		if newBlockPos != nil {
			// Проверяем, чтобы новый блок не заменял существующий
			existingBlock := w.GetBlock(newBlockPos[0], newBlockPos[1], newBlockPos[2])
			if existingBlock.Id == 0 {
				// Добавляем новый блок
				fmt.Printf("SetBlock %d %d %d (Normal: %v)\n", newBlockPos[0], newBlockPos[1], newBlockPos[2], normal)
				w.SetBlock(newBlockPos[0], newBlockPos[1], newBlockPos[2], world.Block{
					Id:    1,                         // Пример ID блока
					Color: [3]float32{0.8, 0.6, 0.4}, // Пример цвета
				})
				cam.lastPlaceAction = currentTime
			}
		}
	}
}

func (cam *Camera) raycast(w *world.World, maxDistance float64) (*[3]int, mgl32.Vec3, *[3]int) {
	// Расчет направления трассировки луча
	direction := mgl32.Vec3{
		float32(math.Cos(float64(mgl32.DegToRad(float32(cam.Yaw)))) * math.Cos(float64(mgl32.DegToRad(float32(cam.Pitch))))),
		float32(math.Sin(float64(mgl32.DegToRad(float32(cam.Pitch))))),
		float32(math.Sin(float64(mgl32.DegToRad(float32(cam.Yaw)))) * math.Cos(float64(mgl32.DegToRad(float32(cam.Pitch))))),
	}.Normalize()

	origin := cam.Position
	step := 0.0001 // Шаг трассировки
	for t := float64(0); t < maxDistance; t += step {
		pos := origin.Add(direction.Mul(float32(t)))
		x, y, z := int(math.Floor(float64(pos.X()))), int(math.Floor(float64(pos.Y()))), int(math.Floor(float64(pos.Z())))

		block := w.GetBlock(x, y, z)
		if block.Id != 0 {
			// Вычисляем направление нормали к грани блока
			epsilon := float32(0.001)
			dx := pos.X() - float32(x)
			dy := pos.Y() - float32(y)
			dz := pos.Z() - float32(z)

			normal := mgl32.Vec3{}
			if dx < epsilon {
				normal = mgl32.Vec3{-1, 0, 0}
			} else if dx > 1-epsilon {
				normal = mgl32.Vec3{1, 0, 0}
			}
			if dy < epsilon {
				normal = mgl32.Vec3{0, -1, 0}
			} else if dy > 1-epsilon {
				normal = mgl32.Vec3{0, 1, 0}
			}
			if dz < epsilon {
				normal = mgl32.Vec3{0, 0, -1}
			} else if dz > 1-epsilon {
				normal = mgl32.Vec3{0, 0, 1}
			}

			// Если нормаль все еще (0, 0, 0), значит произошел сбой, пропускаем
			if normal.Len() == 0 {
				continue
			}

			// Вычисляем координаты нового блока на основе нормали
			newBlockPos := &[3]int{
				x + int(normal.X()),
				y + int(normal.Y()),
				z + int(normal.Z()),
			}

			return &[3]int{x, y, z}, normal, newBlockPos
		}
	}
	return nil, mgl32.Vec3{}, nil
}
