package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Regular struct {
	gorm.Model
	Name            string `gorm:"size:100;not null"`
	Email           string `gorm:"size:100;unique;not null"`
	Password        string `gorm:"size:255;not null"`
	RegularAccessID uint
	RegularSession  RegularSession
	AccessLogs      []AccessLog
}

type RegularAccess struct {
	gorm.Model
	Access   string `gorm:"size:100;not null"`
	Regulars []Regular
}

type RegularSession struct {
	gorm.Model
	Token          string    `gorm:"size:255;not null"`
	LastAccessedAt time.Time `gorm:"type:timestamp;not null"`
	IPAddress      string    `gorm:"size:45"`
	RememberMe     bool      `gorm:"type:bool"`
	RegularState   RegularState
	RegularID      uint
}

type RegularState struct {
	gorm.Model
	LoggedIn         bool                   `gorm:"type:bool;default:false"`
	Page             string                 `gorm:"size:100;default:'index'"`
	PageState        string                 `gorm:"size:100;default:'default'"`
	PageDataStore    datatypes.JSON         `gorm:"type:jsonb"`
	PageData         map[string]interface{} `gorm:"-"`
	Timestamp        int64                  `gorm:"-"`
	Tokens           map[string]interface{} `gorm:"-"`
	RegularSessionID uint
}

func MigrateRegular(db *gorm.DB) {
	db.AutoMigrate(Regular{})
	db.AutoMigrate(RegularAccess{})
	db.AutoMigrate(RegularSession{})
	db.AutoMigrate(RegularState{})
}

func SeedRegular(db *gorm.DB) error {
	if err := seedAccess(db); err != nil {
		return err
	}
	if err := seedRegular(db); err != nil {
		return err
	}

	return nil
}

func seedAccess(db *gorm.DB) error {
	accesses := []RegularAccess{
		{Access: "Administrator"},
		{Access: "Developer"},
		{Access: "Enforcer"},
		{Access: "Regular"},
		{Access: "Guest"},
	}

	tx := db.Begin()
	for _, access := range accesses {
		if err := tx.Create(&access).Error; err != nil {
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

func seedRegular(db *gorm.DB) error {
	regulars := []Regular{
		{
			Name:     "Guest",
			Email:    "guest@alpha.com",
			Password: "guestAlpha",
		},
		{
			Name:     "Nuntius Agent",
			Email:    "nuntius@alpha.com",
			Password: "nuntiusAgent",
		},
	}

	accesses := []string{
		"Guest",
		"Regular",
	}

	tx := db.Begin()
	for i, regular := range regulars {
		var (
			access RegularAccess
		)
		// TODO: Hash Password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(regular.Password), 8)
		if err != nil {
			return err
		}
		regular.Password = string(hashedPassword)

		// TODO: Define Default Access
		if err := db.Where("access = ?", accesses[i]).First(&access).Error; err != nil {
			return err
		}
		regular.RegularAccessID = access.ID

		if err := tx.Create(&regular).Error; err != nil {
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
