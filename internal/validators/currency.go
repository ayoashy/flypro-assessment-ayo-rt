package validators

import (
	"fmt"
	"strings"
	"github.com/go-playground/validator/v10"
)

var ValidCurrencyCodes = map[string]bool{
	"USD": true, "EUR": true, "GBP": true, "JPY": true, "AUD": true,
	"CAD": true, "CHF": true, "CNY": true, "INR": true, "SGD": true,
	"NZD": true, "MXN": true, "HKD": true, "NOK": true, "SEK": true,
	"KRW": true, "TRY": true, "RUB": true, "ZAR": true, 	"BRL": true,
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
