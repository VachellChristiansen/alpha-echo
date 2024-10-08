package handlers

import (
	"alpha-echo/dtos"
	"alpha-echo/models"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type IndexHandler interface {
	Index(c echo.Context) error
	Default(c echo.Context) error
	About(c echo.Context) error
	Projects(c echo.Context) error
	Gate(c echo.Context) error
	GateSwitch(c echo.Context) error
	GatePassing(c echo.Context) error
}

type IndexHandlerImpl struct {
	db       *gorm.DB
	validate *validator.Validate
	logger   map[string]*log.Logger
}

func NewIndexHandler(db *gorm.DB, validate *validator.Validate, logger map[string]*log.Logger) IndexHandler {
	return &IndexHandlerImpl{
		db:       db,
		validate: validate,
		logger:   logger,
	}
}

func (h *IndexHandlerImpl) Index(c echo.Context) error {
	regular := c.Get("regular").(models.Regular)

	if regular.RegularSession.RegularState.PageDataStore != nil {
		if err := json.Unmarshal(regular.RegularSession.RegularState.PageDataStore, &regular.RegularSession.RegularState.PageData); err != nil {
			h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
			errorData := dtos.Error{
				Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusInternalServerError),
				Message: "Loading Page Data Error",
				Error:   err.Error(),
			}
			return c.Render(http.StatusInternalServerError, "error", errorData)
		}
		regular.RegularSession.RegularState.PageData["Refresh"] = false
	}

	regular.RegularSession.RegularState.Timestamp = time.Now().Unix()
	regular.RegularSession.RegularState.Tokens = map[string]interface{}{
		"FontAwesome": os.Getenv("TOKEN_FONT_AWESOME"),
	}

	if err := saveState(c, &regular, h.db, h.logger); err != nil {
		return err
	}

	return c.Render(http.StatusOK, "index", regular.RegularSession.RegularState)
}

func (h *IndexHandlerImpl) Default(c echo.Context) error {
	regular := c.Get("regular").(models.Regular)

	regular.RegularSession.RegularState.Page = "index"
	regular.RegularSession.RegularState.PageState = "default"

	if err := saveState(c, &regular, h.db, h.logger); err != nil {
		return err
	}

	regular.RegularSession.RegularState.Timestamp = time.Now().Unix()
	return c.Render(http.StatusOK, "body", regular.RegularSession.RegularState)
}

func (h *IndexHandlerImpl) About(c echo.Context) error {
	regular := c.Get("regular").(models.Regular)

	regular.RegularSession.RegularState.Page = "index"
	regular.RegularSession.RegularState.PageState = "about"

	if err := saveState(c, &regular, h.db, h.logger); err != nil {
		return err
	}

	regular.RegularSession.RegularState.Timestamp = time.Now().Unix()
	return c.Render(http.StatusOK, "body", regular.RegularSession.RegularState)
}

func (h *IndexHandlerImpl) Projects(c echo.Context) error {
	var (
		projects []models.Project
	)

	regular := c.Get("regular").(models.Regular)

	if err := h.db.Preload("ProjectTags").Where("regular_access_id >= ?", regular.RegularAccessID).Find(&projects).Error; err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusInternalServerError),
			Message: "Fetching Projects Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}

	regular.RegularSession.RegularState.PageData = map[string]interface{}{
		"Projects": projectsToMap(projects),
	}

	regular.RegularSession.RegularState.Page = "index"
	regular.RegularSession.RegularState.PageState = "projects"
	regular.RegularSession.RegularState.PageDataStore = convertToDatabyte(regular.RegularSession.RegularState.PageData, h.logger)

	if err := saveState(c, &regular, h.db, h.logger); err != nil {
		return err
	}

	regular.RegularSession.RegularState.Timestamp = time.Now().Unix()
	return c.Render(http.StatusOK, "body", regular.RegularSession.RegularState)
}

