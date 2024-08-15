package validation

import "github.com/go-playground/validator/v10"

var Validation *validator.Validate

func SingletonValidation() {
	Validation = validator.New(validator.WithRequiredStructEnabled())
}
