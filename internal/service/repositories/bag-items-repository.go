package repositories

import (
	"gorm.io/gorm"

	"github.com/neh333/KaininAPIV2/internal/model"
)

type BagItemsRepository struct {
	*CrudService[model.BagItem]
}

func NewBagItemsRepository(db *gorm.DB) *BagItemsRepository {
	return &BagItemsRepository{
		CrudService: NewCrudService[model.BagItem](db),
	}
}
