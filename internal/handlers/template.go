package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"secuarchive-web/internal/models"
	"secuarchive-web/internal/services"

	"github.com/gin-gonic/gin"
)

func GetTemplates(c *gin.Context) {
	var templates []models.Template
	services.DB.Order("created_at DESC").Find(&templates)
	c.JSON(http.StatusOK, templates)
}

func GetTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var template models.Template
	if err := services.DB.First(&template, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

func ImportTemplate(c *gin.Context) {
	name := c.PostForm("name")
	category := c.PostForm("category")
	description := c.PostForm("description")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	uploadDir := "./data/templates"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
	filePath := filepath.Join(uploadDir, filename)

	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	template := models.Template{
		Name:        name,
		FileName:    header.Filename,
		FilePath:    filePath,
		FileSize:    header.Size,
		Category:    category,
		Description: description,
	}

	if err := services.DB.Create(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save template"})
		return
	}

	c.JSON(http.StatusCreated, template)
}

func DownloadTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var template models.Template
	if err := services.DB.First(&template, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", template.FileName))
	c.Header("Content-Type", "application/octet-stream")
	c.File(template.FilePath)
}

func DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var template models.Template
	if err := services.DB.First(&template, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	if err := os.Remove(template.FilePath); err != nil {
		fmt.Printf("Failed to delete file: %v\n", err)
	}

	services.DB.Delete(&template)
	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}

func GetTemplateCategories(c *gin.Context) {
	categories := []string{
		"渗透测试",
		"代码审计",
		"基线核查",
		"漏洞扫描",
		"安全评估",
		"应急响应",
		"风险评估",
		"合规审计",
		"其他",
	}
	c.JSON(http.StatusOK, categories)
}
