package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	gorm.Model
	Name      string `gorm:"size:255"`
	Priority  int    `gorm:"default:0"`
	Status    int    `gorm:"default:0"`
	Tasks     []Task
	RegularID uint
}

type Task struct {
	gorm.Model
	Title         string    `gorm:"size:255;not null"`
	Details       string    `gorm:"type:text"`
	Notes         string    `gorm:"type:text"`
	Priority      int       `gorm:"default:0"`
	Inset         int       `gorm:"default:1"`
	DoneAt        time.Time `gorm:"type:timestamp"`
	StartDate     time.Time `gorm:"type:timestamp"`
	EndDate       time.Time `gorm:"type:timestamp"`
	Status        int       `gorm:"default:0"`
	ParentTask    *Task     `gorm:"foreignKey:ParentID"`
	ChildrenTasks []Task    `gorm:"foreignKey:ParentID"`
	CategoryID    uint
	ParentID      *uint
	TaskGoals     []TaskGoal `gorm:"foreignKey:TaskID"`
}

type TaskGoal struct {
	gorm.Model
	GoalText  string    `gorm:"type:text"`
	DoneAt    time.Time `gorm:"type:timestamp"`
	StartDate time.Time `gorm:"type:timestamp"`
	EndDate   time.Time `gorm:"type:timestamp"`
	Status    int       `gorm:"default:0"`
	TaskID    uint      `gorm:"index"`
}

// Note:

// Possible Status for Category:
// 0: Active
// 1: Archived (TBA)

// Possible Status for Task:
// 0: Not Done
// 1: Done
// 2: Deleted

// Possible Status for TaskGoal:
// 0: Not Done
// 1: Done
