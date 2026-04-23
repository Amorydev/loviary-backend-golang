package validator

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Validate validates a struct using go-playground/validator
func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		return &ValidationError{Validator: err}
	}
	return nil
}

// ValidationError wraps validator.ValidationErrors
type ValidationError struct {
	Validator error
}

func (e *ValidationError) Error() string {
	return e.Validator.Error()
}
