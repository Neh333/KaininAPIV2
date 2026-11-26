package resources

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/neh333/KaininAPIV2/internal/model"
	"gorm.io/gorm"
)

func RegisterRoutes(api huma.API, db *gorm.DB) {
	registerResource(api, db, resourceSpec[model.Bag, BagRequest, BagResponse]{
		name:        "bags",
		path:        "/api/bags",
		parseID:     parseInt64ID,
		readAccess:  ownerAccess,
		writeAccess: ownerAccess,
		scope:       scopeOwnedBag,
		toResponse:  bagResponse,
		toModel: func(_ BagRequest, userID int64) model.Bag {
			return model.Bag{UserID: userID}
		},
		updateFields: bagUpdates,
		create:       true,
		delete:       true,
	})

	registerResource(api, db, resourceSpec[model.BagItem, BagItemRequest, BagItemResponse]{
		name:        "bag-items",
		path:        "/api/bag-items",
		parseID:     parseInt64ID,
		readAccess:  ownerAccess,
		writeAccess: ownerAccess,
		scope:       scopeOwnedBagItem,
		toResponse:  bagItemResponse,
		toModel: func(value BagItemRequest, _ int64) model.BagItem {
			return bagItemModel(value)
		},
		updateFields: func(value BagItemRequest) map[string]any {
			return map[string]any{"ProductID": value.ProductID}
		},
		beforeCreate: func(ctx context.Context, db *gorm.DB, value BagItemRequest, userID int64) error {
			return ensureOwnedBag(ctx, db, value.BagID, userID)
		},
		create: true,
		update: true,
		delete: true,
	})

	registerResource(api, db, resourceSpec[model.OptionType, OptionTypeRequest, OptionTypeResponse]{
		name:         "option-types",
		path:         "/api/option-types",
		parseID:      parseInt64ID,
		readAccess:   publicAccess,
		writeAccess:  adminAccess,
		toResponse:   optionTypeResponse,
		toModel:      func(value OptionTypeRequest, _ int64) model.OptionType { return optionTypeModel(value) },
		updateFields: optionTypeUpdates,
		create:       true,
		update:       true,
		delete:       true,
	})

	registerResource(api, db, resourceSpec[model.OptionValue, OptionValueRequest, OptionValueResponse]{
		name:         "option-values",
		path:         "/api/option-values",
		parseID:      parseInt64ID,
		readAccess:   publicAccess,
		writeAccess:  adminAccess,
		toResponse:   optionValueResponse,
		toModel:      func(value OptionValueRequest, _ int64) model.OptionValue { return optionValueModel(value) },
		updateFields: optionValueUpdates,
		create:       true,
		update:       true,
		delete:       true,
	})

	registerResource(api, db, resourceSpec[model.Order, OrderRequest, OrderResponse]{
		name:         "orders",
		path:         "/api/orders",
		parseID:      parseInt64ID,
		readAccess:   ownerOrAdminAccess,
		writeAccess:  adminAccess,
		scope:        scopeOwnedOrder,
		toResponse:   orderResponse,
		toModel:      func(value OrderRequest, _ int64) model.Order { return orderModel(value) },
		updateFields: orderUpdates,
		create:       true,
		update:       true,
		delete:       true,
	})

	registerResource(api, db, resourceSpec[model.OrderItem, OrderItemRequest, OrderItemResponse]{
		name:         "order-items",
		path:         "/api/order-items",
		parseID:      parseInt64ID,
		readAccess:   ownerOrAdminAccess,
		writeAccess:  adminAccess,
		scope:        scopeOwnedOrderItem,
		toResponse:   orderItemResponse,
		toModel:      func(value OrderItemRequest, _ int64) model.OrderItem { return orderItemModel(value) },
		updateFields: orderItemUpdates,
		create:       true,
		update:       true,
		delete:       true,
	})

	registerResource(api, db, resourceSpec[model.OrderStatus, OrderStatusRequest, OrderStatusResponse]{
		name:         "order-statuses",
		path:         "/api/order-statuses",
		parseID:      parseInt64ID,
		readAccess:   publicAccess,
		writeAccess:  adminAccess,
		toResponse:   orderStatusResponse,
		toModel:      func(value OrderStatusRequest, _ int64) model.OrderStatus { return orderStatusModel(value) },
		updateFields: orderStatusUpdates,
		create:       true,
		update:       true,
		delete:       true,
	})

	registerResource(api, db, resourceSpec[model.ProductOption, ProductOptionRequest, ProductOptionResponse]{
		name:         "product-options",
		path:         "/api/product-options",
		parseID:      parseInt64ID,
		readAccess:   publicAccess,
		writeAccess:  adminAccess,
		toResponse:   productOptionResponse,
		toModel:      func(value ProductOptionRequest, _ int64) model.ProductOption { return productOptionModel(value) },
		updateFields: productOptionUpdates,
		create:       true,
		update:       true,
		delete:       true,
	})

	registerResource(api, db, resourceSpec[model.Product, ProductRequest, ProductResponse]{
		name:         "products",
		path:         "/api/products",
		parseID:      parseInt64ID,
		readAccess:   publicAccess,
		writeAccess:  adminAccess,
		toResponse:   productResponse,
		toModel:      func(value ProductRequest, _ int64) model.Product { return productModel(value) },
		updateFields: productUpdates,
		create:       true,
		update:       true,
		delete:       true,
	})

	registerResource(api, db, resourceSpec[model.ProductVariation, ProductVariationRequest, ProductVariationResponse]{
		name:        "product-variations",
		path:        "/api/product-variations",
		parseID:     parseInt64ID,
		readAccess:  publicAccess,
		writeAccess: adminAccess,
		toResponse:  productVariationResponse,
		toModel: func(value ProductVariationRequest, _ int64) model.ProductVariation {
			return productVariationModel(value)
		},
		updateFields: productVariationUpdates,
		create:       true,
		update:       true,
		delete:       true,
	})

	registerResource(api, db, resourceSpec[model.Role, RoleRequest, RoleResponse]{
		name:         "roles",
		path:         "/api/roles",
		parseID:      parseInt64ID,
		readAccess:   adminAccess,
		writeAccess:  adminAccess,
		toResponse:   roleResponse,
		toModel:      func(value RoleRequest, _ int64) model.Role { return roleModel(value) },
		updateFields: roleUpdates,
		create:       true,
		update:       true,
		delete:       true,
	})

	registerResource(api, db, resourceSpec[model.NewsletterSubscription, NewsletterSubscriptionRequest, NewsletterSubscriptionResponse]{
		name:        "newsletter-subscriptions",
		path:        "/api/newsletter-subscriptions",
		parseID:     parseStringID,
		readAccess:  adminAccess,
		writeAccess: adminAccess,
		toResponse:  newsletterSubscriptionResponse,
		toModel: func(value NewsletterSubscriptionRequest, _ int64) model.NewsletterSubscription {
			return newsletterSubscriptionModel(value, "")
		},
		updateFields: func(value NewsletterSubscriptionRequest) map[string]any {
			return map[string]any{"Email": value.Email}
		},
		delete: true,
	})

	profileRoutes(api, db)
}
