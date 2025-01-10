package render

import "github.com/go-gl/mathgl/mgl32"

func GetLightProjection(playerPos mgl32.Vec3) mgl32.Mat4 {
	// Допустим, хотим ±200 вокруг игрока
	left := playerPos.X() - 200
	right := playerPos.X() + 200
	bottom := playerPos.Z() - 200
	top := playerPos.Z() + 200

	// Или Y можно взять в зависимости от высоты мира
	return mgl32.Ortho(left, right, bottom, top, 0.1, 500.0)
}
func GetDynamicLightPos(playerPos mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		playerPos.X() + 100,
		200,
		playerPos.Z() + 100,
	}
}
