package render

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

func GetLightProjection() mgl32.Mat4 {
	size := float32(100.0)
	near := float32(0.1)
	far := float32(1000.0)
	return mgl32.Ortho(-size, size, -size, size, near, far)
}
func GetDynamicLightPos(playerPos mgl32.Vec3, timeOfDay float64) mgl32.Vec3 {
	radius := float32(300.0) // Радиус вращения
	height := float32(300.0) // Высота солнца
	x := playerPos.X() + radius*float32(math.Cos(timeOfDay))
	z := playerPos.Z() + radius*float32(math.Sin(timeOfDay))
	y := playerPos.Y() + height
	return mgl32.Vec3{x, y, z}
}
