// validator.go
package jsonutil

import (
	"reflect"

	"github.com/go-playground/validator"
)

type ValidationErrors map[string]string

func (e ValidationErrors) Error() string {
	return "validation error"
}

type Validator interface {
	Validate(s any) error
}

type toMapStructValidator struct {
	validate *validator.Validate
}

func DefaultValidator() *toMapStructValidator {
	return &toMapStructValidator{
		validate: validator.New(),
	}
}

func (sv *toMapStructValidator) Validate(s any) error {
	err := sv.validate.Struct(s)
	if verrs, ok := err.(validator.ValidationErrors); ok {
		return sv.toMap(verrs, s)
	}
	return err
}

func (sv *toMapStructValidator) toMap(errs validator.ValidationErrors, obj any) ValidationErrors {
	result := make(ValidationErrors, 0)
	t := reflect.TypeOf(obj)

	for _, err := range errs {
		fieldName := err.Field()
		if field, found := t.Elem().FieldByName(fieldName); found {
			if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				fieldName = parseJSONTag(jsonTag)
			}
		}
		result[fieldName] = err.Tag()
	}
	return result
}

func parseJSONTag(tag string) string {
	for i := 0; i < len(tag); i++ {
		if tag[i] == ',' {
			return tag[:i]
		}
	}
	return tag
}
