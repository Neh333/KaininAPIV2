package resources

import "github.com/neh333/KaininAPIV2/internal/model"

func bagItemResponse(value model.BagItem) BagItemResponse {
	return BagItemResponse{BagItemID: value.BagItemID, BagID: value.BagID, ProductID: value.ProductID}
}

func bagItemModel(value BagItemRequest) model.BagItem {
	return model.BagItem{BagID: value.BagID, ProductID: value.ProductID}
}

func bagItemUpdates(value BagItemRequest) map[string]any {
	return map[string]any{"BagID": value.BagID, "ProductID": value.ProductID}
}

func bagResponse(value model.Bag) BagResponse {
	return BagResponse{BagID: value.BagID, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}

func bagModel(BagRequest) model.Bag {
	return model.Bag{}
}

func bagUpdates(BagRequest) map[string]any {
	return map[string]any{}
}

func loginMethodResponse(value model.LoginMethod) LoginMethodResponse {
	return LoginMethodResponse{
		ID:        value.ID,
		Provider:  value.Provider,
		Email:     value.Email,
		CreatedAt: value.CreatedAt,
	}
}

func loginMethodModel(value LoginMethodRequest, supertokensUserID, providerUserID string) model.LoginMethod {
	return model.LoginMethod{
		SupertokensUserID: supertokensUserID,
		Provider:          value.Provider,
		ProviderUserID:    providerUserID,
		Email:             value.Email,
	}
}

func newsletterSubscriptionResponse(value model.NewsletterSubscription) NewsletterSubscriptionResponse {
	return NewsletterSubscriptionResponse{
		ID:             value.ID,
		Email:          value.Email,
		CreatedAtUtc:   value.CreatedAtUtc,
		Confirmed:      value.Confirmed,
		ConfirmedAtUtc: value.ConfirmedAtUtc,
	}
}

func newsletterSubscriptionModel(value NewsletterSubscriptionRequest, token string) model.NewsletterSubscription {
	return model.NewsletterSubscription{Email: value.Email, Token: token}
}

func optionTypeResponse(value model.OptionType) OptionTypeResponse {
	return OptionTypeResponse{OptionTypeID: value.OptionTypeID, OptionTypeName: value.OptionTypeName}
}

func optionTypeModel(value OptionTypeRequest) model.OptionType {
	return model.OptionType{OptionTypeName: value.OptionTypeName}
}

func optionTypeUpdates(value OptionTypeRequest) map[string]any {
	return map[string]any{"OptionTypeName": value.OptionTypeName}
}

func optionValueResponse(value model.OptionValue) OptionValueResponse {
	return OptionValueResponse{
		OptionValueID:   value.OptionValueID,
		OptionTypeID:    value.OptionTypeID,
		OptionValueName: value.OptionValueName,
	}
}

func optionValueModel(value OptionValueRequest) model.OptionValue {
	return model.OptionValue{OptionTypeID: value.OptionTypeID, OptionValueName: value.OptionValueName}
}

func optionValueUpdates(value OptionValueRequest) map[string]any {
	return map[string]any{"OptionTypeID": value.OptionTypeID, "OptionValueName": value.OptionValueName}
}

func orderItemResponse(value model.OrderItem) OrderItemResponse {
	return OrderItemResponse{
		OrderItemID: value.OrderItemID,
		OrderID:     value.OrderID,
		ProductID:   value.ProductID,
		Quantity:    value.Quantity,
		UnitPrice:   value.UnitPrice,
	}
}

func orderItemModel(value OrderItemRequest) model.OrderItem {
	return model.OrderItem{
		OrderID:   value.OrderID,
		ProductID: value.ProductID,
		Quantity:  value.Quantity,
		UnitPrice: value.UnitPrice,
	}
}

func orderItemUpdates(value OrderItemRequest) map[string]any {
	return map[string]any{
		"OrderID": value.OrderID, "ProductID": value.ProductID,
		"Quantity": value.Quantity, "UnitPrice": value.UnitPrice,
	}
}

func orderResponse(value model.Order) OrderResponse {
	return OrderResponse{
		OrderID:         value.OrderID,
		CustomerID:      value.CustomerID,
		Amount:          value.Amount,
		OrderDate:       value.OrderDate,
		ExpectedArrival: value.ExpectedArrival,
		OrderStatus:     value.OrderStatus,
		ShippingAddress: value.ShippingAddress,
	}
}

func orderModel(value OrderRequest) model.Order {
	return model.Order{
		CustomerID:      value.CustomerID,
		Amount:          value.Amount,
		OrderDate:       value.OrderDate,
		ExpectedArrival: value.ExpectedArrival,
		OrderStatus:     value.OrderStatus,
		ShippingAddress: value.ShippingAddress,
	}
}

func orderUpdates(value OrderRequest) map[string]any {
	return map[string]any{
		"CustomerID": value.CustomerID, "Amount": value.Amount,
		"OrderDate": value.OrderDate, "ExpectedArrival": value.ExpectedArrival,
		"OrderStatus": value.OrderStatus, "ShippingAddress": value.ShippingAddress,
	}
}

func orderStatusResponse(value model.OrderStatus) OrderStatusResponse {
	return OrderStatusResponse{
		OrderStatusID:   value.OrderStatusID,
		OrderStatusName: value.OrderStatusName,
		Description:     value.Description,
	}
}

func orderStatusModel(value OrderStatusRequest) model.OrderStatus {
	return model.OrderStatus{OrderStatusName: value.OrderStatusName, Description: value.Description}
}

func orderStatusUpdates(value OrderStatusRequest) map[string]any {
	return map[string]any{"OrderStatusName": value.OrderStatusName, "Description": value.Description}
}

func productOptionResponse(value model.ProductOption) ProductOptionResponse {
	return ProductOptionResponse{
		ProductOptionID: value.ProductOptionID,
		ProductID:       value.ProductID,
		OptionValueID:   value.OptionValueID,
		StockQuantity:   value.StockQuantity,
	}
}

func productOptionModel(value ProductOptionRequest) model.ProductOption {
	return model.ProductOption{
		ProductID: value.ProductID, OptionValueID: value.OptionValueID, StockQuantity: value.StockQuantity,
	}
}

func productOptionUpdates(value ProductOptionRequest) map[string]any {
	return map[string]any{
		"ProductID": value.ProductID, "OptionValueID": value.OptionValueID, "StockQuantity": value.StockQuantity,
	}
}

func productResponse(value model.Product) ProductResponse {
	return ProductResponse{
		ProductID: value.ProductID, Description: value.Description, Price: value.Price,
		SquareID: value.SquareID, Name: value.Name, Category: value.Category, ImageURI: value.ImageURI,
	}
}

func productModel(value ProductRequest) model.Product {
	return model.Product{
		Description: value.Description, Price: value.Price, SquareID: value.SquareID,
		Name: value.Name, Category: value.Category, ImageURI: value.ImageURI,
	}
}

func productUpdates(value ProductRequest) map[string]any {
	return map[string]any{
		"Description": value.Description, "Price": value.Price, "SquareID": value.SquareID,
		"Name": value.Name, "Category": value.Category, "image_uri": value.ImageURI,
	}
}

func productVariationResponse(value model.ProductVariation) ProductVariationResponse {
	return ProductVariationResponse{
		ProductVariationID: value.ProductVariationID, ProductID: value.ProductID,
		SKU: value.SKU, Allocated: value.Allocated, Available: value.Available,
		InStock: value.InStock, SquareID: value.SquareID, ProductOptionID: value.ProductOptionID,
		Price: value.Price,
	}
}

func productVariationModel(value ProductVariationRequest) model.ProductVariation {
	return model.ProductVariation{
		ProductID: value.ProductID, SKU: value.SKU, Allocated: value.Allocated,
		Available: value.Available, InStock: value.InStock, SquareID: value.SquareID,
		ProductOptionID: value.ProductOptionID, Price: value.Price,
	}
}

func productVariationUpdates(value ProductVariationRequest) map[string]any {
	return map[string]any{
		"ProductID": value.ProductID, "SKU": value.SKU, "Allocated": value.Allocated,
		"Available": value.Available, "InStock": value.InStock, "SquareID": value.SquareID,
		"ProductOptionID": value.ProductOptionID, "Price": value.Price,
	}
}

func roleResponse(value model.Role) RoleResponse {
	return RoleResponse{RoleID: value.RoleID, RoleName: value.RoleName}
}

func roleModel(value RoleRequest) model.Role {
	return model.Role{RoleName: value.RoleName}
}

func roleUpdates(value RoleRequest) map[string]any {
	return map[string]any{"RoleName": value.RoleName}
}

func userResponse(value model.User) UserResponse {
	return UserResponse{
		UserID: value.UserID, Email: value.Email, FirstName: value.FirstName,
		LastName: value.LastName, IsAdmin: value.IsAdmin, PhoneNumber: value.PhoneNumber,
		CreatedAt: value.CreatedAt, RoleID: value.RoleID, SignupMethod: value.SignupMethod,
	}
}

func userUpdates(value UserRequest) map[string]any {
	fields := make(map[string]any, 3)
	if value.FirstName != nil {
		fields["FirstName"] = *value.FirstName
	}
	if value.LastName != nil {
		fields["LastName"] = *value.LastName
	}
	if value.PhoneNumber != nil {
		fields["PhoneNumber"] = *value.PhoneNumber
	}
	return fields
}
