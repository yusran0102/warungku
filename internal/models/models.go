package models

import (
	"time"

	"gorm.io/gorm"
)

// ── Enums ──────────────────────────────────────────────────────────────────

type Role string

const (
	RoleAdministrator Role = "ADMINISTRATOR"
	RoleManager       Role = "MANAGER"
	RoleMember        Role = "MEMBER"
)

type ProductType string

const (
	ProductTypeReadyStock ProductType = "READY_STOCK"
	ProductTypePreorder   ProductType = "PREORDER"
)

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "INCOME"
	TransactionTypeExpense TransactionType = "EXPENSE"
)

type PaymentStatus string

const (
	PaymentStatusPaid    PaymentStatus = "PAID"
	PaymentStatusPending PaymentStatus = "PENDING"
	PaymentStatusUnpaid  PaymentStatus = "UNPAID"
)

// ── Models ─────────────────────────────────────────────────────────────────

// User → maps to "users" table (same as Prisma User model)
type User struct {
	ID        string    `gorm:"primaryKey;type:text"          json:"id"`
	Email     *string   `gorm:"uniqueIndex;type:text"         json:"email"`
	Password  string    `gorm:"type:text;not null"            json:"-"`
	Name      string    `gorm:"type:text;not null"            json:"name"`
	Phone     string    `gorm:"uniqueIndex;type:text;not null" json:"phone"`
	Birthday  *time.Time `                                    json:"birthday"`
	PhotoURL  *string   `gorm:"type:text"                     json:"photoUrl"`
	Role      Role      `gorm:"type:text;default:'MEMBER'"    json:"role"`
	Points    int       `gorm:"default:0"                     json:"points"`
	Address   *string   `gorm:"type:text"                     json:"address"`
	CreatedAt time.Time `                                     json:"createdAt"`
	UpdatedAt time.Time `                                     json:"updatedAt"`

	// Relations
	CreatedProducts []Product      `gorm:"foreignKey:CreatedByID"        json:"createdProducts,omitempty"`
	UpdatedProducts []Product      `gorm:"foreignKey:UpdatedByID"        json:"updatedProducts,omitempty"`
	CashierSales    []Sale         `gorm:"foreignKey:CashierID"          json:"cashierSales,omitempty"`
	CustomerSales   []Sale         `gorm:"foreignKey:CustomerID"         json:"customerSales,omitempty"`
	PointsHistory   []PointHistory `gorm:"foreignKey:UserID"             json:"pointsHistory,omitempty"`
	Cashflows       []Cashflow     `gorm:"foreignKey:CreatedByID"        json:"cashflows,omitempty"`
}

func (User) TableName() string { return "users" }

// Product → maps to "products" table
type Product struct {
	ID          string      `gorm:"primaryKey;type:text"       json:"id"`
	Name        string      `gorm:"type:text;not null"         json:"name"`
	Description *string     `gorm:"type:text"                  json:"description"`
	SKU         string      `gorm:"uniqueIndex;type:text"      json:"sku"`
	Type        ProductType `gorm:"type:text;default:'READY_STOCK'" json:"type"`
	IsActive    bool        `gorm:"default:true"               json:"isActive"`
	CreatedByID string      `gorm:"type:text;not null"         json:"createdById"`
	UpdatedByID *string     `gorm:"type:text"                  json:"updatedById"`
	CreatedAt   time.Time   `                                  json:"createdAt"`
	UpdatedAt   time.Time   `                                  json:"updatedAt"`

	// Relations
	Variants  []ProductVariant `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"variants,omitempty"`
	CreatedBy User             `gorm:"foreignKey:CreatedByID"                          json:"createdBy,omitempty"`
	UpdatedBy *User            `gorm:"foreignKey:UpdatedByID"                          json:"updatedBy,omitempty"`
}

func (Product) TableName() string { return "products" }

