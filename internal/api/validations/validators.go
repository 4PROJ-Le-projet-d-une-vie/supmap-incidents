package validations

import (
	"github.com/go-playground/validator/v10"
	"reflect"
)

type ValidationError struct {
	Message string            `json:"message"`
	Details map[string]string `json:"data"`
}

func (e ValidationError) Error() string {
	return e.Message
}

type CreateIncidentValidator struct {
	TypeId    int64    `json:"type_id" validate:"required"`
	Latitude  *float64 `json:"lat" validate:"required,latitude"`
	Longitude *float64 `json:"lon" validate:"required,longitude"`
}

func (civ CreateIncidentValidator) Validate() error {
	validate := validator.New()
	if err := validate.RegisterValidation("latitude", validateLatitude); err != nil {
		return err
	}
	if err := validate.RegisterValidation("longitude", validateLongitude); err != nil {
		return err
	}

	if err := validate.Struct(civ); err != nil {
		return err
	}
	return nil
}

func validateLatitude(fl validator.FieldLevel) bool {
	field := fl.Field()
	if !field.IsValid() {
		return false
	}

	var value float64
	if field.Kind() == reflect.Pointer {
		if field.IsNil() {
			return false
		}
		value = field.Elem().Float()
	} else {
		value = field.Float()
	}

	return value >= -90 && value <= 90
}

func validateLongitude(fl validator.FieldLevel) bool {
	field := fl.Field()
	if !field.IsValid() {
		return false
	}

	var value float64
	if field.Kind() == reflect.Pointer {
		if field.IsNil() {
			return false
		}
		value = field.Elem().Float()
	} else {
		value = field.Float()
	}

	return value >= -180 && value <= 180
}

type CreateInteractionValidator struct {
	IncidentID     int64 `json:"incident_id" validate:"required"`
	IsStillPresent *bool `json:"is_still_present" validate:"required"`
}

func (civ CreateInteractionValidator) Validate() error {
	validate := validator.New()
	if err := validate.Struct(civ); err != nil {
		return err
	}
	return nil
}
