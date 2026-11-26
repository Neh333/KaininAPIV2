package model

type ProductVariation struct {
	ProductVariationID int64   `json:"productVariationId" db:"product_variation_id" gorm:"column:ProductVariationID;primaryKey;autoIncrement"`
	ProductID          int64   `json:"productId" db:"product_id" gorm:"column:ProductID"`
	SKU                string  `json:"variationName" db:"variation_name" gorm:"column:SKU"`
	Allocated          int64   `json:"allocated" db:"allocated" gorm:"column:Allocated"`
	Available          int64   `json:"available" db:"available" gorm:"column:Available"`
	InStock            bool    `json:"inStock" db:"in_stock" gorm:"column:InStock"`
	SquareID           string  `json:"squareId" db:"square_id" gorm:"column:SquareID"`
	ProductOptionID    int64   `json:"productOptionId" db:"product_option_id" gorm:"column:ProductOptionID"`
	Price              float64 `json:"price" db:"price" gorm:"column:Price"`
}
