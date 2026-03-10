package validateutil

import (
	"errors"
	"net/mail"
	"strings"
	"unicode"

	"github.com/nyaruka/phonenumbers"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func EmailValid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// PhoneValid checks if a phone number is valid in E.164 format (any country).
// Example accepted: +14155552671, +2348012345678, etc.
func PhoneValid(phone string) bool {
	num, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return false
	}
	return phonenumbers.IsValidNumber(num)
}

// ValidatePassword ensures password is strong
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var hasLetter, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsLetter(ch):
			hasLetter = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return errors.New("password must contain both letters and numbers")
	}
	return nil
}

func ToTitleCase(s string) string {
	caser := cases.Title(language.English)
	return caser.String(strings.ToLower(s))
}
