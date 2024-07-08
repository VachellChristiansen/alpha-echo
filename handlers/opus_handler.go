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
	regular.RegularSession.RegularState.PageData = map[string]interface{}{
		"Refresh":        true,
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
	regular.RegularSession.RegularState.PageData = h.extractTaskDate(regular.RegularSession.RegularState.PageData, &task)
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
		parsedStartDate, err := time.Parse("2006-01-02T15:04", req.StartDate)
		if err != nil {
			errorData := dtos.Error{
				Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
				Message: "Parsing Start Date Error",
				Error:   err.Error(),
			}
			return c.Render(http.StatusInternalServerError, "error", errorData)
		}
		parsedEndDate, err := time.Parse("2006-01-02T15:04", req.EndDate)
		if err != nil {
			errorData := dtos.Error{
				Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
				Message: "Parsing Start Date Error",
				Error:   err.Error(),
			}
			return c.Render(http.StatusInternalServerError, "error", errorData)
		}
		task.StartDate = parsedStartDate
		task.EndDate = parsedEndDate
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

	if err := json.Unmarshal(regular.RegularSession.RegularState.PageDataStore, &regular.RegularSession.RegularState.PageData); err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusInternalServerError),
			Message: "Loading Page Data errorData",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}
	regular.RegularSession.RegularState.PageData["Task"] = task
	regular.RegularSession.RegularState.PageData = h.extractTaskDate(regular.RegularSession.RegularState.PageData, &task)
	regular.RegularSession.RegularState.PageDataStore = h.convertToDatabyte(regular.RegularSession.RegularState.PageData)

	if err := h.saveState(c, &regular); err != nil {
		return err
	}

	return c.Render(http.StatusOK, "opus-main", regular.RegularSession.RegularState)
}

func (h *OpusHandlerImpl) UpdateState(c echo.Context) error {
	var (
		req  dtos.UpdateOpusStateRequest
		task models.Task
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

	if err := json.Unmarshal(regular.RegularSession.RegularState.PageDataStore, &regular.RegularSession.RegularState.PageData); err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusInternalServerError),
			Message: "Loading Page Data errorData",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	switch req.Section {
	case "detail":
		regular.RegularSession.RegularState.PageData["TaskDetail"] = req.State
	case "goals":
		regular.RegularSession.RegularState.PageData["TaskGoals"] = req.State
	case "completion":
		regular.RegularSession.RegularState.PageData["TaskCompletion"] = req.State
	case "notes":
		regular.RegularSession.RegularState.PageData["TaskNotes"] = req.State
	}

	if err := h.db.Where("id = ?", req.Id).First(&task).Error; err != nil {
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v-OPUS", http.StatusInternalServerError),
			Message: "Fetching Task Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}
	regular.RegularSession.RegularState.PageData["Task"] = task
	regular.RegularSession.RegularState.PageData = h.extractTaskDate(regular.RegularSession.RegularState.PageData, &task)
	regular.RegularSession.RegularState.PageDataStore = h.convertToDatabyte(regular.RegularSession.RegularState.PageData)

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

func (h *OpusHandlerImpl) extractTaskDate(data map[string]interface{}, task *models.Task) map[string]interface{} {
	data["StartDate"] = task.StartDate.Format("2006-01-02T15:04")
	data["EndDate"] = task.EndDate.Format("2006-01-02T15:04")

	dayNow := time.Now().YearDay()
	dayStart := task.StartDate.YearDay()
	dayEnd := task.EndDate.YearDay()

	dayOffset := dayNow - dayStart
	dayProgress := make([]bool, dayEnd-dayStart)
	for i := range dayProgress {
		if i <= dayOffset {
			dayProgress[i] = true
		} else {
			dayProgress[i] = false
		}
	}

	data["DayProgress"] = dayProgress
	return data
}
