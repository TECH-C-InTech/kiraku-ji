package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var loadDotEnvOnce sync.Once

// LoadDotEnv は .env が存在すれば 1 度だけ読み込む。
func LoadDotEnv() {
	loadDotEnvOnce.Do(func() {
		if _, err := os.Stat(".env"); err != nil {
			return
		}
		if err := godotenv.Load(); err != nil {
			log.Printf("dotenv: failed to load .env: %v", err)
		}
	})
}
