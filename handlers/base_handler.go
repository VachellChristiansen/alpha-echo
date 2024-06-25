package handlers

import (
	"alpha-echo/dtos"
	"alpha-echo/models"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler struct {
	db           *gorm.DB
	validate     *validator.Validate
	logger       map[string]*log.Logger
	IndexHandler IndexHandler
	OpusHandler  OpusHandler
}

func NewHandler(db *gorm.DB, validate *validator.Validate, logger map[string]*log.Logger) Handler {
	return Handler{
		db:           db,
		validate:     validate,
		logger:       logger,
		IndexHandler: NewIndexHandler(db, validate, logger),
		OpusHandler:  NewOpusHandler(db, validate, logger),
	}
}

func (h *Handler) AccessLogMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			regular models.Regular
		)
		startOverall := time.Now()

		// TODO: Check Token
		token, err := c.Cookie("token")
		if err != nil && err != http.ErrNoCookie {
			h.logger["ERROR"].Printf("Fetching Session Token Error: %s", err.Error())
			errorData := dtos.Error{
				Code:    fmt.Sprintf("IE-Middleware-%v", http.StatusInternalServerError),
				Message: "Fetching Session Token Error",
				Error:   err.Error(),
			}
			return c.Render(http.StatusInternalServerError, "error", errorData)
		}

		if err == http.ErrNoCookie {
			// TODO: Get Guest Session
			if err := h.db.Preload("RegularSession").Where("name = ?", "Guest").First(&regular).Error; err != nil {
				h.logger["ERROR"].Printf("Fetching Guest Data Error: %s", err.Error())
				errorData := dtos.Error{
					Code:    fmt.Sprintf("IE-Middleware-%v", http.StatusInternalServerError),
					Message: "Fetching Guest Data Error",
					Error:   err.Error(),
				}
				return c.Render(http.StatusInternalServerError, "error", errorData)
			}

			// TODO: Generate Guest Session Token
			token := make([]byte, 64)
			if _, err := rand.Read(token); err != nil {
				h.logger["ERROR"].Printf("Guest Token Generation Error: %s", err.Error())
				errorData := dtos.Error{
					Code:    fmt.Sprintf("IE-Middleware-%v", http.StatusInternalServerError),
					Message: "Guest Token Generation Error",
					Error:   err.Error(),
				}
				return c.Render(http.StatusInternalServerError, "error", errorData)
			}
			tokenStr := base64.URLEncoding.EncodeToString(token)

			// TODO: Generate Guest Session
			guestSession := models.RegularSession{
				Token:          tokenStr,
				LastAccessedAt: time.Now(),
				IPAddress:      getIPAddress(c.Request()),
				RegularID:      regular.ID,
			}

			if err := h.db.Create(&guestSession).Error; err != nil {
				h.logger["ERROR"].Printf("Creating Guest Session Error: %s", err.Error())
				errorData := dtos.Error{
					Code:    fmt.Sprintf("IE-Middleware-%v", http.StatusInternalServerError),
					Message: "Creating Guest Session Error",
					Error:   err.Error(),
				}
				return c.Render(http.StatusInternalServerError, "error", errorData)
			}

			// TODO: Init Guest State
			guestState := models.RegularState{
				RegularSessionID: guestSession.ID,
			}

			if err := h.db.Create(&guestState).Error; err != nil {
				h.logger["ERROR"].Printf("Creating Guest State Error: %s", err.Error())
				errorData := dtos.Error{
					Code:    fmt.Sprintf("IE-Middleware-%v", http.StatusInternalServerError),
					Message: "Creating Guest State Error",
					Error:   err.Error(),
				}
				return c.Render(http.StatusInternalServerError, "error", errorData)
			}

			// TODO: Set Guest Token to Cookie
			guestCookie := new(http.Cookie)
			guestCookie.Name = "token"
			guestCookie.Value = tokenStr
			guestCookie.MaxAge = 3600 * 3
			guestCookie.Path = "/"
			guestCookie.Domain = os.Getenv("SERVER_DOMAIN")
			guestCookie.Secure = false
			guestCookie.HttpOnly = false

			guestSession.RegularState = guestState
			regular.RegularSession = guestSession

			c.SetCookie(guestCookie)
		} else {
			if err := h.db.Joins("JOIN regular_sessions ON regular_sessions.regular_id = regulars.id").Where("regular_sessions.token = ?", token.Value).Last(&regular).Error; err != nil {
				h.logger["ERROR"].Printf("Fetching Regular Information Error: %s", err.Error())
				errorData := dtos.Error{
					Code:    fmt.Sprintf("IE-Middleware-%v", http.StatusInternalServerError),
					Message: "Fetching Regular Information Error [Session Might Be Invalid]",
					Error:   err.Error(),
				}
				return c.Render(http.StatusInternalServerError, "error", errorData)
			}
			if err := h.db.Preload("RegularState").Where("regular_sessions.token = ?", token.Value).First(&regular.RegularSession).Error; err != nil {
				h.logger["ERROR"].Printf("Fetching Regular Session Information Error: %s", err.Error())
				errorData := dtos.Error{
					Code:    fmt.Sprintf("IE-Middleware-%v", http.StatusInternalServerError),
					Message: "Fetching Regular Session Information Error [Session Might Be Invalid]",
					Error:   err.Error(),
				}
				return c.Render(http.StatusInternalServerError, "error", errorData)
			}
		}
		c.Set("regular", regular)

		// TODO: Count Access Latency
		startAPI := time.Now()

		err = next(c)

		apiLatency := time.Since(startAPI).Microseconds()

		// TODO: Create Log
		overallLatency := time.Since(startOverall).Microseconds()
		log := models.AccessLog{
			Method:         c.Request().Method,
			Path:           c.Request().URL.Path,
			APILatency:     apiLatency,
			OverallLatency: overallLatency,
			RegularID:      regular.ID,
		}
		if err := h.db.Create(&log).Error; err != nil {
			h.logger["ERROR"].Printf("Inserting Access Log Error: %s", err.Error())
			errorData := dtos.Error{
				Code:    fmt.Sprintf("IE-Middleware-%v", http.StatusInternalServerError),
				Message: "Inserting Access Log Error",
				Error:   err.Error(),
			}
			return c.Render(http.StatusInternalServerError, "error", errorData)
		}

		return err
	}
}

