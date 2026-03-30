package handlers

import (
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

func autoCategorizeReport(filename string) (category, subCategory string) {
	filenameLower := strings.ToLower(filename)
	
	if strings.Contains(filenameLower, "渗透") || strings.Contains(filenameLower, "pentest") || 
	   strings.Contains(filenameLower, "penetration") || strings.Contains(filenameLower, "web") && strings.Contains(filenameLower, "渗透") {
		category = "渗透测试报告"
		if strings.Contains(filenameLower, "web") || strings.Contains(filenameLower, "网站") {
			subCategory = "Web渗透"
		} else if strings.Contains(filenameLower, "app") || strings.Contains(filenameLower, "移动") {
			subCategory = "APP渗透"
		} else if strings.Contains(filenameLower, "内网") || strings.Contains(filenameLower, "域") {
			subCategory = "内网渗透"
		} else if strings.Contains(filenameLower, "红队") || strings.Contains(filenameLower, "red") {
			subCategory = "红队评估"
		} else {
			subCategory = "边界安全"
		}
	} else if strings.Contains(filenameLower, "代码") || strings.Contains(filenameLower, "审计") || 
		strings.Contains(filenameLower, "code") || strings.Contains(filenameLower, "audit") ||
		strings.Contains(filenameLower, "源码") {
		category = "代码审计报告"
		if strings.Contains(filenameLower, "web") {
			subCategory = "Web代码审计"
		} else if strings.Contains(filenameLower, "app") || strings.Contains(filenameLower, "移动") {
			subCategory = "APP代码审计"
		} else if strings.Contains(filenameLower, "sdk") {
			subCategory = "SDK审计"
		} else {
			subCategory = "源码审计"
		}
	} else if strings.Contains(filenameLower, "基线") || strings.Contains(filenameLower, "核查") || 
		strings.Contains(filenameLower, "baseline") || strings.Contains(filenameLower, "配置") {
		category = "基础环境测试"
		if strings.Contains(filenameLower, "基线") {
			subCategory = "基线检查"
		} else if strings.Contains(filenameLower, "配置") {
			subCategory = "配置核查"
		} else if strings.Contains(filenameLower, "漏洞扫描") || strings.Contains(filenameLower, "扫描") {
			subCategory = "漏洞扫描"
		} else if strings.Contains(filenameLower, "加固") {
			subCategory = "安全加固"
		} else {
			subCategory = "基线检查"
		}
	} else if strings.Contains(filenameLower, "漏洞扫描") || strings.Contains(filenameLower, "vulnerability") || 
		strings.Contains(filenameLower, "scan") {
		category = "基础环境测试"
		subCategory = "漏洞扫描"
	} else if strings.Contains(filenameLower, "应急") || strings.Contains(filenameLower, "incident") ||
		strings.Contains(filenameLower, "响应") || strings.Contains(filenameLower, "事件") {
		category = "应急响应报告"
		if strings.Contains(filenameLower, "溯源") {
			subCategory = "溯源分析"
		} else if strings.Contains(filenameLower, "处置") {
			subCategory = "处置报告"
		} else if strings.Contains(filenameLower, "事后") || strings.Contains(filenameLower, "总结") {
			subCategory = "事后分析"
		} else {
			subCategory = "事件分析"
		}
	} else if strings.Contains(filenameLower, "风险") || strings.Contains(filenameLower, "risk") {
		category = "风险评估报告"
		if strings.Contains(filenameLower, "资产") {
			subCategory = "资产识别"
		} else if strings.Contains(filenameLower, "威胁") {
			subCategory = "威胁评估"
		} else if strings.Contains(filenameLower, "脆弱性") || strings.Contains(filenameLower, "漏洞") {
			subCategory = "脆弱性评估"
		} else {
			subCategory = "风险矩阵"
		}
	} else if strings.Contains(filenameLower, "合规") || strings.Contains(filenameLower, "等保") ||
		strings.Contains(filenameLower, "iso") || strings.Contains(filenameLower, "compliance") ||
		strings.Contains(filenameLower, "审计") {
		category = "合规审计报告"
		if strings.Contains(filenameLower, "等保") {
			subCategory = "等保测评"
		} else if strings.Contains(filenameLower, "iso") {
			subCategory = "ISO27001"
		} else if strings.Contains(filenameLower, "pci") {
			subCategory = "PCI-DSS"
		} else {
			subCategory = "合规审计"
		}
	} else {
		category = "其他报告"
		subCategory = "其他"
	}

	isRetest := strings.Contains(filename, "复测")
	if isRetest {
		subCategory = subCategory + "-复测"
	}

	return category, subCategory
}

type ImportReportRequest struct {
	Name        string `json:"name" binding:"required"`
	ProjectID   *uint  `json:"project_id"`
	Category    string `json:"category"`
	SubCategory string `json:"sub_category"`
	Tags        string `json:"tags"`
	Description string `json:"description"`
}

func GetReports(c *gin.Context) {
	var reports []models.Report
	query := services.DB.Preload("Project")

	if projectID := c.Query("project_id"); projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if subCategory := c.Query("sub_category"); subCategory != "" {
		query = query.Where("sub_category = ?", subCategory)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR tags LIKE ? OR description LIKE ?", 
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Order("upload_time DESC").Find(&reports)
	c.JSON(http.StatusOK, reports)
}

func GetReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	var report models.Report
	if err := services.DB.Preload("Project").First(&report, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	c.JSON(http.StatusOK, report)
}

func ImportReport(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	name := c.PostForm("name")
	if name == "" {
		name = header.Filename
	}

	projectIDStr := c.PostForm("project_id")
	var projectID *uint
	if projectIDStr != "" {
		id, _ := strconv.ParseUint(projectIDStr, 10, 32)
		uid := uint(id)
		projectID = &uid
	}

	category := c.PostForm("category")
	subCategory := c.PostForm("sub_category")
	tags := c.PostForm("tags")
	description := c.PostForm("description")

	ext := filepath.Ext(header.Filename)
	fileUUID := uuid.New().String()
	fileName := fmt.Sprintf("%s%s", fileUUID, ext)
	filePath := filepath.Join(config.AppConfig.ReportsDir, fileName)

	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	report := models.Report{
		UUID:        fileUUID,
		Name:        name,
		FileName:    header.Filename,
		FilePath:    filePath,
		FileSize:    header.Size,
		FileType:    ext[1:],
		Category:    category,
		SubCategory: subCategory,
		ProjectID:   projectID,
		Tags:        tags,
		Description: description,
		UploadTime:  time.Now(),
	}

	if err := services.DB.Create(&report).Error; err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create report record"})
		return
	}

	c.JSON(http.StatusCreated, report)
}

func UpdateReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	var report models.Report
	if err := services.DB.First(&report, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	var req ImportReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	report.Name = req.Name
	report.Category = req.Category
	report.SubCategory = req.SubCategory
	report.Tags = req.Tags
	report.Description = req.Description
	report.ProjectID = req.ProjectID

	services.DB.Save(&report)
	c.JSON(http.StatusOK, report)
}

func DeleteReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	var report models.Report
	if err := services.DB.First(&report, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	if report.FilePath != "" {
		os.Remove(report.FilePath)
	}

	services.DB.Delete(&report)
	c.JSON(http.StatusOK, gin.H{"message": "Report deleted successfully"})
}

func DownloadReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	var report models.Report
	if err := services.DB.First(&report, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	if report.FilePath == "" || !fileExists(report.FilePath) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+report.FileName)
	c.Header("Content-Type", "application/octet-stream")
	c.File(report.FilePath)
}

func PreviewReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	var report models.Report
	if err := services.DB.First(&report, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	if report.FilePath == "" || !fileExists(report.FilePath) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.File(report.FilePath)
}

func GetReportCategories(c *gin.Context) {
	categories := []struct {
		Name         string   `json:"name"`
		SubCategories []string `json:"sub_categories"`
	}{
		{
			Name: "渗透测试报告",
			SubCategories: []string{"Web渗透", "APP渗透", "内网渗透", "红队评估", "边界安全"},
		},
		{
			Name: "代码审计报告",
			SubCategories: []string{"Web代码审计", "APP代码审计", "SDK审计", "源码审计"},
		},
		{
			Name: "基础环境测试",
			SubCategories: []string{"基线检查", "配置核查", "漏洞扫描", "安全加固"},
		},
		{
			Name: "应急响应报告",
			SubCategories: []string{"事件分析", "溯源分析", "处置报告", "事后分析"},
		},
		{
			Name: "风险评估报告",
			SubCategories: []string{"资产识别", "威胁评估", "脆弱性评估", "风险矩阵"},
		},
		{
			Name: "合规审计报告",
			SubCategories: []string{"等保测评", "ISO27001", "PCI-DSS", "GDPR"},
		},
	}
	c.JSON(http.StatusOK, categories)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
