package main

import (
	"alpha-echo/dtos"

	"github.com/go-playground/validator/v10"
)

func GateNameRegisterValidation(fl validator.FieldLevel) bool {
    // Get the struct that is being validated
    request := fl.Parent().Interface().(dtos.GateRequest)

    // Check if 'From' is 'register'
    if request.From == "register" {
        // Validate that 'Name' contains only alphabetic characters
        for _, char := range request.Name {
            if !('a' <= char && char <= 'z') && !('A' <= char && char <= 'Z') {
                return false
            }
        }
    }
    return true
}