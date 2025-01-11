package render

import (
	"engine/src/config"
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

func GetLightProjection(Config *config.Config) mgl32.Mat4 {
	size := Config.ShadowDist
	near := float32(0.1)
	far := float32(1000.0)
	return mgl32.Ortho(-size, size, -size, size, near, far)
}
func GetDynamicLightPos(playerPos mgl32.Vec3, timeOfDay float64) mgl32.Vec3 {
	// Константы для настроек движения солнца
	radius := float32(300.0)      // Радиус вращения
	fixedHeight := float32(500.0) // Фиксированная высота солнца

	// Вычисляем позицию солнца
	x := playerPos.X() + radius*float32(math.Cos(timeOfDay))
	y := playerPos.Y() + fixedHeight // Высота фиксирована
	z := playerPos.Z() + radius*float32(math.Sin(timeOfDay))

	return mgl32.Vec3{x, y, z}
}
