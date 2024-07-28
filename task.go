package main

import (
	"alpha-echo/models"
	"os"
	"time"

	"gorm.io/gorm"
)

type Tasks func()

func RunTasks(ts []string, db *gorm.DB, logger Logger) {
	funcMap := map[string]Tasks{
		"MigrateModels": migrateModels(db),
		"SeedModels":    seedModels(db, logger),
	}

	for _, t := range ts {
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
		// Order is important
		models.MigrateProgram(db)
		models.MigrateRegular(db)
		models.MigrateProjects(db)
		models.MigrateOpus(db)
		models.MigrateVacuus(db)
		models.MigrateNuntius(db)
	}
}

func seedModels(db *gorm.DB, logger Logger) func() {
	return func() {
		logger["TASK"].Printf("Seeding Regular Models Begin...")
		if err := models.SeedRegular(db); err != nil {
			logger["TASK"].Printf("Seeding Regular Models Failure, Error: %v", err)
			os.Exit(1)
		}
		logger["TASK"].Printf("Seeding Regular Models Finished!")
		logger["TASK"].Printf("Seeding Projects Models Begin...")
		if err := models.SeedProjects(db); err != nil {
			logger["TASK"].Printf("Seeding Projects Models Failure, Error: %v", err)
			os.Exit(1)
		}
		logger["TASK"].Printf("Seeding Projects Models Finished!")
		logger["TASK"].Printf("Seeding Vacuus Models Begin...")
		if err := models.SeedVacuus(db); err != nil {
			logger["TASK"].Printf("Seeding Vacuus Models Failure, Error: %v", err)
			os.Exit(1)
		}
		logger["TASK"].Printf("Seeding Vacuus Models Finished!")
	}
}
