package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Username  string         `gorm:"unique;size:50;not null" json:"username"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Nickname  string         `gorm:"size:100" json:"nickname"`
	Role      string         `gorm:"size:20;default:'user'" json:"role"`
}

type Project struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Name        string         `gorm:"size:200;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Category    string         `gorm:"size:100" json:"category"`
	ClientName  string         `gorm:"size:200" json:"client_name"`
	Contract    string         `gorm:"size:200" json:"contract"`
	ContractNo  string         `gorm:"size:100" json:"contract_no"`
	Status      string         `gorm:"size:20;default:'active'" json:"status"`
	Reports     []Report       `gorm:"foreignKey:ProjectID" json:"reports,omitempty"`
}

type Report struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	UUID        string         `gorm:"unique;size:36" json:"uuid"`
	Name        string         `gorm:"size:500;not null" json:"name"`
	FileName    string         `gorm:"size:500" json:"file_name"`
	FilePath    string         `gorm:"size:1000" json:"file_path"`
	FileSize    int64          `json:"file_size"`
	FileType    string         `gorm:"size:50" json:"file_type"`
	Category    string         `gorm:"size:100" json:"category"`
	SubCategory string         `gorm:"size:100" json:"sub_category"`
	ProjectID   *uint          `json:"project_id"`
	Project     *Project       `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Tags        string         `gorm:"size:500" json:"tags"`
	Description string         `gorm:"type:text" json:"description"`
	UploadTime  time.Time      `json:"upload_time"`
}

type Backup struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	FileName    string         `gorm:"size:200;not null" json:"file_name"`
	FilePath    string         `gorm:"size:1000;not null" json:"file_path"`
	FileSize    int64          `json:"file_size"`
	Description string         `gorm:"type:text" json:"description"`
	Type        string         `gorm:"size:20;default:'full'" json:"type"`
}

type AIConfig struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Provider     string         `gorm:"size:50;default:'openai'" json:"provider"`
	APIKey       string         `gorm:"size:500" json:"api_key"`
	APIEndpoint  string         `gorm:"size:500" json:"api_endpoint"`
	Model        string         `gorm:"size:100;default:'gpt-3.5-turbo'" json:"model"`
	CustomModel  string         `gorm:"size:100" json:"custom_model"`
	IsEnabled    bool           `gorm:"default:true" json:"is_enabled"`
}

type Template struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Name        string         `gorm:"size:200;not null" json:"name"`
	FileName    string         `gorm:"size:500" json:"file_name"`
	FilePath    string         `gorm:"size:1000" json:"file_path"`
	FileSize    int64          `json:"file_size"`
	Category    string         `gorm:"size:100" json:"category"`
	Description string         `gorm:"type:text" json:"description"`
}

func (r *Report) BeforeCreate(tx *gorm.DB) error {
	if r.UUID == "" {
		r.UUID = uuid.New().String()
	}
	if r.UploadTime.IsZero() {
		r.UploadTime = time.Now()
	}
	return nil
}
