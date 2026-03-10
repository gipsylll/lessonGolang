package handlers

import (
	"fmt"
	"net/mail"
)

func validateLen(field, value string, minLen, maxLen int) error {
	l := len(value)
	if l < minLen {
		return fmt.Errorf("%s must be at least %d characters", field, minLen)
	}
	if l > maxLen {
		return fmt.Errorf("%s must be at most %d characters", field, maxLen)
	}
	return nil
}

func validateEmail(value string) error {
	if _, err := mail.ParseAddress(value); err != nil {
		return fmt.Errorf("must be a valid email address")
	}
	return nil
}
