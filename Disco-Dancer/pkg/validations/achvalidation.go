package validations

import (
	"fmt"
	"strings"
)

func IsValidRoutingNumber(routingNumber string, fieldName string) error {

	if len(routingNumber) != 9 {
		return fmt.Errorf("%s must be 9 digits", strings.ToLower(fieldName))
	}

	if !IsNumeric(routingNumber) {
		return fmt.Errorf("%s must be a number", strings.ToLower(fieldName))
	}

	return nil
}

func IsValidAccountNumber(accountNumber string, fieldName string) error {
	if len(accountNumber) < 1 || len(accountNumber) > 17 {

		return fmt.Errorf("%s must be between 1 to 17 digits", strings.ToLower(fieldName))
	}

	if !IsNumeric(accountNumber) {
		return fmt.Errorf("%s must be a number", strings.ToLower(fieldName))
	}
	return nil
}
