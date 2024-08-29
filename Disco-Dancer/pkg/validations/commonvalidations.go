package validations

import (
	"fmt"
	"regexp"
	"strings"
)

func IsAlphanumeric(inputString string) bool {
	var alphanumericRegex = regexp.MustCompile("^[a-zA-Z0-9]+$")
	return alphanumericRegex.MatchString(inputString)
}
func IsNumeric(inputString string) bool {
	return regexp.MustCompile(`^[0-9]+$`).MatchString(inputString)
}
func CheckEmptyFields(fields map[string]string) error {
	for fieldName, fieldValue := range fields {
		if fieldValue == "" {
			return fmt.Errorf("%s cannot be empty", strings.ToLower(fieldName))
		}
	}
	return nil
}

func IsEmpty(inputString string, inputFieldName string) error {
	if inputString == "" {
		return fmt.Errorf("%s cannot be empty", strings.ToLower(inputFieldName))
	}
	return nil
}

func IsValidProductIdentifier(inputString string, inputFieldName string) error {
	if inputString == "" || inputString == "ALL" || IsAlphanumeric(inputString) {
		return nil
	}
	return fmt.Errorf("%s can only contain alphabets and numbers", strings.ToLower(inputFieldName))
}

func IsValidNumeric(inputString string, inputFieldName string) error {

	IsNumeric := IsNumeric(inputString)
	if !IsNumeric {
		return fmt.Errorf("%s should be numeric", strings.ToLower(inputFieldName))
	}
	return nil
}

func IsValidBooleanOrEmpty(inputString string, inputFieldName string) error {
	if inputString != "" && inputString != "true" && inputString != "false" {
		return fmt.Errorf("%s should be either true, false or an empty string", strings.ToLower(inputFieldName))
	}
	return nil
}

func IsGreaterThanZero(id int64, fieldName string) error {
	if id <= 0 {
		return fmt.Errorf("%s is not valid", strings.ToLower(fieldName))
	}
	return nil
}

func IsValidAlphanumeric(inputString string, fieldName string) error {

	if !IsAlphanumeric(inputString) {
		return fmt.Errorf("%s can only contain alphabets and numbers", strings.ToLower(fieldName))
	}
	return nil
}

// CheckIntBelongstoList checks if an integer belongs to a list of integers
func CheckIntBelongstoList(inputInteger int, list []int) bool {
	for _, value := range list {
		if inputInteger == value {
			return true
		}
	}
	return false
}
