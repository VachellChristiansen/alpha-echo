package handlers

import (
	"alpha-echo/dtos"
	"alpha-echo/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type VacuusHandler interface {
	Default(c echo.Context) error
	UpdateAnimation(c echo.Context) error
}

type VacuusHandlerImpl struct {
	db       *gorm.DB
	validate *validator.Validate
	logger   map[string]*log.Logger
}

func NewVacuusHandler(db *gorm.DB, validate *validator.Validate, logger map[string]*log.Logger) VacuusHandler {
	return &VacuusHandlerImpl{
		db:       db,
		validate: validate,
		logger:   logger,
	}
}

func (h *VacuusHandlerImpl) Default(c echo.Context) error {
	var (
		animations []models.VacuusAnimation
	)
	regular := c.Get("regular").(models.Regular)

	if err := h.db.Find(&animations).Error; err != nil {
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-DB-%v-VACUUS", http.StatusInternalServerError),
			Message: "Fetching Animations List Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	regular.RegularSession.RegularState.Page = "vacuus"
	regular.RegularSession.RegularState.PageData = h.extractAnimationList(make(map[string]interface{}), animations)
	regular.RegularSession.RegularState.PageDataStore = convertToDatabyte(regular.RegularSession.RegularState.PageData, h.logger)

	if err := saveState(c, &regular, h.db, h.logger); err != nil {
		return err
	}

	return c.Render(http.StatusOK, "body", regular.RegularSession.RegularState)
}

func (h *VacuusHandlerImpl) UpdateAnimation(c echo.Context) error {
	var (
		req dtos.UpdateVacuusAnimationRequest
	)

	regular := c.Get("regular").(models.Regular)

	if err := c.Bind(&req); err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-VACUUS", http.StatusBadRequest),
			Message: "Bad Request",
			Error:   err.Error(),
		}
		return c.Render(http.StatusBadRequest, "error", errorData)
	}

	if err := json.Unmarshal(regular.RegularSession.RegularState.PageDataStore, &regular.RegularSession.RegularState.PageData); err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusInternalServerError),
			Message: "Loading Page Data errorData",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	if req.Category == "Background" {
		regular.RegularSession.RegularState.PageData["BackgroundAnimation"] = req.Name
	}
	regular.RegularSession.RegularState.PageDataStore = convertToDatabyte(regular.RegularSession.RegularState.PageData, h.logger)

	if err := saveState(c, &regular, h.db, h.logger); err != nil {
		return err
	}

	return c.Render(http.StatusOK, "vacuus-script", regular.RegularSession.RegularState)
}

func (h *VacuusHandlerImpl) extractAnimationList(data map[string]interface{}, animations []models.VacuusAnimation) map[string]interface{} {
	result := make(map[string]interface{})

	for _, animation := range animations {
		if _, exists := result[animation.Category]; exists {
			result[animation.Category] = append(result[animation.Category].([]string), animation.Name)
		} else {
			result[animation.Category] = []string{animation.Name}
		}
	}

	data["Animations"] = result
	return data
}
