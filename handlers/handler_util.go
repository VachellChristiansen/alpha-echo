package handlers

import (
	"alpha-echo/dtos"
	"alpha-echo/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func convertToDatabyte(obj interface{}, logger map[string]*log.Logger) (result []byte) {
	dataByte, err := json.Marshal(obj)
	if err != nil {
		logger["ERROR"].Printf("Converting To Byte Error")
		return nil
	}
	return dataByte
}

func saveState(c echo.Context, regular *models.Regular, db *gorm.DB, logger map[string]*log.Logger) error {
	if err := db.Save(&regular.RegularSession.RegularState).Error; err != nil {
		logger["ERROR"].Printf("URL: %v, Error: %v", c.Request().URL.Path, err.Error())
		errorData := dtos.Error{
			Code:    fmt.Sprintf("IE-DB-%v-OPUS", http.StatusInternalServerError),
			Message: "Saving State Error",
			Error:   err.Error(),
		}
		return c.Render(http.StatusInternalServerError, "error", errorData)
	}
	return nil
}
