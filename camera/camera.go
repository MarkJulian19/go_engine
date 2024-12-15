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
)

var wireframeMode = false

type Camera struct {
	Position    mgl32.Vec3
	Yaw         float64
	Pitch       float64
	Speed       float32
	Sensitivity float64
	lastX       float64
	lastY       float64
	mu          sync.Mutex
}

func NewCamera(position mgl32.Vec3) *Camera {
	return &Camera{
		Position:    position,
		Yaw:         -90.0,
		Pitch:       0.0,
		Speed:       35.0,
		Sensitivity: 0.05,
	}
}

func (cam *Camera) GetViewMatrix() mgl32.Mat4 {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	direction := mgl32.Vec3{
		float32(math.Cos(float64(mgl32.DegToRad(float32(cam.Yaw)))) * math.Cos(float64(mgl32.DegToRad(float32(cam.Pitch))))),
		float32(math.Sin(float64(mgl32.DegToRad(float32((cam.Pitch)))))),
		float32(math.Sin(float64(mgl32.DegToRad(float32(cam.Yaw)))) * math.Cos(float64(mgl32.DegToRad(float32(cam.Pitch))))),
	}.Normalize()
	return mgl32.LookAtV(cam.Position, cam.Position.Add(direction), mgl32.Vec3{0, 1, 0})
}

func (cam *Camera) ProcessKeyboard(window *glfw.Window, deltaTime float64) {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	speed := cam.Speed * float32(deltaTime)

	// Вычисляем фронтальный вектор (куда смотрит камера)
	forward := mgl32.Vec3{
		float32(math.Cos(float64(mgl32.DegToRad(float32(cam.Yaw)))) * math.Cos(float64(mgl32.DegToRad(float32(cam.Pitch))))),
		0,
		float32(math.Sin(float64(mgl32.DegToRad(float32(cam.Yaw)))) * math.Cos(float64(mgl32.DegToRad(float32(cam.Pitch))))),
	}.Normalize()

	// Вычисляем правый вектор (перпендикуляр к фронтальному и вверх)
	right := forward.Cross(mgl32.Vec3{0, 1, 0}).Normalize()

	// Вычисляем вверхний вектор
	up := right.Cross(forward).Normalize()

	// Обрабатываем ввод
	if window.GetKey(glfw.KeyW) == glfw.Press {
		cam.Position = cam.Position.Add(forward.Mul(speed)) // Движение вперёд
	}
	if window.GetKey(glfw.KeyS) == glfw.Press {
		cam.Position = cam.Position.Sub(forward.Mul(speed)) // Движение назад
	}
	if window.GetKey(glfw.KeyA) == glfw.Press {
		cam.Position = cam.Position.Sub(right.Mul(speed)) // Движение влево
	}
	if window.GetKey(glfw.KeyD) == glfw.Press {
		cam.Position = cam.Position.Add(right.Mul(speed)) // Движение вправо
	}
	if window.GetKey(glfw.KeyQ) == glfw.Press {
		cam.Position = cam.Position.Add(up.Mul(speed)) // Движение вверх
	}
	if window.GetKey(glfw.KeyE) == glfw.Press {
		cam.Position = cam.Position.Sub(up.Mul(speed)) // Движение вниз
	}
	if window.GetKey(glfw.KeyEscape) == glfw.Press {
		os.Exit(1) // Движение вниз
	}
	if window.GetKey(glfw.KeyM) == glfw.Press {
		wireframeMode = !wireframeMode
		if wireframeMode {
			fmt.Println(1)
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE) // Включаем wireframe
		} else {
			fmt.Println(2)
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL) // Возвращаем обычный режим
		}
		time.Sleep(100 * time.Millisecond) // Добавляем задержку для предотвращения многократного срабатывания
	}
}

func (cam *Camera) ProcessMouse(xpos, ypos float64) {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	xOffset := xpos - cam.lastX
	yOffset := cam.lastY - ypos // Инвертируем, так как y-координата в OpenGL работает наоборот
	cam.lastX = xpos
	cam.lastY = ypos

	// Применяем чувствительность
	xOffset *= cam.Sensitivity
	yOffset *= cam.Sensitivity

	// Обновляем углы поворота камеры
	cam.Yaw += xOffset
	cam.Pitch += yOffset

	// Ограничиваем наклон камеры
	if cam.Pitch > 89.0 {
		cam.Pitch = 89.0
	}
	if cam.Pitch < -89.0 {
		cam.Pitch = -89.0
	}

}
