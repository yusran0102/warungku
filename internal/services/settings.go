package services

import (
	"errors"
	"strconv"

	"github.com/yusran0102/warungku/internal/database"
	"github.com/yusran0102/warungku/internal/models"
)

const (
	DefaultPointsConversionRate    = 1000
	DefaultMinPointsForRedemption  = 10
	DefaultMaxPointsPerTransaction = 1000
)

// GetSetting retrieves a setting value by key
func GetSetting(key string) (string, error) {
	var s models.Settings
	if err := database.DB.Where("key = ?", key).First(&s).Error; err != nil {
		return "", err
	}
	return s.Value, nil
}

// GetPointsConversionRate returns 1 point = X rupiah (default 1000)
func GetPointsConversionRate() int {
	val, err := GetSetting("pointsConversionRate")
	if err != nil {
		return DefaultPointsConversionRate
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return DefaultPointsConversionRate
	}
	return n
}

func GetMinPointsForRedemption() int {
	val, err := GetSetting("minPointsForRedemption")
	if err != nil {
		return DefaultMinPointsForRedemption
	}
	n, _ := strconv.Atoi(val)
	return n
}

func GetMaxPointsPerTransaction() int {
	val, err := GetSetting("maxPointsPerTransaction")
	if err != nil {
		return DefaultMaxPointsPerTransaction
	}
	n, _ := strconv.Atoi(val)
	return n
}

type AllSettings struct {
	PointsConversionRate    int
	MinPointsForRedemption  int
	MaxPointsPerTransaction int
}

func GetAllSettings() AllSettings {
	return AllSettings{
		PointsConversionRate:    GetPointsConversionRate(),
		MinPointsForRedemption:  GetMinPointsForRedemption(),
		MaxPointsPerTransaction: GetMaxPointsPerTransaction(),
	}
}

// UpsertSetting creates or updates a setting
func UpsertSetting(key, value, description string) error {
	var s models.Settings
	err := database.DB.Where("key = ?", key).First(&s).Error
	if err != nil {
		// Create
		s = models.Settings{Key: key, Value: value}
		if description != "" {
			s.Description = &description
		}
		return database.DB.Create(&s).Error
	}
	// Update
	s.Value = value
	return database.DB.Save(&s).Error
}

func UpdatePointsConversionRate(rate int) error {
	if rate < 100 || rate > 10000 {
		return errors.New("conversion rate must be between 100 and 10,000")
	}
	return UpsertSetting("pointsConversionRate", strconv.Itoa(rate), "Points to Rupiah conversion rate")
}

func UpdateMinPointsForRedemption(min int) error {
	if min < 1 || min > 1000 {
		return errors.New("minimum points must be between 1 and 1,000")
	}
	return UpsertSetting("minPointsForRedemption", strconv.Itoa(min), "Minimum points required to redeem")
}

func UpdateMaxPointsPerTransaction(max int) error {
	if max < 10 {
		return errors.New("maximum points must be at least 10")
	}
	return UpsertSetting("maxPointsPerTransaction", strconv.Itoa(max), "Maximum points per transaction")
}
