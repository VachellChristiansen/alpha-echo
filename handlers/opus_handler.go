package handlers

import (
	"log"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type OpusHandler interface {
}

type OpusHandlerImpl struct {
	db       *gorm.DB
	validate *validator.Validate
	logger   map[string]*log.Logger
}

func NewOpusHandler(db *gorm.DB, validate *validator.Validate, logger map[string]*log.Logger) OpusHandler {
	return &OpusHandlerImpl{
		db:       db,
		validate: validate,
		logger:   logger,
	}
}