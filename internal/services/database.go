package services

import (
	"secuarchive-web/internal/config"
	"secuarchive-web/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	var err error
	DB, err = gorm.Open(sqlite.Open(config.AppConfig.DBPath), &gorm.Config{})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Report{},
		&models.Backup{},
		&models.AIConfig{},
		&models.Template{},
	)
	if err != nil {
		return err
	}

	createDefaultUser()

	return nil
}

func createDefaultUser() {
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		defaultUser := models.User{
			Username: "admin",
			Password: "$2b$10$UrXiP4Fb7E7/S.QSoB9V.OtydhDsLSEuV1HiCaYWaZ8ls8rltnclS",
			Nickname: "管理员",
			Role:     "admin",
		}
		DB.Create(&defaultUser)

		defaultUser2 := models.User{
			Username: "user",
			Password: "$2b$10$UrXiP4Fb7E7/S.QSoB9V.OtydhDsLSEuV1HiCaYWaZ8ls8rltnclS",
			Nickname: "普通用户",
			Role:     "user",
		}
		DB.Create(&defaultUser2)
	}

	var aiConfigCount int64
	DB.Model(&models.AIConfig{}).Count(&aiConfigCount)
	if aiConfigCount == 0 {
		aiConfig := models.AIConfig{
			Provider:  "openai",
			Model:     "gpt-3.5-turbo",
			IsEnabled: false,
		}
		DB.Create(&aiConfig)
	}
}
