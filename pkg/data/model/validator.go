package model

import (
    "github.com/go-playground/validator/v10"
    "github.com/google/uuid"
)

var validate *validator.Validate

func init() {
    validate = validator.New(validator.WithRequiredStructEnabled())
    if err := validate.RegisterValidation("uuid", isValidUUIDValidator); err != nil {
        panic(err)
    }
}

func isValidUUIDValidator(fl validator.FieldLevel) bool {
    return IsValidUUID(fl.Field().String())
}
func IsValidUUID(u string) bool {
    _, err := uuid.Parse(u)
    return err == nil
}
