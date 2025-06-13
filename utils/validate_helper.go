package utils

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateStruct(s any) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		field := e.Field()
		tag := e.Tag()

		if jsonTag := getJSONFieldName(s, field); jsonTag != "" {
			field = jsonTag
		}

		var message string
		switch tag {
		case "required":
			message = fmt.Sprintf("%s is required", field)
		case "required_without":
			message = fmt.Sprintf("%s is required when %s is not provided", field, e.Param())
		case "min":
			message = fmt.Sprintf("%s must be at least %s characters", field, e.Param())
		case "max":
			message = fmt.Sprintf("%s must be at most %s characters", field, e.Param())
		case "email":
			message = fmt.Sprintf("%s must be a valid email", field)
		default:
			message = fmt.Sprintf("%s is not valid", field)
		}

		errors[field] = message
	}

	return errors
}

func getJSONFieldName(s any, fieldName string) string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if f, ok := t.FieldByName(fieldName); ok {
		jsonTag := f.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			return splitJSONTag(jsonTag)
		}
	}
	return fieldName
}

func splitJSONTag(tag string) string {
	for i, c := range tag {
		if c == ',' {
			return tag[:i]
		}
	}
	return tag
}

func BindJSONWithValidation(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return false
	}

	if validationErrors := ValidateStruct(obj); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErrors})
		return false
	}
	return true
}
