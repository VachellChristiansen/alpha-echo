package models

import (
	"fmt"
	"os"
	"strings"

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

func MigrateProjects(db *gorm.DB) {
	db.AutoMigrate(ProjectTag{})
	db.AutoMigrate(Project{})
}

func SeedProjects(db *gorm.DB) error {
	if err := seedProjectTags(db); err != nil {
		return err
	}
	if err := seedProjects(db); err != nil {
		return err
	}

	return nil
}

func seedProjects(db *gorm.DB) error {
	projects := []Project{
		{
			Name:            "Proximus",
			Description:     "Machine learning projects.",
			PagePath:        "/a/proximus",
			ExternalPath:    fmt.Sprintf("%s/proximus", os.Getenv("ML_DOMAIN")),
			RegularAccessID: 1,
		},
		{
			Name:            "Opus",
			Description:     "Task management tool.",
			PagePath:        "/r/opus",
			ExternalPath:    fmt.Sprintf("%s/opus", os.Getenv("ML_DOMAIN")),
			RegularAccessID: 4,
		},
		{
			Name:            "Nuntius",
			Description:     "A Text To Speech System for Final Year Project.",
			PagePath:        "/r/nuntius",
			ExternalPath:    fmt.Sprintf("%s/nuntius", os.Getenv("ML_DOMAIN")),
			RegularAccessID: 4,
		},
		{
			Name:            "Vacuus",
			Description:     "Sandbox of whatever.",
			PagePath:        "/r/vacuus",
			ExternalPath:    fmt.Sprintf("%s/vacuus", os.Getenv("ML_DOMAIN")),
			RegularAccessID: 4,
		},
		{
			Name:            "Chrysus",
			Description:     "Finance Management tool.",
			PagePath:        "/r/chrysus",
			ExternalPath:    fmt.Sprintf("%s/chrysus", os.Getenv("ML_DOMAIN")),
			RegularAccessID: 4,
		},
		{
			Name:            "Elpida",
			Description:     "Finding more efficient means to run a program.",
			PagePath:        "/r/elpida",
			ExternalPath:    fmt.Sprintf("%s/elpida", os.Getenv("ML_DOMAIN")),
			RegularAccessID: 4,
		},
	}
	tags := []string{
		"ML,Tools",
		"Tools",
		"ML,Tools,Experimental",
		"Experimental",
		"Tools",
		"Tools,Experimental",
	}

	tx := db.Begin()
	for index, project := range projects {
		var tagData []ProjectTag
		if err := db.Where("name IN ?", strings.Split(tags[index], ",")).Find(&tagData).Error; err != nil {
			tx.Rollback()
			return err
		}

		project.ProjectTags = tagData

		if err := tx.Create(&project).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func seedProjectTags(db *gorm.DB) error {
	projectTags := []ProjectTag{
		{
			Name:  "Tools",
			Color: "amber",
		},
		{
			Name:  "ML",
			Color: "rose",
		},
		{
			Name:  "Games",
			Color: "emerald",
		},
		{
			Name:  "Experimental",
			Color: "sky",
		},
	}

	tx := db.Begin()
	for _, projectTag := range projectTags {
		if err := tx.Create(&projectTag).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