func (h *Handler) AccessMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		regularFromContext := c.Get("regular")
		regular := regularFromContext.(models.Regular)

		accessMap := map[string]int{
			"a": 1,
			"d": 2,
			"e": 3,
			"r": 4,
		}

		accessLevel := strings.Split(c.Request().URL.Path, "/")[1]
		if accessMap[accessLevel] == 0 {
			h.logger["INFO"].Printf("%v Accessing Path: %v", regular.Name, c.Request().URL.Path)
			return next(c)
		}

		if int(regular.RegularAccessID) > accessMap[accessLevel] {
			h.logger["ERROR"].Printf("Access Not Sufficient: %v, Path: %v", regular, c.Request().URL.Path)
			errorData := dtos.Error{
				Code:    fmt.Sprintf("IE-Middleware-%v", http.StatusUnauthorized),
				Message: "Access Not Sufficient",
				Error:   "access not sufficient",
			}
			return c.Render(http.StatusInternalServerError, "error", errorData)
		}
		h.logger["INFO"].Printf("%v Accessing Path: %v", regular.Name, c.Request().URL.Path)
		return next(c)
	}
}

func (h *Handler) IPFilterMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the client's IP address
		clientIP := c.RealIP()

		// Allow only requests from localhost
		if clientIP != "127.0.0.1" && clientIP != "::1" {
			return echo.NewHTTPError(http.StatusForbidden, "Access forbidden")
		}

		return next(c)
	}
}

func (h *Handler) ErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	if code == http.StatusNotFound {
		h.logger["ERROR"].Printf("Error Code: %v, Page Not Found", code)
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-Middleware-%v", code),
			Message: "Page Not Found",
			Error:   "page not found",
		}
		c.Render(code, "error", errorData)
	}
}

func getIPAddress(req *http.Request) string {
	ip := req.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = req.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = req.RemoteAddr
	}

	return ip
}
