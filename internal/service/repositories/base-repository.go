package repositories

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CrudService[T any] struct {
	db *gorm.DB
}

type QueryScope func(*gorm.DB) *gorm.DB

func NewCrudService[T any](db *gorm.DB) *CrudService[T] {
	return &CrudService[T]{db: db}
}

// Create inserts a new record into the database.
func (s *CrudService[T]) Create(ctx context.Context, entity *T) error {
	return s.db.WithContext(ctx).Create(entity).Error
}

// GetByID retrieves a single record by its primary key.
func (s *CrudService[T]) GetByID(ctx context.Context, id any) (T, error) {
	var entity T
	err := s.db.WithContext(ctx).First(&entity, id).Error
	return entity, err
}

// GetAll retrieves all records of type T.
func (s *CrudService[T]) GetAll(ctx context.Context) ([]T, error) {
	var entities []T
	err := s.db.WithContext(ctx).Find(&entities).Error
	return entities, err
}

// Update updates a record matching the given ID with the new struct values.
func (s *CrudService[T]) Update(ctx context.Context, id any, entity T) (int, error) {
	statement := &gorm.Statement{DB: s.db}
	if err := statement.Parse(new(T)); err != nil {
		return 0, err
	}
	if statement.Schema.PrioritizedPrimaryField == nil {
		return 0, gorm.ErrPrimaryKeyRequired
	}
	result := s.db.WithContext(ctx).
		Model(new(T)).
		Where(clause.Eq{Column: clause.PrimaryColumn, Value: id}).
		Select("*").
		Omit(statement.Schema.PrioritizedPrimaryField.Name).
		Updates(entity)
	return int(result.RowsAffected), result.Error
}

// Delete removes a record by its primary key.
func (s *CrudService[T]) Delete(ctx context.Context, id any) (int, error) {
	result := s.db.WithContext(ctx).Delete(new(T), id)
	return int(result.RowsAffected), result.Error
}

func (s *CrudService[T]) ListScoped(ctx context.Context, scope QueryScope, limit, offset int) ([]T, error) {
	query := s.db.WithContext(ctx).Model(new(T))
	if scope != nil {
		query = scope(query)
	}
	var entities []T
	err := query.Limit(limit).Offset(offset).Find(&entities).Error
	return entities, err
}

func (s *CrudService[T]) GetByIDScoped(ctx context.Context, id any, scope QueryScope) (T, error) {
	var entity T
	query := s.db.WithContext(ctx).Model(&entity)
	if scope != nil {
		query = scope(query)
	}
	err := query.First(&entity, id).Error
	return entity, err
}

func (s *CrudService[T]) UpdateFieldsScoped(ctx context.Context, id any, fields map[string]any, scope QueryScope) (int, error) {
	query := s.db.WithContext(ctx).Model(new(T))
	if scope != nil {
		query = scope(query)
	}
	result := query.
		Where(clause.Eq{Column: clause.PrimaryColumn, Value: id}).
		Updates(fields)
	return int(result.RowsAffected), result.Error
}

func (s *CrudService[T]) DeleteScoped(ctx context.Context, id any, scope QueryScope) (int, error) {
	query := s.db.WithContext(ctx)
	if scope != nil {
		query = scope(query)
	}
	result := query.Delete(new(T), id)
	return int(result.RowsAffected), result.Error
}
