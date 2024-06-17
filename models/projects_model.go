package models

import "gorm.io/gorm"

type Project struct {
	gorm.Model
	Name        string `gorm:"size:100"`
	Description string `gorm:"type:text"`
	Path        string `gorm:"size:255"`
	ProjectTags []ProjectTag
}

type ProjectTag struct {
	gorm.Model
	Name      string `gorm:"size:100"`
	ProjectID uint
}
