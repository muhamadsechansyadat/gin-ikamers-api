package validation

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
	"strings"
)

var (
	rxUpper   = regexp.MustCompile("[A-Z]")
	rxLower   = regexp.MustCompile("[a-z]")
	rxDigit   = regexp.MustCompile("[0-9]")
	rxSpecial = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{}|;:,.<>?/\\~"'` + "`" + `]`)
)

func Init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" || name == "_" || name == "" {
				return fld.Name
			}
			return name
		})

		_ = v.RegisterValidation("strong_password", validateStrongPassword)
	}
}

func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	return rxUpper.MatchString(password) &&
		rxLower.MatchString(password) &&
		rxDigit.MatchString(password) &&
		rxSpecial.MatchString(password)
}
