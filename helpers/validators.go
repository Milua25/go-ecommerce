package helpers

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func ValidateStruct(s interface{}) error {
	var validate = validator.New()
	return validate.Struct(s)
}

type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

func ExtractValidationErrors(err error) []FieldError {
	var fieldErrors []FieldError
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrors {
			fieldErrors = append(fieldErrors, FieldError{
				Field: fieldErr.Field(),
				Error: fieldErrorToString(fieldErr),
			})
		}
	}
	return fieldErrors
}

func fieldErrorToString(fe validator.FieldError) string {
	fmt.Printf("Validation error on field '%s': %s\n", fe.Field(), fe.Error())
	fmt.Printf("Field Tags '%s': %s\n", fe.Field(), fe.Tag())
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "email":
		return fe.Field() + " must be a valid email address"
	case "min":
		return fe.Field() + " must be at least " + fe.Param() + " characters"
	case "max":
		return fe.Field() + " must be at most " + fe.Param() + " characters"
	case "gt":
		return fe.Field() + " must be greater than " + fe.Param()
	case "lt":
		return fe.Field() + " must be less than " + fe.Param()
	default:
		return "Invalid value"
	}
}
