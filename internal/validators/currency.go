package validators

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

var ValidCurrencyCodes = map[string]bool{
	"USD": true, "EUR": true, "GBP": true, "JPY": true, "AUD": true,
	"CAD": true,
	"NZD": true, "NGN": true,
}

func ValidateCurrency(fl validator.FieldLevel) bool {
	currency := strings.ToUpper(fl.Field().String())
	return ValidCurrencyCodes[currency]
}

func RegisterCustomValidators(v *validator.Validate) error {
	if err := v.RegisterValidation("currency", ValidateCurrency); err != nil {
		return fmt.Errorf("failed to register currency validator: %w", err)
	}
	return nil
}