func (h *IndexHandlerImpl) Gate(c echo.Context) error {
	regular := c.Get("regular").(models.Regular)

	regular.RegularSession.RegularState.Page = "index"
	regular.RegularSession.RegularState.PageState = "gate"
	regular.RegularSession.RegularState.PageData = map[string]interface{}{
		"InnerState": "register",
	}
	regular.RegularSession.RegularState.PageDataStore = convertToDatabyte(regular.RegularSession.RegularState.PageData, h.logger)

	if err := saveState(c, &regular, h.db, h.logger); err != nil {
		return err
	}

	regular.RegularSession.RegularState.Timestamp = time.Now().Unix()
	return c.Render(http.StatusOK, "body", regular.RegularSession.RegularState)
}

func (h *IndexHandlerImpl) GateSwitch(c echo.Context) error {
	var (
		req dtos.GateSwitchRequest
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

	regular.RegularSession.RegularState.PageData = map[string]interface{}{
		"InnerState": req.To,
	}
	regular.RegularSession.RegularState.PageDataStore = convertToDatabyte(regular.RegularSession.RegularState.PageData, h.logger)

	if err := saveState(c, &regular, h.db, h.logger); err != nil {
		return err
	}

	regular.RegularSession.RegularState.Timestamp = time.Now().Unix()
	return c.Render(http.StatusOK, "main", regular.RegularSession.RegularState)
}

func (h *IndexHandlerImpl) GatePassing(c echo.Context) error {
	var (
		req dtos.GateRequest
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

	if err := h.validate.Struct(req); err != nil {
		h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusBadRequest),
			Message: "Input Validation Error",
			Error:   "bad inputs",
		}

		return c.Render(http.StatusBadRequest, "error", errorData)
	}

	regular, err := h.gate(c, req, regular)
	if err != nil {
		return err
	}

	if req.From == "register" {
		regular.RegularSession.RegularState.PageData = map[string]interface{}{
			"InnerState": "login",
		}
		regular.RegularSession.RegularState.PageDataStore = convertToDatabyte(regular.RegularSession.RegularState.PageData, h.logger)

		if err := saveState(c, &regular, h.db, h.logger); err != nil {
			return err
		}

		return c.Render(http.StatusOK, "main", regular.RegularSession.RegularState)
	} else if req.From == "login" {
		regular.RegularSession.RegularState.PageState = "default"

		return c.Render(http.StatusOK, "body", regular.RegularSession.RegularState)
	}

	h.logger["ERROR"].Printf("URL: %v, Error: bad request", c.Request().URL.Path)
	errorData := dtos.Error{
		Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusBadRequest),
		Message: "Bad Request",
		Error:   "bad request",
	}

	return c.Render(http.StatusBadRequest, "error", errorData)
}

