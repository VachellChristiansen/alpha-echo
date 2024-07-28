package models

import "gorm.io/gorm"

type NuntiusChat struct {
	gorm.Model
	Type      string `gorm:"type:varchar(50);not null"`
	Text      string `gorm:"type:text"`
	AudioPath string `gorm:"type:text"`
	RegularID uint
}

func MigrateNuntius(db *gorm.DB) {
	db.AutoMigrate(NuntiusChat{})
}