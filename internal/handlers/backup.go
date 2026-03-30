package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"secuarchive-web/internal/config"
	"secuarchive-web/internal/models"
	"secuarchive-web/internal/services"

	"github.com/gin-gonic/gin"
)

func GetBackups(c *gin.Context) {
	var backups []models.Backup
	services.DB.Order("created_at DESC").Find(&backups)
	c.JSON(http.StatusOK, backups)
}

func CreateBackup(c *gin.Context) {
	var req struct {
		Description string `json:"description"`
		Type        string `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Type = "full"
	}

	backupData := map[string]interface{}{
		"timestamp":     time.Now().Format("2006-01-02 15:04:05"),
		"projects":      []models.Project{},
		"reports":       []models.Report{},
		"ai_config":     models.AIConfig{},
		"description":  req.Description,
	}

	var projects []models.Project
	services.DB.Preload("Reports").Find(&projects)
	backupData["projects"] = projects

	var aiConfig models.AIConfig
	services.DB.First(&aiConfig)
	backupData["ai_config"] = aiConfig

	jsonData, err := json.MarshalIndent(backupData, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate backup data"})
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("secuarchive_backup_%s.json", timestamp)
	filePath := filepath.Join(config.AppConfig.BackupsDir, fileName)

	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save backup file"})
		return
	}

	backup := models.Backup{
		FileName:    fileName,
		FilePath:    filePath,
		FileSize:    int64(len(jsonData)),
		Description: req.Description,
		Type:        req.Type,
	}

	services.DB.Create(&backup)

	c.JSON(http.StatusCreated, backup)
}

func RestoreBackup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backup ID"})
		return
	}

	var backup models.Backup
	if err := services.DB.First(&backup, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
		return
	}

	data, err := os.ReadFile(backup.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read backup file"})
		return
	}

	var backupData map[string]interface{}
	err = json.Unmarshal(data, &backupData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid backup file format"})
		return
	}

	services.DB.Exec("DELETE FROM reports")
	services.DB.Exec("DELETE FROM projects")

	if projectsRaw, ok := backupData["projects"]; ok {
		projectsJSON, _ := json.Marshal(projectsRaw)
		var projects []models.Project
		json.Unmarshal(projectsJSON, &projects)
		for i := range projects {
			projects[i].ID = 0
			services.DB.Create(&projects[i])
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Backup restored successfully"})
}

func DeleteBackup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backup ID"})
		return
	}

	var backup models.Backup
	if err := services.DB.First(&backup, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
		return
	}

	if backup.FilePath != "" {
		os.Remove(backup.FilePath)
	}

	services.DB.Delete(&backup)
	c.JSON(http.StatusOK, gin.H{"message": "Backup deleted successfully"})
}

func DownloadBackup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backup ID"})
		return
	}

	var backup models.Backup
	if err := services.DB.First(&backup, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
		return
	}

	if backup.FilePath == "" || !fileExists(backup.FilePath) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+backup.FileName)
	c.Header("Content-Type", "application/json")
	c.File(backup.FilePath)
}

func ImportBackup(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	var backupData map[string]interface{}
	err = json.Unmarshal(data, &backupData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backup file format"})
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("secuarchive_backup_imported_%s.json", timestamp)
	filePath := filepath.Join(config.AppConfig.BackupsDir, fileName)

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save backup file"})
		return
	}

	backup := models.Backup{
		FileName:    fileName,
		FilePath:    filePath,
		FileSize:    header.Size,
		Description: "Imported backup",
		Type:        "imported",
	}

	services.DB.Create(&backup)

	services.DB.Exec("DELETE FROM reports")
	services.DB.Exec("DELETE FROM projects")

	if projectsRaw, ok := backupData["projects"]; ok {
		projectsJSON, _ := json.Marshal(projectsRaw)
		var projects []models.Project
		json.Unmarshal(projectsJSON, &projects)
		for i := range projects {
			projects[i].ID = 0
			services.DB.Create(&projects[i])
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Backup imported and restored successfully", "backup": backup})
}
