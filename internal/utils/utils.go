package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateSaleNumber mirrors generateSaleNumber() in your utils.ts
func GenerateSaleNumber() string {
	ts := time.Now().UnixMilli()
	r := rand.Intn(1000)
	return fmt.Sprintf("SALE-%d-%03d", ts, r)
}

// FormatRupiah formats a float64 as Indonesian Rupiah string
func FormatRupiah(amount float64) string {
	// Simple formatter: Rp 1.234.567
	intVal := int64(amount)
	s := fmt.Sprintf("%d", intVal)
	result := ""
	n := len(s)
	for i, c := range s {
		if i > 0 && (n-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return "Rp " + result
}

// FormatDate formats a time.Time as Indonesian date string
func FormatDate(t time.Time) string {
	months := []string{"", "Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	return fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()], t.Year())
}

// FormatDateTime formats a time.Time with time
func FormatDateTime(t time.Time) string {
	months := []string{"", "Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	return fmt.Sprintf("%d %s %d, %02d:%02d", t.Day(), months[t.Month()], t.Year(), t.Hour(), t.Minute())
}

// PointsExpiryDate returns Dec 31 of the given year (mirrors getPointsExpiryDate)
func PointsExpiryDate(earned time.Time) time.Time {
	return time.Date(earned.Year(), 12, 31, 23, 59, 59, 0, earned.Location())
}