func (h *IndexHandlerImpl) gate(c echo.Context, req dtos.GateRequest, guest models.Regular) (regular models.Regular, err error) {
	if req.From == "register" {
		var (
			access models.RegularAccess
		)

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
		if err != nil {
			h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
			errorData := dtos.Error{
				Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusInternalServerError),
				Message: "Hashing Password Error",
				Error:   err.Error(),
			}
			return guest, c.Render(http.StatusInternalServerError, "error", errorData)
		}

		if err := h.db.Where("access = ?", "Regular").First(&access).Error; err != nil {
			h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
			errorData := dtos.Error{
				Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusInternalServerError),
				Message: "Hashing Password Error",
				Error:   err.Error(),
			}
			return guest, c.Render(http.StatusInternalServerError, "error", errorData)
		}

		newRegular := models.Regular{
			Name:            req.Name,
			Email:           req.Email,
			Password:        string(hashedPassword),
			RegularAccessID: access.ID,
		}

		if err := h.db.Create(&newRegular).Error; err != nil {
			h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
			errorData := dtos.Error{
				Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusInternalServerError),
				Message: "Storing Data to Database Error",
				Error:   err.Error(),
			}
			return guest, c.Render(http.StatusInternalServerError, "error", errorData)
		}

		return guest, nil
	} else if req.From == "login" {
		var (
			remember = false
		)

		if err := h.db.Where("email = ?", req.Email).First(&regular).Error; err != nil {
			h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
			errorData := dtos.Error{
				Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusNotFound),
				Message: "Email is Not Registered",
				Error:   err.Error(),
			}
			return regular, c.Render(http.StatusNotFound, "error", errorData)
		}

		if err := bcrypt.CompareHashAndPassword([]byte(regular.Password), []byte(req.Password)); err != nil {
			h.logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
			errorData := dtos.Error{
				Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusUnauthorized),
				Message: "Password Incorrect",
				Error:   err.Error(),
			}
			return regular, c.Render(http.StatusUnauthorized, "error", errorData)
		}

		token := make([]byte, 64)
		if _, err := rand.Read(token); err != nil {
			h.logger["ERROR"].Printf("Guest Token Generation Error: %s", err.Error())
			errorData := dtos.Error{
				Code:    fmt.Sprintf("IE-Middleware-%v", http.StatusInternalServerError),
				Message: "Guest Token Generation Error",
				Error:   err.Error(),
			}
			return regular, c.Render(http.StatusInternalServerError, "error", errorData)
		}
		tokenStr := base64.URLEncoding.EncodeToString(token)

		remember = (req.Remember == "remember")

		guestSession := models.RegularSession{
			Token:          tokenStr,
			LastAccessedAt: time.Now(),
			IPAddress:      getIPAddress(c.Request()),
			RegularID:      regular.ID,
			RememberMe:     remember,
		}

		if err := h.db.Create(&guestSession).Error; err != nil {
			h.logger["ERROR"].Printf("Creating Guest Session Error: %s", err.Error())
			errorData := dtos.Error{
				Code:    fmt.Sprintf("IE-Middleware-%v", http.StatusInternalServerError),
				Message: "Creating Guest Session Error",
				Error:   err.Error(),
			}
			return regular, c.Render(http.StatusInternalServerError, "error", errorData)
		}

		guestState := models.RegularState{
			LoggedIn:         true,
			RegularSessionID: guestSession.ID,
		}

		if err := h.db.Create(&guestState).Error; err != nil {
			h.logger["ERROR"].Printf("Creating Guest State Error: %s", err.Error())
			errorData := dtos.Error{
				Code:    fmt.Sprintf("IE-Middleware-%v", http.StatusInternalServerError),
				Message: "Creating Guest State Error",
				Error:   err.Error(),
			}
			return regular, c.Render(http.StatusInternalServerError, "error", errorData)
		}

		maxAge := 3600 * 3
		if remember {
			maxAge = 3600 * 24 * 30
		}

		guestCookie := new(http.Cookie)
		guestCookie.Name = "token"
		guestCookie.Value = tokenStr
		guestCookie.MaxAge = maxAge
		guestCookie.Path = "/"
		guestCookie.Domain = os.Getenv("SERVER_DOMAIN")
		guestCookie.Secure = false
		guestCookie.HttpOnly = false
		guestCookie.SameSite = http.SameSiteLaxMode

		guestSession.RegularState = guestState
		regular.RegularSession = guestSession

		c.SetCookie(guestCookie)
		return regular, nil
	}

	h.logger["ERROR"].Printf("URL: %v, Error: bad request %v", c.Request().URL.Path, req.From)
	errorData := dtos.Error{
		Code:    fmt.Sprintf("IE-Endpoint-%v", http.StatusUnauthorized),
		Message: "Bad Request",
		Error:   "bad request",
	}
	return regular, c.Render(http.StatusUnauthorized, "error", errorData)
}

func projectsToMap(p []models.Project) (projectsMap []map[string]interface{}) {
	for _, i := range p {
		result := make(map[string]interface{})

		v := reflect.ValueOf(i)
		t := reflect.TypeOf(i)

		for i := 0; i < v.NumField(); i++ {
			fieldName := t.Field(i).Name
			fieldValue := v.Field(i).Interface()
			result[fieldName] = fieldValue
		}
		projectsMap = append(projectsMap, result)
	}
	return projectsMap
}
