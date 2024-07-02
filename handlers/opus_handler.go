package handlers

import (
	"alpha-echo/dtos"
	"alpha-echo/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type OpusHandler interface {
	Default(c echo.Context) error
	GetTasks(c echo.Context) error
	GetTaskByID(c echo.Context) error
	AddCategory(c echo.Context) error
	AddTask(c echo.Context) error
	UpdateTask(c echo.Context) error
	UpdateState(c echo.Context) error
	DeleteCategory(c echo.Context) error
	DeleteTask(c echo.Context) error
}

type OpusHandlerImpl struct {
	db       *gorm.DB
	validate *validator.Validate
	location *time.Location
	logger   map[string]*log.Logger
}

func NewOpusHandler(db *gorm.DB, validate *validator.Validate, logger map[string]*log.Logger) OpusHandler {
	location, _ := time.LoadLocation("Asia/Jakarta")
	return &OpusHandlerImpl{
		db:       db,
		validate: validate,
		logger:   logger,
		location: location,
	}
}

func (h *OpusHandlerImpl) Default(c echo.Context) error {
	regular := c.Get("regular").(models.Regular)

	regular.RegularSession.RegularState.Page = "opus"
	regular.RegularSession.RegularState.PageData = map[string]interface{}{
		"TaskOpen":       false,
		"TaskDetail":     "default",
		"TaskGoals":      "default",
		"TaskCompletion": "default",
		"TaskNotes":      "default",
	}
	regular.RegularSession.RegularState.PageDataStore = h.convertToDatabyte(regular.RegularSession.RegularState.PageData)

	if err := h.saveState(c, &regular); err != nil {
		return err
	}

	return c.Render(http.StatusOK, "body", regular.RegularSession.RegularState)
}

func (h *OpusHandlerImpl) GetTasks(c echo.Context) error {
	var (
		categories []models.Category
	)

	regular := c.Get("regular").(models.Regular)

	if err := h.db.Preload(h.generatePreloadTask(10)).Where("regular_id = ?", regular.ID).Order("priority asc").Find(&categories).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Fetching Categories Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	return c.Render(http.StatusOK, "opus-category", categories)
}

func (h *OpusHandlerImpl) GetTaskByID(c echo.Context) error {
	var (
		task models.Task
	)

	regular := c.Get("regular").(models.Regular)

	id := c.Param("id")

	if err := h.db.Preload("TaskGoals").First(&task, id).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Fetching Categories Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	regular.RegularSession.RegularState.PageData = map[string]interface{}{
		"Task":           task,
		"TaskOpen":       true,
		"TaskDetail":     "default",
		"TaskGoals":      "default",
		"TaskCompletion": "default",
		"TaskNotes":      "default",
	}
	regular.RegularSession.RegularState.PageDataStore = h.convertToDatabyte(regular.RegularSession.RegularState.PageData)

	if err := h.saveState(c, &regular); err != nil {
		return err
	}

	return c.Render(http.StatusOK, "opus-main", regular.RegularSession.RegularState)
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
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusBadRequest),
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
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Creating Category Data Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	if err := h.db.Preload(h.generatePreloadTask(10)).Where("regular_id = ?", regular.ID).Order("priority asc").Find(&categories).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Fetching Categories Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	return c.Render(http.StatusOK, "opus-category", categories)
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
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusBadRequest),
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
				Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
				Message: "Fetching Parent Task Error",
				Error:   err.Error(),
			}
			return c.Render(http.StatusInternalServerError, "error", errorData)
		}

		newTask.ParentTask = &parentTask
		newTask.ParentID = &parentTask.ID
		newTask.Inset = parentTask.Inset + 1
	}

	if err := h.db.Create(&newTask).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Fetching Parent Task Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	if err := h.db.Preload(h.generatePreloadTask(10)).Where("regular_id = ?", regular.ID).Order("priority asc").Find(&categories).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Fetching Categories Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	return c.Render(http.StatusOK, "opus-category", categories)
}

func (h *OpusHandlerImpl) UpdateTask(c echo.Context) error {
	var (
		req  dtos.UpdateTaskRequest
		task models.Task
	)

	if err := c.Bind(&req); err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusBadRequest),
			Message: "Bad Request",
			Error:   err.Error(),
		}
		return c.Render(http.StatusBadRequest, "error", errorData)
	}

	if err := h.db.Where("id = ?", req.Id).First(&task).Error; err != nil {
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Fetching Task Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	if req.Updating == "details" {
		task.Details = req.Details
		task.StartDate = req.StartDate.In(h.location)
		task.EndDate = req.EndDate.In(h.location)
	} else if req.Updating == "notes" {
		task.Notes = req.Notes
	}

	if err := h.db.Save(&task).Error; err != nil {
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Saving Task Data Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	return nil
}

func (h *OpusHandlerImpl) UpdateState(c echo.Context) error {
	var (
		req      dtos.UpdateOpusStateRequest
		pageData interface{}
	)

	regular := c.Get("regular").(models.Regular)

	if err := c.Bind(&req); err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusBadRequest),
			Message: "Bad Request",
			Error:   err.Error(),
		}
		return c.Render(http.StatusBadRequest, "error", errorData)
	}

	if err := json.Unmarshal(regular.RegularSession.RegularState.PageDataStore, &pageData); err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusInternalServerError),
			Message: "Loading Page Data errorData",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	tmp := pageData.(map[string]interface{})
	switch req.Section {
	case "detail":
		tmp["TaskDetail"] = req.State
	case "goals":
		tmp["TaskGoals"] = req.State
	case "completion":
		tmp["TaskCompletion"] = req.State
	case "notes":
		tmp["TaskNotes"] = req.State
	}

	regular.RegularSession.RegularState.PageData = tmp
	regular.RegularSession.RegularState.PageDataStore = h.convertToDatabyte(tmp)

	if err := h.saveState(c, &regular); err != nil {
		return err
	}

	return c.Render(http.StatusOK, "opus-main", regular.RegularSession.RegularState)
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

	if err := h.db.Preload(h.generatePreloadTask(10)).Where("regular_id = ?", regular.ID).Order("priority asc").Find(&categories).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Fetching Categories Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	return c.Render(http.StatusOK, "opus-category", categories)
}

func (h *OpusHandlerImpl) DeleteTask(c echo.Context) error {
	var (
		categories []models.Category
	)
	regular := c.Get("regular").(models.Regular)

	taskID := c.Param("id")

	if err := h.db.Delete(&(models.Task{}), taskID).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Deleting Category Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	if err := h.db.Preload(h.generatePreloadTask(10)).Where("regular_id = ?", regular.ID).Order("priority asc").Find(&categories).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Fetching Categories Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	return c.Render(http.StatusOK, "opus-category", categories)
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

func (h *OpusHandlerImpl) convertToDatabyte(obj interface{}) (result []byte) {
	dataByte, err := json.Marshal(obj)
	if err != nil {
		h.logger["ERROR"].Printf("Converting To Byte Error")
		return nil
	}
	return dataByte
}

func (h *OpusHandlerImpl) generatePreloadTask(depth int) string {
	return fmt.Sprintf("Tasks%s", strings.Repeat(".ChildrenTasks", depth))
}
