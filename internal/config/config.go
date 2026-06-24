package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds application configuration.
type Config struct {
	ServerAddr                     string `mapstructure:"server_addr"`
	LogLevel                       string `mapstructure:"log_level"`
	DatabaseURL                    string `mapstructure:"database_url"`
	DatabaseInstanceConnectionName string `mapstructure:"database_instance_connection_name"`
	DatabaseUser                   string `mapstructure:"database_user"`
	DatabasePassword               string `mapstructure:"database_password"`
	DatabaseName                   string `mapstructure:"database_name"`
	JWTSecret                      string `mapstructure:"jwt_secret"`
}

// Load reads configuration from .env file (if present) and environment variables.
func Load() (*Config, error) {
	loadDotEnv()

	v := viper.New()
	v.SetDefault("server_addr", ":8080")
	v.SetDefault("log_level", "info")
	v.SetDefault("database_url", "")
	v.SetDefault("database_instance_connection_name", "")
	v.SetDefault("database_user", "postgres")
	v.SetDefault("database_password", "")
	v.SetDefault("database_name", "postgres")
	v.SetDefault("jwt_secret", "")
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// loadDotEnv 尝试从当前工作目录或项目根目录加载 .env 文件。
// 这样无论从项目根目录（go run ./cmd/server）还是 cmd/server 目录
// （go run main.go）启动，都能正确读取配置。
func loadDotEnv() {
	// 先尝试当前工作目录
	_ = godotenv.Load()

	// 再根据当前源文件位置推算项目根目录
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return
	}
	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
	projectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return
	}

	envPath := filepath.Join(projectRoot, ".env")
	if _, err := os.Stat(envPath); err == nil {
		_ = godotenv.Load(envPath)
	}
}
