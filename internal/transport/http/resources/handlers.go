package resources

import (
	"context"
	"errors"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/neh333/KaininAPIV2/internal/model"
	"github.com/neh333/KaininAPIV2/internal/service/repositories"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"gorm.io/gorm"
)

type accessLevel uint8

const (
	publicAccess accessLevel = iota
	ownerAccess
	ownerOrAdminAccess
	adminAccess
)

type resourceSpec[T any, Request any, Response any] struct {
	name         string
	path         string
	parseID      func(string) (any, error)
	readAccess   accessLevel
	writeAccess  accessLevel
	scope        func(*gorm.DB, int64) *gorm.DB
	toResponse   func(T) Response
	toModel      func(Request, int64) T
	updateFields func(Request) map[string]any
	beforeCreate func(context.Context, *gorm.DB, Request, int64) error
	create       bool
	update       bool
	delete       bool
}

type listInput struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

type idInput struct {
	ID string `path:"id"`
}

type createInput[D any] struct {
	Body D
}

type updateInput[D any] struct {
	ID   string `path:"id"`
	Body D
}

type itemOutput[D any] struct {
	Body D
}

type listBody[D any] struct {
	Items  []D `json:"items"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type listOutput[D any] struct {
	Body listBody[D]
}

type emptyOutput struct {
	Status int
}

func currentUser(ctx context.Context, db *gorm.DB) (model.User, error) {
	sessionContainer := session.GetSessionFromRequestContext(ctx)
	if sessionContainer == nil {
		return model.User{}, huma.Error401Unauthorized("Authentication required")
	}

	var user model.User
	err := db.WithContext(ctx).
		Where("supertokens_user_id = ?", sessionContainer.GetUserID()).
		Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, huma.Error403Forbidden("No application user is linked to this session")
	}
	if err != nil {
		return model.User{}, huma.Error500InternalServerError("Could not load the current user", err)
	}
	return user, nil
}

func authorize(ctx context.Context, db *gorm.DB, level accessLevel) (int64, error) {
	if level == publicAccess {
		return 0, nil
	}

	user, err := currentUser(ctx, db)
	if err != nil {
		return 0, err
	}
	if level == ownerOrAdminAccess && user.IsAdmin {
		return 0, nil
	}
	if level == adminAccess && !user.IsAdmin {
		return 0, huma.Error403Forbidden("Administrator access required")
	}
	return user.UserID, nil
}

func parseInt64ID(value string) (any, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 1 {
		return nil, huma.Error400BadRequest("ID must be a positive integer")
	}
	return id, nil
}

func parseStringID(value string) (any, error) {
	if value == "" {
		return nil, huma.Error400BadRequest("ID is required")
	}
	return value, nil
}

func registerResource[T any, Request any, Response any](api huma.API, db *gorm.DB, spec resourceSpec[T, Request, Response]) {
	store := repositories.NewCrudService[T](db)

	huma.Register(api, huma.Operation{
		OperationID: "list-" + spec.name,
		Method:      "GET",
		Path:        spec.path,
		Summary:     "List " + spec.name,
	}, func(ctx context.Context, input *listInput) (*listOutput[Response], error) {
		actorID, err := authorize(ctx, db, spec.readAccess)
		if err != nil {
			return nil, err
		}
		limit := input.Limit
		if limit == 0 {
			limit = 50
		}
		if limit < 1 || limit > 100 || input.Offset < 0 {
			return nil, huma.Error400BadRequest("limit must be between 1 and 100 and offset cannot be negative")
		}

		var scope repositories.QueryScope
		if spec.scope != nil && (spec.readAccess == ownerAccess || (spec.readAccess == ownerOrAdminAccess && actorID != 0)) {
			scope = func(query *gorm.DB) *gorm.DB { return spec.scope(query, actorID) }
		}
		values, err := store.ListScoped(ctx, scope, limit, input.Offset)
		if err != nil {
			return nil, huma.Error500InternalServerError("Could not list "+spec.name, err)
		}

		out := &listOutput[Response]{}
		out.Body.Limit = limit
		out.Body.Offset = input.Offset
		out.Body.Items = make([]Response, 0, len(values))
		for _, value := range values {
			out.Body.Items = append(out.Body.Items, spec.toResponse(value))
		}
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-" + spec.name,
		Method:      "GET",
		Path:        spec.path + "/{id}",
		Summary:     "Get " + spec.name + " by ID",
	}, func(ctx context.Context, input *idInput) (*itemOutput[Response], error) {
		actorID, err := authorize(ctx, db, spec.readAccess)
		if err != nil {
			return nil, err
		}
		id, err := spec.parseID(input.ID)
		if err != nil {
			return nil, err
		}
		var scope repositories.QueryScope
		if spec.scope != nil && (spec.readAccess == ownerAccess || (spec.readAccess == ownerOrAdminAccess && actorID != 0)) {
			scope = func(query *gorm.DB) *gorm.DB { return spec.scope(query, actorID) }
		}
		value, err := store.GetByIDScoped(ctx, id, scope)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, huma.Error404NotFound(spec.name + " not found")
		} else if err != nil {
			return nil, huma.Error500InternalServerError("Could not load "+spec.name, err)
		}
		return &itemOutput[Response]{Body: spec.toResponse(value)}, nil
	})

	if spec.create {
		huma.Register(api, huma.Operation{
			OperationID:   "create-" + spec.name,
			Method:        "POST",
			Path:          spec.path,
			Summary:       "Create " + spec.name,
			DefaultStatus: 201,
		}, func(ctx context.Context, input *createInput[Request]) (*itemOutput[Response], error) {
			actorID, err := authorize(ctx, db, spec.writeAccess)
			if err != nil {
				return nil, err
			}
			if spec.beforeCreate != nil {
				if err := spec.beforeCreate(ctx, db, input.Body, actorID); err != nil {
					return nil, err
				}
			}
			value := spec.toModel(input.Body, actorID)
			if err := store.Create(ctx, &value); err != nil {
				return nil, huma.Error500InternalServerError("Could not create "+spec.name, err)
			}
			return &itemOutput[Response]{Body: spec.toResponse(value)}, nil
		})
	}

	if spec.update {
		huma.Register(api, huma.Operation{
			OperationID: "update-" + spec.name,
			Method:      "PUT",
			Path:        spec.path + "/{id}",
			Summary:     "Replace " + spec.name,
		}, func(ctx context.Context, input *updateInput[Request]) (*itemOutput[Response], error) {
			actorID, err := authorize(ctx, db, spec.writeAccess)
			if err != nil {
				return nil, err
			}
			id, err := spec.parseID(input.ID)
			if err != nil {
				return nil, err
			}
			var scope repositories.QueryScope
			if spec.scope != nil && spec.writeAccess == ownerAccess {
				scope = func(query *gorm.DB) *gorm.DB { return spec.scope(query, actorID) }
			}
			rowsAffected, err := store.UpdateFieldsScoped(ctx, id, spec.updateFields(input.Body), scope)
			if err != nil {
				return nil, huma.Error500InternalServerError("Could not update "+spec.name, err)
			}
			if rowsAffected == 0 {
				return nil, huma.Error404NotFound(spec.name + " not found")
			}

			value, err := store.GetByIDScoped(ctx, id, scope)
			if err != nil {
				return nil, huma.Error500InternalServerError("Could not reload "+spec.name, err)
			}
			return &itemOutput[Response]{Body: spec.toResponse(value)}, nil
		})
	}

	if spec.delete {
		huma.Register(api, huma.Operation{
			OperationID: "delete-" + spec.name,
			Method:      "DELETE",
			Path:        spec.path + "/{id}",
			Summary:     "Delete " + spec.name,
		}, func(ctx context.Context, input *idInput) (*emptyOutput, error) {
			actorID, err := authorize(ctx, db, spec.writeAccess)
			if err != nil {
				return nil, err
			}
			id, err := spec.parseID(input.ID)
			if err != nil {
				return nil, err
			}
			var scope repositories.QueryScope
			if spec.scope != nil && spec.writeAccess == ownerAccess {
				scope = func(query *gorm.DB) *gorm.DB { return spec.scope(query, actorID) }
			}
			rowsAffected, err := store.DeleteScoped(ctx, id, scope)
			if err != nil {
				return nil, huma.Error500InternalServerError("Could not delete "+spec.name, err)
			}
			if rowsAffected == 0 {
				return nil, huma.Error404NotFound(spec.name + " not found")
			}
			return &emptyOutput{Status: 204}, nil
		})
	}
}

func scopeOwnedBag(query *gorm.DB, userID int64) *gorm.DB {
	return query.Where("UserID = ?", userID)
}

func scopeOwnedBagItem(query *gorm.DB, userID int64) *gorm.DB {
	ownedBags := query.Session(&gorm.Session{NewDB: true}).
		Model(&model.Bag{}).
		Select("BagID").
		Where("UserID = ?", userID)
	return query.Where("BagID IN (?)", ownedBags)
}

func scopeOwnedOrder(query *gorm.DB, userID int64) *gorm.DB {
	return query.Where("CustomerID = ?", userID)
}

func scopeOwnedOrderItem(query *gorm.DB, userID int64) *gorm.DB {
	ownedOrders := query.Session(&gorm.Session{NewDB: true}).
		Model(&model.Order{}).
		Select("OrderID").
		Where("CustomerID = ?", userID)
	return query.Where("OrderID IN (?)", ownedOrders)
}

func ensureOwnedBag(ctx context.Context, db *gorm.DB, bagID, userID int64) error {
	var count int64
	if err := db.WithContext(ctx).Model(&model.Bag{}).
		Where("BagID = ? AND UserID = ?", bagID, userID).
		Count(&count).Error; err != nil {
		return huma.Error500InternalServerError("Could not verify bag ownership", err)
	}
	if count == 0 {
		return huma.Error404NotFound("Bag not found")
	}
	return nil
}

func profileRoutes(api huma.API, db *gorm.DB) {
	huma.Register(api, huma.Operation{
		OperationID: "get-current-user",
		Method:      "GET",
		Path:        "/api/users/me",
		Summary:     "Get the current user's profile",
	}, func(ctx context.Context, _ *struct{}) (*itemOutput[UserResponse], error) {
		user, err := currentUser(ctx, db)
		if err != nil {
			return nil, err
		}
		return &itemOutput[UserResponse]{Body: userResponse(user)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-current-user",
		Method:      "PATCH",
		Path:        "/api/users/me",
		Summary:     "Update the current user's profile",
	}, func(ctx context.Context, input *createInput[UserRequest]) (*itemOutput[UserResponse], error) {
		user, err := currentUser(ctx, db)
		if err != nil {
			return nil, err
		}
		updates := userUpdates(input.Body)
		if len(updates) == 0 {
			return nil, huma.Error400BadRequest("At least one profile field must be provided")
		}
		userStore := repositories.NewCrudService[model.User](db)
		rowsAffected, err := userStore.UpdateFieldsScoped(ctx, user.UserID, updates, nil)
		if err != nil {
			return nil, huma.Error500InternalServerError("Could not update the current user", err)
		}
		if rowsAffected == 0 {
			return nil, huma.Error404NotFound("User not found")
		}
		user, err = userStore.GetByID(ctx, user.UserID)
		if err != nil {
			return nil, huma.Error500InternalServerError("Could not reload the current user", err)
		}
		return &itemOutput[UserResponse]{Body: userResponse(user)}, nil
	})
}
