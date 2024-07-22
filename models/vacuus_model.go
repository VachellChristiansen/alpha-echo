package models

import "gorm.io/gorm"

type VacuusAnimation struct {
	gorm.Model
	Name     string `gorm:"size:255"`
	Category string `gorm:"size:255"`
}

func MigrateVacuus(db *gorm.DB) {
	db.AutoMigrate(VacuusAnimation{})
}

func SeedVacuus(db *gorm.DB) error {
	if err := seedAnimation(db); err != nil {
		return err
	}

	return nil
}

func seedAnimation(db *gorm.DB) error {
	animations := []VacuusAnimation{
		{
			Name:     "random-bouncing-circles",
			Category: "Background",
		},
		{
			Name:     "circling-circles",
			Category: "Background",
		},
	}

	tx := db.Begin()
	for _, animation := range animations {
		if err := tx.Create(&animation).Error; err != nil {
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
