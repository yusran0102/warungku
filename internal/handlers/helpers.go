package handlers

import (
	"regexp"
	"strings"
)

// normalizePhone strips spaces and non-numeric chars (shared by auth, customers, settings)
var nonPhone = regexp.MustCompile(`[^0-9+]`)

func normalizePhone(phone string) string {
	phone = strings.ReplaceAll(phone, " ", "")
	return nonPhone.ReplaceAllString(phone, "")
}

// normalizeEmail trims and lowercases an email
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// strPtr returns a pointer to a string (shared by products, sales service)
func strPtr(s string) *string { return &s }
