package utils

import netmail "net/mail"

func IsValidEmail(value string) bool {
	if value == "" {
		return false
	}
	_, err := netmail.ParseAddress(value)
	return err == nil
}
