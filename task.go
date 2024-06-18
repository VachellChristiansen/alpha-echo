package main

import (
	"alpha-echo/models"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Tasks func()

func RunTasks(ts []string, db *gorm.DB, logger Logger) {
	// TODO: Define Tasks
	funcMap := map[string]Tasks{
		"MigrateModels": migrateModels(db),
		"SeedModels":    seedModels(db, logger),
	}

	// TODO: Run and Log Execution Time
	for _, t := range ts {
		// TODO: Run Task
		if function, exists := funcMap[t]; exists {
			logger["TASK"].Printf("Running Task: %s\n", t)
			start := time.Now()
			function()
			logger["TASK"].Printf("Time Elapsed: %v\n", time.Since(start))
		} else {
			logger["TASK"].Printf("Function %s not found\n", t)
		}
	}
}

func migrateModels(db *gorm.DB) func() {
	return func() {
		db.AutoMigrate(models.AccessLog{})
		db.AutoMigrate(models.Regular{})
		db.AutoMigrate(models.RegularAccess{})
		db.AutoMigrate(models.RegularSession{})
		db.AutoMigrate(models.RegularState{})
		db.AutoMigrate(models.Project{})
		db.AutoMigrate(models.ProjectTag{})
	}
}

func seedModels(db *gorm.DB, logger Logger) func() {
	return func() {
		accessSeeder(db, logger)
		regularSeeder(db, logger)
		projectTagSeeder(db, logger)
		projectSeeder(db, logger)
	}
}

func accessSeeder(db *gorm.DB, logger Logger) {
	accesses := []models.RegularAccess{
		{Access: "Administrator"},
		{Access: "Developer"},
		{Access: "Enforcer"},
		{Access: "Regular"},
	}

	tx := db.Begin()
	for _, access := range accesses {
		if err := tx.Create(&access).Error; err != nil {
			tx.Rollback()
			logger["TASK"].Fatalf("[accessSeeder] Seeding Failure, Value: %v", access)
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		logger["TASK"].Fatalf("[accessSeeder] Transaction Commit Failure")
	}
}

func regularSeeder(db *gorm.DB, logger Logger) {
	regulars := []models.Regular{
		{
			Name:     "Guest",
			Email:    "guest@alpha.com",
			Password: "guestAlpha",
		},
	}

	var (
		access models.RegularAccess
	)

	tx := db.Begin()
	for _, regular := range regulars {
		// TODO: Hash Password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(regular.Password), 8)
		if err != nil {
			logger["TASK"].Fatalf("[regularSeeder] Generating Hashed Password Failure")
			return
		}
		regular.Password = string(hashedPassword)

		// TODO: Define Default Access
		if err := db.Where("access = ?", "Regular").First(&access).Error; err != nil {
			logger["TASK"].Fatalf("[regularSeeder] Generating Default Access Failure")
			return
		}
		regular.RegularAccessID = access.ID

		if err := tx.Create(&regular).Error; err != nil {
			tx.Rollback()
			logger["TASK"].Fatalf("[regularSeeder] Seeding Failure, Value: %v", regular)
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		logger["TASK"].Fatalf("[regularSeeder] Transaction Commit Failure")
	}
}

func projectSeeder(db *gorm.DB, logger Logger) {
	projects := []models.Project{
		{
			Name:        "Proximus",
			Description: "Machine learning projects.",
			Path:        fmt.Sprintf("%s/proximus", os.Getenv("ML_URL")),
		},
		{
			Name:        "Opus",
			Description: "Task management tool.",
			Path:        fmt.Sprintf("%s/opus", os.Getenv("SERVER_URL")),
		},
	}
	tags := []string{
		"ML,Tools",
		"Tools",
	}

	tx := db.Begin()
	for index, project := range projects {
		var tagData []models.ProjectTag
		if err := db.Where("name IN ?", strings.Split(tags[index], ",")).Find(&tagData).Error; err != nil {
			tx.Rollback()
			logger["TASK"].Fatalf("[projectSeeder] Fetching Tags Failure, value: %v", tags[index])
		}

		project.ProjectTags = tagData

		if err := tx.Create(&project).Error; err != nil {
			tx.Rollback()
			logger["TASK"].Fatalf("[projectSeeder] Seeding Failure, Value: %v", project)
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		logger["TASK"].Fatalf("[projectSeeder] Transaction Commit Failure")
	}
}

func projectTagSeeder(db *gorm.DB, logger Logger) {
	projectTags := []models.ProjectTag{
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
	}

	tx := db.Begin()
	for _, projectTag := range projectTags {
		if err := tx.Create(&projectTag).Error; err != nil {
			tx.Rollback()
			logger["TASK"].Fatalf("[projectTagSeeder] Seeding Failure, Value: %v", projectTag)
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		logger["TASK"].Fatalf("[projectTagSeeder] Transaction Commit Failure")
	}
}