// ProductVariant → maps to "product_variants" table
type ProductVariant struct {
	ID        string    `gorm:"primaryKey;type:text"      json:"id"`
	ProductID string    `gorm:"type:text;not null"        json:"productId"`
	Name      string    `gorm:"type:text;not null"        json:"name"`
	SKU       string    `gorm:"uniqueIndex;type:text"     json:"sku"`
	Price     float64   `gorm:"type:decimal(10,2)"        json:"price"`
	Cost      float64   `gorm:"type:decimal(10,2)"        json:"cost"`
	Stock     int       `gorm:"default:0"                 json:"stock"`
	LowStock  int       `gorm:"default:10"                json:"lowStock"`
	IsActive  bool      `gorm:"default:true"              json:"isActive"`
	Points    int       `gorm:"default:0"                 json:"points"`
	CreatedAt time.Time `                                 json:"createdAt"`
	UpdatedAt time.Time `                                 json:"updatedAt"`

	// Relations
	Product        Product         `gorm:"foreignKey:ProductID"          json:"product,omitempty"`
	SaleItems      []SaleItem      `gorm:"foreignKey:VariantID"          json:"saleItems,omitempty"`
	StockMovements []StockMovement `gorm:"foreignKey:VariantID"          json:"stockMovements,omitempty"`
}

func (ProductVariant) TableName() string { return "product_variants" }

// StockMovement → maps to "stock_movements" table
type StockMovement struct {
	ID        string    `gorm:"primaryKey;type:text"  json:"id"`
	VariantID string    `gorm:"type:text;not null"    json:"variantId"`
	Quantity  int       `gorm:"not null"              json:"quantity"`
	Type      string    `gorm:"type:text;not null"    json:"type"`
	Notes     *string   `gorm:"type:text"             json:"notes"`
	CreatedAt time.Time `                             json:"createdAt"`

	// Relations
	Variant ProductVariant `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
}

func (StockMovement) TableName() string { return "stock_movements" }

// Customer → maps to "customers" table (non-member walk-in customers)
type Customer struct {
	ID        string    `gorm:"primaryKey;type:text"      json:"id"`
	Name      string    `gorm:"type:text;not null"        json:"name"`
	Phone     string    `gorm:"uniqueIndex;type:text"     json:"phone"`
	Address   *string   `gorm:"type:text"                 json:"address"`
	CreatedAt time.Time `                                 json:"createdAt"`
	UpdatedAt time.Time `                                 json:"updatedAt"`

	// Relations
	Sales []Sale `gorm:"foreignKey:NonMemberCustomerID" json:"sales,omitempty"`
}

func (Customer) TableName() string { return "customers" }

// Sale → maps to "sales" table
type Sale struct {
	ID                  string        `gorm:"primaryKey;type:text"          json:"id"`
	SaleNumber          string        `gorm:"uniqueIndex;type:text"         json:"saleNumber"`
	CustomerID          *string       `gorm:"type:text"                     json:"customerId"`
	NonMemberCustomerID *string       `gorm:"type:text"                     json:"nonMemberCustomerId"`
	CashierID           string        `gorm:"type:text;not null"            json:"cashierId"`
	Subtotal            float64       `gorm:"type:decimal(10,2)"            json:"subtotal"`
	Discount            float64       `gorm:"type:decimal(10,2);default:0"  json:"discount"`
	Tax                 float64       `gorm:"type:decimal(10,2);default:0"  json:"tax"`
	Ongkir              float64       `gorm:"type:decimal(10,2);default:0"  json:"ongkir"`
	Total               float64       `gorm:"type:decimal(10,2)"            json:"total"`
	PaymentMethod       string        `gorm:"type:text;not null"            json:"paymentMethod"`
	PaymentStatus       PaymentStatus `gorm:"type:text;default:'PAID'"      json:"paymentStatus"`
	Notes               *string       `gorm:"type:text"                     json:"notes"`
	PointsEarned        int           `gorm:"default:0"                     json:"pointsEarned"`
	PointsRedeemed      int           `gorm:"default:0"                     json:"pointsRedeemed"`
	CreatedAt           time.Time     `                                     json:"createdAt"`

	// Relations
	Items               []SaleItem `gorm:"foreignKey:SaleID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	Cashier             User       `gorm:"foreignKey:CashierID"                          json:"cashier,omitempty"`
	Customer            *User      `gorm:"foreignKey:CustomerID"                         json:"customer,omitempty"`
	NonMemberCustomer   *Customer  `gorm:"foreignKey:NonMemberCustomerID"                json:"nonMemberCustomer,omitempty"`
}

