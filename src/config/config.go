package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type Config struct {
	ContextVersionMajor int     `json:"ContextVersionMajor"`
	ContextVersionMinor int     `json:"ContextVersionMinor"`
	Width               int     `json:"Width"`
	Height              int     `json:"Height"`
	Title               string  `json:"Title"`
	ChunkDist           int     `json:"ChunkDist"`
	NumWorkers          int     `json:"NumWorkers"`
	ChunkX              int     `json:"ChunkX"`
	ChunkY              int     `json:"ChunkY"`
	ChunkZ              int     `json:"ChunkZ"`
	FogStartLoc         float32 `json:"FogStartLoc"`
	FogEndLoc           float32 `json:"FogEndLoc"`
	ShadowDist          float32 `json:"ShadowDist"`
	ShadowHeight        int32   `json:"ShadowHeight"`
	ShadowWidth         int32   `json:"ShadowWidth"`
	WarpScale           float64 `json:"WarpScale"`
	WarpAmp             float64 `json:"WarpAmp"`
	MaxTerrainHeight    float64 `json:"MaxTerrainHeight"`
	SeaLevel            float64 `json:"SeaLevel"`
}

func LoadConfigFromFile(filePath string) (*Config, error) {
	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer file.Close()

	// Читаем содержимое файла
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	// Парсим JSON в структуру
	var config Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("не удалось распарсить JSON: %w", err)
	}

	return &config, nil
}
func LoadConfig(file string) *Config {
	config, err := LoadConfigFromFile(file)
	if err != nil {
		log.Fatalln("Error reading config file:", err)
	}
	return config
}
