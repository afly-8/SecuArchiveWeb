package config

import (
	"os"
	"time"
)

type Config struct {
	ServerPort     string
	JWT_SECRET     string
	JWT_EXPIRE     time.Duration
	DataDir        string
	ReportsDir     string
	BackupsDir     string
	ProjectsDir    string
	DBPath         string
}

var AppConfig *Config

func Init() {
	baseDir := "./data"
	if envDir := os.Getenv("DATA_DIR"); envDir != "" {
		baseDir = envDir
	}

	AppConfig = &Config{
		ServerPort: "8080",
		JWT_SECRET: "secuarchive-secret-key-2024",
		JWT_EXPIRE: 24 * time.Hour * 7,
		DataDir:    baseDir,
		ReportsDir: baseDir + "/reports",
		BackupsDir: baseDir + "/backups",
		ProjectsDir: baseDir + "/projects",
		DBPath:     baseDir + "/secuarchive.db",
	}

	os.MkdirAll(AppConfig.ReportsDir, 0755)
	os.MkdirAll(AppConfig.BackupsDir, 0755)
	os.MkdirAll(AppConfig.ProjectsDir, 0755)
}
