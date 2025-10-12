package utils

import "strings"

func NormalizeEmail(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
func Trim(s string) string           { return strings.TrimSpace(s) }
