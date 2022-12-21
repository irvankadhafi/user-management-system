package helper

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/ttacon/libphonenumber"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

// RemoveLeadingZeroPhoneNumber remove leading 0 if any
func RemoveLeadingZeroPhoneNumber(phone *string) error {
	if phone == nil || *phone == "" {
		// Skip process
		return nil
	}

	if (*phone)[0] == '0' {
		*phone = strings.Join(strings.Split(*phone, "")[1:], "")
	}

	return nil
}

// FormatPhoneNumberWithCountryCode formating phone number with country code
func FormatPhoneNumberWithCountryCode(phone *string, countryCode string) error {
	if phone == nil {
		// Skip process
		return nil
	}

	libPhone, err := libphonenumber.Parse(*phone, countryCode)
	if err != nil {
		return err
	}
	ph := libphonenumber.Format(libPhone, libphonenumber.E164)
	*phone = ph

	return nil
}

// FormatEmail converts email string to lower case
// and trim trailing and leading space
func FormatEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// HashString encrypt given text
func HashString(text string) (string, error) {
	bt, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bt), nil
}

// IsHashedStringMatch check the plain against the cipher using bcrypt.
// If they don't match, will return false
func IsHashedStringMatch(plain, cipher []byte) bool {
	err := bcrypt.CompareHashAndPassword(cipher, plain)
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false
	}
	if err != nil {
		logrus.Error(err)
		return false
	}
	return true
}

// WrapCloser call close and log the error
func WrapCloser(close func() error) {
	if close == nil {
		return
	}
	if err := close(); err != nil {
		logrus.Error(err)
	}
}
