package models

import "gorm.io/gorm"

type VacuusAnimation struct {
	gorm.Model
	Name     string `gorm:"size:255"`
	Category string `gorm:"size:255"`
}