package models

import (
	"gorm.io/gorm"
)

type AccessLog struct {
	gorm.Model
	Method         string `gorm:"size:20;not null"`
	Path           string `gorm:"size:150;not null"`
	APILatency     int64  `gorm:"not null"`
	OverallLatency int64  `gorm:"not null"`
	RegularID      uint
}

func MigrateProgram(db *gorm.DB) {
	db.AutoMigrate(AccessLog{})
}