package helpers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func FormatValidationError(err error, obj interface{}) map[string]string {
	res := make(map[string]string)

	// Get the struct type via reflection
	objType := reflect.TypeOf(obj)

	for _, e := range err.(validator.ValidationErrors) {

		// Get field name from struct (Go field name)
		fieldName := e.Field()

		// Find the actual struct field
		if field, ok := objType.FieldByName(fieldName); ok {
			// Get json tag
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" {
				fieldName = jsonTag
			} else {
				fieldName = strings.ToLower(fieldName)
			}
		}

		// Switch error message based on validation tag
		switch e.Tag() {
		case "required":
			res[fieldName] = fmt.Sprintf("%s is required", fieldName)
		case "email":
			res[fieldName] = "invalid email format"
		case "min":
			res[fieldName] = fmt.Sprintf("%s must be at least %s characters", fieldName, e.Param())
		case "eqfield":
			res[fieldName] = fmt.Sprintf("%s must match %s", fieldName, strings.ToLower(e.Param()))
		default:
			res[fieldName] = "invalid value"
		}
	}

	return res
}
