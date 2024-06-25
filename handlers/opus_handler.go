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
	GetTasks(c echo.Context) error
	AddCategory(c echo.Context) error
	AddTask(c echo.Context) error
	DeleteCategory(c echo.Context) error
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

func (h *OpusHandlerImpl) GetTasks(c echo.Context) error {
	var (
		categories []models.Category
	)

	regular := c.Get("regular").(models.Regular)

	if err := h.db.Preload("Tasks").Where("regular_id = ?", regular.ID).Order("priority asc").Find(&categories).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Fetching Categories Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	return c.Render(http.StatusOK, "opus-task", categories)
}

func (h *OpusHandlerImpl) AddCategory(c echo.Context) error {
	var (
		req        dtos.AddCategoryRequest
		categories []models.Category
	)

	regular := c.Get("regular").(models.Regular)

	if err := c.Bind(&req); err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusBadRequest),
			Message: "Bad Request",
			Error:   err.Error(),
		}
		return c.Render(http.StatusBadRequest, "error", errorData)
	}

	newCategory := models.Category{
		Name:      req.Name,
		Priority:  req.Priority,
		RegularID: regular.ID,
	}

	if err := h.db.Create(&newCategory).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusInternalServerError),
			Message: "Creating Category Data Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	if err := h.db.Preload("Tasks").Where("regular_id = ?", regular.ID).Order("priority asc").Find(&categories).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Fetching Categories Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	return c.Render(http.StatusOK, "opus-task", categories)
}

func (h *OpusHandlerImpl) AddTask(c echo.Context) error {
	var (
		req        dtos.AddTaskRequest
		categories []models.Category
		newTask    models.Task
	)

	regular := c.Get("regular").(models.Regular)

	if err := c.Bind(&req); err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusBadRequest),
			Message: "Bad Request",
			Error:   err.Error(),
		}
		return c.Render(http.StatusBadRequest, "error", errorData)
	}

	newTask = models.Task{
		Title:      req.Title,
		Priority:   req.Priority,
		CategoryID: req.CategoryID,
	}

	if req.ParentType == "task" {
		var parentTask models.Task

		if err := h.db.Where("id = ?", req.ParentID).First(&parentTask).Error; err != nil {
			h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
			errorData := dtos.Error{
				Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusInternalServerError),
				Message: "Fetching Parent Task Error",
				Error:   err.Error(),
			}
			return c.Render(http.StatusInternalServerError, "error", errorData)
		}

		newTask.ParentID = &parentTask.ID

	}

	if err := h.db.Create(&newTask).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusInternalServerError),
			Message: "Fetching Parent Task Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	if err := h.db.Preload("Tasks").Where("regular_id = ?", regular.ID).Order("priority asc").Find(&categories).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Fetching Categories Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	return c.Render(http.StatusOK, "opus-task", categories)
}

func (h *OpusHandlerImpl) DeleteCategory(c echo.Context) error {
	var (
		categories []models.Category
	)
	regular := c.Get("regular").(models.Regular)

	categoryID := c.Param("id")

	if err := h.db.Delete(&(models.Category{}), categoryID).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Deleting Category Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	if err := h.db.Preload("Tasks").Where("regular_id = ?", regular.ID).Order("priority asc").Find(&categories).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Fetching Categories Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	return c.Render(http.StatusOK, "opus-task", categories)
}

func (h *OpusHandlerImpl) saveState(c echo.Context, regular *models.Regular) error {
	if err := h.db.Save(&regular.RegularSession.RegularState).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Fetching Regular Information Error [Session Might Be Invalid]",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}
	return nil
}
