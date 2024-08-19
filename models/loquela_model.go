package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type LoquelaLanguage struct {
	gorm.Model
	Language            string `gorm:"size:255"`
	LoquelaVocabularies []LoquelaVocabulary
}

type LoquelaVocabulary struct {
	gorm.Model
	Word              string                 `gorm:"type:text"`
	Meaning           string                 `gorm:"type:text"`
	Reading           string                 `gorm:"type:text"`
	AudioPath         string                 `gorm:"type:text"`
	MetadataStore     datatypes.JSON         `gorm:"type:jsonb"`
	Metadata          map[string]interface{} `gorm:"-"`
	LoquelaLanguageID uint
}

func MigrateLoquela(db *gorm.DB) {
	db.AutoMigrate(LoquelaLanguage{})
	db.AutoMigrate(LoquelaVocabulary{})
}
