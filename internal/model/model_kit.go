package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

type BagItem struct {
	BagItemID int64 `gorm:"column:BagItemID;primaryKey;autoIncrement"`
	BagID     int64 `gorm:"column:BagID"`
	ProductID int64 `gorm:"column:ProductID"`
}

type Bag struct {
	BagID     int64     `gorm:"column:BagID;primaryKey;autoIncrement"`
	UserID    int64     `gorm:"column:UserID"`
	SessionID string    `gorm:"column:SessionID"`
	CreatedAt time.Time `gorm:"column:CreatedAt"`
	UpdatedAt time.Time `gorm:"column:UpdatedAt"`
}

type LoginMethod struct {
	ID                int64     `gorm:"column:id;primaryKey;autoIncrement"`
	SupertokensUserID string    `gorm:"column:supertokens_user_id"`
	Provider          string    `gorm:"column:provider"`
	ProviderUserID    string    `gorm:"column:provider_user_id"`
	Email             string    `gorm:"column:email"`
	CreatedAt         time.Time `gorm:"column:created_at"`
}

type NewsletterSubscription struct {
	ID             string    `gorm:"column:Id;primaryKey"`
	Email          string    `gorm:"column:Email"`
	Token          string    `gorm:"column:Token"`
	CreatedAtUtc   time.Time `gorm:"column:CreatedAtUtc"`
	Confirmed      bool      `gorm:"column:Confirmed"`
	ConfirmedAtUtc time.Time `gorm:"column:ConfirmedAtUtc"`
}

type OptionType struct {
	OptionTypeID   int64  `gorm:"column:OptionTypeID;primaryKey;autoIncrement"`
	OptionTypeName string `gorm:"column:OptionTypeName"`
}

type OptionValue struct {
	OptionValueID   int64  `gorm:"column:OptionValueID;primaryKey;autoIncrement"`
	OptionTypeID    int64  `gorm:"column:OptionTypeID"`
	OptionValueName string `gorm:"column:OptionValueName"`
}

type OrderItem struct {
	OrderItemID int64   `gorm:"column:OrderItemID;primaryKey;autoIncrement"`
	OrderID     int64   `gorm:"column:OrderID"`
	ProductID   int64   `gorm:"column:ProductID"`
	Quantity    int64   `gorm:"column:Quantity"`
	UnitPrice   float64 `gorm:"column:UnitPrice"`
}

type Order struct {
	OrderID         int64     `gorm:"column:OrderID;primaryKey;autoIncrement"`
	CustomerID      int64     `gorm:"column:CustomerID"`
	Amount          float64   `gorm:"column:Amount"`
	OrderDate       time.Time `gorm:"column:OrderDate"`
	ExpectedArrival time.Time `gorm:"column:ExpectedArrival"`
	OrderStatus     string    `gorm:"column:OrderStatus"`
	ShippingAddress string    `gorm:"column:ShippingAddress"`
}

type OrderStatus struct {
	OrderStatusID   int64  `gorm:"column:OrderStatusID;primaryKey;autoIncrement"`
	OrderStatusName string `gorm:"column:OrderStatusName"`
	Description     string `gorm:"column:Description"`
}

type ProductOption struct {
	ProductOptionID int64 `gorm:"column:ProductOptionID;primaryKey;autoIncrement"`
	ProductID       int64 `gorm:"column:ProductID"`
	OptionValueID   int64 `gorm:"column:OptionValueID"`
	StockQuantity   int64 `gorm:"column:StockQuantity"`
}

type Product struct {
	ProductID   int64   `gorm:"column:ProductID;primaryKey;autoIncrement"`
	Description string  `gorm:"column:Description"`
	Price       float64 `gorm:"column:Price"`
	SquareID    string  `gorm:"column:SquareID"`
	Name        string  `gorm:"column:Name"`
	Category    string  `gorm:"column:Category"`
	ImageURI    string  `gorm:"column:image_uri"`
}

type Role struct {
	RoleID   int64  `gorm:"column:RoleID;primaryKey;autoIncrement"`
	RoleName string `gorm:"column:RoleName"`
}

type User struct {
	UserID            int64     `gorm:"column:UserID;primaryKey;autoIncrement"`
	Email             string    `gorm:"column:Email"`
	FirstName         string    `gorm:"column:FirstName"`
	LastName          string    `gorm:"column:LastName"`
	IsAdmin           bool      `gorm:"column:IsAdmin"`
	PhoneNumber       string    `gorm:"column:PhoneNumber"`
	CreatedAt         time.Time `gorm:"column:CreatedAt"`
	RoleID            int64     `gorm:"column:RoleID"`
	SignupMethod      string    `gorm:"column:SignupMethod"`
	PhoneVerified     bool      `gorm:"column:PhoneVerified"`
	LastLogin         time.Time `gorm:"column:LastLogin"`
	SupertokensUserID string    `gorm:"column:supertokens_user_id"`
	SquareCustomerID  *string   `gorm:"column:SquareCustomerID"`
}

// NullString is a helper wrapper around sql.NullString that handles the null case of the value when marshalling to JSON.
type NullString sql.NullString

// Scan implements the sql.Scanner interface so database/sql (and sqlx) can scan NULLs and strings into this type.
func (ns *NullString) Scan(value any) error {
	var s sql.NullString
	if err := s.Scan(value); err != nil {
		return err
	}
	*ns = NullString(s)
	return nil
}

// Value implements the driver.Valuer interface so this type can be used in query parameters.
func (ns NullString) Value() (driver.Value, error) {
	s := sql.NullString(ns)
	if !s.Valid {
		return nil, nil
	}
	return s.String, nil
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	}
	return json.Marshal(nil)
}

func (ns *NullString) UnmarshalJSON(b []byte) error {
	var s *string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s != nil {
		ns.Valid = true
		ns.String = *s
	} else {
		ns.Valid = false
	}
	return nil
}

// NullInt64 is a helper wrapper around sql.NullInt64 that handles the null case of the value when marshalling to JSON.
type NullInt64 sql.NullInt64

// Scan implements the sql.Scanner interface so database/sql (and sqlx) can scan NULLs and integers into this type.
func (ni *NullInt64) Scan(value any) error {
	var i sql.NullInt64
	if err := i.Scan(value); err != nil {
		return err
	}
	*ni = NullInt64(i)
	return nil
}

// Value implements the driver.Valuer interface so this type can be used in query parameters.
func (ni NullInt64) Value() (driver.Value, error) {
	i := sql.NullInt64(ni)
	if !i.Valid {
		return nil, nil
	}
	return i.Int64, nil
}

func (ni NullInt64) MarshalJSON() ([]byte, error) {
	if ni.Valid {
		return json.Marshal(ni.Int64)
	}
	return json.Marshal(nil)
}

func (ni *NullInt64) UnmarshalJSON(b []byte) error {
	var s *int64
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s != nil {
		ni.Valid = true
		ni.Int64 = *s
	} else {
		ni.Valid = false
	}
	return nil
}
