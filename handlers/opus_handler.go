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

type OpusHandler interface {
	Default(c echo.Context) error
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

func (h *OpusHandlerImpl) Default(c echo.Context) error {
	regular := c.Get("regular").(models.Regular)

	regular.RegularSession.RegularState.Page = "opus"
	if err := h.saveState(c, &regular); err != nil {
		return err
	}

	return c.Render(http.StatusOK, "opus", regular.RegularSession.RegularState)
}

func (h *OpusHandlerImpl) saveState(c echo.Context, regular *models.Regular) error {
	if err := h.db.Save(&regular.RegularSession.RegularState).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusInternalServerError),
			Message: "Fetching Regular Information Error [Session Might Be Invalid]",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}
	return nil
}
