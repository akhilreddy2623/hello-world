package validations

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func IsValidMonth(month string, fieldName string) error {
	re := regexp.MustCompile(`^(0[1-9]|1[0-2])$`)
	IsValid := re.MatchString(month)
	if IsValid {
		return nil
	}
	return fmt.Errorf("%s must be in the format 'MM' and valid", strings.ToLower(fieldName))
}

func IsValidYear(year string, fieldName string) error {
	if len(year) != 4 {

		return fmt.Errorf("%s must be in the format 'yyyy'", strings.ToLower(fieldName))
	}

	yearInt, err := strconv.Atoi(year)
	if err != nil {

		return fmt.Errorf("%s year must be a number", strings.ToLower(fieldName))
	}

	currentYear := time.Now().Year()
	if yearInt < currentYear || yearInt > currentYear+20 {

		return fmt.Errorf("%s year must be within the next 20 years", strings.ToLower(fieldName))
	}

	return nil
}

func IsValidCVV(cvv string, fieldName string) error {
	if len(cvv) != 3 && len(cvv) != 4 {

		return fmt.Errorf("%s must be 3 or 4 digits", strings.ToLower(fieldName))
	}
	if !IsNumeric(cvv) {
		return fmt.Errorf("%s must be a number", strings.ToLower(fieldName))
	}

	return nil
}

func IsValidZIP(zip string, fieldName string) error {
	if len(zip) != 5 && len(zip) != 10 {
		return fmt.Errorf("%s must be 5 or 10 characters long", strings.ToLower(fieldName))

	}

	if len(zip) == 10 && zip[5] != '-' {
		return fmt.Errorf("%s+4 format must include a hyphen after the 5th digit", strings.ToLower(fieldName))
	}

	for i, c := range zip {
		if c < '0' || c > '9' {
			if !(i == 5 && c == '-') {
				return fmt.Errorf("%s must be a number or a number followed by a hyphen and four more numbers", strings.ToLower(fieldName))
			}
		}
	}

	return nil
}

func IsValidCardNumber(cardNumber string, fieldName string) error {
	if len(cardNumber) < 13 || len(cardNumber) > 19 {
		return fmt.Errorf("%s must be between 13 to 19 digits", strings.ToLower(fieldName))
	}

	if !IsNumeric(cardNumber) {
		return fmt.Errorf("%s must be a number", strings.ToLower(fieldName))

	}
	return nil
}
