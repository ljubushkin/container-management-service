package httptransport

import "github.com/go-playground/validator/v10"

var validate = validator.New()

func validateRequest(v any) error {
	return validate.Struct(v)
}
