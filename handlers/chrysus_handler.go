package handlers

import (
	"alpha-echo/dtos"
	"alpha-echo/models"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ChrysusHandler interface {
	Default(c echo.Context) error
}

type ChrysusHandlerImpl struct {
	db       *gorm.DB
	validate *validator.Validate
	logger   map[string]*log.Logger
}

func NewChrysusHandler(db *gorm.DB, validate *validator.Validate, logger map[string]*log.Logger) ChrysusHandler {
	return &ChrysusHandlerImpl{
		db:       db,
		validate: validate,
		logger:   logger,
	}
}

func (h *ChrysusHandlerImpl) Default(c echo.Context) error {
	regular := c.Get("regular").(models.Regular)

	regular.RegularSession.RegularState.Page = "chrysus"

	if err := h.saveState(c, &regular); err != nil {
		return err
	}

	return c.Render(http.StatusOK, "body", regular.RegularSession.RegularState)
}

func (h *ChrysusHandlerImpl) saveState(c echo.Context, regular *models.Regular) error {
	if err := h.db.Save(&regular.RegularSession.RegularState).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-DB-%v-OPUS", http.StatusInternalServerError),
			Message: "Saving State Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}
	return nil
}
