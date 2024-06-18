package models

import (
	"encoding/json"
	"time"

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
	Page             string                 `gorm:"size:100;default:'default'"`
	PageState        string                 `gorm:"size:100;default:'default'"`
	PageDataStore    json.RawMessage        `gorm:"type:jsonb"`
	PageData         map[string]interface{} `gorm:"-"`
	RegularSessionID uint
}
