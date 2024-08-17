package handlers

import (
	"alpha-echo/models"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type LoquelaHandler interface {
	Default(c echo.Context) error
}

type LoquelaHandlerImpl struct {
	db       *gorm.DB
	validate *validator.Validate
	logger   map[string]*log.Logger
}

func NewLoquelaHandler(db *gorm.DB, validate *validator.Validate, logger map[string]*log.Logger) LoquelaHandler {
	return &LoquelaHandlerImpl{
		db:       db,
		validate: validate,
		logger:   logger,
	}
}

func (h *LoquelaHandlerImpl) Default(c echo.Context) error {
	regular := c.Get("regular").(models.Regular)

	regular.RegularSession.RegularState.Page = "loquela"

	if err := saveState(c, &regular, h.db, h.logger); err != nil {
		return err
	}

	return c.Render(http.StatusOK, "body", regular.RegularSession.RegularState)
}