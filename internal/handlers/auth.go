package handlers

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yusran0102/warung-ku/internal/database"
	"github.com/yusran0102/warung-ku/internal/middleware"
	"github.com/yusran0102/warung-ku/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// ShowLogin renders GET /login
func ShowLogin(c *fiber.Ctx) error {
	registered := c.Query("registered") == "1"
	return c.Render("pages/auth/login", fiber.Map{
		"Title":      "Login – Warung-Ku",
		"Registered": registered,
	}, "layouts/base")
}

// ShowRegister renders GET /register
func ShowRegister(c *fiber.Ctx) error {
	return c.Render("pages/auth/register", fiber.Map{
		"Title": "Daftar Member – Warung-Ku",
	}, "layouts/base")
}

// Login handles POST /login
func Login(c *fiber.Ctx) error {
	identifier := strings.TrimSpace(c.FormValue("identifier"))
	password := c.FormValue("password")

	if identifier == "" || password == "" {
		return c.Render("pages/auth/login", fiber.Map{
			"Title": "Login – Warung-Ku",
			"Error": "Email/telepon dan password wajib diisi",
		}, "layouts/base")
	}

	var user models.User
	result := database.DB.
		Where("phone = ? OR email = ?", normalizePhone(identifier), normalizeEmail(identifier)).
		First(&user)

	if result.Error != nil {
		return c.Render("pages/auth/login", fiber.Map{
			"Title": "Login – Warung-Ku",
			"Error": "Identitas atau password salah",
		}, "layouts/base")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return c.Render("pages/auth/login", fiber.Map{
			"Title": "Login – Warung-Ku",
			"Error": "Identitas atau password salah",
		}, "layouts/base")
	}

	token, err := createToken(&user)
	if err != nil {
		return c.Status(500).SendString("Gagal membuat sesi")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		SameSite: "Lax",
	})

	switch user.Role {
	case models.RoleAdministrator, models.RoleManager:
		return c.Redirect("/admin")
	default:
		return c.Redirect("/member")
	}
}

// Logout handles POST /logout
func Logout(c *fiber.Ctx) error {
	c.ClearCookie("auth_token")
	return c.Redirect("/login")
}

// Register handles POST /register
func Register(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.FormValue("name"))
	rawPhone := c.FormValue("phone")
	rawEmail := c.FormValue("email")
	address := strings.TrimSpace(c.FormValue("address"))
	password := c.FormValue("password")
	birthday := c.FormValue("birthday")

	phone := normalizePhone(rawPhone)
	email := normalizeEmail(rawEmail)

	renderErr := func(msg string) error {
		return c.Render("pages/auth/register", fiber.Map{
			"Title": "Daftar – Warung-Ku",
			"Error": msg,
		}, "layouts/base")
	}

	if name == "" || phone == "" || password == "" || address == "" {
		return renderErr("Nama, telepon, password, dan alamat wajib diisi")
	}
	if len(password) < 6 {
		return renderErr("Password minimal 6 karakter")
	}
	if len(phone) < 10 || len(phone) > 15 {
		return renderErr("Masukkan nomor telepon yang valid")
	}

	var existing models.User
	if database.DB.Where("phone = ?", phone).First(&existing).Error == nil {
		return renderErr("Nomor telepon sudah terdaftar")
	}
	if email != "" {
		if database.DB.Where("email = ?", email).First(&existing).Error == nil {
			return renderErr("Email sudah terdaftar")
		}
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return renderErr("Gagal membuat akun. Coba lagi.")
	}

	newUser := models.User{
		Name:     name,
		Phone:    phone,
		Password: string(hashed),
		Role:     models.RoleMember,
		Points:   0,
	}
	if email != "" {
		newUser.Email = &email
	}
	if address != "" {
		newUser.Address = &address
	}
	if birthday != "" {
		t, err := time.Parse("2006-01-02", birthday)
		if err == nil {
			newUser.Birthday = &t
		}
	}

	if err := database.DB.Create(&newUser).Error; err != nil {
		return renderErr("Gagal membuat akun. Coba lagi.")
	}

	return c.Redirect("/login?registered=1")
}

// ── helpers ────────────────────────────────────────────────────────────────

func createToken(user *models.User) (string, error) {
	claims := middleware.Claims{
		UserID: user.ID,
		Name:   user.Name,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
