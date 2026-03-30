package handlers

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"secuarchive-web/internal/config"
	"secuarchive-web/internal/models"
	"secuarchive-web/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category"`
	ClientName  string `json:"client_name"`
	Contract    string `json:"contract"`
	ContractNo  string `json:"contract_no"`
}

func GetProjects(c *gin.Context) {
	var projects []models.Project
	query := services.DB.Preload("Reports")

	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ? OR client_name LIKE ?", 
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Order("created_at DESC").Find(&projects)
	c.JSON(http.StatusOK, projects)
}

func GetProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var project models.Project
	if err := services.DB.Preload("Reports").First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	project := models.Project{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		ClientName:  req.ClientName,
		Contract:    req.Contract,
		ContractNo:  req.ContractNo,
		Status:      "active",
	}

	if err := services.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

func UpdateProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var project models.Project
	if err := services.DB.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	project.Name = req.Name
	project.Description = req.Description
	project.Category = req.Category
	project.ClientName = req.ClientName
	project.Contract = req.Contract
	project.ContractNo = req.ContractNo

	services.DB.Save(&project)
	c.JSON(http.StatusOK, project)
}

func DeleteProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	if err := services.DB.Delete(&models.Project{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

func GetProjectCategories(c *gin.Context) {
	categories := []string{
		"渗透测试",
		"代码审计",
		"安全评估",
		"应急响应",
		"安全培训",
		"风险评估",
		"合规审计",
		"其他",
	}
	c.JSON(http.StatusOK, categories)
}

func CreateProjectWithZip(c *gin.Context) {
	name := c.PostForm("name")
	description := c.PostForm("description")
	category := c.PostForm("category")
	clientName := c.PostForm("client_name")
	contract := c.PostForm("contract")
	contractNo := c.PostForm("contract_no")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project name is required"})
		return
	}

	project := models.Project{
		Name:        name,
		Description: description,
		Category:    category,
		ClientName:  clientName,
		Contract:    contract,
		ContractNo:  contractNo,
		Status:      "active",
	}

	if err := services.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	reportCount := 0
	file, header, err := c.Request.FormFile("file")
	if err == nil {
		defer file.Close()

		if strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
			tempFile, err := os.CreateTemp("", "report_*.zip")
			if err == nil {
				defer os.Remove(tempFile.Name())
				io.Copy(tempFile, file)
				tempFile.Close()

				reportCount, err = extractAndImportReports(tempFile.Name(), project.ID)
				if err != nil {
					fmt.Printf("Error extracting reports: %v\n", err)
				}
			}
		}
	}

	services.DB.Preload("Reports").First(&project, project.ID)
	c.JSON(http.StatusCreated, gin.H{
		"project":     project,
		"reportCount": reportCount,
	})
}

func extractAndImportReports(zipPath string, projectID uint) (int, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, err
	}
	defer reader.Close()

	count := 0
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.FileInfo().Name()))
		if ext != ".pdf" && ext != ".docx" && ext != ".doc" && ext != ".xlsx" && ext != ".xls" && ext != ".txt" && ext != ".html" {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			continue
		}
		defer rc.Close()

		fileUUID := uuid.New().String()
		savePath := filepath.Join(config.AppConfig.ReportsDir, fileUUID+ext)
		out, err := os.Create(savePath)
		if err != nil {
			rc.Close()
			continue
		}

		io.Copy(out, rc)
		out.Close()
		rc.Close()

		reportName := strings.TrimSuffix(file.FileInfo().Name(), ext)
		category, subCategory := autoCategorizeReport(reportName)

		report := models.Report{
			UUID:        fileUUID,
			Name:        reportName,
			FileName:    file.FileInfo().Name(),
			FilePath:    savePath,
			FileSize:    int64(file.FileInfo().Size()),
			FileType:    ext[1:],
			Category:    category,
			SubCategory: subCategory,
			ProjectID:  &projectID,
			UploadTime:  time.Now(),
		}

		services.DB.Create(&report)
		count++
	}

	return count, nil
}
