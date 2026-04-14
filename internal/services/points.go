package services

import (
	"time"

	"github.com/yusran0102/warungku/internal/database"
	"github.com/yusran0102/warungku/internal/models"
)

// GetAvailablePoints calculates non-expired points for a user (mirrors getAvailablePoints)
func GetAvailablePoints(userID string) (int, error) {
	var history []models.PointHistory
	if err := database.DB.Where("user_id = ?", userID).Order("created_at ASC").Find(&history).Error; err != nil {
		return 0, err
	}

	now := time.Now()
	total := 0
	for _, entry := range history {
		if entry.ExpiresAt != nil && entry.ExpiresAt.Before(now) {
			continue // skip expired
		}
		total += entry.Points
	}
	if total < 0 {
		return 0, nil
	}
	return total, nil
}
