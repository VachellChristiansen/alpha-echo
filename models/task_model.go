package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	gorm.Model
	Name      string `gorm:"size:255"`
	Priority  int    `gorm:"default:0"`
	Tasks     []Task
	RegularID uint
}

type Task struct {
	gorm.Model
	Title         string    `gorm:"size:255;not null"`
	Priority      int       `gorm:"default:0"`
	DoneAt        time.Time `gorm:"type:timestamp"`
	ParentTask    *Task     `gorm:"foreignKey:ParentID"`
	ChildrenTasks []Task    `gorm:"foreignKey:ParentID"`
	CategoryID    uint
	ParentID      *uint
}
