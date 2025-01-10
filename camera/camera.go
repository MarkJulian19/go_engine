package camera

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"

	// Предположим, что в пакете world у нас есть метод GetBlock.
	"engine/world"
)

var (
	wireframeMode   = false
	gravity         = 30.0 // сила гравитации
	jumpSpeed       = 20.0 // сила прыжка
	playerHeight    = 1.8  // высота игрока (примерно 2 блока)
	playerEyeOffset = 1.7  // где «глаза» относительно нижней точки
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

	velocityY    float32
	isOnGround   bool
	creativeMode bool // если true — режим «креатива» (полёт, нет коллизий)
}

func NewCamera(position mgl32.Vec3) *Camera {
	return &Camera{
		Position:     position,
		Yaw:          -90.0,
		Pitch:        0.0,
		Speed:        25.0,
		Sensitivity:  0.05,
		velocityY:    0,
		isOnGround:   false,
		creativeMode: false,
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
			float32(math.Sin(pitchRad)),
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
		return pos
	}
	return candidate
}

// checkCollision — проверяем коллизию «капсулы» игрока
func (cam *Camera) checkCollision(pos mgl32.Vec3, w *world.World) bool {
	stepCount := int(playerHeight * 2.0) // ~3
	heightStep := float32(playerHeight) / float32(stepCount)

	for i := 0; i <= stepCount; i++ {
		checkY := float64(pos.Y()) - playerEyeOffset + (float64(i) * float64(heightStep))
		if isSolidAt(w, float64(pos.X()), checkY, float64(pos.Z())) {
			return true
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
		return true
	}
	block := w.GetBlock(x, y, z)
	return block.Id != 0
}
