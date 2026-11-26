package resources

import "time"

type BagItemRequest struct {
	BagID     int64 `json:"bagId"`
	ProductID int64 `json:"productId"`
}

type BagItemResponse struct {
	BagItemID int64 `json:"bagItemId"`
	BagID     int64 `json:"bagId"`
	ProductID int64 `json:"productId"`
}

type BagRequest struct{}

type BagResponse struct {
	BagID     int64     `json:"bagId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type LoginMethodRequest struct {
	Provider string `json:"provider"`
	Email    string `json:"email"`
}

type LoginMethodResponse struct {
	ID        int64     `json:"id"`
	Provider  string    `json:"provider"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type NewsletterSubscriptionRequest struct {
	Email string `json:"email"`
}

type NewsletterSubscriptionResponse struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	CreatedAtUtc   time.Time `json:"createdAtUtc"`
	Confirmed      bool      `json:"confirmed"`
	ConfirmedAtUtc time.Time `json:"confirmedAtUtc"`
}

type OptionTypeRequest struct {
	OptionTypeName string `json:"optionTypeName"`
}

type OptionTypeResponse struct {
	OptionTypeID   int64  `json:"optionTypeId"`
	OptionTypeName string `json:"optionTypeName"`
}

type OptionValueRequest struct {
	OptionTypeID    int64  `json:"optionTypeId"`
	OptionValueName string `json:"optionValueName"`
}

type OptionValueResponse struct {
	OptionValueID   int64  `json:"optionValueId"`
	OptionTypeID    int64  `json:"optionTypeId"`
	OptionValueName string `json:"optionValueName"`
}

type OrderItemRequest struct {
	OrderID   int64   `json:"orderId"`
	ProductID int64   `json:"productId"`
	Quantity  int64   `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
}

type OrderItemResponse struct {
	OrderItemID int64   `json:"orderItemId"`
	OrderID     int64   `json:"orderId"`
	ProductID   int64   `json:"productId"`
	Quantity    int64   `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
}

type OrderRequest struct {
	CustomerID      int64     `json:"customerId"`
	Amount          float64   `json:"amount"`
	OrderDate       time.Time `json:"orderDate"`
	ExpectedArrival time.Time `json:"expectedArrival"`
	OrderStatus     string    `json:"orderStatus"`
	ShippingAddress string    `json:"shippingAddress"`
}

type OrderResponse struct {
	OrderID         int64     `json:"orderId"`
	CustomerID      int64     `json:"customerId"`
	Amount          float64   `json:"amount"`
	OrderDate       time.Time `json:"orderDate"`
	ExpectedArrival time.Time `json:"expectedArrival"`
	OrderStatus     string    `json:"orderStatus"`
	ShippingAddress string    `json:"shippingAddress"`
}

type OrderStatusRequest struct {
	OrderStatusName string `json:"orderStatusName"`
	Description     string `json:"description"`
}

type OrderStatusResponse struct {
	OrderStatusID   int64  `json:"orderStatusId"`
	OrderStatusName string `json:"orderStatusName"`
	Description     string `json:"description"`
}

type ProductOptionRequest struct {
	ProductID     int64 `json:"productId"`
	OptionValueID int64 `json:"optionValueId"`
	StockQuantity int64 `json:"stockQuantity"`
}

type ProductOptionResponse struct {
	ProductOptionID int64 `json:"productOptionId"`
	ProductID       int64 `json:"productId"`
	OptionValueID   int64 `json:"optionValueId"`
	StockQuantity   int64 `json:"stockQuantity"`
}

type ProductRequest struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	SquareID    string  `json:"squareId"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	ImageURI    string  `json:"imageUri"`
}

type ProductResponse struct {
	ProductID   int64   `json:"productId"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	SquareID    string  `json:"squareId"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	ImageURI    string  `json:"imageUri"`
}

type ProductVariationRequest struct {
	ProductID       int64   `json:"productId"`
	SKU             string  `json:"variationName"`
	Allocated       int64   `json:"allocated"`
	Available       int64   `json:"available"`
	InStock         bool    `json:"inStock"`
	SquareID        string  `json:"squareId"`
	ProductOptionID int64   `json:"productOptionId"`
	Price           float64 `json:"price"`
}

type ProductVariationResponse struct {
	ProductVariationID int64   `json:"productVariationId"`
	ProductID          int64   `json:"productId"`
	SKU                string  `json:"variationName"`
	Allocated          int64   `json:"allocated"`
	Available          int64   `json:"available"`
	InStock            bool    `json:"inStock"`
	SquareID           string  `json:"squareId"`
	ProductOptionID    int64   `json:"productOptionId"`
	Price              float64 `json:"price"`
}

type RoleRequest struct {
	RoleName string `json:"roleName"`
}

type RoleResponse struct {
	RoleID   int64  `json:"roleId"`
	RoleName string `json:"roleName"`
}

type UserRequest struct {
	FirstName   *string `json:"firstName,omitempty"`
	LastName    *string `json:"lastName,omitempty"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
}

type UserResponse struct {
	UserID       int64     `json:"userId"`
	Email        string    `json:"email"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	IsAdmin      bool      `json:"isAdmin"`
	PhoneNumber  string    `json:"phoneNumber"`
	CreatedAt    time.Time `json:"createdAt"`
	RoleID       int64     `json:"roleId"`
	SignupMethod string    `json:"signupMethod"`
}
