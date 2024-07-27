package handlers

import (
	"alpha-echo/models"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type NuntiusHandler interface {
	Default(c echo.Context) error
}

type NuntiusHandlerImpl struct {
	db       *gorm.DB
	validate *validator.Validate
	logger   map[string]*log.Logger
}

func NewNuntiusHandler(db *gorm.DB, validate *validator.Validate, logger map[string]*log.Logger) NuntiusHandler {
	return &NuntiusHandlerImpl{
		db:       db,
		validate: validate,
		logger:   logger,
	}
}

func (h *NuntiusHandlerImpl) Default(c echo.Context) error {
	regular := c.Get("regular").(models.Regular)

	regular.RegularSession.RegularState.Page = "nuntius"

	if err := saveState(c, &regular, h.db, h.logger); err != nil {
		return err
	}

	return c.Render(http.StatusOK, "body", regular.RegularSession.RegularState)
}
