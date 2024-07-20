package main

import (
	"alpha-echo/models"
	"os"
	"time"

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
		models.MigrateProgram(db)
		models.MigrateRegular(db)
		models.MigrateProjects(db)
		models.MigrateOpus(db)
		models.MigrateVacuus(db)
	}
}

func seedModels(db *gorm.DB, logger Logger) func() {
	return func() {
		if err := models.SeedRegular(db); err != nil {
			logger["TASK"].Printf("Seeding Regular Models Failure, Error: %v", err)
			os.Exit(1)
		}
		if err := models.SeedProjects(db); err != nil {
			logger["TASK"].Printf("Seeding Opus Models Failure")
			os.Exit(1)
		}
	}
}
