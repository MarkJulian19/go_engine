package src

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type Config struct {
	ContextVersionMajor int    `json:"ContextVersionMajor"`
	ContextVersionMinor int    `json:"ContextVersionMinor"`
	Width               int    `json:"Width"`
	Height              int    `json:"Height"`
	Title               string `json:"Title"`
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
