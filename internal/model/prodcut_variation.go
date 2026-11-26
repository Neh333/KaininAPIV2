package model

type ProductVariation struct {
	ProductVariationID string  `json:"productVariationId" db:"product_variation_id"`
	ProductID          int     `json:"productId" db:"product_id"`
	SKU                string  `json:"variationName" db:"variation_name"`
	Allocated          int     `json:"allocated" db:"allocated"`
	Available          int     `json:"available" db:"available"`
	InStock            bool    `json:"inStock" db:"in_stock"`
	SquareID           string  `json:"squareId" db:"square_id"`
	ProductOptionID    int     `json:"productOptionId" db:"product_option_id"`
	Price              float64 `json:"price" db:"price"`
}