func (Sale) TableName() string { return "sales" }

// SaleItem → maps to "sale_items" table
type SaleItem struct {
	ID        string    `gorm:"primaryKey;type:text"   json:"id"`
	SaleID    string    `gorm:"type:text;not null"     json:"saleId"`
	VariantID string    `gorm:"type:text;not null"     json:"variantId"`
	Quantity  int       `gorm:"not null"               json:"quantity"`
	Price     float64   `gorm:"type:decimal(10,2)"     json:"price"`
	Subtotal  float64   `gorm:"type:decimal(10,2)"     json:"subtotal"`

	// Relations
	Sale    Sale           `gorm:"foreignKey:SaleID"    json:"sale,omitempty"`
	Variant ProductVariant `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
}

func (SaleItem) TableName() string { return "sale_items" }

// PointHistory → maps to "point_history" table
type PointHistory struct {
	ID          string     `gorm:"primaryKey;type:text"  json:"id"`
	UserID      string     `gorm:"type:text;not null"    json:"userId"`
	Points      int        `gorm:"not null"              json:"points"`
	Type        string     `gorm:"type:text;not null"    json:"type"`
	Description string     `gorm:"type:text;not null"    json:"description"`
	CreatedAt   time.Time  `                             json:"createdAt"`
	ExpiresAt   *time.Time `                             json:"expiresAt"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (PointHistory) TableName() string { return "point_history" }

// Cashflow → maps to "cashflows" table
type Cashflow struct {
	ID          string          `gorm:"primaryKey;type:text"  json:"id"`
	Type        TransactionType `gorm:"type:text;not null"    json:"type"`
	Category    string          `gorm:"type:text;not null"    json:"category"`
	Amount      float64         `gorm:"type:decimal(10,2)"    json:"amount"`
	Description string          `gorm:"type:text;not null"    json:"description"`
	Date        time.Time       `gorm:"default:now()"         json:"date"`
	CreatedByID string          `gorm:"type:text;not null"    json:"createdById"`
	CreatedAt   time.Time       `                             json:"createdAt"`

	// Relations
	CreatedBy User `gorm:"foreignKey:CreatedByID" json:"createdBy,omitempty"`
}

func (Cashflow) TableName() string { return "cashflows" }

// Settings → maps to "settings" table
type Settings struct {
	ID          string    `gorm:"primaryKey;type:text"   json:"id"`
	Key         string    `gorm:"uniqueIndex;type:text"  json:"key"`
	Value       string    `gorm:"type:text;not null"     json:"value"`
	Description *string   `gorm:"type:text"              json:"description"`
	UpdatedAt   time.Time `                              json:"updatedAt"`
}

func (Settings) TableName() string { return "settings" }

// ── BeforeCreate hooks (generate CUID like Prisma) ─────────────────────────

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = generateID()
	}
	return nil
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = generateID()
	}
	return nil
}

func (pv *ProductVariant) BeforeCreate(tx *gorm.DB) error {
	if pv.ID == "" {
		pv.ID = generateID()
	}
	return nil
}

func (s *Sale) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = generateID()
	}
	return nil
}

func (si *SaleItem) BeforeCreate(tx *gorm.DB) error {
	if si.ID == "" {
		si.ID = generateID()
	}
	return nil
}

func (sm *StockMovement) BeforeCreate(tx *gorm.DB) error {
	if sm.ID == "" {
		sm.ID = generateID()
	}
	return nil
}

func (ph *PointHistory) BeforeCreate(tx *gorm.DB) error {
	if ph.ID == "" {
		ph.ID = generateID()
	}
	return nil
}

func (cf *Cashflow) BeforeCreate(tx *gorm.DB) error {
	if cf.ID == "" {
		cf.ID = generateID()
	}
	return nil
}

func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = generateID()
	}
	return nil
}

func (s *Settings) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = generateID()
	}
	return nil
}
