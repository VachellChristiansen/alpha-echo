package models

import (
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	Name            string `gorm:"size:100"`
	Description     string `gorm:"type:text"`
	PagePath        string `gorm:"size:255"`
	ExternalPath    string `gorm:"size:255"`
	RegularAccessID uint
	ProjectTags     []ProjectTag `gorm:"many2many:project_projecttags"`
}

type ProjectTag struct {
	gorm.Model
	Name  string `gorm:"size:100"`
	Color string `gorm:"size:20"`
}
