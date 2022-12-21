package model

import (
	"github.com/go-playground/validator"
	"regexp"
	"sync"
)

// validate singleton, it's thread safe and cached the struct validation rules
var validate *validator.Validate

// singleton regex
var phoneNumberRgx *regexp.Regexp

var initOnce sync.Once

func init() {
	initOnce.Do(func() {
		validate = validator.New()

		_ = validate.RegisterValidation("phonenumber", isPhoneValid)

		phoneNumberRgx = regexp.MustCompile(`(^(\+)|^[0-9]+$)`)
	})
}

// isPhoneValid implements validator.Func for check phone number
func isPhoneValid(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	if len(phone) < 3 {
		return false
	}

	return phoneNumberRgx.MatchString(phone)
}
