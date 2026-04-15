package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yusran0102/warung-ku/internal/database"
	"github.com/yusran0102/warung-ku/internal/middleware"
	"github.com/yusran0102/warung-ku/internal/models"
	"github.com/yusran0102/warung-ku/internal/services"
	"golang.org/x/crypto/bcrypt"
)

// ShowSettings renders GET /admin/settings/points
func ShowSettings(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	settings := services.GetAllSettings()

	return c.Render("pages/admin/settings_points", fiber.Map{
		"Title":      "Pengaturan Poin – Warung-Ku",
		"User":       claims,
		"ActivePage": "settings-points",
		"Settings":   settings,
	}, "layouts/admin")
}

// UpdateSettings handles POST /admin/settings/points
func UpdateSettings(c *fiber.Ctx) error {
	if rate, err := strconv.Atoi(c.FormValue("pointsConversionRate")); err == nil {
		services.UpdatePointsConversionRate(rate)
	}
	if min, err := strconv.Atoi(c.FormValue("minPointsForRedemption")); err == nil {
		services.UpdateMinPointsForRedemption(min)
	}
	if max, err := strconv.Atoi(c.FormValue("maxPointsPerTransaction")); err == nil {
		services.UpdateMaxPointsPerTransaction(max)
	}
	return c.Redirect("/admin/settings/points")
}

// ShowProfile renders GET /admin/settings/profile
func ShowProfile(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)

	var admin models.User
	database.DB.First(&admin, "id = ?", claims.UserID)

	// Managers list (admin only)
	var managers []models.User
	if claims.Role == models.RoleAdministrator {
		database.DB.Where("role = ?", models.RoleManager).Order("name ASC").Find(&managers)
	}

	return c.Render("pages/admin/settings_profile", fiber.Map{
		"Title":      "Profil – Warung-Ku",
		"User":       claims,
		"ActivePage": "settings-profile",
		"Admin":      admin,
		"Managers":   managers,
	}, "layouts/admin")
}

// UpdateProfile handles POST /admin/settings/profile
func UpdateProfile(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)

	name := strings.TrimSpace(c.FormValue("name"))
	email := normalizeEmail(c.FormValue("email"))
	phone := normalizePhone(c.FormValue("phone"))
	address := strings.TrimSpace(c.FormValue("address"))

	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if phone != "" {
		// Check uniqueness
		var existing models.User
		if database.DB.Where("phone = ? AND id != ?", phone, claims.UserID).First(&existing).Error == nil {
			return c.Status(400).JSON(fiber.Map{"error": "Nomor telepon sudah digunakan"})
		}
		updates["phone"] = phone
	}
	if email != "" {
		var existing models.User
		if database.DB.Where("email = ? AND id != ?", email, claims.UserID).First(&existing).Error == nil {
			return c.Status(400).JSON(fiber.Map{"error": "Email sudah digunakan"})
		}
		updates["email"] = email
	}
	updates["address"] = address

	database.DB.Model(&models.User{}).Where("id = ?", claims.UserID).Updates(updates)
	return c.Redirect("/admin/settings/profile")
}

// UpdatePassword handles POST /admin/settings/profile/password
func UpdatePassword(c *fiber.Ctx) error {
	claims := middleware.GetClaims(c)

	currentPw := c.FormValue("currentPassword")
	newPw := c.FormValue("newPassword")

	if len(newPw) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "Password baru minimal 6 karakter"})
	}

	var user models.User
	database.DB.First(&user, "id = ?", claims.UserID)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPw)); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Password saat ini tidak benar"})
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(newPw), bcrypt.DefaultCost)
	database.DB.Model(&models.User{}).Where("id = ?", claims.UserID).Update("password", string(hashed))

	return c.Redirect("/admin/settings/profile")
}

// CreateManager handles POST /admin/settings/profile/managers
func CreateManager(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.FormValue("name"))
	phone := normalizePhone(c.FormValue("phone"))
	password := c.FormValue("password")
	email := normalizeEmail(c.FormValue("email"))
	address := strings.TrimSpace(c.FormValue("address"))

	if name == "" || phone == "" || len(password) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "Nama, telepon, dan password (min 6 karakter) wajib diisi"})
	}

	var existing models.User
	if database.DB.Where("phone = ?", phone).First(&existing).Error == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Nomor telepon sudah digunakan"})
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	manager := models.User{
		Name:     name,
		Phone:    phone,
		Password: string(hashed),
		Role:     models.RoleManager,
	}
	if email != "" {
		manager.Email = &email
	}
	if address != "" {
		manager.Address = &address
	}

	database.DB.Create(&manager)
	return c.Redirect("/admin/settings/profile")
}

// DeleteManager handles POST /admin/settings/profile/managers/:id/delete
func DeleteManager(c *fiber.Ctx) error {
	id := c.Params("id")
	database.DB.Where("id = ? AND role = ?", id, models.RoleManager).Delete(&models.User{})
	return c.Redirect("/admin/settings/profile")
}
