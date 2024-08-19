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

func SeedLoquela(db *gorm.DB) error {
	if err := seedLoquelaLanguage(db); err != nil {
		return err
	}

	return nil
}

func seedLoquelaLanguage(db *gorm.DB) error {
	languages := []LoquelaLanguage{
		{Language: "Mandarin"},
		{Language: "English"},
		{Language: "Russian"},
	}

	tx := db.Begin()
	for _, language := range languages {
		if err := tx.Create(&language).Error; err != nil {
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
